package caching

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/eymyong/drop/model"
	"github.com/eymyong/drop/repo"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

const (
	CONFIG_EXP = 3600
	keyCache   = "caching-id:"
)

func keyCachingID(id string) string {
	return keyCache + id
}

type cacheRedis interface {
	sync(ctx context.Context, id string, clip model.Clipboard) error
	syncMany(ctx context.Context, data []model.Clipboard) error

	getFromCache(ctx context.Context, key string) (model.Clipboard, error)
	getAllFromCache(ctx context.Context) ([]model.Clipboard, error)
	deleteFromCache(ctx context.Context, key string) error
	deleteAllFromCache(ctx context.Context) error
}

type RepoCachingClipboard struct {
	cache cacheRedis
	db    repo.RepositoryClipboard
}

type cacheRedisImpl struct {
	rd *redis.Client
}

func NewRepoCachingClipboard(rd *redis.Client, db repo.RepositoryClipboard) repo.RepositoryClipboard {
	return &RepoCachingClipboard{
		cache: &cacheRedisImpl{
			rd: rd,
		},
		db: db,
	}
}

// ที่ไม่ใช้ []เพราะคิดว่า ที่สร้าง Cache ขึ้นมาสิ่งที่จะได้ใช้งานโดยตรงจริงๆมีแค่ GetByID ซึ่งไม่มีคามจำเป็นต้องเป็น [] ขนาดนั้น
func (repo *cacheRedisImpl) sync(ctx context.Context, id string, clip model.Clipboard) error {
	rawClips, err := json.Marshal(clip)
	if err != nil {
		return fmt.Errorf("cannot unmarshal clips: %w", err)
	}

	keyCache := keyCachingID(id)
	if err = repo.rd.Set(ctx, keyCache, string(rawClips), CONFIG_EXP).Err(); err != nil {
		return fmt.Errorf("cannot set clips: %w", err)
	}

	return nil
}

func (c *cacheRedisImpl) syncMany(ctx context.Context, data []model.Clipboard) error {
	eg, egCtx := errgroup.WithContext(ctx)

	for i := range data {
		func(clip model.Clipboard) {
			eg.Go(func() error {
				return c.sync(egCtx, clip.Id, clip)
			})
		}(data[i])
	}

	return eg.Wait()
}

func (c *cacheRedisImpl) getAllFromCache(ctx context.Context) ([]model.Clipboard, error) {
	keyCacheing, err := c.rd.Keys(ctx, keyCache+"*").Result()
	if err != nil {
		return []model.Clipboard{}, fmt.Errorf("keys redis err: %w", err)
	}

	var clipboards []model.Clipboard
	for _, v := range keyCacheing {
		clip, err := c.getFromCache(ctx, v)
		if err != nil {
			return []model.Clipboard{}, fmt.Errorf("get caching err: %w", err)
		}
		clipboards = append(clipboards, clip)
	}

	return clipboards, nil
}

// key ="caching-id:" + id
func (c *cacheRedisImpl) getFromCache(ctx context.Context, key string) (model.Clipboard, error) {
	clipStr, err := c.rd.Get(ctx, key).Result()
	if err != nil {
		return model.Clipboard{}, fmt.Errorf("get redis err: %w", err)
	}

	var clipboard model.Clipboard
	err = json.Unmarshal([]byte(clipStr), &clipboard)
	if err != nil {
		return model.Clipboard{}, fmt.Errorf("unmarshal err: %w", err)
	}

	return clipboard, nil
}

func (c *cacheRedisImpl) deleteFromCache(ctx context.Context, key string) error {
	value, err := c.rd.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("delete redis err: %w", err)
	}

	if value != 1 {
		return fmt.Errorf("delete caching incorrrect")
	}

	return nil
}

func (c *cacheRedisImpl) deleteAllFromCache(ctx context.Context) error {
	keys, err := c.rd.Keys(ctx, keyCache+"*").Result()
	if err != nil {
		return err
	}

	eg, egCtx := errgroup.WithContext(ctx)
	for _, v := range keys {
		func(k string) {
			eg.Go(func() error {
				return c.deleteFromCache(egCtx, k)
			})
		}(v)
	}

	err = eg.Wait()
	if err != nil {
		errClear := c.rd.Del(ctx, keyCache+"*").Err()
		if errClear != nil {
			return fmt.Errorf("failed to clear all cache after error '%s': %w", err, errClear)
		}

		return fmt.Errorf("cache has been cleared due to error: %w", err)
	}

	return nil
}

//=================================================================================================

func (r *RepoCachingClipboard) Create(ctx context.Context, clip model.Clipboard) error {
	err := r.db.Create(ctx, clip)
	if err != nil {
		return err
	}

	keyCache := keyCachingID(clip.Id)

	err = r.cache.sync(ctx, keyCache, clip)
	if err != nil {
		slog.Error("failed to create in cache", "clipboard_id", clip.Id)
	}

	return nil
}

func (r *RepoCachingClipboard) GetAll(ctx context.Context) ([]model.Clipboard, error) {
	//cache
	clipboards, err := r.cache.getAllFromCache(ctx)
	if err != nil {
		fmt.Println("cannot getall cache", err)
	}

	if len(clipboards) > 0 {
		return clipboards, nil
	}

	//db
	clipboards, err = r.db.GetAll(ctx)
	if err != nil {
		return []model.Clipboard{}, fmt.Errorf("getall posgest err: %w", err)
	}

	if len(clipboards) <= 0 {
		return []model.Clipboard{}, fmt.Errorf("'clipboard' incorrect 'clipboards' Should have length more than 0")

	}

	// sync
	for _, v := range clipboards {
		keyCache := keyCachingID(v.Id)
		err := r.cache.sync(ctx, keyCache, v)
		if err != nil {
			fmt.Println("cannot sync cache", err)
			break
		}
	}

	return clipboards, nil
}

func (r *RepoCachingClipboard) GetById(ctx context.Context, id string) (model.Clipboard, error) {
	//cache
	keyCache := keyCachingID(id)
	clip, err := r.cache.getFromCache(ctx, keyCache)
	if err != nil {
		// fmt.Println("cannot get cache", err)
		fmt.Printf("cannot get cache\nerr where: %s", err) //ดีไหม
		// fmt.Errorf("cannot get cache\nerr where: %w", err)
	}

	if clip.Id != "" && clip.UserId != "" && clip.Text != "" {
		return clip, nil
	}

	//db
	clip, err = r.db.GetById(ctx, id)
	if err != nil {
		return model.Clipboard{}, err
	}

	//ควรจะต้อง sync ใหม่หรือไม่
	err = r.cache.sync(ctx, keyCache, clip)
	if err != nil {
		fmt.Println("cannot sync cache:", err)
	}

	return clip, nil
}

func (r *RepoCachingClipboard) Update(ctx context.Context, id string, newdata string) error {
	//db
	err := r.db.Update(ctx, id, newdata)
	if err != nil {
		return fmt.Errorf("update posgest err: %w", err)
	}

	clip, err := r.db.GetById(ctx, id)
	if err != nil {
		return fmt.Errorf("get-by-id posgest err: %w", err)
	}

	//cache
	keyCache := keyCachingID(id)
	err = r.cache.sync(ctx, keyCache, clip)
	if err != nil {
		slog.Error("cannot update cache", "clipboard_id", clip.Id)
	}

	return nil
}

func (r *RepoCachingClipboard) Delete(ctx context.Context, id string) error {
	//db
	err := r.db.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("delete posgest err: %w", err)
	}

	//cache
	keyCache := keyCachingID(id)
	err = r.cache.deleteFromCache(ctx, keyCache)
	if err != nil {
		slog.Error("cannot delete cache", "clipboard_id", id)
	}

	return nil
}
func (r *RepoCachingClipboard) DeleteAll(ctx context.Context) error {
	//db
	err := r.db.DeleteAll(ctx)
	if err != nil {
		return fmt.Errorf("delete-all posgest err: %w", err)
	}

	//cachee
	err = r.cache.deleteAllFromCache(ctx)
	if err != nil {
		fmt.Println("cannot delete-all cache", err)

	}

	return nil
}

func (r *RepoCachingClipboard) GetAllUserClipboards(ctx context.Context, userID string) ([]model.Clipboard, error) {
	// cache
	clipboards, err := r.cache.getAllFromCache(ctx)
	if err != nil {
		fmt.Println("cannot getall cache", err)
	}

	// chech userID
	/*
	 */
	if len(clipboards) > 0 {
		return clipboards, nil
	}

	//db
	clipboards, err = r.db.GetAllUserClipboards(ctx, userID)
	if err != nil {
		return []model.Clipboard{}, fmt.Errorf("getall-user-clipboards posgest err: %w", err)
	}

	if len(clipboards) <= 0 {
		return []model.Clipboard{}, fmt.Errorf("'clipboard' incorrect 'clipboards' Should have length more than 0")
	}

	//sync
	for _, v := range clipboards {
		keyCache := keyCachingID(v.Id)
		err := r.cache.sync(ctx, keyCache, v)
		if err != nil {
			fmt.Println("cannot sync cache", err)
			break
		}
	}

	return nil, nil

}

func (r *RepoCachingClipboard) GetUserClipboard(ctx context.Context, id string, userID string) (model.Clipboard, error) {
	//cache
	keyCache := keyCachingID(id)
	clip, err := r.cache.getFromCache(ctx, keyCache)
	if err != nil {
		fmt.Println("cannot get cache", err)
	}

	if clip.Id != "" && clip.UserId != "" && clip.Text != "" {
		if clip.UserId != userID {
			return model.Clipboard{}, fmt.Errorf("no such clipboard for userID: %s", userID)
			// return model.Clipboard{}, fmt.Errorf("not allowed to get clipboard")
		}

		return clip, nil
	}

	//db
	clip, err = r.db.GetUserClipboard(ctx, id, userID)
	if err != nil {
		return model.Clipboard{}, fmt.Errorf("get-user-clipboard err: %w", err)
	}

	//ควรจะต้อง sync ใหม่หรือไม่
	err = r.cache.sync(ctx, keyCache, clip)
	if err != nil {
		fmt.Println("cannot sync cache:", err)
	}

	return clip, nil
}
func (r *RepoCachingClipboard) UpdateUserClipboard(ctx context.Context, id string, userID string, text string) error {
	//db
	err := r.db.UpdateUserClipboard(ctx, id, userID, text)
	if err != nil {
		return fmt.Errorf("update-user-clipboard err: %w", err)
	}

	//cache_1
	clip, err := r.db.GetById(ctx, id)
	if err != nil {
		fmt.Println("get posgest err: %w", err)
	}

	err = r.cache.sync(ctx, id, clip)
	if err != nil {
		fmt.Println("cannot sync cache:", err)
	}

	//cache_2
	clip2, err := r.cache.getFromCache(ctx, id)
	if err != nil {
		fmt.Println("cannot sync cache:", err)
	}

	keyCachee := keyCachingID(id)
	err = r.cache.sync(ctx, keyCachee, clip2)
	if err != nil {
		fmt.Println("cannot sync cache:", err)
	}

	//cache_3
	clip3 := model.Clipboard{
		Id:     id,
		UserId: userID,
		Text:   text,
	}

	err = r.cache.sync(ctx, keyCachee, clip3)
	if err != nil {
		fmt.Println("cannot sync cache:", err)
	}

	return nil
}
func (r *RepoCachingClipboard) DeleteUserClipboard(ctx context.Context, id string, userID string) error {
	//db
	err := r.db.DeleteUserClipboard(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("delete-user-clipboard err: %w", err)
	}

	//cache
	err = r.cache.deleteFromCache(ctx, id)
	if err != nil {
		fmt.Println("cannot delete cache", err)
	}

	return nil
}

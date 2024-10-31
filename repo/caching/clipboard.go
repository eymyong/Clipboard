package caching

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/eymyong/drop/model"
	"github.com/eymyong/drop/repo"
	"github.com/redis/go-redis/v9"
)

const (
	CONFIG_EXP = 3600
)

func KeyCachingID(id string) string {
	return "caching-id:" + id
}

type RepoCache interface {
	Sync(ctx context.Context, id string, clip model.Clipboard) error
	GetCaching(ctx context.Context, key string) (model.Clipboard, error)
	DeleteCaching(ctx context.Context, key string) error
	//
	GetAllCaching(ctx context.Context) ([]model.Clipboard, error)
	DeleteCachingAll(ctx context.Context) error
}

type RepoCachingClipboard struct {
	cache RepoCache
	db    repo.RepositoryClipboard
}

type RepoCacheImpl struct {
	rd *redis.Client
}

func NewRepoCachingClipboard(rd *redis.Client, db repo.RepositoryClipboard) repo.RepositoryClipboard {
	return &RepoCachingClipboard{
		cache: &RepoCacheImpl{
			rd: rd,
		},
		db: db,
	}
}

// ที่ไม่ใช้ []เพราะคิดว่า ที่สร้าง Cache ขึ้นมาสิ่งที่จะได้ใช้งานโดยตรงจริงๆมีแค่ GetByID ซึ่งไม่มีคามจำเป็นต้องเป็น [] ขนาดนั้น
func (repo *RepoCacheImpl) Sync(ctx context.Context, id string, clip model.Clipboard) error {
	rawClips, err := json.Marshal(clip)
	if err != nil {
		return fmt.Errorf("cannot unmarshal clips: %w", err)
	}

	keyCache := KeyCachingID(id)
	if err = repo.rd.Set(ctx, keyCache, string(rawClips), CONFIG_EXP).Err(); err != nil {
		return fmt.Errorf("cannot set clips: %w", err)
	}

	return nil
}

func (repo *RepoCacheImpl) GetAllCaching(ctx context.Context) ([]model.Clipboard, error) {
	keyCacheing, err := repo.rd.Keys(ctx, "caching-id:*").Result()
	if err != nil {
		return []model.Clipboard{}, fmt.Errorf("keys redis err: %w", err)
	}

	var clipboards []model.Clipboard
	for _, v := range keyCacheing {
		clip, err := repo.GetCaching(ctx, v)
		if err != nil {
			return []model.Clipboard{}, fmt.Errorf("get caching err: %w", err)
		}
		clipboards = append(clipboards, clip)
	}

	return clipboards, nil
}

// key ="caching-id:" + id
func (repo *RepoCacheImpl) GetCaching(ctx context.Context, key string) (model.Clipboard, error) {
	clipStr, err := repo.rd.Get(ctx, key).Result()
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

func (repo *RepoCacheImpl) DeleteCaching(ctx context.Context, key string) error {
	value, err := repo.rd.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("delete redis err: %w", err)
	}

	if value != 1 {
		return fmt.Errorf("delete caching incorrrect")
	}

	return nil
}

func (repo *RepoCacheImpl) DeleteCachingAll(ctx context.Context) error {
	keys, err := repo.rd.Keys(ctx, "caching-id:*").Result()
	if err != nil {
		return err
	}

	for _, v := range keys {
		err := repo.DeleteCaching(ctx, v)
		if err != nil {
			return fmt.Errorf("delete caching err: %w", err)
		}
	}

	return nil
}

//=================================================================================================

func (r *RepoCachingClipboard) Create(ctx context.Context, clip model.Clipboard) error {
	err := r.db.Create(ctx, clip)
	if err != nil {
		return err
	}

	keyCache := KeyCachingID(clip.Id)

	err = r.cache.Sync(ctx, keyCache, clip)
	if err != nil {
		slog.Error("failed to create in cache", "clipboard_id", clip.Id)
	}

	return nil
}

func (r *RepoCachingClipboard) GetAll(ctx context.Context) ([]model.Clipboard, error) {
	//cache
	clipboards, err := r.cache.GetAllCaching(ctx)
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
		keyCache := KeyCachingID(v.Id)
		err := r.cache.Sync(ctx, keyCache, v)
		if err != nil {
			fmt.Println("cannot sync cache", err)
			break
		}
	}

	return clipboards, nil
}

func (r *RepoCachingClipboard) GetById(ctx context.Context, id string) (model.Clipboard, error) {
	//cache
	keyCache := KeyCachingID(id)
	clip, err := r.cache.GetCaching(ctx, keyCache)
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
	err = r.cache.Sync(ctx, keyCache, clip)
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
	keyCache := KeyCachingID(id)
	err = r.cache.Sync(ctx, keyCache, clip)
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
	keyCache := KeyCachingID(id)
	err = r.cache.DeleteCaching(ctx, keyCache)
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
	err = r.cache.DeleteCachingAll(ctx)
	if err != nil {
		fmt.Println("cannot delete-all cache", err)

	}

	return nil
}

func (r *RepoCachingClipboard) GetAllUserClipboards(ctx context.Context, userID string) ([]model.Clipboard, error) {
	// cache
	clipboards, err := r.cache.GetAllCaching(ctx)
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
		keyCache := KeyCachingID(v.Id)
		err := r.cache.Sync(ctx, keyCache, v)
		if err != nil {
			fmt.Println("cannot sync cache", err)
			break
		}
	}

	return nil, nil

}

func (r *RepoCachingClipboard) GetUserClipboard(ctx context.Context, id string, userID string) (model.Clipboard, error) {
	//cache
	keyCache := KeyCachingID(id)
	clip, err := r.cache.GetCaching(ctx, keyCache)
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
	err = r.cache.Sync(ctx, keyCache, clip)
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

	err = r.cache.Sync(ctx, id, clip)
	if err != nil {
		fmt.Println("cannot sync cache:", err)
	}

	//cache_2
	clip2, err := r.cache.GetCaching(ctx, id)
	if err != nil {
		fmt.Println("cannot sync cache:", err)
	}

	keyCachee := KeyCachingID(id)
	err = r.cache.Sync(ctx, keyCachee, clip2)
	if err != nil {
		fmt.Println("cannot sync cache:", err)
	}

	//cache_3
	clip3 := model.Clipboard{
		Id:     id,
		UserId: userID,
		Text:   text,
	}

	err = r.cache.Sync(ctx, keyCachee, clip3)
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
	err = r.cache.DeleteCaching(ctx, id)
	if err != nil {
		fmt.Println("cannot delete cache", err)
	}

	return nil
}

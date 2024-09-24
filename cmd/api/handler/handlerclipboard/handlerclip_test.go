package handlerclipboard_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eymyong/drop/cmd/api/handler/handlerclipboard"
	"github.com/eymyong/drop/model"
	"github.com/eymyong/drop/repo"
	"github.com/eymyong/drop/repo/redisclipboard"
	"github.com/redis/go-redis/v9"
)

func flush(rd *redis.Client) {
	ctx := context.Background()
	err := rd.FlushDB(ctx).Err()
	if err != nil {
		panic(err)
	}
}

func Test_CreateClipHappy(t *testing.T) {
	rd := repo.NewRedis("167.179.66.149:6379", "", "Eepi2geeque2ahCo", 3)

	flush(rd)

	repo := redisclipboard.New(rd)
	handlerClipboard := handlerclipboard.NewClipboard(repo)

	response := httptest.NewRecorder()

	text := "clip-1"
	body := bytes.NewBufferString(text)
	request, err := http.NewRequest(http.MethodPost, "", body)
	if err != nil {
		t.Errorf("unexpected request err: %s", err.Error())
	}

	userID := "test-user-1"
	request.Header.Set("jwt-clipboard-user-id", userID)

	handlerClipboard.CreateClip(response, request)

	responseBody := response.Result().Body

	var result struct {
		Created model.Clipboard `json:"created"`
	}

	err = json.NewDecoder(responseBody).Decode(&result)
	if err != nil {
		panic(err)
	}

	// if result.Created.Id != "" {
	// 	t.Error("response body:", result)
	// }

	t.Log("response body:", result)

	ctx := context.Background()

	clip, err := repo.GetById(ctx, result.Created.Id)
	if err != nil {
		t.Errorf("failed to get clipboard from redis: %v", err)
	}

	t.Log("db data:", clip)

	if clip.UserId != userID {
		t.Errorf("unexpected user-id: expected='%s', actual='%s'", userID, clip.UserId)
	}

	if clip.Text != text {
		t.Errorf("unexpected text: expected='%s', actual='%s'", text, clip.Text)
	}

	// flush(rd)
}

func Test_GetAllClipHappy(t *testing.T) {
	rd := repo.NewRedis("167.179.66.149:6379", "", "Eepi2geeque2ahCo", 3)

	flush(rd)

	repo := redisclipboard.New(rd)
	handlerClipboard := handlerclipboard.NewClipboard(repo)

	clipsTest := []model.Clipboard{
		{
			Id:     "1",
			UserId: "xxx",
			Text:   "one",
		},
		{
			Id:     "2",
			UserId: "xxx",
			Text:   "two",
		},
	}

	ctx := context.Background()

	for _, v := range clipsTest {
		err := repo.Create(ctx, v)
		if err != nil {
			t.Errorf("unexpected err: %s", err.Error())
			return
		}
	}

	response := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		t.Errorf("unexpected err: %s", err.Error())
		return
	}

	handlerClipboard.GetAllClips(response, request)

	responseBody := response.Result().Body

	var result struct {
		Created []model.Clipboard `json:"created"`
	}

	err = json.NewDecoder(responseBody).Decode(&result)
	if err != nil {
		panic(err)
	}

	// for i, v := range result.Created {
	// 	if v != clipsTest[i] {
	// 		t.Errorf("expected cliptest:'%v',but got %v", clipsTest[i], v)
	// 	}
	// }

	if clipsTest[0] != result.Created[0] {
		t.Errorf("expected cliptest:'%v',but got %v", clipsTest[0], result.Created[0])
	}

	// flush(rd)

}

func Test_GetClipByIDHappy(t *testing.T) {
	rd := repo.NewRedis("167.179.66.149:6379", "", "Eepi2geeque2ahCo", 3)

	flush(rd)

	repo := redisclipboard.New(rd)
	handlerClipboard := handlerclipboard.NewClipboard(repo)

	clipExpexted := model.Clipboard{
		Id:     "2",
		UserId: "zzz",
		Text:   "two",
	}

	ctx := context.Background()
	err := repo.Create(ctx, clipExpexted)
	if err != nil {
		t.Errorf("unexpected err: %s", err.Error())
		return
	}

	response := httptest.NewRecorder()

	prat := clipExpexted.Id
	request, err := http.NewRequest(http.MethodGet, prat, nil)
	if err != nil {
		t.Errorf("unexpected request err: %s", err.Error())
	}

	userID := clipExpexted.UserId
	request.Header.Set("jwt-clipboard-user-id", userID)

	handlerClipboard.GetClipById(response, request)

	//t.Log("responseBody:", responseBody)

	responseBody := response.Result().Body

	buf := bytes.NewBuffer(nil)
	io.Copy(buf, responseBody)

	t.Log("buf:", string(buf.Bytes()))

	// 	var result struct {
	// 		Clipboard model.Clipboard `json:"created"`
	// 	}

	// 	err = json.NewDecoder(responseBody).Decode(&result)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	if result.Clipboard.Id == "" {
	// 		t.Error("result:", result.Clipboard)
	// 	}
	// 	//t.Log("result.Clipboard:", result.Clipboard)

	// 	if result.Clipboard != clipExpexted {
	// 		t.Errorf("expected clipExpexted:'%v' ,but got result:'%v'", clipExpexted, result.Clipboard)
	// 	}

}

func Test_UpdateClipByIDHappy(t *testing.T) {

}

func Test_DeleteClipHappy(t *testing.T) {

}

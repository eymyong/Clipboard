package handlerclipboard_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"

	"github.com/eymyong/drop/cmd/api/config"
	"github.com/eymyong/drop/cmd/api/handler/handlerclipboard"
	"github.com/eymyong/drop/model"
	"github.com/eymyong/drop/repo"
	"github.com/eymyong/drop/repo/redisclipboard"
)

func flush(rd *redis.Client) {
	ctx := context.Background()
	err := rd.FlushDB(ctx).Err()
	if err != nil {
		panic(err)
	}
}

func mockRegisterRoutesClipboardAPI(rd *redis.Client) *mux.Router {

	repo := redisclipboard.New(rd)
	handlerClipboard := handlerclipboard.NewClipboard(repo)

	r := mux.NewRouter()
	routerClip := r.PathPrefix("/clipboards").Subrouter()
	handlerclipboard.RegisterRoutesClipboardAPI(routerClip, handlerClipboard)

	return r

}

func parseClipboardForTest(rd *redis.Client, clipboard model.Clipboard, oder string, text string) (clips []model.Clipboard, clip model.Clipboard) {
	repo := redisclipboard.New(rd)

	ctx := context.Background()

	switch oder {
	case "create":
		err := repo.Create(ctx, clipboard)
		if err != nil {
			panic(err)
		}
		return

	case "get-all":
		clips, err := repo.GetAll(ctx)
		if err != nil {
			panic(err)
		}
		return clips, model.Clipboard{}

	case "get-by-id":
		clip, err := repo.GetById(ctx, clipboard.Id)
		if err != nil {
			panic(err)
		}
		return nil, clip

	case "update":
		err := repo.Update(ctx, clipboard.Id, text)
		if err != nil {
			panic(err)
		}
		return

	case "delete":
		err := repo.Delete(ctx, clipboard.Id)
		if err != nil {
			panic(err)
		}
		return
	}

	return nil, model.Clipboard{}

}

func Test_CreateClipErr(t *testing.T) {
	rd := repo.NewRedis("167.179.66.149:6379", "", "Eepi2geeque2ahCo", 3)
	flush(rd)

	//router := mockRegisterRoutesClipboardAPI(rd)
}

func ReadConfig(fileName string) config.Config {

	b, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	var env config.Config
	err = json.Unmarshal(b, &env)
	if err != nil {
		panic(err)
	}

	return env
}

func Test_CreateClipHappy(t *testing.T) {
	fileName := "../../../../config.json"
	conF := ReadConfig(fileName)

	rd := repo.NewRedis(conF.RedisAddr, conF.RedisUsername, conF.RedisPassword, conF.RedisDb)

	flush(rd)

	router := mockRegisterRoutesClipboardAPI(rd)

	response := httptest.NewRecorder()

	text := "clip-1"
	body := bytes.NewBufferString(text)
	request, err := http.NewRequest(http.MethodPost, "/clipboards/create", body)
	if err != nil {
		t.Errorf("unexpected request err: %s", err.Error())
	}
	// Set Header เนื่องจาก main API need use playload in request
	userID := "test-user-1"
	request.Header.Set("jwt-clipboard-user-id", userID)

	//call API
	router.ServeHTTP(response, request)

	responseBody := response.Result().Body

	var result struct {
		Created model.Clipboard `json:"created"`
	}

	err = json.NewDecoder(responseBody).Decode(&result)
	if err != nil {
		panic(err)
	}

	_, clip := parseClipboardForTest(rd, result.Created, "get-by-id", "")

	if clip.UserId != userID {
		t.Errorf("unexpected user-id: expected='%s', actual='%s'", userID, clip.UserId)
	}

	if clip.Text != text {
		t.Errorf("unexpected text: expected='%s', actual='%s'", text, clip.Text)
	}

}

func Test_GetClipByIDHappy(t *testing.T) {
	fileName := "../../../../config.json"
	conF := ReadConfig(fileName)

	rd := repo.NewRedis(conF.RedisAddr, conF.RedisUsername, conF.RedisPassword, conF.RedisDb)

	flush(rd)

	router := mockRegisterRoutesClipboardAPI(rd)

	clipExpected := model.Clipboard{
		Id:     "1",
		UserId: "zzz",
		Text:   "one",
	}

	parseClipboardForTest(rd, clipExpected, "create", "")

	response := httptest.NewRecorder()

	request, err := http.NewRequest(http.MethodGet, "/clipboards/get/"+clipExpected.Id, nil)
	if err != nil {
		t.Errorf("unexpected request err: %s", err.Error())
	}

	router.ServeHTTP(response, request)

	responseBody := response.Result().Body
	var result model.Clipboard
	err = json.NewDecoder(responseBody).Decode(&result)
	if err != nil {
		panic(err)
	}

	if clipExpected != result {
		t.Errorf("expected clipExpected: '%v' but got '%v'", clipExpected, result)
	}

	if clipExpected.Id != result.Id {
		t.Errorf("unexpected clipExpected.Id: '%v' != result.Id: '%v'", clipExpected.Id, result.Id)
	}

	if clipExpected.UserId != result.UserId {
		t.Errorf("unexpected clipExpected.UserId: '%v' != result.UserId: '%v'", clipExpected.UserId, result.UserId)
	}

	if clipExpected.Text != result.Text {
		t.Errorf("unexpected clipExpected.Text: '%v' != result.Text: '%v'", clipExpected.Text, result.Text)
	}
}

func Test_GetAllClipHappy(t *testing.T) {
	fileName := "../../../../config.json"
	conF := ReadConfig(fileName)

	rd := repo.NewRedis(conF.RedisAddr, conF.RedisUsername, conF.RedisPassword, conF.RedisDb)

	flush(rd)

	router := mockRegisterRoutesClipboardAPI(rd)

	clipExpected := []model.Clipboard{
		{
			Id:     "1",
			UserId: "xxx",
			Text:   "one",
		},
		{
			Id:     "2",
			UserId: "yyy",
			Text:   "two",
		},
	}

	for i := range clipExpected {
		parseClipboardForTest(rd, clipExpected[i], "create", "")
	}

	// //log มาดูแล้ว create มันไป create 'clipExpected[1]' ก่อน ,ซึ่งมันควนจะ create 'clipExpected[0]' ก่อน

	response := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/clipboards/get-all", nil)
	if err != nil {
		t.Errorf("unexpected err: %s", err.Error())
		return
	}

	router.ServeHTTP(response, request)

	responseBody := response.Result().Body

	var result []model.Clipboard
	err = json.NewDecoder(responseBody).Decode(&result)
	if err != nil {
		panic(err)
	}

	if len(clipExpected) != len(result) {
		t.Errorf("unexpect len 'result': %d != len 'clipExpected': %d", len(result), len(clipExpected))
		return
	}

	resultMap := make(map[string]model.Clipboard)

	for _, v := range result {
		key := v.Id
		resultMap[key] = v
	}

	for _, v := range clipExpected {
		clip, ok := resultMap[v.Id]
		if !ok {
			t.Errorf("unexpected not found Id: %s", v.Id)
		}

		if clip.Id != v.Id {
			t.Errorf("unexpected clipExpected.Id: '%s' != clip.Id: '%s'", v.Id, clip.Id)
		}

		if clip.UserId != v.UserId {
			t.Errorf("unexpected clipExpected.UserId: '%s' != clip.UserId: '%s'", v.UserId, clip.UserId)
		}

		if clip.Text != v.Text {
			t.Errorf("unexpected clipExpected.Text: '%s' != clip.Text: '%s'", v.Text, clip.Text)
		}

	}

}

func Test_UpdateClipByIDHappy(t *testing.T) {
	fileName := "../../../../config.json"
	conF := ReadConfig(fileName)

	rd := repo.NewRedis(conF.RedisAddr, conF.RedisUsername, conF.RedisPassword, conF.RedisDb)

	flush(rd)

	router := mockRegisterRoutesClipboardAPI(rd)

	clipExpected := model.Clipboard{
		Id:     "1",
		UserId: "zzz",
		Text:   "one",
	}

	parseClipboardForTest(rd, clipExpected, "create", "")

	response := httptest.NewRecorder()

	newText := "newONE"
	body := bytes.NewBufferString(newText)
	request, err := http.NewRequest(http.MethodPatch, "/clipboards/update/"+clipExpected.Id, body)
	if err != nil {
		t.Errorf("unexpected err: %s", err.Error())
	}

	router.ServeHTTP(response, request)

	_, actual := parseClipboardForTest(rd, clipExpected, "get-by-id", "")

	if actual.Text != newText {
		t.Errorf("expected newText: %s but got %s", newText, actual.Text)
	}

}

func Test_DeleteClipHappy(t *testing.T) {
	fileName := "../../../../config.json"
	conF := ReadConfig(fileName)

	rd := repo.NewRedis(conF.RedisAddr, conF.RedisUsername, conF.RedisPassword, conF.RedisDb)

	flush(rd)

	router := mockRegisterRoutesClipboardAPI(rd)

	clipExpected := model.Clipboard{
		Id:     "1",
		UserId: "zzz",
		Text:   "one",
	}

	parseClipboardForTest(rd, clipExpected, "create", "")

	response := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodDelete, "/clipboards/delete/"+clipExpected.Id, nil)
	if err != nil {
		t.Errorf("unexpected err: %s", err.Error())
	}

	router.ServeHTTP(response, request)

	ctx := context.Background()
	keys, err := rd.Keys(ctx, "*").Result()
	if err != nil {
		t.Errorf("unexpect err: %s", err.Error())
		return
	}

	if len(keys) != 0 {
		t.Errorf("unexpected leagth: %d", len(keys))
	}

}

func Test_GetClipByIDHappy0(t *testing.T) {
	fileName := "../../../../config.json"
	conF := ReadConfig(fileName)

	rd := repo.NewRedis(conF.RedisAddr, conF.RedisUsername, conF.RedisPassword, conF.RedisDb)

	flush(rd)

	repo := redisclipboard.New(rd)
	handlerClipboard := handlerclipboard.NewClipboard(repo)

	r := mux.NewRouter()
	routerClip := r.PathPrefix("/clipboards").Subrouter()
	handlerclipboard.RegisterRoutesClipboardAPI(routerClip, handlerClipboard)

	clipExpected := model.Clipboard{
		Id:     "1",
		UserId: "zzz",
		Text:   "one",
	}

	ctx := context.Background()
	err := repo.Create(ctx, clipExpected)
	if err != nil {
		t.Errorf("unexpected err: %s", err.Error())
		return
	}

	response := httptest.NewRecorder()

	request, err := http.NewRequest(http.MethodGet, "/clipboards/get/"+clipExpected.Id, nil)
	if err != nil {
		t.Errorf("unexpected request err: %s", err.Error())
	}

	r.ServeHTTP(response, request)

	// //read responseBody
	// responseBody := response.Result().Body

	// //สร้าง buf เพื่อจะ t.Log ดูของก่อน
	// buf := bytes.NewBuffer(nil)
	// io.Copy(buf, responseBody)

	// t.Log("buf:", string(buf.Bytes()))
	// t.Log("status", response.Result().Status)

	// var result model.Clipboard

	// //ที่ใส่ 'buf' เพราะอ่าน body ได้แค่ครั้งเดียวซึ่ง body ถูก responseBody อ่านไปแล้ว
	// err = json.NewDecoder(buf).Decode(&result)
	// if err != nil {
	// 	panic(err)
	// }

	// t.Log("result:", result)

	responseBody := response.Result().Body
	var result model.Clipboard
	err = json.NewDecoder(responseBody).Decode(&result)
	if err != nil {
		panic(err)
	}

	if clipExpected != result {
		t.Errorf("expected clipExpected: '%v' but got '%v'", clipExpected, result)
	}

	if clipExpected.Id != result.Id {
		t.Errorf("unexpected clipExpected.Id: '%v' != result.Id: '%v'", clipExpected.Id, result.Id)
	}

	if clipExpected.UserId != result.UserId {
		t.Errorf("unexpected clipExpected.UserId: '%v' != result.UserId: '%v'", clipExpected.UserId, result.UserId)
	}

	if clipExpected.Text != result.Text {
		t.Errorf("unexpected clipExpected.Text: '%v' != result.Text: '%v'", clipExpected.Text, result.Text)
	}
}

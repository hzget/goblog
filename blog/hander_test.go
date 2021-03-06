package blog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

const (
	cookie = "session_token=5a844737-62f2-4121-9402-3a538684e0d9; user=admin"
)

func TestMain(m *testing.M) {
	// <setup code>
	initGlobals()

	code := m.Run()

	// <tear-down code>

	// exit
	os.Exit(code)
}

func TestHandler(t *testing.T) {
	t.Run("Viewjs", func(t *testing.T) {
		doATest(t, makePageHandler(viewjsHandler), encodeJson(viewReq{1}), &viewResp{})
	})
	t.Run("Savejs", func(t *testing.T) {
		doATest(t, makePageHandler(savejsHandler), encodeJson(saveReq{1, "S", "nihao"}), &saveResp{})
	})
}

func doARequest(handler http.HandlerFunc, url, bodyJson string, data interface{}) error {

	// create req
	body := strings.NewReader(bodyJson)
	req := httptest.NewRequest("POST", url, body)
	req.Header.Set("Cookie", cookie)

	// send req
	w := httptest.NewRecorder()
	handler(w, req)

	// get response
	resp := w.Result()

	// check auth status
	if resp.StatusCode == http.StatusUnauthorized {
		return errors.New("StatusUnauthorized")
	}

	// read resp
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// decode resp
	if err := decodeJsonResp(bodyText, data); err != nil {
		return err
	}

	return nil
}

/*
 * doATest send a request and verify successful operation confirmed from server side
 *
 *      params:
 *              - url, bodyJson: is used to send request
 *              - data: response text will be decoded to its dynamic type value
 */
func doATest(i interface{}, handler http.HandlerFunc, bodyJson string, data interface{}) {

	if err := doARequest(handler, "/", bodyJson, data); err != nil {
		handleError(i, err)
		return
	}

	if !getRespStatus(data) {
		message := getRespMessage(data)
		err := fmt.Errorf("test failed: req [%s] and response status: false, message: %s",
			bodyJson, message)
		handleError(i, err)
	}
}

func handleError(i interface{}, err error) {

	if err == nil {
		return
	}

	switch v := i.(type) {
	case *testing.T:
		v.Fatal(err.Error())
	case *testing.B:
		v.Fatal(err.Error())
	default:
		fmt.Printf("unknown type %T, %v\n", v, v)
	}
}

func getRespStatus(data interface{}) bool {
	success := reflect.ValueOf(data).Elem().FieldByName("Success")
	return success.Bool()
}

func getRespMessage(data interface{}) string {
	message := reflect.ValueOf(data).Elem().FieldByName("Message")
	return message.String()
}

func decodeJsonResp(body []byte, data interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(data); err != nil {
		return err
	}

	return nil
}

package blog

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
)

var (
	cookie      = "session_token=5a844737-62f2-4121-9402-3a538684e0d9; user=admin"
	signin_url  = "http://127.0.0.1:8080/signin"
	signin_user = "admin"
	signin_pwd  = "admin"
)

func TestMain(m *testing.M) {
	// <setup code>
	// be carefull! you should change config/config.json according to reality
	//       address of mysql, redis, analysis center, and log file path
	// and then do a test
	initGlobals()

	// signin to get a token saving in the cookies
	if err := signinAndSaveCookie(); err != nil {
		fmt.Println(err)
		os.Exit(1)
		return
	}

	code := m.Run()

	// <tear-down code>

	closeLogFile()
	// exit
	os.Exit(code)
}

func doASignin(url, bodyJson string) *http.Response {

	// create req
	body := strings.NewReader(bodyJson)
	req := httptest.NewRequest("POST", url, body)

	// mock a signin req
	w := httptest.NewRecorder()
	makeAuthHandler(signinHandler)(w, req)

	// get response
	return w.Result()
}

func signinAndSaveCookie() error {

	resp := doASignin(signin_url, encodeJson(Credentials{signin_user, signin_pwd}))

	// read resp
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// decode resp
	result := &jsonResp{}
	if err := decodeJson(bodyText, result); err != nil {
		return err
	}

	// check result
	if !result.Success {
		return errors.New(result.Message)
	}

	//Cookies() []*Cookie
	cookies := resp.Cookies()
	var token, user string
	for _, v := range cookies {
		switch v.Name {
		case "session_token":
			token = v.Value
		case "user":
			user = v.Value
		}
	}

	if token != "" && user != "" {
		cookie = "session_token=" + token + "; user=" + user
		return nil
	}

	return fmt.Errorf("cookie is unexpectied:%v", cookies)

}

func TestJSHandler(t *testing.T) {
	t.Run("Viewjs", func(t *testing.T) {
		doATest(t, makePageHandler(viewjsHandler), encodeJson(viewReq{1}), &viewResp{})
	})
	t.Run("Savejs", func(t *testing.T) {
		doATest(t, makePageHandler(savejsHandler), encodeJson(saveReq{1, "S", "nihao"}), &saveResp{})
	})
}

func TestPressureViewJs(t *testing.T) {
	// update cache before check to avoid frequent update
	doATest(t, makePageHandler(viewjsHandler), encodeJson(viewReq{1}), &viewResp{})
	var wg sync.WaitGroup
	N := 1000
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			doATest(t, makePageHandler(viewjsHandler), encodeJson(viewReq{1}), &viewResp{})
		}()
	}
	wg.Wait()
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
	if err := decodeJson(bodyText, data); err != nil {
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

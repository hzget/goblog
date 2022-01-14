package clienttest

import (
    "testing"
    "fmt"
    "net/http"
    "strings"
    "io/ioutil"
    "encoding/json"
    "time"
    "bytes"
    "reflect"
)

const (
    cookie = "session_token=8adff996-4bab-43e7-b92e-ffe311ffd21a; user=admin"
    viewjs_url = "http://127.0.0.1:8080/viewjs"
    savejs_url = "http://127.0.0.1:8080/savejs"
)

type viewResp struct {
    Success  bool      `json:"success"`
    Message  string    `json:"message"`
    ID       int       `json:"id"`
    Title    string    `json:"title"`
    Author   string    `json:"author"`
    Date     time.Time `json:"date"`
    Modified time.Time `json:"modified"`
    Body     string    `json:"body"`
    Star    [5]int       `json:"star"`
}

type saveResp struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	ID      int    `json:"id"`
}

func TestView(t *testing.T) {
	client := &http.Client{}
	var data = strings.NewReader(`{"id":22}`)
	req, err := http.NewRequest("POST", viewjs_url, data)
	if err != nil {
		t.Fatalf(err.Error())
	}
	req.Header.Set("Cookie", cookie)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer resp.Body.Close()
    bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf(err.Error())
	}

    vdata := viewResp{}

    if err:= verifyJsonResp(bodyText, &vdata); err != nil {
		t.Fatalf(err.Error())
    }

    fmt.Println(vdata)
}

func TestCreate(t *testing.T) {
	client := &http.Client{}
	var data = strings.NewReader(`{"id":0, "title":"hello", "body":"你好"}`)
	req, err := http.NewRequest("POST", savejs_url, data)
	if err != nil {
		t.Fatalf(err.Error())
	}
	req.Header.Set("Cookie", cookie)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf(err.Error())
	}

    if err:= verifyJsonResp(bodyText, &saveResp{}); err != nil {
		t.Fatalf(err.Error())
    }
}

func verifyJsonResp(body []byte, resp interface{}) error {

    if err := decodeJsonResp(body, resp); err != nil {
        return err
    }

    success := reflect.ValueOf(resp).Elem().FieldByName("Success")

    if !success.Bool() {
        message := reflect.ValueOf(resp).Elem().FieldByName("Message")
        return fmt.Errorf("response return success-false, message: %v", message.String())
    }

    return nil
}

func decodeJsonResp(body []byte, resp interface{}) error {
    decoder := json.NewDecoder(bytes.NewReader(body))
    decoder.DisallowUnknownFields()
    if err := decoder.Decode(resp); err != nil {
        return err
    }

    return nil
}


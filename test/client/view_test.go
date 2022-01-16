package clienttest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	cookie     = "session_token=b0a7abe6-3296-46f1-8138-9dd04bd3243d; user=admin"
	viewjs_url = "http://127.0.0.1:8080/viewjs"
	savejs_url = "http://127.0.0.1:8080/savejs"
)

type viewReq struct {
	ID int64 `json:"id"`
}

type saveReq struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

type viewResp struct {
	Success  bool      `json:"success"`
	Message  string    `json:"message"`
	ID       int64     `json:"id"`
	Title    string    `json:"title"`
	Author   string    `json:"author"`
	Date     time.Time `json:"date"`
	Modified time.Time `json:"modified"`
	Body     string    `json:"body"`
	Star     [5]int    `json:"star"`
}

type saveResp struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	ID      int64  `json:"id"`
}

func encodeJson(data interface{}) string {
	b, err := json.MarshalIndent(data, "", "")
	if err != nil {
		panic(err)
	}
	return string(b)
}

func TestViewNtimes(t *testing.T) {
	N := 10_000
	for i := 0; i < N; i++ {
		doTestView(t, 22)
	}
}

func TestViewCases(t *testing.T) {

	cases := []struct {
		Body           string
		ExpectedStatus bool
	}{
		{`{"id", :22}`, false}, // negative case:  malformed json format
		{`{"i" :22}`, false},   // negative case:  unknown json key
		{`{"id": -1}`, false},  // negative case:  invalid id
		{`{"id": 0}`, false},   // negative case:  invalid id
		{`{"id": 23}`, true},   // positive case:  valid id
	}

	for _, c := range cases {
		doTestAndVerifyStatus(t, viewjs_url, c.Body, &viewResp{}, c.ExpectedStatus)
	}
}

func TestSaveAndViewNtimes(t *testing.T) {

	N := 10
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			doTestSaveAndView(t, 23, "hello", "hei")
		}()
	}
	wg.Wait()
}

func TestSaveAndViewCases(t *testing.T) {

	return
	cases := []saveReq{
		{0, "hello", "你好"}, // create a post
		{23, "hi", "你好\n"}, // modify a post
		{23, "hen 好", "你\t好\n"},
	}

	for _, c := range cases {
		doTestSaveAndView(t, c.ID, c.Title, c.Body)
	}
}

/*
 * doARequest send a request and decode the response
 *
 *      params: data: to store decoded response
 *
 */
func doARequest(t *testing.T, url, bodyJson string, data interface{}) {

	// create req
	client := &http.Client{}
	body := strings.NewReader(bodyJson)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		t.Fatalf(err.Error())
	}
	req.Header.Set("Cookie", cookie)

	// send req
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer resp.Body.Close()

	// check auth status
	if resp.StatusCode == http.StatusUnauthorized {
		t.Fatalf("StatusUnauthorized")
	}

	// read resp
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf(err.Error())
	}

	//t.Logf("req: [%s]", bodyJson)
	// decode resp and verify the success status field
	if err := decodeJsonResp(bodyText, data); err != nil {
		t.Fatalf(err.Error())
	}
	//t.Logf("resp: [%v]", data)
}

/*
 * doATest send a request and verify successful operation confirmed from server side
 *
 *      params:
 *              - url, bodyJson: is used to send request
 *              - data: response text will be decoded to its dynamic type value
 */
func doATest(t *testing.T, url, bodyJson string, data interface{}) {

	doARequest(t, url, bodyJson, data)

	if err := verifyStatusOK(data); err != nil {
		t.Fatalf(err.Error())
	}
}

/*
 * send a request and verify the status field
 *   it can be used to test kinds of different input and output
 *
 */
func doTestAndVerifyStatus(t *testing.T, url, bodyJson string, data interface{}, expectedStatus bool) {

	doARequest(t, url, bodyJson, data)

	if verified, err := verifyStatus(data, expectedStatus); !verified {
		t.Fatalf("response expected status %v is not met. message %v ", expectedStatus, err)
	}
}

func doTestSave(t *testing.T, id int64, title, bodytext string) {

	sreq := saveReq{id, title, bodytext}
	body := encodeJson(sreq)
	data := &saveResp{}
	doATest(t, savejs_url, body, data)
}

func doTestView(t *testing.T, id int64) {

	vreq := viewReq{id}
	body := encodeJson(vreq)
	data := &viewResp{}
	doATest(t, viewjs_url, body, data)
}

/*
 * id:  == 0 - it is a create operation
 *       > 0 - it is a save operation
 */
func doTestSaveAndView(t *testing.T, id int64, title, body string) {

	// save a post
	sreq := saveReq{id, title, body}
	sbody := encodeJson(sreq)
	sdata := &saveResp{}
	doATest(t, savejs_url, sbody, sdata)

	var testId int64
	if id == 0 {
		testId = sdata.ID
	} else {
		testId = id
	}

	// view a post
	greq := viewReq{sdata.ID}
	gbody := encodeJson(greq)
	gdata := &viewResp{}
	doATest(t, viewjs_url, gbody, gdata)

	// compare create and view
	if testId != gdata.ID {
		t.Fatalf("save and view id is different: [%d] vs [%d]", testId, gdata.ID)
	}

	if title != gdata.Title {
		t.Fatalf("save and view title is different: [%s] vs [%s]", title, gdata.Title)
	}

	if body != gdata.Body {
		t.Fatalf("save and view body is different: [%s] vs [%s]", body, gdata.Body)
	}
}

func verifyStatusOK(data interface{}) error {

	success := reflect.ValueOf(data).Elem().FieldByName("Success")

	if !success.Bool() {
		message := reflect.ValueOf(data).Elem().FieldByName("Message")
		return fmt.Errorf("response return success-false, message: %v", message.String())
	}

	return nil
}

func verifyStatus(data interface{}, expectedStatus bool) (bool, error) {

	success := reflect.ValueOf(data).Elem().FieldByName("Success")
	var err error

	if !success.Bool() {
		message := reflect.ValueOf(data).Elem().FieldByName("Message")
		err = fmt.Errorf("response return success-false, message: %v", message.String())
	}

	return success.Bool() == expectedStatus, err
}

func decodeJsonResp(body []byte, data interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(data); err != nil {
		return err
	}

	return nil
}

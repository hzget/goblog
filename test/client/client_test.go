package clienttest

import (
	"bytes"
	"encoding/json"
	"errors"
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
	cookie     = "session_token=5a844737-62f2-4121-9402-3a538684e0d9; user=admin"
	viewjs_url = "http://127.0.0.1:8080/viewjs"
	savejs_url = "http://127.0.0.1:8080/savejs"
)

type viewReq struct {
	Id int64 `json:"id"`
}

type saveReq struct {
	Id    int64  `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

type viewResp struct {
	Success  bool      `json:"success"`
	Message  string    `json:"message"`
	Id       int64     `json:"id"`
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
	Id      int64  `json:"id"`
}

func BenchmarkViewN(b *testing.B) {
	for i := 0; i < b.N; i++ {
		doATest(b, viewjs_url, encodeJson(viewReq{1}), &viewResp{})
	}
}

func TestPressure(t *testing.T) {
	// <setup code>
	t.Run("NParallelView", func(t *testing.T) {
		doTestNParallel(t, viewjs_url, encodeJson(viewReq{1}), &viewResp{}, 150)
	})
	t.Run("NParallelSave", func(t *testing.T) {
		doTestNParallel(t, savejs_url, encodeJson(saveReq{2, "morning", "zao"}), &saveResp{}, 150)
	})
	t.Run("NSequentialView", func(t *testing.T) {
		doTestNSequential(t, viewjs_url, encodeJson(viewReq{1}), &viewResp{}, 150)
	})
	t.Run("NSequentialSave", func(t *testing.T) {
		doTestNSequential(t, savejs_url, encodeJson(saveReq{2, "morning", "zao"}), &saveResp{}, 150)
	})
	// <tear-down code>
}

func TestViewCases(t *testing.T) {

	cases := []struct {
		Body           string
		ExpectedStatus bool
	}{
		{`{"id", :1}`, false}, // negative case:  malformed json format
		{`{"i" :1}`, false},   // negative case:  unknown json key
		{`{"id": -1}`, false}, // negative case:  invalid id
		{`{"id": 0}`, false},  // negative case:  invalid id
		{`{"id": 1}`, true},   // positive case:  valid id
	}

	for _, c := range cases {
		doTestAndVerifyStatus(t, viewjs_url, c.Body, &viewResp{}, c.ExpectedStatus)
	}
}

func TestSaveCases(t *testing.T) {

	cases := []struct {
		Body           string
		ExpectedStatus bool
	}{
		{`{"id": 2, "title":, "body":"??????"}`, false},           // negative case:  malformed json format
		{`{"id": 2, "title", "body":"??????"}`, false},            // negative case:  malformed json format
		{`{"id": 2, "title"body":"??????"}`, false},               // negative case:  malformed json format
		{`{"id": 2, "body":"??????", "name":"hehe"}`, false},      // negative case:  unknown json key
		{`{"id": -2, "title":"hello", "body":"??????"}`, false},   // negative case:  invalid id
		{`{"id": "yo", "title":"hello", "body":"??????"}`, false}, // negative case:  invalid id
		{`{"id": 2, "title":"", "body":"??????"}`, false},         // negative case:  invalid title
		{`{"id": 2, "title":" ", "body":"??????"}`, false},        // negative case:  invalid title
		{`{"id": 2, "title":"hello", "body":"??????"}`, true},     // positive case:  valid id, title, body
		{`{"id": 0, "title":"hello", "body":"??????"}`, true},     // positive case:  create
	}

	for _, c := range cases {
		doTestAndVerifyStatus(t, savejs_url, c.Body, &saveResp{}, c.ExpectedStatus)
	}
}

func TestSaveAndViewCases(t *testing.T) {

	cases := []saveReq{
		{0, "hello", "??????"}, // create a post
		{2, "hen ???", "???\t???\n"},
	}

	for _, c := range cases {
		doTestSaveAndView(t, c.Id, c.Title, c.Body)
	}
}

/**************** helper function *****************/

/*
 * doARequest send a request and decode the response
 *
 *      params: data: to store decoded response
 *
 */
func doARequest(url, bodyJson string, data interface{}) error {

	// create req
	client := &http.Client{}
	body := strings.NewReader(bodyJson)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Cookie", cookie)

	// send req
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

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
func doATest(i interface{}, url, bodyJson string, data interface{}) {

	if err := doARequest(url, bodyJson, data); err != nil {
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

/*
 * send a request and verify the status field
 *   it can be used to test kinds of different input and output
 *
 */
func doTestAndVerifyStatus(i interface{}, url, bodyJson string, data interface{}, expectedStatus bool) {

	if err := doARequest(url, bodyJson, data); err != nil {
		handleError(i, err)
		return
	}

	if getRespStatus(data) != expectedStatus {
		message := getRespMessage(data)
		err := fmt.Errorf("req [%s] expect status '%v' is not met. message: %s",
			bodyJson, expectedStatus, message)
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

func doTestNParallel(t *testing.T, url, bodyJson string, data interface{}, N int) {
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			doATest(t, url, bodyJson, data)
		}()
	}
	wg.Wait()
}

func doTestNSequential(t *testing.T, url, bodyJson string, data interface{}, N int) {
	for i := 0; i < N; i++ {
		doATest(t, url, bodyJson, data)
	}
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
		testId = sdata.Id
	} else {
		testId = id
	}

	// view a post
	greq := viewReq{sdata.Id}
	gbody := encodeJson(greq)
	gdata := &viewResp{}
	doATest(t, viewjs_url, gbody, gdata)

	// compare create and view
	if testId != gdata.Id {
		t.Fatalf("save and view id is different: [%d] vs [%d]", testId, gdata.Id)
	}

	if title != gdata.Title {
		t.Fatalf("save and view title is different: [%s] vs [%s]", title, gdata.Title)
	}

	if body != gdata.Body {
		t.Fatalf("save and view body is different: [%s] vs [%s]", body, gdata.Body)
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

func encodeJson(data interface{}) string {
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return string(b)
}

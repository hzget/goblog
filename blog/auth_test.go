package blog

import (
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestCredentialsValidate(t *testing.T) {
	cases := []struct {
		Credentials
		want bool
	}{
		{Credentials{"a", "pwd"}, false},
		{Credentials{"abc12345678", "pwd"}, false},
		{Credentials{"a_bc", "pwd"}, false},

		{Credentials{"abc", "pwd"}, true},
		{Credentials{"1234", "pwd"}, true},
		{Credentials{"abc1234567", "pwd"}, true},
	}

	for _, v := range cases {
		got, err := v.Credentials.Validate()
		if err != nil {
			t.Error(err)
			continue
		}
		if got != v.want {
			t.Errorf("creds %#v: got %v, want %v",
				v.Credentials, got, v.want)
		}
	}
}

func TestCredentialsSave(t *testing.T) {
	creds := Credentials{"Jack&Lucy", "abc123"}
	err := creds.save()
	if err != nil {
		t.Fatal(err)
	}

	existed, err := checkUserExist(creds.Username)
	if err != nil {
		t.Fatal(err)
	}
	if !existed {
		t.Fatalf("does not find user %s in the datastore",
			creds.Username)
	}

	pwd, err := getPassword(creds.Username)
	if err != nil {
		t.Error(err)
		goto end
	}

	if err := validateHash(creds.Password, pwd); err != nil {
		t.Errorf("save credentials %#v, but "+
			"get wrong password %s, error is %v",
			creds, pwd, err)
		goto end
	}

end:
	if err := creds.remove(); err != nil {
		t.Log(err)
	}
}

func TestCredentialsRemove(t *testing.T) {
	creds := Credentials{"Lily", "abc123"}
	if err := creds.save(); err != nil {
		t.Fatal(err)
	}

	existed, err := checkUserExist(creds.Username)
	if err != nil {
		t.Fatal(err)
	}
	if !existed {
		t.Fatalf("does not find user %s in the datastore", creds.Username)
	}

	if err := creds.remove(); err != nil {
		t.Fatal(err)
	}

	existed, err = checkUserExist(creds.Username)
	if err != nil {
		t.Fatal(err)
	}
	if existed {
		t.Fatalf("find user %s in the datastore", creds.Username)
	}
}

func TestSignupHandlerWrapper(t *testing.T) {
	body := `{"username":"Lucy", "password":"12345"}`
	code := http.StatusOK
	expected := jsonResp{true, "signup success"}
	t.Run("Success", signupWrapper(t, body, code, &expected))

	code = http.StatusBadRequest
	expected = jsonResp{false, "user already exists, please choose another name"}
	t.Run("UserExists", signupWrapper(t, body, code, &expected))
	(&Credentials{Username: "Lucy"}).remove()

	body = `{"username":"1234567890a", "password":"12345"}`
	expected = jsonResp{false, "invalid username"}
	t.Run("InvalidUsername", signupWrapper(t, body, code, &expected))

	body = `{"usernames":"abc", "password":"12345"}`
	expected = jsonResp{false, "invalid username"}
	t.Run("UnknownJsonField", signupWrapper(t, body, code, nil))

	body = `{"usernam:"abc", "password":"12345"}`
	expected = jsonResp{false, "invalid username"}
	t.Run("MalFormedRequest", signupWrapper(t, body, code, nil))
}

// if expected is nil, do not verify expected msg
func signupWrapper(t *testing.T, body string, code int, expected *jsonResp) func(*testing.T) {

	return func(t *testing.T) {

		handler := makeAuthHandler(signupHandler)

		reqbody := strings.NewReader(body)
		req := httptest.NewRequest("", "/signup", reqbody)
		w := httptest.NewRecorder()
		handler(w, req)

		resp := w.Result()
		body, _ := io.ReadAll(resp.Body)
		//t.Log(string(body))
		if resp.StatusCode != code {
			t.Fatalf("code expected %v, got %v, body: %v",
				code, resp.StatusCode, string(body))
		}

		if expected == nil {
			return
		}

		got := &jsonResp{}
		if err := decodeJsonResp(body, got); err != nil {
			t.Fatalf("fail to decode body %s, error: %v",
				string(body), err)
		}
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("body is unexpected. want %v, but got %v",
				encodeJson(expected), string(body))
		}
	}
}

func TestSigninHandlerWrapper(t *testing.T) {

	// signup a new user
	creds := Credentials{"Lucy", "123"}
	if err := creds.save(); err != nil {
		t.Fatal(err)
	}
	defer creds.remove()

	body := `{"username":"Lucy", "password":"123"}`
	code := http.StatusOK
	expected := jsonResp{true, "signin success"}
	t.Run("Success", signinWrapper(t, body, code, &expected))

	body = `{"username":"Lucya", "password":"123"}`
	code = http.StatusUnauthorized
	expected = jsonResp{false, "no such user Lucya"}
	t.Run("UserNotExist", signinWrapper(t, body, code, &expected))

	body = `{"username":"Lucy", "password":"1234"}`
	code = http.StatusUnauthorized
	expected = jsonResp{false, "failed to validate password"}
	t.Run("PasswordFailed", signinWrapper(t, body, code, &expected))

}

// if expected is nil, do not verify expected msg
func signinWrapper(t *testing.T, body string, code int, expected *jsonResp) func(*testing.T) {

	return func(t *testing.T) {

		handler := makeAuthHandler(signinHandler)

		creds := Credentials{}
		decoder := json.NewDecoder(strings.NewReader(body))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&creds); err != nil {
			t.Errorf("fail to decode creds: %v", err)
			return
		}

		reqbody := strings.NewReader(body)
		req := httptest.NewRequest("", "/signin", reqbody)
		w := httptest.NewRecorder()
		handler(w, req)

		resp := w.Result()
		body, _ := io.ReadAll(resp.Body)
		// t.Logf("body: %s", string(body))
		if resp.StatusCode != code {
			t.Fatalf("code expected %v, got %v, body: %v",
				code, resp.StatusCode, string(body))
		}

		if expected != nil {

			got := &jsonResp{}
			if err := decodeJsonResp(body, got); err != nil {
				t.Fatalf("fail to decode body %s, error: %v",
					string(body), err)
			}
			if !reflect.DeepEqual(got, expected) {
				t.Fatalf("body is unexpected. want %v, but got %v",
					encodeJson(expected), string(body))
			}

		}

		if resp.StatusCode != http.StatusOK {
			return
		}

		// check cookies
		cookies := resp.Cookies()
		//t.Logf("cookies: %v", cookies)
		var token, user, tokenpath, userpath string
		for _, v := range cookies {
			switch v.Name {
			case "session_token":
				token = v.Value
				tokenpath = v.Path
			case "user":
				user = v.Value
				userpath = v.Path
			}
		}

		if token == "" {
			t.Errorf("not get session_token inside resp cookies")
			return
		}

		if tokenpath != "/" {
			t.Errorf("cookie: session_token(%s)'s path "+
				"expected %s, got %s", token, "/", tokenpath)
			return
		}

		if user != creds.Username {
			t.Errorf("not get user [%s] inside resp cookies", creds.Username)
			return
		}

		if userpath != "/" {
			t.Errorf("cookie: user(%s)'s path "+
				"expected %s, got %s", user, "/", userpath)
			return
		}

		// check session data in datastore
		tokenid, err := checkKey(user)
		switch {
		case err == redis.Nil:
			t.Errorf("no session data for %s", user)
			return
		case err != nil:
			t.Fatal(err)
		}

		if token != tokenid {
			t.Errorf("session data for %s has wrong token: "+
				"client get %v, in datastore %v", user, token, tokenid)
			return
		}

		defer removeKey(user)
	}
}
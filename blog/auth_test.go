package blog

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
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
		if err := decodeJson(body, got); err != nil {
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
			if err := decodeJson(body, got); err != nil {
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

func signin(t *testing.T, username, password string) {
	body := `{"username":"` + username + `", "password":"` + password + `"}`
	req := httptest.NewRequest("", "/signin", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	makeAuthHandler(signinHandler)(w, req)
	if resp := w.Result(); resp.StatusCode != http.StatusOK {
		t.Fatalf("signin failed, status: %d", resp.StatusCode)
	}
}

func doALogout(mycookie string) *http.Response {
	req := httptest.NewRequest("", "/logout", new(bytes.Buffer))
	req.Header.Set("Cookie", mycookie)
	w := httptest.NewRecorder()
	logoutHandler(w, req)
	return w.Result()
}

func verifyLogoutResp(t *testing.T, resp *http.Response, code int) {
	if resp.StatusCode != code {
		t.Fatalf("expect http code %d, but got %d",
			code, resp.StatusCode)
	}

	cookies := resp.Cookies()
	verifyClearCookies(t, cookies)
}

func TestLogoutHandler(t *testing.T) {

	// signup a new user
	creds := Credentials{"Lucy", "123"}
	if err := creds.save(); err != nil {
		t.Fatal(err)
	}
	defer creds.remove()

	// signin
	signin(t, "Lucy", "123")
	token, err := checkKey("Lucy")
	if err != nil {
		t.Fatal(err)
	}
	defer removeKey("Lucy")

	var resp *http.Response

	t.Run("Unauthorized", func(t *testing.T) {
		resp = doALogout("session_token=2" + token + "; user=Lucy")
		verifyLogoutResp(t, resp, http.StatusUnauthorized)
		if _, err := checkKey("Lucy"); err != nil {
			t.Fatalf("expect nil, but got %v", err)
		}
	})

	t.Run("Unauthorized", func(t *testing.T) {
		resp = doALogout("session_token=" + token + "; user=Lucys")
		verifyLogoutResp(t, resp, http.StatusUnauthorized)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		resp = doALogout("session_token=" + token + ";")
		verifyLogoutResp(t, resp, http.StatusUnauthorized)
	})

	t.Run("StatusOK", func(t *testing.T) {
		resp = doALogout("session_token=" + token + "; user=Lucy")
		verifyLogoutResp(t, resp, http.StatusOK)
		if _, err := checkKey("Lucy"); err != redis.Nil {
			t.Fatalf("expect redis.Nil, but got %v", err)
		}
	})

	t.Run("Unauthorized", func(t *testing.T) {
		resp = doALogout("session_token=" + token + "; user=Lucy")
		verifyLogoutResp(t, resp, http.StatusUnauthorized)
	})
}

func TestValidateSession(t *testing.T) {

	cases := []struct {
		name string
		c    string
		user string
		err  error
	}{
		{"AlreadyLogin", cookie, "admin", nil},
		{"LackOfToken", "", "", http.ErrNoCookie},
		{"LackOfUser", "session_token=5", "", http.ErrNoCookie},
		{"NoSuchUser", "session_token=5; user=whoareyou", "", redis.Nil},
		{"UnmatchToken", "session_token=5; user=admin", "", ErrCacheTokenUnmatch},
	}

	w := httptest.NewRecorder()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			r := httptest.NewRequest("POST", "/", new(bytes.Buffer))
			r.Header.Set("Cookie", tc.c)

			user, err := ValidateSession(w, r)
			if !errors.Is(err, tc.err) {
				t.Fatalf("expect error %v but got %v", tc.err, err)
			}
			if user != tc.user {
				t.Fatalf("expect user %q but got %q", tc.user, user)
			}
		})
	}
}

func TestClearCookies(t *testing.T) {
	w := httptest.NewRecorder()
	clearCookies(w)
	resp := w.Result()
	cookies := resp.Cookies()
	verifyClearCookies(t, cookies)
}

func verifyClearCookies(t *testing.T, cookies []*http.Cookie) {
	var token, user bool
	for _, c := range cookies {
		if c.Name == "session_token" {
			validateCookie(t, c, &http.Cookie{Name: c.Name, Path: "/", MaxAge: -1})
			token = true
		} else if c.Name == "user" {
			validateCookie(t, c, &http.Cookie{Name: c.Name, Path: "/", MaxAge: -1})
			user = true
		}
	}
	if !token || !user {
		t.Fatalf("response cookie not contain \"session_token\" or \"user\""+
			".\n---cookies---\n%v\n---end---\n", cookies)
	}
}

func validateCookie(t *testing.T, c, expected *http.Cookie) {
	if c.Value != expected.Value {
		t.Fatalf("cookie %q expect Value %q, but got %q", c.Name, expected.Value, c.Value)
	}
	if c.Path != expected.Path {
		t.Fatalf("cookie %q expect Path %q, but got %q", c.Name, expected.Path, c.Path)
	}
	expiretime := time.Now().Add(-7 * 24 * time.Hour)
	if expiretime.Before(c.Expires) {
		t.Fatalf("cookie %q expect Expires %v, but got %v", c.Name, expiretime, c.Expires)
	}
	if c.MaxAge != expected.MaxAge {
		t.Fatalf("cookie %q expect MaxAge %d, but got %d", c.Name, expected.MaxAge, c.MaxAge)
	}
}

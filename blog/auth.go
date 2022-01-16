package blog

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"regexp"
	"time"
)

const sessionTimeout = 30 * 24 * time.Hour

type SessionStatus uint

const (
	SessionAuthorized    = 0
	SessionUnauthorized  = 1
	SessionInternalError = 2
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// remaining:
//   add contraint: validate pwd regexp
func (creds *Credentials) Validate() (bool, error) {
	return regexp.MatchString(`^[0-9a-zA-Z]{3,10}$`, creds.Username)
}

func (creds *Credentials) save() error {

	hash, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	q := `INSERT INTO users (username, password, rank) VALUES (?, ?, ?)`

	_, err = db.ExecContext(ctx, q, creds.Username, string(hash), "bronze")

	return err
}

func checkUserExist(username string) (bool, error) {

	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	var exist = false
	q := `SELECT (count(*)>0) from users WHERE username = ?`
	err := db.QueryRowContext(ctx, q, username).Scan(&exist)

	return exist, err
}

func getPassword(username string) (string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	var password string
	q := `SELECT password FROM users WHERE username = ?`
	err := db.QueryRowContext(ctx, q, username).Scan(&password)

	return password, err
}

func makeAuthHandler(fn func(http.ResponseWriter, *http.Request, *Credentials) *appError) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var e *appError
		var creds = &Credentials{}

		var isvalid = false
		var err error

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err = decoder.Decode(creds); err != nil {
			e = &appError{err, http.StatusBadRequest}
			goto Err
		}

		// validate legal username
		if isvalid, err = creds.Validate(); err != nil {
			e = &appError{errors.New("internal error happened when validate credentials"),
				http.StatusInternalServerError}
			goto Err
		}

		if !isvalid {
			e = &appError{errors.New("invalid username"), http.StatusBadRequest}
			goto Err
		}

		if e = fn(w, r, creds); e == nil {
			return
		}

	Err:
		if e.Code == http.StatusInternalServerError {
			fmt.Println(e.Error)
		}
		http.Error(w, encodeJsonResp(false, e.Error.Error()), e.Code)
	}
}

func signupHandler(w http.ResponseWriter, r *http.Request, creds *Credentials) *appError {

	exist, err := checkUserExist(creds.Username)
	if err != nil {
		return &appError{err, http.StatusInternalServerError}
	}

	if exist {
		return &appError{errors.New("user already exists, please choose another name"),
			http.StatusBadRequest}
	}

	if err = creds.save(); err != nil {
		return &appError{fmt.Errorf("fail to save creds %v, err info:%v\n", creds, err),
			http.StatusInternalServerError}
	}

	fmt.Fprintf(w, encodeJsonResp(true, "signup success"))

	return nil
}

func signinHandler(w http.ResponseWriter, r *http.Request, creds *Credentials) *appError {

	hash, err := getPassword(creds.Username)
	switch {
	case err == sql.ErrNoRows:
		return &appError{fmt.Errorf("no such user %v, %v", creds.Username, err), http.StatusUnauthorized}
	case err != nil:
		return &appError{fmt.Errorf("fail to get password for user %v, %v", creds.Username, err),
			http.StatusInternalServerError}
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(creds.Password))
	if err != nil {
		return &appError{fmt.Errorf("failed to validate password"), http.StatusUnauthorized}
	}

	token := uuid.NewString()
	err = rdb.Set(context.Background(), token, creds.Username, sessionTimeout).Err()
	if err != nil {
		return &appError{fmt.Errorf("fail to set token for user %v", creds.Username),
			http.StatusInternalServerError}
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   token,
		Path:    "/",
		Expires: time.Now().Add(sessionTimeout),
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "user",
		Value:   creds.Username,
		Path:    "/",
		Expires: time.Now().Add(sessionTimeout),
	})

	fmt.Fprintf(w, encodeJsonResp(true, "signin success"))
	return nil
}

func ValidateSession(w http.ResponseWriter, r *http.Request) (string, SessionStatus) {
	c, err := r.Cookie("session_token")
	switch {
	case err == http.ErrNoCookie:
		return "", SessionUnauthorized
	case err != nil:
		fmt.Printf("internal error: %v\n", err)
		return "", SessionInternalError
	}

	cuser, err := r.Cookie("user")
	switch {
	case err == http.ErrNoCookie:
		return "", SessionUnauthorized
	case err != nil:
		fmt.Printf("internal error: %v\n", err)
		return "", SessionInternalError
	}

	user, err := rdb.Get(context.Background(), c.Value).Result()
	switch {
	case err == redis.Nil:
		return "", SessionUnauthorized
	case err != nil:
		fmt.Printf("internal error: %v\n", err)
		return "", SessionInternalError
	}

	//fmt.Printf("sid %v, suser %v, user %v\n", c.Value, cuser.Value, user)
	if user != cuser.Value {
		fmt.Printf("error: user unmatched %v, %v\n", user, cuser.Value)
		clearCookies(w)
		return "", SessionUnauthorized
	}

	return user, SessionAuthorized
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {

	var e *appError
	c, err := r.Cookie("session_token")
	switch {
	case err == http.ErrNoCookie:
		clearCookies(w)
		fmt.Fprintf(w, encodeJsonResp(true, "no cookie item, but will clear via set-cookie"))
		return
	case err != nil:
		e = &appError{err, http.StatusInternalServerError}
		goto Err
	}

	err = rdb.Del(context.Background(), c.Value).Err()
	if err != nil {
		e = &appError{fmt.Errorf("%v: fail to del token %v", err, c.Value),
			http.StatusInternalServerError}
		goto Err
	}

	clearCookies(w)

	fmt.Fprintf(w, encodeJsonResp(true, "logout success"))
	return

Err:
	if e.Code == http.StatusInternalServerError {
		fmt.Println(e.Error)
	}
	http.Error(w, encodeJsonResp(false, e.Error.Error()), e.Code)
}

func clearCookies(w http.ResponseWriter) {
	expiretime := time.Now().Add(-7 * 24 * time.Hour)
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   "",
		MaxAge:  -1,
		Path:    "/",
		Expires: expiretime,
	})

	http.SetCookie(w, &http.Cookie{
		Name:    "user",
		Value:   "",
		MaxAge:  -1,
		Path:    "/",
		Expires: expiretime,
	})

}

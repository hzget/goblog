package blog

import (
	"context"
	"database/sql"
	"encoding/json"
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
//
//	add contraint: validate pwd regexp
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

	q := "INSERT INTO users (username, password, `rank`) VALUES (?, ?, ?)"

	_, err = db.ExecContext(ctx, q, creds.Username, string(hash), "bronze")

	return err
}

func (creds *Credentials) remove() error {

	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	q := "DELETE FROM users where username = ?"

	_, err := db.ExecContext(ctx, q, creds.Username)

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

func validateHash(origin, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(origin))
}

func verifyCredentials(r *http.Request) (*Credentials, error) {

	var creds = &Credentials{}

	var isvalid = false
	var err error

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(creds); err != nil {
		return nil, NewRespErr(ErrCredentialFailed, http.StatusBadRequest,
			"fail to decode creds: "+err.Error())
	}

	// validate legal username
	if isvalid, err = creds.Validate(); err != nil {
		return nil, NewRespErr(ErrCredentialFailed, http.StatusInternalServerError,
			err.Error())
	}

	if !isvalid {
		return nil, NewRespErr(ErrCredentialFailed, http.StatusBadRequest,
			"invalid username.")
	}

	return creds, nil
}

func signupHandler(w http.ResponseWriter, r *http.Request) {

	creds, err := verifyCredentials(r)
	if err != nil {
		RespondError(w, err)
		return
	}

	exist, err := checkUserExist(creds.Username)
	if err != nil {
		http.Error(w, encodeJsonResp(false, err.Error()),
			http.StatusInternalServerError)
		return
	}

	if exist {
		http.Error(w, encodeJsonResp(false,
			"user already exists, please choose another name"),
			http.StatusBadRequest)
		return
	}

	if err = creds.save(); err != nil {
		s := fmt.Sprintf("fail to save creds %v, err info:%v\n", creds, err)
		http.Error(w, encodeJsonResp(false, s),
			http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, encodeJsonResp(true, "signup success"))
}

func signinHandler(w http.ResponseWriter, r *http.Request) {

	creds, err := verifyCredentials(r)
	if err != nil {
		RespondError(w, err)
		return
	}

	hash, err := getPassword(creds.Username)
	switch {
	case err == sql.ErrNoRows:
		msg := fmt.Sprintf("no such user %v", creds.Username)
		http.Error(w, encodeJsonResp(false, msg), http.StatusUnauthorized)
		return
	case err != nil:
		msg := fmt.Sprintf("fail to get password for user %v, %v",
			creds.Username, err)
		http.Error(w, encodeJsonResp(false, msg),
			http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(creds.Password))
	if err != nil {
		msg := fmt.Sprintf("failed to validate password: %v", err)
		http.Error(w, encodeJsonResp(false, msg), http.StatusUnauthorized)
		return
	}

	token := uuid.NewString()
	err = rdb.Set(context.Background(), creds.Username, token, sessionTimeout).Err()
	if err != nil {
		msg := fmt.Sprintf("fail to set token for user %q, %v",
			creds.Username, err)
		http.Error(w, encodeJsonResp(false, msg),
			http.StatusInternalServerError)
		return
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
}

func ValidateSession(w http.ResponseWriter, r *http.Request) (string, error) {
	c, err := r.Cookie("session_token")
	switch {
	case err == http.ErrNoCookie:
		return "", NewRespErr(err, http.StatusUnauthorized)
	case err != nil:
		fmt.Printf("internal error: %v\n", err)
		return "", NewRespErr(err, http.StatusInternalServerError)
	}

	cuser, err := r.Cookie("user")
	switch {
	case err == http.ErrNoCookie:
		return "", NewRespErr(err, http.StatusUnauthorized)
	case err != nil:
		fmt.Printf("internal error: %v\n", err)
		return "", NewRespErr(err, http.StatusInternalServerError)
	}

	token, err := checkKey(cuser.Value)
	switch {
	case err == redis.Nil:
		return "", NewRespErr(err, http.StatusUnauthorized)
	case err != nil:
		fmt.Printf("internal error: %v\n", err)
		return "", NewRespErr(err, http.StatusInternalServerError)
	}

	if token != c.Value {
		err := fmt.Errorf("error: %v token unmatched %v != %v\n", cuser.Value, c.Value, token)
		Info(err.Error())
		clearCookies(w)
		return "", NewRespErr(ErrCacheTokenUnmatch, http.StatusUnauthorized)
	}

	return cuser.Value, nil
}

func RespondError(w http.ResponseWriter, err error) {
	if err, ok := err.(*respErr); ok {
		http.Error(w, encodeJsonResp(false, err.Error()), err.Code())
		return
	}
	http.Error(w, encodeJsonResp(false, err.Error()),
		http.StatusInternalServerError)
}

func RespondAlert(w http.ResponseWriter, err error) {
	if err, ok := err.(*respErr); ok {
		printAlert(w, err.Error(), err.Code())
		return
	}
	printAlert(w, err.Error(), http.StatusInternalServerError)
}

func checkKey(name string) (string, error) {
	return rdb.Get(context.Background(), name).Result()
}

func removeKey(name string) error {
	return rdb.Del(context.Background(), name).Err()
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {

	user, err := ValidateSession(w, r)
	clearCookies(w)
	if err != nil {
		RespondError(w, err)
		return
	}

	err = removeKey(user)
	if err == nil {
		fmt.Fprintf(w, encodeJsonResp(true, "logout success"))
		return
	}

	err = fmt.Errorf("%v: fail to del user %v", err, user)

	fmt.Println(err)
	http.Error(w, encodeJsonResp(false, err.Error()),
		http.StatusInternalServerError)
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

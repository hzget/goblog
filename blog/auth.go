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

const sessionTimeout = 60 * 60 * time.Second

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

func makeAuthHandler(fn func(http.ResponseWriter, *http.Request, *Credentials)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var creds = &Credentials{}

		// decode json-format credentials from client
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(creds); err != nil {
			http.Error(w, fmt.Sprintf("failed to decode request %v", err), http.StatusBadRequest)
			return
		}

		//fmt.Println("get creds:", creds)

		// validate legal username
		isvalid, err := creds.Validate()
		if err != nil {
			fmt.Printf("validate err:%v\n", err)
			http.Error(w, "internal error happened when validate credentials",
				http.StatusInternalServerError)
			return
		}

		if !isvalid {
			http.Error(w, "invalid username", http.StatusBadRequest)
			return
		}

		fn(w, r, creds)
	}
}

func signupHandler(w http.ResponseWriter, r *http.Request, creds *Credentials) {

	exist, err := checkUserExist(creds.Username)
	if err != nil {
		fmt.Printf("fail to check user existance in database:%v\n", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	if exist {
		http.Error(w, "user already exists, please choose another name", http.StatusBadRequest)
		return
	}

	if err = creds.save(); err != nil {
		fmt.Printf("fail to save creds %v, err info:%v\n", creds, err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "success!")
}

func signinHandler(w http.ResponseWriter, r *http.Request, creds *Credentials) {

	hash, err := getPassword(creds.Username)
	switch {
	case err == sql.ErrNoRows:
		http.Error(w, fmt.Sprintf("no such user %v", creds.Username), http.StatusUnauthorized)
		return
	case err != nil:
		fmt.Printf("fail to get password for user %v\n", creds.Username)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(creds.Password))
	if err != nil {
		http.Error(w, "failed to validate password", http.StatusUnauthorized)
		return
	}

	// create session token
	token := uuid.NewString()
	// fmt.Println(token, creds.Username, sessionTimeout)
	err = rdb.Set(context.Background(), token, creds.Username, sessionTimeout).Err()
	if err != nil {
		fmt.Printf("fail to set token for user %v\n", creds.Username)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
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

	c, err := r.Cookie("session_token")
	switch {
	case err == http.ErrNoCookie:
		clearCookies(w)
		fmt.Fprintf(w, "success!")
		return
	case err != nil:
		fmt.Printf("internal error: %v\n", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	err = rdb.Del(context.Background(), c.Value).Err()
	if err != nil {
		fmt.Printf("fail to del token %v\n", c.Value)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	clearCookies(w)

	fmt.Fprintf(w, "success!")
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

package blog

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"text/template"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

var rdb *redis.Client
var db *sql.DB
var templates *template.Template

const shortDuration = 3 * time.Second
const uuidRe = `[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}`
const postRe = `/(view|edit|save|delete)/([0-9]+)`

/* not thread-safe */
var debugPage bool
var debugViewCode bool
var funcMap template.FuncMap
var randomSitePrefix = false
var sitePrefix string
var siteRe string

func initDebugMode() {
	debugPage = true
    debugViewCode = true
}

func initGlobals() {
	randomSitePrefix = true
	initPagePrefix()
}

func initPagePrefix() {

	if !randomSitePrefix {
		sitePrefix = ``
		siteRe = `^` + postRe + `$`
		return
	}

	siteRe = `^/` + uuidRe + postRe + `$`
	sitePrefix = `/` + uuid.NewString()
	fmt.Println(sitePrefix)
	match, err := regexp.MatchString(`^/`+uuidRe+`$`, sitePrefix)
	if err != nil {
		log.Fatal(err)
	}
	if !match {
		log.Fatal("uuid regexp failed")
	}
}

func initDBHandler() {
	cfg := mysql.Config{
		// linux cmd to set var: export DBUSER=username
		User:   os.Getenv("DBUSER"),
		Passwd: os.Getenv("DBPASS"),
		Net:    "tcp",
		Addr:   "192.168.1.12:3306",
		//Addr: "10.43.6.18:3306",
		//Addr:                 "10.43.23.193:3306",
		DBName:               "recordings",
		AllowNativePasswords: true,
		ParseTime:            true,
		Loc:                  time.Local,
	}
	// Get a database handle.
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	// Create a child context
	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected!")
}

func initRedisClient() {
	ctx := context.Background()

	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := rdb.Ping(ctx).Result()
	fmt.Println(pong, err)
	if err != nil {
		panic(err)
	}
}

func initFuncMap() {
	funcMap = template.FuncMap{"add": add, "multiple": multiple}
}

func initTemplate() {

	initFuncMap()

	if debugPage {
		return
	}

	t, err := template.New("").Funcs(funcMap).ParseFiles(
		"templ/view.html", "templ/edit.html", "templ/frontpage.html")
	templates = template.Must(t, err)
}

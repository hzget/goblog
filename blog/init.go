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
	"github.com/spf13/viper"
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

func initGlobals() {
	getConfig()
	initDebugMode()
	initPagePrefix()
	initRedisClient()
	initDBHandler()
	initTemplate()
}

func getConfig() {
    // attention: viper is not thread-safe
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath("blog/config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %w \n", err))
	}
}

func initDebugMode() {
	debugPage = viper.GetBool("debug.page")
	debugViewCode = viper.GetBool("debug.viewcode")
}

func initPagePrefix() {

	randomSitePrefix = viper.GetBool("page.randomprefix")

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

	addr := viper.GetString("datastore.mysql.addr")

	var err error
	if err = initMysqlDBHandler(addr); err == nil {
		fmt.Println("Connected!")
		return
	}

	fmt.Printf("db err: %v, now try another db\n", err)

	addr = viper.GetString("datastore.mysql.addr2")
	if err = initMysqlDBHandler(addr); err == nil {
		fmt.Println("Connected!")
		return
	}

	log.Fatal(err)
}

func initMysqlDBHandler(addr string) error {

	cfg := mysql.Config{
		// linux cmd to set var: export DBUSER=username
		User:                 os.Getenv("DBUSER"),
		Passwd:               os.Getenv("DBPASS"),
		Net:                  "tcp",
		Addr:                 addr,
		DBName:               "recordings",
		AllowNativePasswords: true,
		ParseTime:            true,
		Loc:                  time.Local,
	}

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	return db.PingContext(ctx)
}

func initRedisClient() {

	addr := viper.GetString("datastore.redis.addr")
	ctx := context.Background()

	rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
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

package blog

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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
var dataAnalysisAddress string

func initGlobals() {
	getConfig()
	initDebugMode()
	initPagePrefix()
	initRedisClient()
	initDBHandler()
	initDBTables()
	initDataAnalysis()
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

	cfg := mysql.Config{
		Addr:                 viper.GetString("service.mysql.addr"),
		User:                 viper.GetString("service.mysql.user"),
		Passwd:               viper.GetString("service.mysql.passwd"),
		DBName:               viper.GetString("service.mysql.dbname"),
		Net:                  "tcp",
		AllowNativePasswords: true,
		ParseTime:            true,
		Loc:                  time.Local,
		MultiStatements:      true,
	}

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected!")
}

func initRedisClient() {

	ctx := context.Background()

	rdb = redis.NewClient(&redis.Options{
		Addr:     viper.GetString("service.redis.addr"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := rdb.Ping(ctx).Result()
	fmt.Println(pong, err)
	if err != nil {
		panic(err)
	}
}

func initDataAnalysis() {
	dataAnalysisAddress = viper.GetString("service.data-analysis.addr")
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
		"templ/view.html", "templ/edit.html", "templ/frontpage.html",
		"templ/analysis.html", "templ/useradmin.html", "templ/alert.html",
		"templ/inspect.html",
	)
	templates = template.Must(t, err)
}

func initDBTables() {

	if !viper.GetBool("debug.initdbtable") {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	q := `ALTER DATABASE blog CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
        CREATE TABLE IF NOT EXISTS post(
          id        INT AUTO_INCREMENT NOT NULL,
          title     TINYTEXT NOT NULL,
          author    VARCHAR(10) NOT NULL,
          ctime     DATETIME NOT NULL,
          mtime     DATETIME NOT NULL,
          body      LONGTEXT,
          PRIMARY KEY (id)
        );
        CREATE TABLE IF NOT EXISTS poststatistics(
          postid    INT NOT NULL UNIQUE,
          star1    INT NOT NULL DEFAULT 0,
          star2    INT NOT NULL DEFAULT 0,
          star3    INT NOT NULL DEFAULT 0,
          star4    INT NOT NULL DEFAULT 0,
          star5    INT NOT NULL DEFAULT 0,
          PRIMARY KEY (postid)
        );
        CREATE TABLE IF NOT EXISTS users (
          username  VARCHAR(10) NOT NULL,
          password  VARCHAR(1024) NOT NULL,
          rank  ENUM('bronze','silver','gold') NOT NULL,
          PRIMARY KEY (username)
        );`

	if _, err := db.ExecContext(ctx, q); err != nil {
		log.Fatal(err)
	}
}

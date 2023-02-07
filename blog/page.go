package blog

/*
 * data flow:
 *
 *     client <-- tcp/ip --> server <-- router --> handlers <-- --> data structure object <-- interface --> data store
 *
 * generate html:
 *
 *     template files --> template object       --|
 *                                                | --> html --> server
 *     data files     --> data structure object --|
 *
 */

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type appError struct {
	Error error
	Code  int
}

type viewReq struct {
	Id int64 `json:"id"`
}

type saveReq struct {
	Id    int64  `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

type viewResp struct {
	jsonResp
	PostInfo
}

type saveResp struct {
	jsonResp
	Id int64 `json:"id"`
}

type jsonResp struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type PageInfo struct {
	Username string
	Id       int64 `json:"id"`
}

const (
	PermNone   = 0
	PermView   = 1 << iota
	PermEdit   = 1 << iota
	PermDelete = 1 << iota
)

func frontpageHandler(w http.ResponseWriter, r *http.Request) {

	data := struct {
		ViewCode bool
	}{debugViewCode}

	renderTemplate(w, "frontpage.html", data)
}

func postlistHandler(w http.ResponseWriter, r *http.Request) {

	data, err := getPostsInfo()
	if err != nil {
		printAlert(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i := 0; i < len(data); i++ {
		data[i].Body = getHTMLEscapeString(data[i].Body)
	}

	renderTemplate(w, "postlist.html", data)
}

func handleErr(w http.ResponseWriter, r *http.Request, err error) {
	switch err.(type) {
	case *limitErr:
		printAlert(w, err.Error(), http.StatusInternalServerError)
	default:
		printAlert(w, "the user is not allowed to view post", http.StatusBadRequest)

	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, info *PageInfo) {

	perm, err := info.getPermisson()
	if err != nil {
		handleErr(w, r, err)
		return
	}

	canView := perm&PermView > 0
	if !canView {
		printAlert(w, "the user is not allowed to view post", http.StatusBadRequest)
		return
	}

	data, err := getViewData(info)
	switch {
	case err == sql.ErrNoRows:
		printAlert(w, err.Error(), http.StatusBadRequest)
		return
	case err != nil:
		fmt.Println(err)
		printAlert(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderTemplate(w, "view.html", data)

}

func editHandler(w http.ResponseWriter, r *http.Request, info *PageInfo) {

	perm, err := info.getPermisson()
	if err != nil {
		handleErr(w, r, err)
		return
	}

	canEdit := perm&PermEdit > 0
	if !canEdit {
		printAlert(w, "the user is not allowed to edit post", http.StatusBadRequest)
		return
	}

	data, err := loadPost(info.Id)
	if err != nil {
		data = &Post{}
	}

	renderTemplate(w, "edit.html", data)
}

func saveHandler(w http.ResponseWriter, r *http.Request, info *PageInfo) {

	perm, err := info.getPermisson()
	if err != nil {
		handleErr(w, r, err)
		return
	}

	canEdit := perm&PermEdit > 0
	if !canEdit {
		printAlert(w, "the user is not allowed to edit and save post", http.StatusBadRequest)
		return
	}

	// remaining: shall validate if title is "" or " " or others
	title := r.FormValue("title")
	if strings.EqualFold(title, "New") || strings.EqualFold(title, "frontpage") {
		fmt.Fprintf(w, "Please give it a title name other than %v", title)
		return
	}

	p := &Post{Id: info.Id, Title: title, Body: r.FormValue("body"),
		Author: info.Username}
	if err := p.save(); err != nil {
		printAlert(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("../view/%d", p.Id), http.StatusSeeOther)
}

func viewjsHandler(w http.ResponseWriter, r *http.Request, info *PageInfo) *appError {

	var req = &viewReq{}

	// remaining issue: how to handle missing field in json?
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(req); err != nil {
		return &appError{err, http.StatusBadRequest}
	}

	info.Id = req.Id
	perm, err := info.getPermisson()
	if err != nil {
		return &appError{err, http.StatusBadRequest}
	}

	canView := perm&PermView > 0
	if !canView {
		return &appError{errors.New("the user is not allowed to view post"),
			http.StatusBadRequest}
	}

	data, err := getPostInfo(info.Id)
	switch {
	case err == sql.ErrNoRows:
		return &appError{err, http.StatusBadRequest}
	case err != nil:
		return &appError{err, http.StatusInternalServerError}
	}

	fmt.Fprintf(w, encodeJsonViewResp(data))

	return nil
}

func savejsHandler(w http.ResponseWriter, r *http.Request, info *PageInfo) *appError {

	var req = &saveReq{}

	// remaining issue: how to handle missing field in json?
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(req); err != nil {
		return &appError{err, http.StatusBadRequest}
	}

	info.Id = req.Id
	perm, err := info.getPermisson()
	if err != nil {
		return &appError{err, http.StatusBadRequest}
	}

	canEdit := perm&PermEdit > 0
	if !canEdit {
		return &appError{errors.New("the user is not allowed to edit and save post"),
			http.StatusBadRequest}
	}

	var post = &Post{Id: req.Id, Title: req.Title, Body: req.Body, Author: info.Username}
	if err := post.save(); err != nil {
		return &appError{err, http.StatusInternalServerError}
	}

	fmt.Fprintf(w, encodeJsonSaveResp(true, "save success", post.Id))

	return nil
}

func encodeJson(data interface{}) string {
	b, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	return string(b)
}

func decodeJson(body []byte, data interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(data); err != nil {
		return err
	}

	return nil
}

func encodeJsonResp(success bool, msg string) string {
	return encodeJson(&jsonResp{success, msg})
}

func encodeJsonSaveResp(success bool, msg string, id int64) string {
	return encodeJson(&saveResp{jsonResp{success, msg}, id})
}

func encodeJsonViewResp(p PostInfo) string {
	return encodeJson(&viewResp{jsonResp{true, ""}, p})
}

func deleteHandler(w http.ResponseWriter, r *http.Request, info *PageInfo) {

	perm, err := info.getPermisson()
	if err != nil {
		handleErr(w, r, err)
		return
	}

	canDel := perm&PermDelete > 0
	if !canDel {
		printAlert(w, "the user is not allowed to delete post", http.StatusBadRequest)
		return
	}

	if err := DeletePost(info.Id); err != nil {
		printAlert(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "../", http.StatusFound)
}

func getId(s string) (string, error) {
	m := regexp.MustCompile(siteRe).FindStringSubmatch(s)
	if m == nil {
		es := fmt.Sprintf("pathname is invalid: %s", s)
		return "", errors.New(es)
	}
	return m[2], nil
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, *PageInfo)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title, err := getId(r.URL.Path)
		if err != nil {
			printAlert(w, "the pathname is invalid: "+r.URL.Path, http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(title)

		username, err2 := ValidateSession(w, r)
		if err2 != nil {
			printAlert(w, err2.Error(), err2.Code())
			return
		}

		info := &PageInfo{username, int64(id)}

		fn(w, r, info)
	}
}

func makePageHandler(fn func(http.ResponseWriter, *http.Request, *PageInfo) *appError) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var e *appError
		var info *PageInfo

		username, err := ValidateSession(w, r)
		if err != nil {
			e = &appError{err, err.Code()}
			goto Err
		}

		info = &PageInfo{Username: username}

		if e = fn(w, r, info); e == nil {
			return
		}

	Err:
		if e.Code == http.StatusInternalServerError {
			fmt.Println(e.Error)
		}
		http.Error(w, encodeJsonResp(false, e.Error.Error()), e.Code)
	}
}

/*
 * view: id exists
 * edit: (id exists and user == author) or ( id == 0 and user is not superadmin)
 * del : id exists and is superadmin
 */
func (info *PageInfo) getPermisson() (int, error) {

	if info.Id < 0 {
		return PermNone, nil
	}

	// create a post
	if info.Id == 0 {
		if info.Username == "superadmin" {
			return PermNone, nil
		}
		return PermEdit, nil
	}

	post, err := loadPost(info.Id)
	if err != nil {
		return PermNone, err
	}

	if info.Username == "superadmin" {
		return PermDelete | PermView, nil
	}

	if info.Username == post.Author {
		return PermEdit | PermView, nil
	}

	return PermView, nil
}

func getViewData(info *PageInfo) (interface{}, error) {

	pi, err := getPostInfo(info.Id)
	if err != nil {
		return nil, err
	}

	perm, err := info.getPermisson()
	if err != nil {
		return nil, err
	}

	pi.Body = getHTMLEscapeString(pi.Body)

	data := struct {
		PostInfo
		CanEdit   bool
		CanDelete bool
	}{pi, perm&PermEdit > 0, perm&PermDelete > 0}

	return data, nil
}

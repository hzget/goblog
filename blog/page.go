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
	PermView   = 1
	PermEdit   = 2
	PermDelete = 3
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

func viewHandler(w http.ResponseWriter, r *http.Request, info *PageInfo) {

	canView := info.getPermisson() >= PermView
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

	canEdit := info.getPermisson() == PermEdit
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

	canEdit := info.getPermisson() == PermEdit
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

	var post = &PageInfo{}

	// remaining issue: how to handle missing field in json?
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(post); err != nil {
		return &appError{err, http.StatusBadRequest}
	}

	info.Id = post.Id
	canView := info.getPermisson() >= PermView
	if !canView {
		return &appError{errors.New("the user is not allowed to view post"),
			http.StatusBadRequest}
	}

	data, err := getPostInfo(post.Id)
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

	var post = &Post{}

	// remaining issue: how to handle missing field in json?
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(post); err != nil {
		return &appError{err, http.StatusBadRequest}
	}

	info.Id = post.Id
	canEdit := info.getPermisson() == PermEdit
	if !canEdit {
		return &appError{errors.New("the user is not allowed to edit and save post"),
			http.StatusBadRequest}
	}

	post.Author = info.Username
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

	canDel := info.getPermisson() == PermDelete
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

		username, status := ValidateSession(w, r)
		switch status {
		case SessionUnauthorized:
			printAlert(w, "please log in first", http.StatusUnauthorized)
			return
		case SessionInternalError:
			printAlert(w, "internal error", http.StatusInternalServerError)
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

		username, status := ValidateSession(w, r)
		switch status {
		case SessionUnauthorized:
			e = &appError{errors.New("please log in first"), http.StatusUnauthorized}
			goto Err
		case SessionInternalError:
			e = &appError{errors.New("internal error"), http.StatusInternalServerError}
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

func (info *PageInfo) getPermisson() int {
	if info.Username == "superadmin" {
		return PermDelete
	}

	// create a post
	if info.Id == 0 {
		return PermEdit
	}

	post, err := loadPost(info.Id)
	if err != nil {
		return PermNone
	}

	if info.Username == post.Author {
		return PermEdit
	}

	return PermView
}

func getViewData(info *PageInfo) (interface{}, error) {

	pi, err := getPostInfo(info.Id)
	if err != nil {
		return nil, err
	}

	perm := info.getPermisson()

	pi.Body = getHTMLEscapeString(pi.Body)

	data := struct {
		PostInfo
		CanEdit   bool
		CanDelete bool
	}{pi, perm == PermEdit, perm == PermDelete}

	return data, nil
}

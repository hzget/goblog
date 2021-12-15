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
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
    "log"
)

type PageInfo struct {
	Username string
	Id       int64
}

const (
	PermNone   = 0
	PermView   = 1
	PermEdit   = 2
	PermDelete = 3
)

func frontpageHandler(w http.ResponseWriter, r *http.Request) {

	data, err := getPostsInfo()
	if err != nil {
		fmt.Fprintf(w, "load post info failed: %v", err)
		return
	}

	renderTemplate(w, "frontpage.html", data)
}

func viewHandler(w http.ResponseWriter, r *http.Request, info *PageInfo) {

	canView := info.getPermisson() >= PermView
	if !canView {
		http.Error(w, "the user is not allowed to view post", http.StatusBadRequest)
		return
	}

	data, err := getViewData(info)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "load page failed: %v", err)
		return
	}

	renderTemplate(w, "view.html", data)

}

func editHandler(w http.ResponseWriter, r *http.Request, info *PageInfo) {

	canEdit := info.getPermisson() == PermEdit
	if !canEdit {
		http.Error(w, "the user is not allowed to edit post", http.StatusBadRequest)
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
		http.Error(w, "the user is not allowed to edit and save post", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	if strings.EqualFold(title, "New") || strings.EqualFold(title, "frontpage") {
		fmt.Fprintf(w, "Please give it a title name other than %v", title)
		return
	}

	p := &Post{Id: info.Id, Title: title, Body: r.FormValue("body"),
		Author: info.Username}
	if err := p.save(); err != nil {
		fmt.Fprintf(w, "save file failed: %v", err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("../view/%d", p.Id), http.StatusSeeOther)
}

func deleteHandler(w http.ResponseWriter, r *http.Request, info *PageInfo) {

	canDel := info.getPermisson() == PermDelete
	if !canDel {
		http.Error(w, "the user is not allowed to delete post", http.StatusBadRequest)
		return
	}

	if err := DeletePost(info.Id); err != nil {
		fmt.Fprintf(w, "delete post failed: %v", err)
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
			http.Error(w, "the pathname is invalid: "+r.URL.Path, http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(title)

		username, status := ValidateSession(w, r)
		switch status {
		case SessionUnauthorized:
			http.Error(w, "please log in first", http.StatusUnauthorized)
			return
		case SessionInternalError:
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		info := &PageInfo{username, int64(id)}

		fn(w, r, info)
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

	p, err := loadPost(info.Id)
	if err != nil {
		return nil, err
	}

	s, err := loadPostStatistics(p.Id)
	vote := 0
	percent := []float64{0, 0, 0, 0, 0}
	if err == nil {
		vote, percent = s.getVote()
	}

	canEdit := info.getPermisson() == PermEdit
	canDel := info.getPermisson() == PermDelete
	body := strings.Split(string(p.Body), "\n")
	data := struct {
		Id      int64
		Title   string
		Body    []string
		Author  string
		Edit    bool
		Del     bool
		Vote    int
		Percent []float64
	}{p.Id, p.Title, body, p.Author, canEdit, canDel, vote, percent}

	return data, nil
}

func Run(addr string) {

	initDebugMode()
	initGlobals()
	initTemplate()
	initRedisClient()
	initDBHandler()

    if (debugViewCode) {
	    http.Handle(sitePrefix+"/code/", http.StripPrefix(
		    sitePrefix+"/code/", http.FileServer(http.Dir("./"))))
    }

	http.HandleFunc(sitePrefix+"/", frontpageHandler)
	http.HandleFunc(sitePrefix+"/view/", makeHandler(viewHandler))
	http.HandleFunc(sitePrefix+"/edit/", makeHandler(editHandler))
	http.HandleFunc(sitePrefix+"/save/", makeHandler(saveHandler))
	http.HandleFunc(sitePrefix+"/delete/", makeHandler(deleteHandler))
	http.Handle(sitePrefix+"/templ/rs/", http.StripPrefix(
	    sitePrefix+"/templ/rs/", http.FileServer(http.Dir("./templ/resource/"))))

	http.HandleFunc(sitePrefix+"/signup", makeAuthHandler(signupHandler))
	http.HandleFunc(sitePrefix+"/signin", makeAuthHandler(signinHandler))
	http.HandleFunc(sitePrefix+"/logout", logoutHandler)

	http.HandleFunc(sitePrefix+"/vote", voteHandler)

	http.HandleFunc(sitePrefix+"/analysis", analysisHandler)

	http.HandleFunc(sitePrefix+"/superadmin", makeAdminHandler(superadminHandler))
	http.HandleFunc(sitePrefix+"/saveranks", makeAdminHandler(saveranksHandler))

	log.Fatal(http.ListenAndServe(addr, nil))

}

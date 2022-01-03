package blog

import (
	"fmt"
	"net/http"
	"text/template"
)

func add(a, b int) int              { return a + b }
func multiple(a, b float64) float64 { return a * b }

func printAlert(w http.ResponseWriter, msg string, code int) {

	data := struct {
		Prefix string
		Info   string
	}{sitePrefix, msg}

	if code == http.StatusUnauthorized {
		clearCookies(w)
	}

	w.WriteHeader(code)
	renderTemplate(w, "alert.html", data)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	if debugPage {
		t, err := template.New(tmpl).Funcs(funcMap).ParseFiles("templ/" + tmpl)
		if err != nil {
			printAlert(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err = t.Execute(w, data); err != nil {
			fmt.Println(err)
		}
		return
	}

	if err := templates.ExecuteTemplate(w, tmpl, data); err != nil {
		fmt.Println(err)
		printAlert(w, err.Error(), http.StatusInternalServerError)
	}
}

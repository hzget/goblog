package blog

import (
	"fmt"
	"net/http"
	"text/template"
)

func add(a, b int) int              { return a + b }
func multiple(a, b float64) float64 { return a * b }

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	if debugPage {
		t, err := template.New(tmpl).Funcs(funcMap).ParseFiles("templ/" + tmpl)
		if err != nil {
			fmt.Fprintf(w, "load file failed: %v", err)
			return
		}

		if err = t.Execute(w, data); err != nil {
			fmt.Println(err)
		}
		return
	}

	if err := templates.ExecuteTemplate(w, tmpl, data); err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

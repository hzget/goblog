package blog

import (
	"log"
	"net/http"
)

func Run(addr string) {

	initGlobals()

	if debugViewCode {
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

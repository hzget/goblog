package blog

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
)

func Run(addr string) {

	initGlobals()

	var srv = &http.Server{Addr: addr}

	go startHttpServer(srv)

	gracefullyShutdown(srv)
}

func gracefullyShutdown(srv *http.Server) {

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
	s := <-sigint

	log.Printf("receive signal %v\n", s)

	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP server Shutdown: %v\n", err)
	}

	log.Printf("HTTP server is shutdown\n")
}

func startHttpServer(srv *http.Server) {

	if debugViewCode {
		http.Handle(sitePrefix+"/code/", http.StripPrefix(
			sitePrefix+"/code/", http.FileServer(http.Dir("./"))))
	}

	http.HandleFunc(sitePrefix+"/", frontpageHandler)
	http.HandleFunc(sitePrefix+"/postlist", postlistHandler)
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
	http.HandleFunc(sitePrefix+"/analyze", analyzeHandler)

	http.HandleFunc(sitePrefix+"/superadmin", makeAdminHandler(superadminHandler))
	http.HandleFunc(sitePrefix+"/saveranks", makeAdminHandler(saveranksHandler))

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	log.Printf("HTTP server stop receiving new request\n")
}

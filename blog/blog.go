package blog

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Run(addr string) {

	initGlobals()

	var srv = &http.Server{Addr: addr}

	go startHttpServer(srv)

	gracefulShutdown(srv)
}

func gracefulShutdown(srv *http.Server) {

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
	s := <-sigint

	log.Printf("received signal %v, the type is %T\n", s, s)

	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP server Shutdown: %v\n", err)
	}

	log.Printf("the HTTP server Shutdown\n")
}

func startHttpServer(srv *http.Server) {

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

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	log.Printf("HTTP server stop recieving new request\n")
}

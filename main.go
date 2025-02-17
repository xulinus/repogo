package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/xulinus/repogo/pkg/global"
	"github.com/xulinus/repogo/pkg/handlers"
)

var port = "8085"

func main() {
	global.BRANCH = "main"
	global.REPO = "xulinus/policy-docs/"
	global.FOLDER = ""
	global.GH_BEARER_TOKEN = os.Getenv("GITHUB_BEARER_TOKEN")

	router := mux.NewRouter().StrictSlash(true)

	router.PathPrefix("/fonts/").
		Handler(http.StripPrefix("/fonts", handlers.NonListFileServer(http.FileServer(http.Dir("./tmpl/fonts/")))))

	router.PathPrefix("/css/").
		Handler(http.StripPrefix("/css", handlers.NonListFileServer(http.FileServer(http.Dir("./tmpl/css/")))))

	router.HandleFunc("/", handlers.Main)
	router.HandleFunc("/doc/{sha:[a-f0-9]{40}}/{doc:.*}", handlers.Doc)
	router.HandleFunc("/doc/{doc:.*}", handlers.Doc)

	log.Printf("Webserver listening on port %s", port)
	http.ListenAndServe(":"+port, router)
}

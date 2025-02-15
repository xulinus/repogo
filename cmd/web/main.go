package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/xulinus/repogo/pkg/handlers"
)

var port = "8085"

func main() {
	router := mux.NewRouter().StrictSlash(true)

	router.PathPrefix("/css/").
		Handler(http.StripPrefix("/css", handlers.NonListFileServer(http.FileServer(http.Dir("./tmpl/css/")))))

	router.HandleFunc("/doc", handlers.Doc)

	log.Printf("Webserver listening on port %s", port)
	http.ListenAndServe(":"+port, router)
}

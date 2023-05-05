package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const (
	PORT = ":8080"
)

func main() {
	
	fmt.Println("Running the server on port:", PORT)
	http.HandleFunc("/", baseHandler)
	http.Handle("/api", apiRouter{}.Use())
	log.Fatal(http.ListenAndServe(PORT, nil))
}

func baseHandler(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) > 1 {
		http.NotFound(w, r)
		return
	}
	documentationHandler(w, r)
}

func documentationHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world!")
}

type apiRouter map[string]http.HandlerFunc
func (apir apiRouter) Use() apiRouter {
	apir["GET"] = getHandler
	apir["POST"] = postHandler
	return apir
}

func (apir apiRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, ok := apir[r.Method]
	if !ok {
		apir.ReturnMethodNotAllowed(w, r)
		return
	}
	handler(w, r)
}

func (apir apiRouter) ReturnMethodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	allowedArr := make([]string, len(apir))
	i := 0
	for method := range apir {
		allowedArr[i] = method
		i++
	}
	w.Header().Set("Allow", strings.Join(allowedArr, ", "))
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello GEt"))
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("HEELO post"))
}

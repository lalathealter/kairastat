package controllers

import "net/http"

func GetHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello GEt"))
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("HEELO post"))
}

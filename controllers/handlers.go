package controllers

import (
	"net/http"

	"github.com/lalathealter/kairastat/postgre"
)

func GetHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello GEt"))
}


func PostHandler(w http.ResponseWriter, r *http.Request) {
	queries := []string{
		"event_name",
	}
	_ = parseUrlQuery(r, queries)

	_ = postgre.GetDB()
	// dbr, err := db.Query(postgre.SelectEventQuery, vals[0])
	// if err != nil {
	// 	log.Panicln(err)
	// }

	w.Write([]byte("HEELO post"))
}

func parseUrlQuery(r *http.Request, loadArr []string) []string {
	urlvals := r.URL.Query()
	for i, key := range loadArr {
		loadArr[i] = urlvals.Get(key)
	}
	return loadArr
}

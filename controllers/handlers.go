package controllers

import (
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/lalathealter/kairastat/postgre"
)

func GetHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello GEt"))
}
// 1 метод:
// Принимать название и статус пользоватля (авторизован или нет)
// Добавитьк этому вспомогательную информацию 
// Сохранить событие 


func PostHandler(w http.ResponseWriter, r *http.Request) {

	urlvals := r.URL.Query()
	eventName := urlvals.Get("event")
	if eventName == "" {
		log.Panicln("event name wasn't provided;")
	} else if len(eventName) > 128  {
		log.Panicln("given event name is too long;")
	}
	isAuthorized := urlvals.Has("authorized")

	db := postgre.GetWrapper()


	clientIP := getClientIP(r)
	userID := db.GetUserFor(clientIP)

	db.SetUserAuthorized(userID, isAuthorized)

	db.SaveEvent(eventName, userID)
	
	w.WriteHeader(http.StatusNoContent)
}

func getClientIP(r *http.Request) string {
	usedIPs := r.Header.Get("X-Forwarded-For")
	originIP := strings.Split(usedIPs, ", ")[0]
	if originIP == "" {
		originIP, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	return originIP
}


package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/lalathealter/kairastat/postgre"
)

func GetHandler(w http.ResponseWriter, r *http.Request) {
	urlvals := r.URL.Query()

	db := postgre.GetWrapper()
	var dbresArr any

	switch {
		case urlvals.Has("event"):
			passedEventName := urlvals.Get("event")
			dbresArr = db.GetEventsByName(passedEventName)
		case urlvals.Has("user-ip"):
			passedIP := urlvals.Get("user-ip")
			ip := parseIPAddress(passedIP)
			dbresArr = db.GetEventsByUserIP(ip)
		case urlvals.Has("is-authorized"):
			passedArg := urlvals.Get("is-authorized")
			isAuth := parseBoolString(passedArg)
			dbresArr = db.GetEventsByAuthorized(isAuth)
		default:
			dbresArr = db.GetEventsAll()
	}


	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(dbresArr); err != nil {
		log.Panicln(err)
	}
}

func parseIPAddress(input string) string {
	ipAddr := net.ParseIP(input)
	if ipAddr == nil {
		log.Panicln("given user-ip isn't a valid IP address;")
	}
	fmt.Println(ipAddr.String())
	return ipAddr.String() 
}

var parseBoolString = func() func(string)bool {
	trueDefinitions := map[string]bool{
		"true": true,
		"t": true,
		"1": true,
		"yes": true,
		"y": true,
	}
	return func(input string) bool {
		linput := strings.ToLower(input)
		_, defined := trueDefinitions[linput]
		return defined
	}
}()

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


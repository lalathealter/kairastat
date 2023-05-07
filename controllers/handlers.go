package controllers

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lalathealter/kairastat/postgre"
)

func GetHandler(w http.ResponseWriter, r *http.Request) {
	urlvals := r.URL.Query()
	db := postgre.GetWrapper()

	events := findEventsFromURLQuery(db, urlvals)
	eventsFiltered := filterEventsFromURLQuery(db, urlvals, events)
	resultRows := db.MakeNestedQuery(eventsFiltered)
	resultData := db.ParseEventRows(resultRows)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resultData); err != nil {
		log.Panicln(err)
	}
}


func findEventsFromURLQuery(db postgre.Wrapper, urlvals url.Values ) postgre.SubQueryCB {
	var baseQuery postgre.SubQueryCB 
	switch {
	case urlvals.Has("event"):
		passedEventName := urlvals.Get("event")
		baseQuery = db.GetEventsByName(passedEventName)
	case urlvals.Has("user-ip"):
		passedIP := urlvals.Get("user-ip")
		ip := parseIPAddress(passedIP)
		baseQuery = db.GetEventsByUserIP(ip)
	case urlvals.Has("is-authorized"):
		passedArg := urlvals.Get("is-authorized")
		isAuth := parseBoolString(passedArg)
		baseQuery = db.GetEventsByAuthorized(isAuth)
	default:
		baseQuery = db.GetEventsAll()
	}
	return baseQuery
}

func filterEventsFromURLQuery(db postgre.Wrapper, urlvals url.Values, prevQuery postgre.SubQueryCB) postgre.SubQueryCB {
	var currQuery postgre.SubQueryCB
	switch {
	case urlvals.Has("starts-with"):
		nameStartParam := urlvals.Get("starts-with")
		currQuery = db.FilterEventsByName(nameStartParam)
	case urlvals.Has("later-than"):
		timeParam := urlvals.Get("later-than")
		laterThan := parseTimestamp(timeParam)
		currQuery = db.FilterEventsByTime(laterThan)
	default:
		return prevQuery
	}
	
	return db.ChainQuery(prevQuery, currQuery)
}

func parseIPAddress(input string) string {
	ipAddr := net.ParseIP(input)
	if ipAddr == nil {
		log.Panicln("given user-ip isn't a valid IP address;")
	}
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

func parseTimestamp(input string) string {
	input = guardTZPlusSign(input)
	timestampTZ, err := time.Parse(time.RFC3339, input)
	if err != nil {
		log.Panicln(err)
	}
	timeTZString := stripDuplicateTZMark(timestampTZ.String())
	return timeTZString
}

func guardTZPlusSign(s string) string {
	escapedPlusInd := strings.LastIndex(s, " ")
	if escapedPlusInd > -1 {
		s =   
			s[:escapedPlusInd] + "+" + 
			s[escapedPlusInd+1:]
	} 
	return s
}

func stripDuplicateTZMark (s string) string {
	lastSpaceInd := strings.LastIndex(s, " ")
	return s[:lastSpaceInd]
}

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


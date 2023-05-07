package postgre

import (
	"database/sql"
	"log"
)

type Wrapper struct {
	db *sql.DB
}


func (wr Wrapper) GetUserFor(clientIP string) int {
	var userID int
	dbrow := wr.db.QueryRow(SelectUserIDQuery, clientIP)
	dbrow.Scan(&userID)
	if userID == 0 {
		dbrow := wr.db.QueryRow(CreateUserQuery, clientIP)
		if err := dbrow.Err(); err != nil {
			log.Panicln(err)
		}
		dbrow.Scan(&userID)
	}

	return userID
}

func (wr Wrapper) SetUserAuthorized(userID int, isAuthorized bool) {
	_, err := wr.db.Exec(UpdateUserAuthorizedQuery, isAuthorized, userID)
	if err != nil {
		log.Panicln(err)
	} 
}

func (wr Wrapper) SaveEvent(eventName string, userID int) {
	_, err := wr.db.Exec(CreateEventQuery, eventName, userID)
	if err != nil {
		log.Panicln(err)
	}
}


type EventObject struct {
	EventName string `field:"event_name"`
	EventCount string `field:"endorsements_count"`
	OldestRecord string `field:"created_at"`
}

func (wr Wrapper) ParseEventRows(eventRows *sql.Rows) []*EventObject {
	return parseSQLRows(eventRows, EventObject{})
}

func (wr Wrapper) GetEventsAll() SubQueryCB  {
	return GenSubQueryWith(SelectAllEvents)()
}

func (wr Wrapper) GetEventsByAuthorized(isAuthorized bool)  SubQueryCB {
	return GenSubQueryWith(SelectEventsByUserAuthorization)(isAuthorized)
}

func (wr Wrapper) GetEventsByName(eventName string) SubQueryCB {
	return GenSubQueryWith(SelectEventsByName)(eventName)
}

func (wr Wrapper) GetEventsByUserIP(ip string) SubQueryCB {
	return GenSubQueryWith(SelectEventsByIP)(ip)
}


var (
	SelectEventsBy = applyTemplate(TemplateSelectEventsBy)
	SelectEventsByName = SelectEventsBy(AggregateEventByName)
	SelectEventsByUserAuthorization = SelectEventsBy(AggregateEventByAthorization)
	SelectEventsByIP = SelectEventsBy(AggregateEventByAuthorIP)
	SelectAllEvents = SelectEventsBy("")
)
func (wr Wrapper) FilterEventsByName(eventName string) SubQueryCB {
	return GenSubQueryWith(FilterEventByName)(eventName)
}

func (wr Wrapper) FilterEventsByTime(time string) SubQueryCB {
	return GenSubQueryWith(FilterEventByTime)(time)
}


const (
	FilterEventByName = `
		SELECT * 
		FROM (%s) R
		WHERE R.event_name LIKE CONCAT($1::text, '%%')
	;`
	FilterEventByTime = `
		SELECT * 
		FROM (%s) R
		WHERE R.created_at >= $1
	;`


	TemplateSelectEventsBy = `
		SELECT 
			event_name, 
			sum(endorsements_count) AS endorsements_count,
			(SELECT min(created_at) AS created_at FROM kairastat.events WHERE event_name=evs.event_name)
		FROM 
			kairastat.events evs
		%s
		GROUP BY 
			event_name
	;`
	
	AggregateEventByAuthorIP = `
		WHERE (
			SELECT ip_address
			FROM kairastat.users
			WHERE user_id = evs.author_id
		) = $1
	`
	AggregateEventByName = `
		WHERE event_name = $1
	`
	AggregateEventByAthorization = `
		WHERE (
			SELECT authorized 
			FROM kairastat.users 
			WHERE user_id = evs.author_id
		) = $1
	`

	SelectUserIDQuery = `
		SELECT user_id 
		FROM kairastat.users
		WHERE ip_address=$1
	;`

	SelectUserIPQuery = `
		SELECT ip_address
		FROM kairastat.users
		WHERE user_id=$1
	;`

	CreateUserQuery = `
		INSERT INTO kairastat.users
		(ip_address)
		VALUES
		($1)
		RETURNING user_id
	;`

	UpdateUserAuthorizedQuery = `
		UPDATE kairastat.users
		SET authorized = $1
		WHERE user_id=$2
	;`

	SelectEventIDQuery = `
		SELECT event_id 
		FROM kairastat.events
		WHERE event_name=$1
	;`

	SelectEventQuery = `
		SELECT * 
		FROM kairastat.events
		WHERE event_name=$1
	;`

	CreateEventQuery = `
		INSERT INTO kairastat.events AS evs
		(event_name, author_id)
		VALUES 
		($1, $2)
		ON CONFLICT (event_name, author_id)
		DO UPDATE
			SET endorsements_count = evs.endorsements_count + 1
	;`
)


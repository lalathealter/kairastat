package postgre

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
)

type wrapper struct {
	db *sql.DB
}

func (wr wrapper) GetUserFor(clientIP string) int {
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

func (wr wrapper) SetUserAuthorized(userID int, isAuthorized bool) {
	_, err := wr.db.Exec(UpdateUserAuthorizedQuery, isAuthorized, userID)
	if err != nil {
		log.Panicln(err)
	} 
}

func (wr wrapper) SaveEvent(eventName string, userID int) {
	_, err := wr.db.Exec(CreateEventQuery, eventName, userID)
	if err != nil {
		log.Panicln(err)
	}
}

func (wr wrapper) GetEventsAll() []*EventObject {
	dbrows, err :=  wr.db.Query(SelectAllEvents)
	if err != nil {
		log.Panicln(err)
	}
	resultsArr := parseSQLRows(dbrows, EventObject{})
	return resultsArr 
}

type EventObject struct {
	EventName string `field:"event_name"`
	EventCount string `field:"endorsements_count"`
	OldestRecord string `field:"created_at"`
}

func (wr wrapper) GetEventsByAuthorized(isAuthorized bool) []*EventObject {
	dbrows, err := wr.db.Query(SelectEventsByUserAuthorization, isAuthorized)
	if err != nil {
		log.Panicln(err)
	}
	return parseSQLRows(dbrows, EventObject{})
}

func (wr wrapper) GetEventsByName(eventName string) []*EventObject {
	dbrows, err := wr.db.Query(SelectEventsByName, eventName)
	if err != nil {
		log.Panicln(err)
	}
	return parseSQLRows(dbrows, EventObject{})
}

func (wr wrapper) GetEventsByUserIP(ip string) []*EventObject {
	dbrows, err := wr.db.Query(SelectEventsByIP, ip)
	if err != nil {
		log.Panicln(err)
	}
	return parseSQLRows(dbrows, EventObject{})
}

func applyFilter(template string) func(string)string {
	return func(filter string) string {
		return fmt.Sprintf(template, filter)
	}
}

var (
	SelectEventsBy = applyFilter(TemplateSelectEventsBy)
	SelectEventsByName = SelectEventsBy(FilterEventByName)
	SelectEventsByUserAuthorization = SelectEventsBy(FilterEventBuAthorization)
	SelectEventsByIP = SelectEventsBy(FilterEventByAuthorIP)
	SelectAllEvents = SelectEventsBy("")
)

const (
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
	
	FilterEventByAuthorIP = `
		WHERE (
			SELECT ip_address
			FROM kairastat.users
			WHERE user_id = evs.author_id
		) = $1
	`
	FilterEventByName = `
		WHERE event_name = $1
	`
	FilterEventBuAthorization = `
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

func parseSQLRows[T any](rows *sql.Rows, outputFormat T) ([]*T) {
	defer rows.Close()

	results := make([]*T, 0)
	i := 0
	for rows.Next() {
		results = append(results, new(T))
		fieldMap := ExtractFieldPointersIntoNamedMap(results[i])
		sqlColumns, err := rows.Columns()
		if err != nil {
			log.Panicln(err)
		}

		orderedPointersArr := make([]any, len(fieldMap))
		for i, column := range sqlColumns {
			orderedPointersArr[i] = fieldMap[column]
		}
		err = rows.Scan(orderedPointersArr...)
		if err != nil {
			log.Panicln(err)
		}
		i++
	}

	if err := rows.Err(); err != nil {
		log.Panicln(err)
	}
	return results
}

func ExtractFieldPointersIntoNamedMap[T any](in *T) (map[string]any) {
	fieldMap := make(map[string]any)
	iter := reflect.ValueOf(in).Elem()
	for i := 0; i < iter.NumField(); i++ {
		currPtr := iter.Field(i).Addr().Interface()

		columnName := iter.Type().Field(i).Tag.Get("field") // sql field tag
		if columnName == "" {
			log.Panicln(fmt.Errorf("Struct type %T doesn't provide the necessary field tags for successful sql parsing", *in))
		}

		fieldMap[columnName] = currPtr
	}
	return fieldMap
}



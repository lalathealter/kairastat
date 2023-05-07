package postgre

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"
)

func applyTemplate(template string) func(string)string {
	return func(filter string) string {
		return fmt.Sprintf(template, filter)
	}
}

type SubQueryCB func()(string, []any)
func GenSubQueryWith(temp string) func(...any)SubQueryCB {
	return func(args ...any) SubQueryCB {
		return func()(string, []any) {
			return temp, args
		}
	}
}

func (wr Wrapper) MakeNestedQuery(readFinalQuery SubQueryCB) *sql.Rows {
	queryStatement, args := readFinalQuery()
	rows, err := wr.db.Query(queryStatement, args...)
	if err != nil {
		log.Panicln(err)
	}
	return rows
}

func (wr Wrapper) ChainQuery(innerSubQuery, outerSubQuery SubQueryCB) SubQueryCB {
	outerStmt, outerArgs := outerSubQuery()
	innerStmt, innerArgs := innerSubQuery()
	innerArgs = append(innerArgs, outerArgs...)

	nextTemplate := applyTemplate(outerStmt)(innerStmt)
	nextTemplate = rewriteTemplateArgs(nextTemplate)
	
	return GenSubQueryWith(nextTemplate)(innerArgs...)
}

// rewrites $-sign sql arguements so they go in order from 1 to N
func rewriteTemplateArgs(temp string) string {
	sb := strings.Builder{}

	dec := 1
	argCount := 0
	charsToSkip := 0
	for _, ch := range temp {
		if charsToSkip > 0 {
			charsToSkip--
			continue 
		}
		if ch == ';' {
			continue
		}
		sb.WriteRune(ch)
		if ch == '$' {
			argCount++
			sb.WriteString(fmt.Sprint(argCount))
			if argCount % 10 == 0 {
				dec++
			}
			charsToSkip = dec
		}
	}
	sb.WriteRune(';')

	return sb.String() 
}


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



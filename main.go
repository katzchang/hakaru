package main

import (
	"context"
	"net/http"
	"log"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"encoding/json"

	"cloud.google.com/go/trace/apiv2"
)

type Event struct {
	Name string `json:"name"`
	Value interface{} `json:"value"`
}


func main() {
	ctx := context.Background()
	_, _ = trace.NewClient(ctx)
	//if err != nil {
	//	// TODO: Handle error.
	//}
	//
	//req := &cloudtracepb.BatchWriteSpansRequest{
	//	// TODO: Fill request struct fields.
	//}
	//err = c.BatchWriteSpans(ctx, req)
	//if err != nil {
	//	// TODO: Handle error.
	//}

	dataSourceName := os.Getenv("HAKARU_DATASOURCENAME")
	if dataSourceName == "" {
		dataSourceName = "root:password@tcp(127.0.0.1:13306)/hakaru"
	}
    logger := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	hakaruHandler := func(w http.ResponseWriter, r *http.Request) {
		db, err := sql.Open("mysql", dataSourceName)
		if err != nil {
			panic(err.Error())
		}
		defer db.Close()

		stmt, e := db.Prepare("INSERT INTO eventlog(at, name, value) values(NOW(), ?, ?)")
		if e != nil {
			panic(e.Error())
		}

		defer stmt.Close()

		name := r.URL.Query().Get("name")
		value := r.URL.Query().Get("value")

		bytes, _ := json.Marshal(Event{
			Name: name,
			Value: value,
		})
		logger.Println(string(bytes))


		_, _ = stmt.Exec(name, value)

		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
	}

	http.HandleFunc("/hakaru", hakaruHandler)
	http.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })

	// start server
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}

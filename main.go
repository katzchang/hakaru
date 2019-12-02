package main

import (
	"net/http"
	"log"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"encoding/json"

	"contrib.go.opencensus.io/integrations/ocsql"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
	"contrib.go.opencensus.io/exporter/stackdriver"

)

type Event struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

func main() {
	// Create and register a OpenCensus Stackdriver Trace exporter.
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
	})
	if err != nil {
		log.Fatal(err)
	}
	trace.RegisterExporter(exporter)

	dataSourceName := os.Getenv("HAKARU_DATASOURCENAME")
	if dataSourceName == "" {
		dataSourceName = "root:password@tcp(127.0.0.1:13306)/hakaru"
	}
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	hakaruHandler := func(w http.ResponseWriter, r *http.Request) {
		driverName, err := ocsql.Register("mysql", ocsql.WithAllTraceOptions())
		if err != nil {
			log.Fatalf("Failed to register the ocsql driver: %v", err)
		}

		db, err := sql.Open(driverName, dataSourceName)
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
			Name:  name,
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

	http.Handle("/hakaru", &ochttp.Handler{
		Handler: http.HandlerFunc(hakaruHandler),
	})
	http.Handle("/ok", &ochttp.Handler{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }),
	})

	// start server
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}

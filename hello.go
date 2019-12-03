package main

import (
	"bytes"
	"go.opencensus.io/stats/view"
	"log"
	"net/http"
	"os"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"

	"github.com/newrelic/go-agent"
)

func main() {
	config := newrelic.NewConfig("Your App Name", os.Getenv("NEW_RELIC_LICENSE_KEY"))
	app, err := newrelic.NewApplication(config)

	//exporter, err := nrcensus.NewExporter("My-OpenCensus-App", os.Getenv("NEW_RELIC_INSIGHTS_API_KEY"))
	//if err != nil {
	//	panic(err)
	//}


	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
	})
	if err != nil {
		log.Fatal(err)
	}
	view.RegisterExporter(exporter)
	trace.RegisterExporter(exporter)

	client := &http.Client{
		Transport: &ochttp.Transport{
			// Use Google Cloud propagation format.
			Propagation: &propagation.HTTPFormat{},
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// New Relic
		txn := app.StartTransaction("myTxn", w, r)
		defer txn.End()

		// OpenCensus
		_, span := trace.StartSpan(r.Context(), "handler")
		defer span.End()

		req, _ := http.NewRequest("GET", "https://xxxxxxxxexample.com", nil)

		// The trace ID from the incoming request will be
		// propagated to the outgoing request.
		req = req.WithContext(r.Context())

		// The outgoing request will be traced with r's trace ID.
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Something bad happened!"))
			txn.NoticeError(err)
		} else {
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			body := buf.String()
			resp.Body.Close()
			m := map[string]interface{}{"body": body}
			txn.Application().RecordCustomEvent("hello custom event", m)
		}
	})

	http.Handle("/foo", handler)
	log.Fatal(http.ListenAndServe(":6060", &ochttp.Handler{}))
}

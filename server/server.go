package eventfullserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	evecli "github.com/zerogvt/eventfull/client"
)

const port = ":8080"

type metrics struct {
	count float64
	sum   float64
}

var slis map[string]*metrics

func ingest(w http.ResponseWriter, r *http.Request) {
	var tmp, body []byte
	var unzipped bytes.Buffer
	var err error
	if tmp, err = ioutil.ReadAll(r.Body); err != nil {
		http.Error(w, "Bad Request.", http.StatusBadRequest)
		return
	}
	switch enc := r.Header.Get("Content-Encoding"); enc {
	case "":
		body = tmp
	case "gzip":
		zipped := bytes.NewBuffer(tmp)
		unzipped, err = evecli.UnzipBuffer(*zipped)
		if err != nil {
			log.Println(err)
			http.Error(w, "Cannot unzip.", http.StatusBadRequest)
			return
		}
		body = unzipped.Bytes()
	default:
		http.Error(w, "Don't know how to unzip.", http.StatusBadRequest)
		return
	}

	var rawevt interface{}
	//fmt.Println(string(body))
	if err = json.Unmarshal(body, &rawevt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// calc moving average for this metric
	evt := rawevt.(map[string]interface{})
	metrickey := evt["service"].(string) + ":" + evt["metric"].(string)
	var metricvalue float64
	if metricvalue, err = strconv.ParseFloat(evt["value"].(string), 64); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if val, ok := slis[metrickey]; ok {
		val.count += 1.0
		val.sum += metricvalue
	} else {
		nm := metrics{count: 1.0, sum: metricvalue}
		slis[metrickey] = &nm
	}

	log.Println(evecli.GenericJSONToStr(evt))
}

func stats(w http.ResponseWriter, r *http.Request) {
	for k, v := range slis {
		io.WriteString(w, fmt.Sprintf("metric: %s, samples: %.0f, sum: %.0f, average: %.2f\n",
			k, v.count, v.sum, v.sum/v.count))
	}
}

//Exec executes the server
func Exec() {
	slis = make(map[string]*metrics)
	http.HandleFunc("/ingest", ingest)
	http.HandleFunc("/stats", stats)
	log.Println("Listening on " + port)
	log.Fatal(http.ListenAndServe(port, nil))
}

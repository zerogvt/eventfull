package eventfullserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	evecli "github.com/zerogvt/eventfull/client"
)

const port = ":8080"

type metrics struct {
	Service     string
	Metric      string
	Count       float64
	Sum         float64
	CutoffValue float64
	_under      float64
	_over       float64
	SLI         float64 //Service Level Indicator (what we really get)
	SLO         float64 //Service Level Objective (what we want)
}

var slis map[string]*metrics

func (m metrics) str() string {
	return fmt.Sprintf("count: %.0f, sum: %.0f, cutoffValue: %.0f, slo: %.4f, sli: %.4f",
		m.Count, m.Sum, m.CutoffValue, m.SLO, m.SLI)
}

func decode(r *http.Request) (map[string]interface{}, error) {
	var tmp, body []byte
	var unzipped bytes.Buffer
	var err error
	if tmp, err = ioutil.ReadAll(r.Body); err != nil {
		return nil, errors.New("Bad Request")
	}
	switch enc := r.Header.Get("Content-Encoding"); enc {
	case "":
		body = tmp
	case "gzip":
		zipped := bytes.NewBuffer(tmp)
		unzipped, err = evecli.UnzipBuffer(*zipped)
		if err != nil {
			log.Println(err)
			return nil, errors.New("Cannot unzip")
		}
		body = unzipped.Bytes()
	default:
		return nil, errors.New("Don't know how to unzip")
	}
	var rawevt interface{}
	//fmt.Println(string(body))
	if err = json.Unmarshal(body, &rawevt); err != nil {
		return nil, err
	}
	//fmt.Println(rawevt)
	var evt map[string]interface{}
	var ok bool
	// calc moving average for this metric
	if evt, ok = rawevt.(map[string]interface{}); !ok {
		return nil, errors.New("Data error")
	}
	return evt, nil
}

func ingest(w http.ResponseWriter, r *http.Request) {
	var evt map[string]interface{}
	var err error
	if evt, err = decode(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if evt["eventType"] == "SLI" {
		if err = updateMetric(evt); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if evt["eventType"] == "registration" {
		if err = register(evt); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if _, debug := os.LookupEnv("DEBUG"); debug {
		log.Println(evecli.GenericJSONToStr(evt))
	}
}

func updateMetric(evt map[string]interface{}) error {
	var metricvalue float64
	var err error
	metrickey := evt["service"].(string) + ":" + evt["metric"].(string)
	if metricvalue, err = strconv.ParseFloat(evt["value"].(string), 64); err != nil {
		return err
	}
	if val, ok := slis[metrickey]; ok {
		val.Count += 1.0
		val.Sum += metricvalue
		if metricvalue <= val.CutoffValue {
			val._under++
		} else {
			val._over++
		}
		val.SLI = 100 * val._under / (val._over + val._under)
	} else {
		nm := metrics{
			Service: evt["service"].(string),
			Metric:  evt["metric"].(string),
			Count:   1.0,
			Sum:     metricvalue}
		slis[metrickey] = &nm
	}
	return nil
}

func register(evt map[string]interface{}) error {
	metrickey := evt["service"].(string) + ":" + evt["metric"].(string)
	metricCutoffValue := evt["cutoff_value"].(float64)
	slo := evt["slo"].(float64)
	if _, ok := slis[metrickey]; !ok {
		nm := metrics{
			Service:     evt["service"].(string),
			Metric:      evt["metric"].(string),
			Count:       1.0,
			Sum:         0.0,
			CutoffValue: metricCutoffValue,
			SLO:         slo,
		}
		slis[metrickey] = &nm
	} else {
		slis[metrickey].CutoffValue = metricCutoffValue
		slis[metrickey].SLO = slo
	}

	fmt.Printf("Registered: %s with %s\n", metrickey, slis[metrickey].str())
	return nil
}

func stats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	for _, v := range slis {
		json.NewEncoder(w).Encode(v)
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

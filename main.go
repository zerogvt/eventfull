package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"text/template"
	"time"
)

func fatalif(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func readWildJSON(path string) map[string]interface{} {
	fbytes, err := ioutil.ReadFile(path)
	fatalif(err)
	var f interface{}
	err = json.Unmarshal(fbytes, &f)
	fatalif(err)
	return f.(map[string]interface{})
}

func printWildJSON(m map[string]interface{}) {
	for k, v := range m {
		switch vv := v.(type) {
		case string:
			fmt.Printf("  %s: %s\n", k, vv)
		case float64:
			fmt.Printf("  %s: %f\n", k, vv)
		case bool:
			fmt.Printf("  %s: %t\n", k, vv)
		case map[string]interface{}:
			fmt.Print("  [")
			printWildJSON(vv)
			fmt.Print("  ]\n")
		default:
			fmt.Println(k, "is of a type I don't know how to handle")
		}
	}
}

func getRandomMetric(SLO float64, cutoff float64) float64 {
	rand.Seed(time.Now().UnixNano())
	if rand.Float64() < (SLO / 100.0) {
		return float64(rand.Intn(int(cutoff)))
	}
	return cutoff + float64(rand.Intn(int(cutoff)))
}

func createEvent(ut *template.Template, conf map[string]interface{}) (bytes.Buffer, error) {
	//Execute template according to configuration conf
	var evt bytes.Buffer
	err := ut.Execute(&evt, conf)
	return evt, err
}

func gzipBuffer(buf bytes.Buffer) (bytes.Buffer, error) {
	var err error
	var zipped bytes.Buffer
	zw := gzip.NewWriter(&zipped)
	if _, err = zw.Write(buf.Bytes()); err != nil {
		log.Fatal(err)
	}
	if err = zw.Close(); err != nil {
		log.Fatal(err)
	}
	return zipped, err
}

func postEventToNR(buf bytes.Buffer) error {
	var err error
	var resp *http.Response
	var body []byte
	nrURL := fmt.Sprintf("https://insights-collector.newrelic.com/v1/accounts/%s/events",
		os.Getenv("NEW_RELIC_ACCOUNT_ID"))
	client := &http.Client{}
	req, err := http.NewRequest("POST", nrURL, bytes.NewReader(buf.Bytes()))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Insert-Key", os.Getenv("NEW_RELIC_INSIGHTS_KEY"))
	req.Header.Add("Content-Encoding", "gzip")
	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	fmt.Printf("Response code: %s, body: %s\n", resp.Status, string(body))
	return err
}

func emitEvent(ut *template.Template, conf map[string]interface{}) error {
	//Get our "value"
	conf["value"] = getRandomMetric(conf["metric_slo"].(float64),
		conf["metric_cuttoff_value"].(float64))

	fmt.Printf("value: %4.0f, ", conf["value"])

	//Execute template according to configuration conf
	evt, err := createEvent(ut, conf)
	if err != nil {
		return err
	}
	//gzip evt json
	zbuf, err := gzipBuffer(evt)
	if err != nil {
		return err
	}
	//send gzipped json to NR
	err = postEventToNR(zbuf)
	return err
}

func main() {
	conf := readWildJSON("conf.json")
	fmt.Println("Configuration:")
	printWildJSON(conf)

	//Read event template
	fbytes, err := ioutil.ReadFile("event.json")
	ut, err := template.New("event").Parse(string(fbytes))
	fatalif(err)

	//create and send an event
	for {
		emitEvent(ut, conf)
		time.Sleep(2 * time.Second)
	}
}

package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"
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

func main() {
	conf := readWildJSON("conf.json")
	fmt.Println("Configuration:")
	printWildJSON(conf)

	nrURL := fmt.Sprintf("https://insights-collector.newrelic.com/v1/accounts/%s/events",
		os.Getenv("NEW_RELIC_ACCOUNT_ID"))

	//Read json template
	fbytes, err := ioutil.ReadFile("event.json")
	ut, err := template.New("event").Parse(string(fbytes))
	fatalif(err)

	//Execute template and produce a specific event
	var evt bytes.Buffer
	err = ut.Execute(&evt, conf)
	fatalif(err)

	//gzip json
	var reqBody bytes.Buffer
	zw := gzip.NewWriter(&reqBody)
	if _, err := zw.Write(evt.Bytes()); err != nil {
		log.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		log.Fatal(err)
	}

	//send gzipped json to NR
	client := &http.Client{}
	req, err := http.NewRequest("POST", nrURL, bytes.NewReader(reqBody.Bytes()))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Insert-Key", os.Getenv("NEW_RELIC_INSIGHTS_KEY"))
	req.Header.Add("Content-Encoding", "gzip")
	resp, err := client.Do(req)
	fatalif(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fatalif(err)
	fmt.Printf("Response code: %s, body: %s\n", resp.Status, string(body))
}

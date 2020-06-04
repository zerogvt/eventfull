package eventfullclient

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
	"strings"
	"text/template"
	"time"
)

func fatalif(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// ReadGenericJSON reads a json file
func ReadGenericJSON(path string) (map[string]interface{}, error) {
	fbytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f interface{}
	err = json.Unmarshal(fbytes, &f)
	if err != nil {
		return nil, err
	}
	return f.(map[string]interface{}), nil
}

// GenericJSONToStr returns generic JSON map as a string
func GenericJSONToStr(m map[string]interface{}) string {
	var res strings.Builder
	res.WriteString(fmt.Sprintf("{\n"))
	for k, v := range m {
		switch vv := v.(type) {
		case string:
			res.WriteString(fmt.Sprintf("  %s: %s,\n", k, vv))
		case float64:
			res.WriteString(fmt.Sprintf("  %s: %.2f,\n", k, vv))
		case bool:
			res.WriteString(fmt.Sprintf("  %s: %t,\n", k, vv))
		case map[string]interface{}:
			res.WriteString(fmt.Sprint("  ["))
			res.WriteString(GenericJSONToStr(vv))
			res.WriteString(fmt.Sprint("  ],\n"))
		default:
			res.WriteString(fmt.Sprintln(k, "is of a type I don't know how to handle"))
		}
	}
	res.WriteString(fmt.Sprintf("\n}\n"))
	return res.String()
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

//UnzipBuffer unzips a buffer previously zipped with gzip
func UnzipBuffer(zipped bytes.Buffer) (bytes.Buffer, error) {
	var err error
	var unzipped []byte
	var zr *gzip.Reader
	if zr, err = gzip.NewReader(&zipped); err != nil {
		log.Fatal(err)
	}
	defer zr.Close()
	if unzipped, err = ioutil.ReadAll(zr); err != nil {
		log.Fatal(err)
	}
	return *bytes.NewBuffer(unzipped), err
}

type kv struct {
	key   string
	value string
}

func postEventToNR(buf bytes.Buffer) error {
	nrURL := fmt.Sprintf("https://insights-collector.newrelic.com/v1/accounts/%s/events",
		os.Getenv("NEW_RELIC_ACCOUNT_ID"))
	nrInsightsKeyHeader := kv{"X-Insert-Key", os.Getenv("NEW_RELIC_INSIGHTS_KEY")}
	return postBuffer(nrURL, buf, nrInsightsKeyHeader)
}

func postBuffer(url string, buf bytes.Buffer, headers ...kv) error {
	var err error
	var resp *http.Response
	var body []byte
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewReader(buf.Bytes()))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Encoding", "gzip")
	for _, h := range headers {
		req.Header.Add(h.key, h.value)
	}
	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("[ERROR] %s\n", err)
		return err
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	fmt.Printf("Response code: %s, body: %s\n", resp.Status, string(body))
	return err
}

func emitEvent(ut *template.Template, conf map[string]interface{}) error {
	//Get our "value"
	conf["value"] = getRandomMetric(conf["slo"].(float64),
		conf["cutoff_value"].(float64))

	//Execute template according to configuration conf
	evt, err := createEvent(ut, conf)
	if err != nil {
		return err
	}
	//gzip evt json
	fmt.Printf("\nWill zip and POST evt:%s", string(evt.Bytes()))
	zbuf, err := gzipBuffer(evt)
	if err != nil {
		return err
	}

	if url, ok := conf["url"]; ok {
		err = postBuffer(url.(string), zbuf)
	} else {
		//send gzipped json to NR
		err = postEventToNR(zbuf)
	}
	return err
}

func postJSON(url string, data map[string]interface{}) error {
	databytes, err := json.Marshal(data)
	fmt.Println(string(databytes))

	//gzip json
	zbuf, err := gzipBuffer(*bytes.NewBuffer(databytes))
	if err != nil {
		return err
	}
	return postBuffer(url, zbuf)
}

// Daemon will loop according to settings in configurationFile and
// send out events cookie cut from eventTemplateFile
func Daemon(configurationFile string, eventTemplateFile string) {
	conf, err := ReadGenericJSON(configurationFile)
	fatalif(err)
	fmt.Println("Configuration:")
	fmt.Print(GenericJSONToStr(conf))

	//Read event template
	fbytes, err := ioutil.ReadFile(eventTemplateFile)
	ut, err := template.New("event").Parse(string(fbytes))
	fatalif(err)

	conf["eventType"] = "registration"
	postJSON(conf["url"].(string), conf)
	//create and send an event
	for repeat := true; repeat; {
		emitEvent(ut, conf)
		repeat = false
		if r, ok := conf["repeat_every_msecs"]; ok {
			if rv, ok := r.(float64); ok {
				time.Sleep(time.Duration(rv) * time.Millisecond)
				repeat = true
			}
		}
	}
}

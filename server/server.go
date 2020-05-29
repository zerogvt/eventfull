package eventfullserver

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	evecli "github.com/zerogvt/eventfull/client"
)

const port = ":8080"

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

	var evt interface{}
	//fmt.Println(string(body))
	if err = json.Unmarshal(body, &evt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println(evecli.GenericJSONToStr(evt.(map[string]interface{})))
}

//Exec executes the server
func Exec() {
	http.HandleFunc("/ingest", ingest)
	log.Println("Listening on " + port)
	log.Fatal(http.ListenAndServe(port, nil))
}

package eventfullserver

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	evecli "github.com/zerogvt/eventfull/client"
)

func ingest(w http.ResponseWriter, r *http.Request) {
	var body []byte
	var err error
	if body, err = ioutil.ReadAll(r.Body); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	var evt interface{}
	if err = json.Unmarshal(body, &evt); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	evecli.PrintGenericJSON(evt.(map[string]interface{}))
	//fmt.Fprintf(w, "all good")
}

//Exec executes the server
func Exec() {
	http.HandleFunc("/ingest", ingest)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

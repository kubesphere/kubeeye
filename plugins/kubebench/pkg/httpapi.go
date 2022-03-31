package pkg

import (
	"encoding/json"
	"net/http"

	log "github.com/golang/glog"
)

func handle(res http.ResponseWriter, req *http.Request) {
	log.Info("get KubeBench result")
	result := *KBResult
	jsonResults, err := json.Marshal(result)
	if err != nil {
		log.Errorf("Marshal KubeBench result failed")
	}
	_, err = res.Write(jsonResults)
	if err != nil {
		log.Errorf("Write KubeBench result failed")
	}
}

func KubeBenchAPI() {
	http.HandleFunc("/plugins", handle)
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Errorf("start KubeBench server failed")
	}
}

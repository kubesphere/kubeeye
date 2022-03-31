package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/golang/glog"
)

func handle(res http.ResponseWriter, req *http.Request) {
	fmt.Println("get result")
	result := *KSResult
	jsonResults, err := json.Marshal(result)
	if err != nil {
		fmt.Println("Marshal result failed")
	}
	_, err = res.Write(jsonResults)
	if err != nil {
		fmt.Println("Write result failed")
	}
}

func KubescapeAPI() {
	http.HandleFunc("/plugins", handle)
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Errorf("start server failed")
	}
}

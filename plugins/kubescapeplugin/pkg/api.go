package pkg

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func KubeScapeAPI() {
	mux := http.NewServeMux()
	mux.Handle("/plugins", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		result, err := KubescapeAudit()
		if err != nil {
			log.Printf("KubeScape audit failed: %+v", err)
		}
		jsonResults, err := json.Marshal(&result)
		if err != nil {
			log.Printf("Marshal KubeScape result failed: %+v", err)
		}
		_, err = writer.Write(jsonResults)
		if err != nil {
			log.Printf("Write KubeScape result to response writer failed: %+v", err)
		}
	}))

	mux.Handle("/start", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		pluginAudit()
	}))

	mux.Handle("/healthz", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))

	log.Println("KubeScape audit API ready")
	log.Fatal(http.ListenAndServe(":80", mux))
}

func pluginAudit() {
	go func() {
		result, err := KubescapeAudit()
		if err != nil {
			log.Printf("KubeScape audit failed: %+v", err)
		}
		jsonResults, err := json.Marshal(&result)
		if err != nil {
			log.Printf("Marshal KubeScape result failed: %+v", err)
		}

		req, err := http.NewRequest("POST", "http://kubeeye-controller-manager-service.kubeeye-system.svc/plugins?name=kubesacpe", bytes.NewReader(jsonResults))
		if err != nil {
			log.Printf("Create request failed: %+v", err)
		}

		tr := &http.Transport{
			IdleConnTimeout:    5 * time.Second, // the maximum amount of time an idle connection will remain idle before closing itself.
			DisableCompression: true,            // prevents the Transport from requesting compression with an "Accept-Encoding: gzip" request header when the Request contains no existing Accept-Encoding value.
			WriteBufferSize:    10 << 10,        // specifies the size of the write buffer to 10KB used when writing to the transport.
		}
		client := &http.Client{Transport: tr}

		_, err = client.Do(req)
		if err != nil {
			log.Printf("Push plugin result to kubeeye failed: %+v", err)
		}
		log.Printf("Push plugin result to kubeeye successful")
	}()
}

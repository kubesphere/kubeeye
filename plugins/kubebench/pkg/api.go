package pkg

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func KubeBenchAPI() {
	mux := http.NewServeMux()
	mux.Handle("/plugins", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		result, err := KubeBenchAudit()
		if err != nil {
			log.Printf( "KubeBench audit failed: %+v", err)
		}
		jsonResults, err := json.Marshal(&result)
		if err != nil {
			log.Printf("Marshal KubeBench result failed: %+v", err)
		}
		_, err = writer.Write(jsonResults)
		if err != nil {
			log.Printf("Write KubeBench result to response writer failed: %+v", err)
		}
	}))

	mux.Handle("/healthz", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		io.WriteString(writer, "health")
	}))

	log.Println("KubeBench audit API ready")
	log.Fatal(http.ListenAndServe(":80", mux))
}
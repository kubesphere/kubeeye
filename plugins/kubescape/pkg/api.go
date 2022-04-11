package pkg

import (
    "encoding/json"
    "io"
    "log"
    "net/http"
)

func KubeScapeAPI() {
    mux := http.NewServeMux()
    mux.Handle("/plugins", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
        result, err := KubescapeAudit()
        if err != nil {
            log.Printf( "KubeScape audit failed: %+v", err)
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

    mux.Handle("/healthz", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
        io.WriteString(writer, "health")
    }))

    log.Println("KubeScape audit API ready")
    log.Fatal(http.ListenAndServe(":80", mux))
}
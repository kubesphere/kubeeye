package pkg

import (
    "encoding/json"
    "io"
    "log"
    "net/http"
)

func KubeHunterAPI()  {
    mux := http.NewServeMux()
    mux.Handle("/plugins", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        result, err := KubeHunterAudit()
        if err != nil {
            log.Printf( "KubeHunter audit failed: %+v", err)
        }
        jsonResults, err := json.Marshal(&result)
        if err != nil {
            log.Printf("Marshal KubeHunter result failed: %+v", err)
        }
        _, err = w.Write(jsonResults)
        if err != nil {
            log.Printf("Write KubeBench result to response writer failed: %+v", err)
        }
    }))

    mux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        io.WriteString(w, "health")
    }))

    log.Println("KubeBench audit API ready")
    log.Fatal(http.ListenAndServe(":80", mux))
}
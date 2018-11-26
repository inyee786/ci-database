package main

import (
	"net/http"

	_ "github.com/lib/pq"
	"github.com/openebs/ci-database/database"
	"github.com/openebs/ci-database/handler"
	log "github.com/sirupsen/logrus"
)

func main() {
	database.InitDb()
	// defer database.Db.Close()
	http.HandleFunc("/gke", handler.Gkehandler)
	http.HandleFunc("/aws", handler.Awshandler)
	http.HandleFunc("/gcp", handler.Gcphandler)
	http.HandleFunc("/azure", handler.Azurehandler)
	http.HandleFunc("/packet", handler.Packethandler)
	http.HandleFunc("/eks", handler.Ekshandler)
	http.HandleFunc("/build", handler.Buildhandler)
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

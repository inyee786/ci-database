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
	// defer db.Close()
	http.HandleFunc("/gke", handler.Gkehandler)
	http.HandleFunc("/aws", handler.Awshandler)
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

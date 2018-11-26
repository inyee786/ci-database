package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
)

var db *sql.DB

const (
	dbhost = "DBHOST"
	dbport = "DBPORT"
	dbuser = "DBUSER"
	dbpass = "DBPASS"
	dbname = "DBNAME"
)

// dashboardSummary contains the details of a gitlab pipelines
type dashboardSummary struct {
	ID     int    `json:"id"`
	Sha    string `json:"sha"`
	Ref    string `json:"ref"`
	Status string `json:"status"`
	WebURL string `json:"web_url"`
}

// Pipeline struct
type Pipeline []struct {
	ID     int    `json:"id"`
	Sha    string `json:"sha"`
	Ref    string `json:"ref"`
	Status string `json:"status"`
	WebURL string `json:"web_url"`
}

type dashboard struct {
	Dashboard []dashboardSummary
}

type dashboardaws struct {
	Dashboardaws []dashboardSummary
}

func main() {
	initDb()
	defer db.Close()
	http.HandleFunc("/gke", gkedata)
	http.HandleFunc("/aws", awsdata)
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

func gkedata(w http.ResponseWriter, r *http.Request) {
	datas := dashboard{}
	err := queryData(&datas)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	out, err := json.Marshal(datas)
	// log.Infof("%+v", string(out))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Fprintf(w, string(out))
	go gkeData()
}

func awsdata(w http.ResponseWriter, r *http.Request) {
	datasaws := dashboardaws{}
	err := queryAwsData(&datasaws)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	outaws, err := json.Marshal(datasaws)
	// log.Infof("%+v", string(out))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Fprintf(w, string(outaws))
	go awsData()
}

// Get data from gitlab api for gke and push to database
func gkeData() {
	url := "https://gitlab.openebs.ci/api/v4/projects/24/pipelines?ref=master"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("PRIVATE-TOKEN", "GN5eUuyg-ybHErwYLR3T")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	// fmt.Println(string(body))
	var y Pipeline
	json.Unmarshal(body, &y)
	for i := range y {
		sqlStatement := `
			INSERT INTO gke (id, sha, ref, status, web_url)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO UPDATE
			SET status = $4
			RETURNING id`
		id := 0
		err = db.QueryRow(sqlStatement, y[i].ID, y[i].Sha, y[i].Ref, y[i].Status, y[i].WebURL).Scan(&id)
		if err != nil {
			panic(err)
		}
		fmt.Println("New record ID for gke:", id)
	}
}

// Get data from gitlab api for aws and push to database
func awsData() {
	url := "https://gitlab.openebs.ci/api/v4/projects/7/pipelines?ref=master"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("PRIVATE-TOKEN", "GN5eUuyg-ybHErwYLR3T")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	// fmt.Println(string(body))
	var y Pipeline
	json.Unmarshal(body, &y)
	for i := range y {
		sqlStatement := `
			INSERT INTO aws (id, sha, ref, status, web_url)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO UPDATE
			SET status = $4
			RETURNING id`
		id := 0
		err = db.QueryRow(sqlStatement, y[i].ID, y[i].Sha, y[i].Ref, y[i].Status, y[i].WebURL).Scan(&id)
		if err != nil {
			panic(err)
		}
		fmt.Println("New record ID for aws:", id)
	}
}

// queryData first fetches the dashboard data from the db
func queryData(datas *dashboard) error {
	rows, err := db.Query(`
		SELECT
			*
		FROM gke
		ORDER BY id DESC`)
	// log.Infof("%+v", rows)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		data := dashboardSummary{}
		err = rows.Scan(
			&data.ID,
			&data.Sha,
			&data.Ref,
			&data.Status,
			&data.WebURL,
		)
		if err != nil {
			return err
		}
		// log.Infof("%+v", data)
		datas.Dashboard = append(datas.Dashboard, data)
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
}

// queryData first fetches the dashboard data from the db
func queryAwsData(datasaws *dashboardaws) error {
	rows, err := db.Query(`
		SELECT
			*
		FROM aws
		ORDER BY id DESC`)
	// log.Infof("%+v", rows)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		data := dashboardSummary{}
		err = rows.Scan(
			&data.ID,
			&data.Sha,
			&data.Ref,
			&data.Status,
			&data.WebURL,
		)
		if err != nil {
			return err
		}
		// log.Infof("%+v", data)
		datasaws.Dashboardaws = append(datasaws.Dashboardaws, data)
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
}

func initDb() {
	config := dbConfig()
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config[dbhost], config[dbport],
		config[dbuser], config[dbpass], config[dbname])

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected to Database!")
}

func dbConfig() map[string]string {
	conf := make(map[string]string)
	host, ok := os.LookupEnv(dbhost)
	if !ok {
		panic("DBHOST environment variable required but not set")
	}
	port, ok := os.LookupEnv(dbport)
	if !ok {
		panic("DBPORT environment variable required but not set")
	}
	user, ok := os.LookupEnv(dbuser)
	if !ok {
		panic("DBUSER environment variable required but not set")
	}
	password, ok := os.LookupEnv(dbpass)
	if !ok {
		panic("DBPASS environment variable required but not set")
	}
	name, ok := os.LookupEnv(dbname)
	if !ok {
		panic("DBNAME environment variable required but not set")
	}
	conf[dbhost] = host
	conf[dbport] = port
	conf[dbuser] = user
	conf[dbpass] = password
	conf[dbname] = name
	return conf
}

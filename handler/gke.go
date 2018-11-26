package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/openebs/ci-database/database"
)

// Gkehandler for fetch data from gitlab api as well as database
func Gkehandler(w http.ResponseWriter, r *http.Request) {
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
		err = database.Db.QueryRow(sqlStatement, y[i].ID, y[i].Sha, y[i].Ref, y[i].Status, y[i].WebURL).Scan(&id)
		if err != nil {
			panic(err)
		}
		fmt.Println("New record ID for gke:", id)
	}
}

// queryData first fetches the dashboard data from the db
func queryData(datas *dashboard) error {
	rows, err := database.Db.Query(`
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

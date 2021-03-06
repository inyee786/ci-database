package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/openebs/ci-database/database"
)

// Ekshandler return eks pipeline data to api
func Ekshandler(w http.ResponseWriter, r *http.Request) {
	datas := dashboard{}
	err := QueryEksData(&datas)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	out, err := json.Marshal(datas)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Fprintf(w, string(out))
	go EksData()
}

// EksData from gitlab api for aws and dump to database
func EksData() {
	url := "https://gitlab.openebs.ci/api/v4/projects/27/pipelines?ref=master"
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
			INSERT INTO eks (id, sha, ref, status, web_url)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO UPDATE
			SET status = $4
			RETURNING id`
		id := 0
		err = database.Db.QueryRow(sqlStatement, y[i].ID, y[i].Sha, y[i].Ref, y[i].Status, y[i].WebURL).Scan(&id)
		if err != nil {
			panic(err)
		}
		fmt.Println("New record ID for aws:", id)
	}
}

// QueryEksData first fetches the dashboard data from the db
func QueryEksData(datas *dashboard) error {
	rows, err := database.Db.Query(`
		SELECT
			*
		FROM eks
		ORDER BY id DESC`)
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
		datas.Dashboard = append(datas.Dashboard, data)
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
}

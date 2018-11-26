package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/openebs/ci-database/database"
)

// Buildhandler return packet pipeline data to api
func Buildhandler(w http.ResponseWriter, r *http.Request) {
	datas := dashboard{}
	err := QueryBuildData(&datas)
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
	go BuildData()
}

// BuildData from gitlab api for packet and dump to database
func BuildData() {
	// Fetch jiva data from gitlab
	jivaURL := "https://gitlab.openebs.ci/api/v4/projects/28/pipelines?ref=master"
	req, _ := http.NewRequest("GET", jivaURL, nil)
	req.Header.Add("PRIVATE-TOKEN", "GN5eUuyg-ybHErwYLR3T")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	jivaData, _ := ioutil.ReadAll(res.Body)
	// fmt.Println(string(jivaData))

	// Fetch maya data from gitlab
	mayaURL := "https://gitlab.openebs.ci/api/v4/projects/31/pipelines?ref=master"
	req, _ = http.NewRequest("GET", mayaURL, nil)
	req.Header.Add("PRIVATE-TOKEN", "GN5eUuyg-ybHErwYLR3T")
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	mayaData, _ := ioutil.ReadAll(res.Body)

	var y, z Pipeline
	json.Unmarshal(jivaData, &y)
	json.Unmarshal(mayaData, &z)
	for i := range z {
		y = append(y, z[i])
	}
	for i := range y {
		sqlStatement := `
			INSERT INTO build (id, sha, ref, status, web_url)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO UPDATE
			SET status = $4
			RETURNING id`
		id := 0
		err = database.Db.QueryRow(sqlStatement, y[i].ID, y[i].Sha, y[i].Ref, y[i].Status, y[i].WebURL).Scan(&id)
		if err != nil {
			panic(err)
		}
		fmt.Println("New record ID for build:", id)
	}
}

// QueryBuildData first fetches the dashboard data from the db
func QueryBuildData(datas *dashboard) error {
	rows, err := database.Db.Query(`
		SELECT
			*
		FROM build
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

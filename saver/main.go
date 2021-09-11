package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-gorp/gorp"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"strconv"
	"time"
)

type dbData struct {
	Serial      int       `db:"serial"`
	DeviceId    int       `db:"deviceId"`
	Date        time.Time `db:"date"`
	Temperature float64   `db:"temperature"`
	Humidity    float64   `db:"humidity"`
}

type rawData struct {
	Device      int     `json:"device"`
	Date        string  `json:"date"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
}

const table = "data"

var db *gorp.DbMap

func initDb(fileName string) *gorp.DbMap {
	// connect to rawDb using standard Go database/sql API
	// use whatever database/sql driver you wish
	rawDb, err := sql.Open("sqlite3", fileName+"?parseTime=true")
	checkErr(err, "sql.Open failed")

	// construct a gorp DbMap
	dbmap := &gorp.DbMap{Db: rawDb, Dialect: gorp.SqliteDialect{}}
	// add a table, setting the table name to 'data' and
	// specifying that the DeviceId property is an auto incrementing PK
	dbmap.AddTableWithName(dbData{}, table).SetKeys(true, "serial")

	// create the table. in a production system you'd generally
	// use a migration tool, or create the tables via scripts
	err = dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")

	return dbmap
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}

func httpErrorHandler(w http.ResponseWriter, err error, StatusCode int) {
	if err != nil {
		http.Error(w, fmt.Sprintf("ERROR is %e\n", err), StatusCode)
	}
}

func save(w http.ResponseWriter, r *http.Request) {
	var d rawData
	err := json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		httpErrorHandler(w, err, http.StatusBadRequest)
		return
	}
	date, err := time.Parse(time.RFC3339, d.Date)
	if err != nil {
		httpErrorHandler(w, err, http.StatusBadRequest)
		return
	}

	dbD := dbData{
		DeviceId:    d.Device,
		Date:        date,
		Temperature: d.Temperature,
		Humidity:    d.Humidity,
	}
	if !dbD.validate() {
		err = errors.New("input value error\n")
		httpErrorHandler(w, err, http.StatusBadRequest)
		return
	}

	err = db.Insert(&dbD)
	if err != nil {
		httpErrorHandler(w, err, http.StatusBadRequest)
		return
	}
	w.Write([]byte(fmt.Sprintf("Success!")))
	w.WriteHeader(http.StatusOK)

}
func (d *dbData) validate() bool {
	lower := map[string]float64{"temp": -10, "hum": 0}
	upper := map[string]float64{"temp": 50, "hum": 100}
	return (lower["temp"] < d.Temperature && d.Temperature < upper["temp"]) && (lower["hum"] < d.Humidity && d.Humidity < upper["hum"])
}

func get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		err := errors.New("Method error. Please Use 'GET' method\n")
		httpErrorHandler(w, err, http.StatusMethodNotAllowed)
		return
	}
	var datas []dbData
	buf := r.URL.Query().Get("deviceId")
	var query string
	if len(buf) != 0 {
		id, err := strconv.Atoi(buf)
		if err != nil {
			httpErrorHandler(w, err, http.StatusBadRequest)
			return
		}
		query = fmt.Sprintf("SELECT * FROM %s WHERE deviceId=%d", "main."+table, id)
	} else {
		query = "SELECT * FROM main.data"
	}
	//buf := dbData{}
	//_, err := db.Get(&datas, &buf)
	_, err := db.Select(&datas, query)
	if err != nil {
		httpErrorHandler(w, err, http.StatusInternalServerError)
		return
	}
	b, err := json.Marshal(datas)
	if err != nil {
		log.Fatal(err)
	}
	_, err = w.Write(b)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	db = initDb("/home/azukibar/uec/koken/koken_contest/2021/kokenDataLogger/entifier.sqlite")
	defer db.Db.Close()

	m := mux.NewRouter()
	m.HandleFunc("/save", save)
	m.HandleFunc("/get", get)
	http.Handle("/", m)
	http.ListenAndServe(":3000", m)
}

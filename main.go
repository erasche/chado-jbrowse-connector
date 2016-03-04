package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
)

var db *sqlx.DB

func okJson(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.WriteHeader(http.StatusOK)
}

func notOkJson(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.WriteHeader(http.StatusBadRequest)
}

func StatsGlobal(w http.ResponseWriter, r *http.Request) {
	okJson(w)
	w.Write([]byte("{\"featureDensity\": 0.1}"))
}

func FeatureSeqHandler(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	start, _ := strconv.Atoi(r.Form.Get("start"))
	end, _ := strconv.Atoi(r.Form.Get("end"))

	vars := mux.Vars(r)
	organism := vars["organism"]
	refseq := vars["refseq"]

	if end < start {
		notOkJson(w)
		return
	}

	if r.Form.Get("sequence") == "true" {
		SequenceHandler(w, r, organism, refseq, start, end)
	} else {
		FeatureHandler(w, r, organism, refseq, start, end)
	}
}

type FeatureContainerRefSeqWithStruct struct {
	Features []refSeqWithSeqStruct `json:"features"`
}
type FeatureContainerFeatures struct {
	Features []ProcessedFeature `json:"features"`
}

func SequenceHandler(w http.ResponseWriter, r *http.Request, organism string, refseq string, start int, end int) {
	seq := []refSeqWithSeqStruct{}
	err := db.Select(&seq, refSeqWithSeqQuery, organism, refseq, start, end-start)
	if err != nil {
		panic(err)
	}
	for idx := range seq {
		seq[idx].Start = start
		seq[idx].End = end
	}

	container := &FeatureContainerRefSeqWithStruct{
		Features: seq,
	}

	okJson(w)
	if err := json.NewEncoder(w).Encode(container); err != nil {
		panic(err)
	}
}

func FeatureHandler(w http.ResponseWriter, r *http.Request, organism string, refseq string, start int, end int) {
	features := []TripFeature{}

	err := db.Select(&features, featureListQuery, start, end)
	if err != nil {
		fmt.Println(err);
	}

	for _, f := range features {
		fmt.Printf("%#v\n", f)
	}

	okJson(w)
	if err := json.NewEncoder(w).Encode(features); err != nil {
		panic(err)
	}
}

func RefSeqs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organism := vars["organism"]
	seqs := []refSeqStruct{}
	db.Select(&seqs, refSeqQuery, organism)

	for idx, _ := range seqs {
		seqs[idx].SeqChunkSize = 20000
		seqs[idx].End = seqs[idx].Length
	}

	okJson(w)
	if err := json.NewEncoder(w).Encode(seqs); err != nil {
		panic(err)
	}
}

func main() {
	// this Pings the database trying to connect, panics on error
	// use sqlx.Open() for sql.Open() semantics
	var err error
	db, err = sqlx.Connect("postgres", "postgres://postgres:postgres@localhost/postgres?sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}

	//FeatureList("yeast", 1000, 20000)
	r := mux.NewRouter()
	r.HandleFunc("/{organism}/refSeqs.json", RefSeqs)
	r.HandleFunc("/{organism}/stats/global", StatsGlobal)
	r.HandleFunc("/{organism}/features/{refseq}", FeatureSeqHandler)
	http.Handle("/", r)
	http.ListenAndServe(":5000", r)
}

package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

var db *sqlx.DB

func okJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.WriteHeader(http.StatusOK)
}

func notOkJSON(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.WriteHeader(http.StatusBadRequest)
}

func StatsGlobal(w http.ResponseWriter, r *http.Request) {
	okJSON(w)
	w.Write([]byte("{\"featureDensity\": 0.01}"))
}

func FeatureSeqHandler(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	start, _ := strconv.Atoi(r.Form.Get("start"))
	end, _ := strconv.Atoi(r.Form.Get("end"))

	vars := mux.Vars(r)
	organism := vars["organism"]
	refseq := vars["refseq"]

	if end < start {
		notOkJSON(w)
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
	Features []SimpleFeature `json:"features"`
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

	okJSON(w)
	if err := json.NewEncoder(w).Encode(container); err != nil {
		panic(err)
	}
}

func FeatureHandler(w http.ResponseWriter, r *http.Request, organism string, refseq string, start int, end int) {
	features := []SimpleFeature{}
	soType := r.Form.Get("soType")

	err := db.Select(&features, simpleFeatQuery, organism, refseq, soType, start, end)
	if err != nil {
		fmt.Println(err)
	}

	for idx := range features {
		features[idx].Subfeatures = []ProcessedFeature{}
	}

	container := &FeatureContainerFeatures{
		Features: features,
	}

	okJSON(w)
	if err := json.NewEncoder(w).Encode(container); err != nil {
		panic(err)
	}
}

func listOrganisms() []Organism {
	data := []Organism{}
	err := db.Select(&data, OrganismQuery)
	if err != nil {
		fmt.Println(err)
	}
	return data
}

func listSoTypes(organism string) []SoType {
	soTypeList := []SoType{}
	err := db.Select(&soTypeList, SoTypeQuery, organism)
	if err != nil {
		fmt.Println(err)
	}
	return soTypeList
}

func refSeqsData(organism string) []refSeqStruct {
	seqs := []refSeqStruct{}
	db.Select(&seqs, refSeqQuery, organism)

	for idx := range seqs {
		seqs[idx].SeqChunkSize = 20000
		seqs[idx].End = seqs[idx].Length
	}
	return seqs
}

func RefSeqs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organism := vars["organism"]
	seqs := refSeqsData(organism)
	okJSON(w)
	if err := json.NewEncoder(w).Encode(seqs); err != nil {
		panic(err)
	}
}

type TrackList struct {
	RefSeqs string        `json:"refSeqs"`
	Names   NameStruct    `json:"names"`
	Tracks  []interface{} `json:"tracks"`
}

type NameStruct struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type TrackListTrack struct {
	Category               string            `json:"category"`
	Label                  string            `json:"label"`
	Type                   string            `json:"type"`
	TrackType              string            `json:"trackType"`
	Key                    string            `json:"key"`
	Query                  map[string]string `json:"query"`
	RegionFeatureDensities bool              `json:"regionFeatureDensities"`
	StoreClass             string            `json:"storeClass"`
}

type SeqTrack struct {
	UseAsRefSeqStore bool              `json:"useAsRefSeqStore"`
	Label            string            `json:"label"`
	Key              string            `json:"key"`
	Type             string            `json:"type"`
	StoreClass       string            `json:"storeClass"`
	BaseURL          string            `json:"baseUrl"`
	Query            map[string]string `json:"query"`
}

func OrgTracksConf(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.WriteHeader(http.StatusOK)
}

func OrgTrackListJSON(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organism := vars["organism"]

	var tracks []interface{}
	queryMap := make(map[string]string)
	queryMap["sequence"] = "true"

	tracks = append(tracks, SeqTrack{
		UseAsRefSeqStore: true,
		Label:            "ref_seq",
		Key:              "REST Reference Sequence",
		Type:             "JBrowse/View/Track/Sequence",
		StoreClass:       "JBrowse/Store/SeqFeature/REST",
		BaseURL:          addr + "/link/" + organism + "/",
		Query:            queryMap,
	})

	for _, sotype := range listSoTypes(organism) {
		tmpMap := make(map[string]string)
		tmpMap["soType"] = sotype.Type
		tracks = append(tracks, TrackListTrack{
			Category: "Generic SO Type Tracks",
			Label:    organism + "_" + sotype.Type,
			Key:      sotype.Type,
			Query:    tmpMap,
			RegionFeatureDensities: true,
			Type:       "JBrowse/View/Track/HTMLFeatures",
			TrackType:  "JBrowse/View/Track/HTMLFeatures",
			StoreClass: "JBrowse/Store/SeqFeature/REST",
		})
	}

	data := &TrackList{
		RefSeqs: addr + "/link/" + organism + "/refSeqs.json",
		Names: NameStruct{
			Type: "REST",
			URL:  addr + "/link/names",
		},
		Tracks: tracks,
	}

	okJSON(w)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		panic(err)
	}
}

func MainHandler(w http.ResponseWriter, r *http.Request) {
	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}
	t, err := template.New("webpage").Parse(homeTemplate)
	check(err)

	orgs := listOrganisms()
	var items []string
	for _, org := range orgs {
		items = append(items, org.CommonName)
	}

	data := struct {
		Title      string
		Items      []string
		FakeDirURL string
	}{
		Title:      "Chado-JBrowse Connector",
		FakeDirURL: addr + "/link",
		Items:      items,
	}

	err = t.Execute(w, data)
	check(err)
}

func Db(dbURL, listen string) {
	var err error
	db, err = sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalln(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/link/{organism}/refSeqs.json", RefSeqs)
	r.HandleFunc("/link/{organism}/stats/global", StatsGlobal)
	r.HandleFunc("/link/{organism}/features/{refseq}", FeatureSeqHandler)
	r.HandleFunc("/link/{organism}/tracks.conf", OrgTracksConf)
	r.HandleFunc("/link/{organism}/trackList.json", OrgTrackListJSON)
	r.HandleFunc("/", MainHandler)

	http.Handle("/", r)
	http.ListenAndServe(listen, r)
}

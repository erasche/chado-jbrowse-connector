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

func statsGlobal(w http.ResponseWriter, r *http.Request) {
	okJSON(w)
	_, _ = w.Write([]byte("{\"featureDensity\": 0.01}"))
}

func featureSeqHandler(w http.ResponseWriter, r *http.Request) {
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
		sequenceHandler(w, r, organism, refseq, start, end)
	} else {
		featureHandler(w, r, organism, refseq, start, end)
	}
}

func sequenceHandler(w http.ResponseWriter, r *http.Request, organism string, refseq string, start int, end int) {
	seq := []refSeqWithSeqStruct{}
	err := db.Select(&seq, refSeqWithSeqQuery, organism, refseq, start, end-start)
	if err != nil {
		panic(err)
	}
	for idx := range seq {
		seq[idx].Start = start
		seq[idx].End = end
	}

	container := &featureContainerRefSeqWithStruct{
		Features: seq,
	}

	okJSON(w)
	if err := json.NewEncoder(w).Encode(container); err != nil {
		panic(err)
	}
}

func featureHandler(w http.ResponseWriter, r *http.Request, organism string, refseq string, start int, end int) {
	features := []simpleFeature{}
	soType := r.Form.Get("soType")

	err := db.Select(&features, simpleFeatQueryWithParent, organism, refseq, soType, start, end)
	if err != nil {
		fmt.Println(err)
	}
	toplevels := []simpleFeature{}

	for idx := range features {
		if !features[idx].NullName.Valid {
			features[idx].Name = "(unnamed)"
		} else {
			features[idx].Name = features[idx].NullName.String
		}
		features[idx].Subfeatures = []simpleFeature{}
		if features[idx].Type == soType {
			toplevels = append(toplevels, features[idx])
		} else {
			for idx2 := range toplevels {
				if features[idx].ParentID == toplevels[idx2].FeatureID {
					toplevels[idx2].Subfeatures = append(toplevels[idx2].Subfeatures, features[idx])
				}
			}
		}
	}

	container := &featureContainerFeatures{
		Features: toplevels,
	}

	okJSON(w)
	if err := json.NewEncoder(w).Encode(container); err != nil {
		panic(err)
	}
}

func listOrganisms() []organism {
	data := []organism{}
	err := db.Select(&data, organismQuery)
	if err != nil {
		fmt.Println(err)
	}
	return data
}

func listSoTypes(organism string) []soType {
	soTypeList := []soType{}
	err := db.Select(&soTypeList, soTypeQuery, organism)
	if err != nil {
		fmt.Println(err)
	}
	return soTypeList
}

func refSeqsData(organism string) []refSeqStruct {
	var err error
	seqs := []refSeqStruct{}
	err = db.Select(&seqs, refSeqQuery, organism)
	if err != nil {
		log.Fatalln(err)
	}

	for idx := range seqs {
		seqs[idx].SeqChunkSize = 20000
		seqs[idx].End = seqs[idx].Length
	}
	return seqs
}

func refSeqs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organism := vars["organism"]
	seqs := refSeqsData(organism)
	okJSON(w)
	if err := json.NewEncoder(w).Encode(seqs); err != nil {
		panic(err)
	}
}

func orgTracksConf(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.WriteHeader(http.StatusOK)
}

func orgTrackListJSON(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organism := vars["organism"]

	var tracks []interface{}
	queryMap := make(map[string]string)
	queryMap["sequence"] = "true"

	tracks = append(tracks, seqTrack{
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
		tracks = append(tracks, trackListTrack{
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

	data := &trackList{
		RefSeqs: addr + "/link/" + organism + "/refSeqs.json",
		Names: nameStruct{
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

func mainHandler(w http.ResponseWriter, r *http.Request) {
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

func connect(dbURL, listen string) {
	var err error
	db, err = sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalln(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/link/{organism}/refSeqs.json", refSeqs)
	r.HandleFunc("/link/{organism}/stats/global", statsGlobal)
	r.HandleFunc("/link/{organism}/features/{refseq}", featureSeqHandler)
	r.HandleFunc("/link/{organism}/tracks.conf", orgTracksConf)
	r.HandleFunc("/link/{organism}/trackList.json", orgTrackListJSON)
	r.HandleFunc("/", mainHandler)

	http.Handle("/", r)
	_ = http.ListenAndServe(listen, r)
}

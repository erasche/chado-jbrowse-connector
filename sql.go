package main

type organism struct {
	Genus      string `db:"genus"`
	Species    string `db:"species"`
	CommonName string `db:"common_name"`
}

var organismQuery = `
SELECT
	genus, species, common_name
FROM
	organism
`

type soType struct {
	Type string `db:"type"`
}

var soTypeQuery = `
SELECT
	cvterm.name as type
FROM
	feature, cvterm
WHERE
	(feature.organism_id = (select organism_id from organism where common_name = $1))
	AND
	feature.type_id = cvterm.cvterm_id
GROUP BY
	cvterm.name
	;

`

type simpleFeature struct {
	Type        string             `db:"feature_type" json:"type"`
	Start       int                `db:"feature_fmin" json:"start"`
	End         int                `db:"feature_fmax" json:"end"`
	Strand      int                `db:"feature_strand" json:"strand"`
	Name        string             `db:"feature_name" json:"name"`
	UniqueID    string             `db:"feature_uniquename" json:"uniqueID"`
	Score       float64            ` json:"score"`
	Subfeatures []processedFeature ` json:"subfeatures"`
}

var simpleFeatQuery = `
SELECT
    cvterm.name AS feature_type,

    featureloc.fmin AS feature_fmin,
    featureloc.fmax AS feature_fmax,
    featureloc.strand AS feature_strand,

    feature.name AS feature_name,
    feature.uniquename AS feature_uniquename

FROM feature
    INNER JOIN
    featureloc ON (feature.feature_id = featureloc.feature_id)
    INNER JOIN
    cvterm ON (feature.type_id = cvterm.cvterm_id)
WHERE
    (feature.organism_id = (select organism_id from organism where common_name = $1))
	AND
    (cvterm.name = $3)
    AND
    (featureloc.srcfeature_id = (select feature_id from feature where name = $2))
	AND
	(featureloc.fmin <= $5 AND $4 <= featureloc.fmax)
;
`

//AND
//(featureloc.srcfeature_id = (select feature_id from feature where name = $2))
//AND
//(featureloc.fmin <= $5 AND $4 <= featureloc.fmax)

type processedFeature struct {
	Type        string             `json:"type"`
	Start       int                `json:"start"`
	End         int                `json:"end"`
	Score       float64            `json:"score"`
	Strand      int                `json:"strand"`
	Name        string             `json:"name"`
	UniqueID    string             `json:"uniqueID"`
	Subfeatures []processedFeature `json:"subfeatures"`
}


type refSeqStruct struct {
	Length       int    `db:"seqlen" json:"length"`
	Name         string `db:"name" json:"name"`
	Start        int    `json:"start"`
	End          int    `json:"end"`
	SeqChunkSize int    `json:"seqChunkSize"`
}

var refSeqQuery = `
SELECT
    seqlen, name
FROM
    feature
WHERE
    feature.organism_id = (select organism_id from organism where common_name = $1)
    AND
    type_id = 455
;
`

type refSeqWithSeqStruct struct {
	Start int    `json:"start"`
	Seq   string `db:"seq" json:"seq"`
	End   int    `json:"end"`
}

var refSeqWithSeqQuery = `
SELECT
    substring(residues, $3::int, $4::int) as seq
FROM
    feature
WHERE
    name = $2
    AND
    feature.organism_id = (select organism_id from organism where common_name = $1)
    AND
    residues is not NULL
;
`

type featureContainerRefSeqWithStruct struct {
	Features []refSeqWithSeqStruct `json:"features"`
}
type featureContainerFeatures struct {
	Features []simpleFeature `json:"features"`
}
type trackList struct {
	RefSeqs string        `json:"refSeqs"`
	Names   nameStruct    `json:"names"`
	Tracks  []interface{} `json:"tracks"`
}

type nameStruct struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type trackListTrack struct {
	Category               string            `json:"category"`
	Label                  string            `json:"label"`
	Type                   string            `json:"type"`
	TrackType              string            `json:"trackType"`
	Key                    string            `json:"key"`
	Query                  map[string]string `json:"query"`
	RegionFeatureDensities bool              `json:"regionFeatureDensities"`
	StoreClass             string            `json:"storeClass"`
}

type seqTrack struct {
	UseAsRefSeqStore bool              `json:"useAsRefSeqStore"`
	Label            string            `json:"label"`
	Key              string            `json:"key"`
	Type             string            `json:"type"`
	StoreClass       string            `json:"storeClass"`
	BaseURL          string            `json:"baseUrl"`
	Query            map[string]string `json:"query"`
}

package main

import "database/sql"

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
    FeatureID   int                `db:"xfeature_id" json:"-"`
	Type        string             `db:"feature_type" json:"type"`
	Start       int                `db:"feature_fmin" json:"start"`
	End         int                `db:"feature_fmax" json:"end"`
	Strand      int                `db:"feature_strand" json:"strand"`
	Name        string             `json:"name"`
	UniqueID    string             `db:"feature_uniquename" json:"uniqueID"`
    Score       float64            `json:"score"`
    ParentID    int                `db:"parent_id" json:"-"`
    Subfeatures []simpleFeature    `json:"subfeatures"`
    NullName    sql.NullString     `db:"feature_name"`
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
    (featureloc.srcfeature_id = (select feature_id from feature where uniquename = $2))
	AND
	(featureloc.fmin <= $5 AND $4 <= featureloc.fmax)
;
`

var simpleFeatQueryWithParent = `
WITH RECURSIVE feature_tree(xfeature_id, feature_type, feature_fmin, feature_fmax, feature_strand, feature_name, feature_uniquename, object_id, parent_id)
AS (
    SELECT
        feature.feature_id as xfeature_id,
        cvterm.name AS feature_type,
        featureloc.fmin AS feature_fmin,
        featureloc.fmax AS feature_fmax,
        featureloc.strand AS feature_strand,
        feature.name AS feature_name,
        feature.uniquename AS feature_uniquename,
        feature.feature_id as object_id,
        feature.feature_id as parent_id
    FROM feature
        LEFT JOIN
        featureloc ON (feature.feature_id = featureloc.feature_id)
        LEFT JOIN
        cvterm ON (feature.type_id = cvterm.cvterm_id)
    WHERE
        (feature.organism_id IN (select organism_id from organism where common_name = $1))
        AND
        -- with queried seqid
        (featureloc.srcfeature_id IN (SELECT feature_id FROM feature WHERE uniquename = $2))
        AND
        -- within queried region
        (featureloc.fmin <= $5 AND $4 <= featureloc.fmax)
        -- top level only
        AND cvterm.name = $3
UNION ALL
    SELECT
        feature.feature_id as xfeature_id,
        cvterm.name AS feature_type,
        featureloc.fmin AS feature_fmin,
        featureloc.fmax AS feature_fmax,
        featureloc.strand AS feature_strand,
        feature.name AS feature_name,
        feature.uniquename AS feature_uniquename,
        feature.feature_id as object_id,
        feature_relationship.object_id as parent_id
    FROM feature_relationship
        LEFT JOIN
        feature ON (feature.feature_id = feature_relationship.subject_id
                    AND feature_relationship.type_id IN (SELECT cvterm_id FROM cvterm WHERE name = 'part_of'))
        LEFT JOIN
        featureloc ON (feature.feature_id = featureloc.feature_id)
        LEFT JOIN
        cvterm ON (feature.type_id = cvterm.cvterm_id)
        JOIN
        cvterm reltype ON (reltype.cvterm_id = feature_relationship.type_id),
        feature_tree
    WHERE
        feature_relationship.object_id = feature_tree.object_id
        AND feature_relationship.type_id IN (SELECT cvterm_id FROM cvterm WHERE name = 'part_of')
)
SELECT xfeature_id, feature_type, feature_fmin, feature_fmax, feature_strand, feature_name, feature_uniquename, parent_id FROM feature_tree;
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
    seqlen, uniquename as name
FROM
    feature
WHERE
    organism_id IN (select organism_id from organism where common_name = $1)
    AND
    type_id IN (select cvterm_id from cvterm join cv using (cv_id)
                where cvterm.name = 'chromosome' and cv.name = 'sequence')
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
    uniquename = $2
    AND
    feature.organism_id IN (select organism_id from organism where common_name = $1)
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

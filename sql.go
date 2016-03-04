package main

type Organism struct {
	Genus      string `db:"genus"`
	Species    string `db:"species"`
	CommonName string `db:"common_name"`
}

var OrganismQuery = `
SELECT
	genus, species, common_name
FROM
	organism
`

type SoType struct {
	Type string `db:"type"`
}

var SoTypeQuery = `
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

type SimpleFeature struct {
	Type        string             `db:"feature_type" json:"type"`
	Start       int                `db:"feature_fmin" json:"start"`
	End         int                `db:"feature_fmax" json:"end"`
	Strand      int                `db:"feature_strand" json:"strand"`
	Name        string             `db:"feature_name" json:"name"`
	UniqueID    string             `db:"feature_uniquename" json:"uniqueID"`
	Score       float64            ` json:"score"`
	Subfeatures []ProcessedFeature ` json:"subfeatures"`
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

type TripFeature struct {
	GeneType   string `db:"gene_type"`
	GeneName   string `db:"gene_name"`
	GeneFmin   string `db:"gene_fmin"`
	GeneFmax   string `db:"gene_fmax"`
	GeneStrand string `db:"gene_strand"`

	Subfeat1Type   string `db:"subfeat1_type"`
	Subfeat1Name   string `db:"subfeat1_name"`
	Subfeat1Fmin   string `db:"subfeat1_fmin"`
	Subfeat1Fmax   string `db:"subfeat1_fmax"`
	Subfeat1Strand string `db:"subfeat1_strand"`

	Subfeat2Type   string `db:"subfeat2_type"`
	Subfeat2Name   string `db:"subfeat2_name"`
	Subfeat2Fmin   string `db:"subfeat2_fmin"`
	Subfeat2Fmax   string `db:"subfeat2_fmax"`
	Subfeat2Strand string `db:"subfeat2_strand"`
}

type ProcessedFeature struct {
	Type        string             `json:"type"`
	Start       int                `json:"start"`
	End         int                `json:"end"`
	Score       float64            `json:"score"`
	Strand      int                `json:"strand"`
	Name        string             `json:"name"`
	UniqueID    string             `json:"uniqueID"`
	Subfeatures []ProcessedFeature `json:"subfeatures"`
}

var featureListQuery = `
SELECT
    genecv.name AS gene_type,
    gene.name AS gene_name,
    geneloc.fmin AS gene_fmin,
    geneloc.fmax AS gene_fmax,
    geneloc.strand AS gene_strand,

    subfeat1cv.name AS subfeat1_type,
    subfeat1.name AS subfeat1_name,
    subfeat1loc.fmin AS subfeat1_fmin,
    subfeat1loc.fmax AS subfeat1_fmax,
    subfeat1loc.strand AS subfeat1_strand,

    subfeat2cv.name AS subfeat2_type,
    subfeat2.name AS subfeat2_name,
    subfeat2loc.fmin AS subfeat2_fmin,
    subfeat2loc.fmax AS subfeat2_fmax,
    subfeat2loc.strand AS subfeat2_strand

FROM feature AS gene
    INNER JOIN
    featureloc AS geneloc ON (gene.feature_id = geneloc.feature_id)
    INNER JOIN
    cvterm AS genecv ON (gene.type_id = genecv.cvterm_id)
    INNER JOIN
    featureloc AS gloc ON (gene.feature_id = gloc.featureloc_id)

    INNER JOIN
    feature_relationship AS feat0 ON (gene.feature_id = feat0.object_id)
    INNER JOIN
    feature AS subfeat1 ON (subfeat1.feature_id = feat0.subject_id)
    INNER JOIN
    featureloc AS subfeat1loc ON (subfeat1.feature_id = subfeat1loc.feature_id)
    INNER JOIN
    cvterm AS subfeat1cv ON (subfeat1cv.cvterm_id = subfeat1.type_id)

    INNER JOIN
    feature_relationship AS feat1 ON (subfeat1.feature_id = feat1.object_id)
    INNER JOIN
    feature AS subfeat2 ON (subfeat2.feature_id = feat1.subject_id)
    INNER JOIN
    featureloc AS subfeat2loc ON (subfeat2.feature_id = subfeat2loc.feature_id)
    INNER JOIN
    cvterm AS subfeat2cv ON (subfeat2cv.cvterm_id = subfeat2.type_id)
WHERE
    (gene.organism_id = (select organism_id from organism where common_name = $1))
    AND
    (gloc.srcfeature_id = (select feature_id from feature where name = $2))
    AND
    (geneloc.fmin <= $1 and $2 <= geneloc.fmax)
;
`

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

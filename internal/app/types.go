package app

type Example struct {
	ID     string `json:"id"`
	Input  string `json:"input"`
	Output string `json:"output"`
	Label  string `json:"label"`
}

type SimilarPair struct {
	LeftID     string  `json:"left_id"`
	RightID    string  `json:"right_id"`
	Similarity float64 `json:"similarity"`
	LeftInput  string  `json:"left_input,omitempty"`
	RightInput string  `json:"right_input,omitempty"`
}

type LengthStats struct {
	AvgInputTokens  float64 `json:"avg_input_tokens"`
	P95InputTokens  int     `json:"p95_input_tokens"`
	AvgOutputTokens float64 `json:"avg_output_tokens"`
	P95OutputTokens int     `json:"p95_output_tokens"`
}

type SplitReport struct {
	Name                 string         `json:"name"`
	Rows                 int            `json:"rows"`
	MissingID            int            `json:"missing_id"`
	EmptyInput           int            `json:"empty_input"`
	EmptyOutput          int            `json:"empty_output"`
	DuplicateInputs      int            `json:"duplicate_inputs"`
	DuplicateSamples     []string       `json:"duplicate_samples,omitempty"`
	NearDuplicatePairs   int            `json:"near_duplicate_pairs"`
	NearDuplicateSamples []SimilarPair  `json:"near_duplicate_samples,omitempty"`
	LabelConflicts       int            `json:"label_conflicts"`
	ConflictSamples      []string       `json:"conflict_samples,omitempty"`
	LabelCounts          map[string]int `json:"label_counts"`
	LengthStats          LengthStats    `json:"length_stats"`
}

type Report struct {
	Train                  SplitReport   `json:"train"`
	Eval                   SplitReport   `json:"eval"`
	OverlapCount           int           `json:"overlap_count"`
	OverlapExamples        []string      `json:"overlap_examples,omitempty"`
	SemanticOverlapCount   int           `json:"semantic_overlap_count"`
	SemanticOverlapSamples []SimilarPair `json:"semantic_overlap_samples,omitempty"`
}

type indexedRow struct {
	example Example
	norm    string
	tokens  map[string]struct{}
}

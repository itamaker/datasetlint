package app

import "testing"

func TestScanDatasetsDetectsOverlap(t *testing.T) {
	t.Parallel()

	train := []Example{
		{ID: "1", Input: "Summarize the paper", Output: "A short summary", Label: "summary"},
		{ID: "2", Input: "Summarize the paper", Output: "Another summary", Label: "summary"},
	}
	eval := []Example{
		{ID: "3", Input: "Summarize the paper", Output: "Held out summary", Label: "summary"},
	}

	report := scanDatasets(train, eval)
	if report.Train.DuplicateInputs != 1 {
		t.Fatalf("Train.DuplicateInputs = %d, want 1", report.Train.DuplicateInputs)
	}
	if report.OverlapCount != 1 {
		t.Fatalf("OverlapCount = %d, want 1", report.OverlapCount)
	}
}

func TestScanDatasetsDetectsSemanticPairsAndConflicts(t *testing.T) {
	t.Parallel()

	train := []Example{
		{ID: "1", Input: "Summarize the retrieval benchmark setup", Output: "A short summary", Label: "summary"},
		{ID: "2", Input: "Summarize retrieval benchmark configuration", Output: "A second summary", Label: "analysis"},
	}
	eval := []Example{
		{ID: "3", Input: "Summarize the benchmark setup for retrieval", Output: "Held out summary", Label: "summary"},
	}

	report := scanDatasetsWithOptions(train, eval, 0.4, 5)
	if report.Train.LabelConflicts != 1 {
		t.Fatalf("Train.LabelConflicts = %d, want 1", report.Train.LabelConflicts)
	}
	if report.SemanticOverlapCount == 0 {
		t.Fatalf("SemanticOverlapCount = %d, want > 0", report.SemanticOverlapCount)
	}
	if len(report.SemanticOverlapSamples) == 0 {
		t.Fatalf("expected semantic overlap samples")
	}
}

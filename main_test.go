package main

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

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

type Example struct {
	ID     string `json:"id"`
	Input  string `json:"input"`
	Output string `json:"output"`
	Label  string `json:"label"`
}

type SplitReport struct {
	Name             string         `json:"name"`
	Rows             int            `json:"rows"`
	MissingID        int            `json:"missing_id"`
	EmptyInput       int            `json:"empty_input"`
	EmptyOutput      int            `json:"empty_output"`
	DuplicateInputs  int            `json:"duplicate_inputs"`
	DuplicateSamples []string       `json:"duplicate_samples,omitempty"`
	LabelCounts      map[string]int `json:"label_counts"`
}

type Report struct {
	Train           SplitReport `json:"train"`
	Eval            SplitReport `json:"eval"`
	OverlapCount    int         `json:"overlap_count"`
	OverlapExamples []string    `json:"overlap_examples,omitempty"`
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}

	switch args[0] {
	case "scan":
		return runScan(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n\n", args[0])
		usage()
		return 2
	}
}

func usage() {
	fmt.Println("datasetlint scans JSONL datasets for basic quality issues.")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  datasetlint scan -train examples/train.jsonl -eval examples/eval.jsonl")
}

func runScan(args []string) int {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	trainPath := fs.String("train", "", "path to a train JSONL file")
	evalPath := fs.String("eval", "", "path to an eval JSONL file")
	strict := fs.Bool("strict", false, "exit with code 1 if any issue is found")
	jsonOutput := fs.Bool("json", false, "emit machine-readable JSON")
	fs.SetOutput(os.Stderr)

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *trainPath == "" || *evalPath == "" {
		fmt.Fprintln(os.Stderr, "both -train and -eval are required")
		return 2
	}

	train, err := loadDataset(*trainPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	eval, err := loadDataset(*evalPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	report := scanDatasets(train, eval)

	if *jsonOutput {
		body, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		fmt.Println(string(body))
		if *strict && hasIssues(report) {
			return 1
		}
		return 0
	}

	printReport(report)
	if *strict && hasIssues(report) {
		return 1
	}
	return 0
}

func loadDataset(path string) ([]Example, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open dataset: %w", err)
	}
	defer file.Close()

	var examples []Example
	scanner := bufio.NewScanner(file)
	line := 0
	for scanner.Scan() {
		line++
		if scanner.Text() == "" {
			continue
		}
		var example Example
		if err := json.Unmarshal(scanner.Bytes(), &example); err != nil {
			return nil, fmt.Errorf("decode line %d: %w", line, err)
		}
		examples = append(examples, example)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan dataset: %w", err)
	}
	return examples, nil
}

func scanDatasets(train []Example, eval []Example) Report {
	trainReport, trainInputs := analyzeSplit("train", train)
	evalReport, evalInputs := analyzeSplit("eval", eval)

	var overlap []string
	for text := range trainInputs {
		if _, ok := evalInputs[text]; ok {
			overlap = append(overlap, text)
		}
	}
	sort.Strings(overlap)
	if len(overlap) > 5 {
		overlap = overlap[:5]
	}

	return Report{
		Train:           trainReport,
		Eval:            evalReport,
		OverlapCount:    len(overlapInputs(trainInputs, evalInputs)),
		OverlapExamples: overlap,
	}
}

func analyzeSplit(name string, rows []Example) (SplitReport, map[string]struct{}) {
	report := SplitReport{
		Name:        name,
		Rows:        len(rows),
		LabelCounts: map[string]int{},
	}

	seen := map[string]int{}
	inputs := map[string]struct{}{}
	for _, row := range rows {
		if strings.TrimSpace(row.ID) == "" {
			report.MissingID++
		}
		if strings.TrimSpace(row.Input) == "" {
			report.EmptyInput++
		}
		if strings.TrimSpace(row.Output) == "" {
			report.EmptyOutput++
		}
		if label := strings.TrimSpace(row.Label); label != "" {
			report.LabelCounts[label]++
		}

		norm := normalize(row.Input)
		if norm == "" {
			continue
		}
		inputs[norm] = struct{}{}
		seen[norm]++
	}

	for text, count := range seen {
		if count > 1 {
			report.DuplicateInputs++
			report.DuplicateSamples = append(report.DuplicateSamples, text)
		}
	}
	sort.Strings(report.DuplicateSamples)
	if len(report.DuplicateSamples) > 5 {
		report.DuplicateSamples = report.DuplicateSamples[:5]
	}

	return report, inputs
}

func normalize(input string) string {
	parts := strings.Fields(strings.ToLower(strings.TrimSpace(input)))
	return strings.Join(parts, " ")
}

func overlapInputs(train map[string]struct{}, eval map[string]struct{}) []string {
	var overlap []string
	for text := range train {
		if _, ok := eval[text]; ok {
			overlap = append(overlap, text)
		}
	}
	sort.Strings(overlap)
	return overlap
}

func hasIssues(report Report) bool {
	return report.Train.MissingID > 0 ||
		report.Train.EmptyInput > 0 ||
		report.Train.EmptyOutput > 0 ||
		report.Train.DuplicateInputs > 0 ||
		report.Eval.MissingID > 0 ||
		report.Eval.EmptyInput > 0 ||
		report.Eval.EmptyOutput > 0 ||
		report.Eval.DuplicateInputs > 0 ||
		report.OverlapCount > 0
}

func printReport(report Report) {
	fmt.Printf("Train rows: %d\n", report.Train.Rows)
	fmt.Printf("Eval rows: %d\n", report.Eval.Rows)
	fmt.Printf("Train issues: missing id=%d empty input=%d empty output=%d duplicate inputs=%d\n",
		report.Train.MissingID, report.Train.EmptyInput, report.Train.EmptyOutput, report.Train.DuplicateInputs)
	fmt.Printf("Eval issues: missing id=%d empty input=%d empty output=%d duplicate inputs=%d\n",
		report.Eval.MissingID, report.Eval.EmptyInput, report.Eval.EmptyOutput, report.Eval.DuplicateInputs)
	fmt.Printf("Cross-split overlap: %d\n", report.OverlapCount)
	if len(report.Train.LabelCounts) > 0 {
		fmt.Printf("Train labels: %v\n", report.Train.LabelCounts)
	}
	if len(report.Eval.LabelCounts) > 0 {
		fmt.Printf("Eval labels: %v\n", report.Eval.LabelCounts)
	}
	if len(report.OverlapExamples) > 0 {
		fmt.Printf("Overlap examples: %s\n", strings.Join(report.OverlapExamples, "; "))
	}
}

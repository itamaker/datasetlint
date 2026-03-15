package app

import (
	"sort"
	"strings"
)

func scanDatasets(train []Example, eval []Example) Report {
	return scanDatasetsWithOptions(train, eval, 0.6, 5)
}

func scanDatasetsWithOptions(train []Example, eval []Example, threshold float64, maxSamples int) Report {
	trainReport, trainRows := analyzeSplit("train", train, threshold, maxSamples)
	evalReport, evalRows := analyzeSplit("eval", eval, threshold, maxSamples)

	var overlap []string
	trainInputs := make(map[string]struct{}, len(trainRows))
	evalInputs := make(map[string]struct{}, len(evalRows))
	for _, row := range trainRows {
		trainInputs[row.norm] = struct{}{}
	}
	for _, row := range evalRows {
		evalInputs[row.norm] = struct{}{}
	}
	for text := range trainInputs {
		if _, ok := evalInputs[text]; ok {
			overlap = append(overlap, text)
		}
	}
	sort.Strings(overlap)
	if len(overlap) > maxSamples {
		overlap = overlap[:maxSamples]
	}

	semanticOverlapCount, semanticOverlapSamples := findCrossSplitNearDuplicates(trainRows, evalRows, threshold, maxSamples)

	return Report{
		Train:                  trainReport,
		Eval:                   evalReport,
		OverlapCount:           len(overlapInputs(trainInputs, evalInputs)),
		OverlapExamples:        overlap,
		SemanticOverlapCount:   semanticOverlapCount,
		SemanticOverlapSamples: semanticOverlapSamples,
	}
}

func analyzeSplit(name string, rows []Example, threshold float64, maxSamples int) (SplitReport, []indexedRow) {
	report := SplitReport{
		Name:        name,
		Rows:        len(rows),
		LabelCounts: map[string]int{},
	}

	seen := map[string]int{}
	inputLabels := map[string]string{}
	lengthInputs := make([]int, 0, len(rows))
	lengthOutputs := make([]int, 0, len(rows))
	indexed := make([]indexedRow, 0, len(rows))

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

		inputTokens := tokenize(row.Input)
		outputTokens := tokenize(row.Output)
		lengthInputs = append(lengthInputs, len(inputTokens))
		lengthOutputs = append(lengthOutputs, len(outputTokens))

		norm := normalize(row.Input)
		tokenSet := tokenSet(row.Input)
		indexed = append(indexed, indexedRow{
			example: row,
			norm:    norm,
			tokens:  tokenSet,
		})

		if norm == "" {
			continue
		}
		seen[norm]++
		if label := strings.TrimSpace(row.Label); label != "" {
			if existing, ok := inputLabels[norm]; ok && existing != label {
				report.LabelConflicts++
				report.ConflictSamples = append(report.ConflictSamples, norm)
			} else if !ok {
				inputLabels[norm] = label
			}
		}
	}

	for text, count := range seen {
		if count > 1 {
			report.DuplicateInputs++
			report.DuplicateSamples = append(report.DuplicateSamples, text)
		}
	}

	report.NearDuplicatePairs, report.NearDuplicateSamples = findNearDuplicates(indexed, threshold, maxSamples)
	semanticConflicts, semanticConflictSamples := findSemanticLabelConflicts(indexed, threshold, maxSamples)
	report.LabelConflicts += semanticConflicts
	report.ConflictSamples = append(report.ConflictSamples, semanticConflictSamples...)
	report.LengthStats = LengthStats{
		AvgInputTokens:  averageInts(lengthInputs),
		P95InputTokens:  percentile(lengthInputs, 0.95),
		AvgOutputTokens: averageInts(lengthOutputs),
		P95OutputTokens: percentile(lengthOutputs, 0.95),
	}

	sort.Strings(report.DuplicateSamples)
	if len(report.DuplicateSamples) > maxSamples {
		report.DuplicateSamples = report.DuplicateSamples[:maxSamples]
	}
	report.ConflictSamples = uniqueStrings(report.ConflictSamples)
	sort.Strings(report.ConflictSamples)
	if len(report.ConflictSamples) > maxSamples {
		report.ConflictSamples = report.ConflictSamples[:maxSamples]
	}

	return report, indexed
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
		report.Train.NearDuplicatePairs > 0 ||
		report.Train.LabelConflicts > 0 ||
		report.Eval.MissingID > 0 ||
		report.Eval.EmptyInput > 0 ||
		report.Eval.EmptyOutput > 0 ||
		report.Eval.DuplicateInputs > 0 ||
		report.Eval.NearDuplicatePairs > 0 ||
		report.Eval.LabelConflicts > 0 ||
		report.OverlapCount > 0 ||
		report.SemanticOverlapCount > 0
}

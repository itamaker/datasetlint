package app

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

func tokenize(text string) []string {
	normalized := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return ' '
	}, text)
	return strings.Fields(normalized)
}

func tokenSet(text string) map[string]struct{} {
	set := map[string]struct{}{}
	for _, token := range tokenize(text) {
		if len(token) < 3 {
			continue
		}
		set[token] = struct{}{}
	}
	return set
}

func findNearDuplicates(rows []indexedRow, threshold float64, maxSamples int) (int, []SimilarPair) {
	return compareIndexedRows(rows, rows, threshold, maxSamples, true)
}

func findCrossSplitNearDuplicates(left []indexedRow, right []indexedRow, threshold float64, maxSamples int) (int, []SimilarPair) {
	return compareIndexedRows(left, right, threshold, maxSamples, false)
}

func findSemanticLabelConflicts(rows []indexedRow, threshold float64, maxSamples int) (int, []string) {
	inverted := map[string][]int{}
	for idx, row := range rows {
		for token := range row.tokens {
			inverted[token] = append(inverted[token], idx)
		}
	}

	count := 0
	var samples []string
	seenPairs := map[string]struct{}{}
	for leftIdx, leftRow := range rows {
		leftLabel := strings.TrimSpace(leftRow.example.Label)
		if leftLabel == "" {
			continue
		}
		candidates := map[int]struct{}{}
		for token := range leftRow.tokens {
			for _, rightIdx := range inverted[token] {
				if rightIdx <= leftIdx {
					continue
				}
				candidates[rightIdx] = struct{}{}
			}
		}

		for rightIdx := range candidates {
			rightRow := rows[rightIdx]
			rightLabel := strings.TrimSpace(rightRow.example.Label)
			if rightLabel == "" || leftLabel == rightLabel || leftRow.norm == rightRow.norm {
				continue
			}
			similarity := jaccard(leftRow.tokens, rightRow.tokens)
			if similarity < threshold {
				continue
			}
			key := pairKey(leftRow.example.ID, rightRow.example.ID, leftIdx, rightIdx, true)
			if _, ok := seenPairs[key]; ok {
				continue
			}
			seenPairs[key] = struct{}{}
			count++
			if len(samples) < maxSamples {
				samples = append(samples, fmt.Sprintf("%s <> %s", leftRow.norm, rightRow.norm))
			}
		}
	}
	return count, samples
}

func compareIndexedRows(left []indexedRow, right []indexedRow, threshold float64, maxSamples int, sameSplit bool) (int, []SimilarPair) {
	inverted := map[string][]int{}
	target := right
	if sameSplit {
		target = left
	}
	for idx, row := range target {
		for token := range row.tokens {
			inverted[token] = append(inverted[token], idx)
		}
	}

	count := 0
	var samples []SimilarPair
	seenPairs := map[string]struct{}{}

	for leftIdx, leftRow := range left {
		candidates := map[int]struct{}{}
		for token := range leftRow.tokens {
			for _, rightIdx := range inverted[token] {
				if sameSplit && rightIdx <= leftIdx {
					continue
				}
				candidates[rightIdx] = struct{}{}
			}
		}

		for rightIdx := range candidates {
			rightRow := right[rightIdx]
			if leftRow.norm == "" || rightRow.norm == "" {
				continue
			}
			if leftRow.norm == rightRow.norm {
				continue
			}
			similarity := jaccard(leftRow.tokens, rightRow.tokens)
			if similarity < threshold {
				continue
			}
			key := pairKey(leftRow.example.ID, rightRow.example.ID, leftIdx, rightIdx, sameSplit)
			if _, ok := seenPairs[key]; ok {
				continue
			}
			seenPairs[key] = struct{}{}
			count++
			if len(samples) < maxSamples {
				samples = append(samples, SimilarPair{
					LeftID:     fallbackID(leftRow.example.ID, leftIdx),
					RightID:    fallbackID(rightRow.example.ID, rightIdx),
					Similarity: similarity,
					LeftInput:  leftRow.example.Input,
					RightInput: rightRow.example.Input,
				})
			}
		}
	}

	sort.Slice(samples, func(i, j int) bool {
		if samples[i].Similarity == samples[j].Similarity {
			if samples[i].LeftID == samples[j].LeftID {
				return samples[i].RightID < samples[j].RightID
			}
			return samples[i].LeftID < samples[j].LeftID
		}
		return samples[i].Similarity > samples[j].Similarity
	})
	return count, samples
}

func jaccard(left map[string]struct{}, right map[string]struct{}) float64 {
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	intersection := 0
	union := len(left)
	for token := range right {
		if _, ok := left[token]; ok {
			intersection++
			continue
		}
		union++
	}
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

func pairKey(leftID string, rightID string, leftIdx int, rightIdx int, sameSplit bool) string {
	if sameSplit {
		return fmt.Sprintf("%d:%d", leftIdx, rightIdx)
	}
	return fallbackID(leftID, leftIdx) + "->" + fallbackID(rightID, rightIdx)
}

func fallbackID(id string, index int) string {
	if strings.TrimSpace(id) != "" {
		return id
	}
	return fmt.Sprintf("row-%d", index+1)
}

func averageInts(values []int) float64 {
	if len(values) == 0 {
		return 0
	}
	total := 0
	for _, value := range values {
		total += value
	}
	return float64(total) / float64(len(values))
}

func percentile(values []int, p float64) int {
	if len(values) == 0 {
		return 0
	}
	cloned := append([]int(nil), values...)
	sort.Ints(cloned)
	index := int(p*float64(len(cloned)-1) + 0.5)
	if index < 0 {
		index = 0
	}
	if index >= len(cloned) {
		index = len(cloned) - 1
	}
	return cloned[index]
}

func uniqueStrings(items []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

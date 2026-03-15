package app

import (
	"fmt"
	"os"
)

func Run(args []string) int {
	if len(args) == 0 {
		return runTUI()
	}

	switch args[0] {
	case "scan":
		return runScan(args[1:])
	case "tui", "interactive":
		return runTUI()
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n\n", args[0])
		usage()
		return 2
	}
}

func usage() {
	fmt.Println("datasetlint scans JSONL datasets for quality, overlap, and semantic duplication issues.")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  datasetlint                # launch Bubble Tea TUI")
	fmt.Println("  datasetlint scan -train examples/train.jsonl -eval examples/eval.jsonl")
}

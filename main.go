package main

import (
	"os"

	"github.com/itamaker/datasetlint/internal/app"
)

func main() {
	os.Exit(app.Run(os.Args[1:]))
}

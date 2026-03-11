# datasetlint

`datasetlint` catches common dataset issues before they leak into training or evaluation.

## Example

```bash
go run . scan -train examples/train.jsonl -eval examples/eval.jsonl
```

## Checks

- missing IDs
- empty inputs and outputs
- duplicate normalized inputs within a split
- cross-split overlap between train and eval
- label-count summary

## Install

From source:

```bash
go install github.com/YOUR_GITHUB_USER/datasetlint@latest
```

From Homebrew after you publish a tap formula:

```bash
brew tap itamaker/tap https://github.com/itamaker/homebrew-tap
brew install itamaker/tap/datasetlint
```

## Repo-Ready Files

- `.github/workflows/ci.yml`
- `.github/workflows/release.yml`
- `.goreleaser.yaml`
- `PUBLISHING.md`
- `scripts/render-homebrew-formula.sh`

## Release

```bash
git tag v0.1.0
git push origin v0.1.0
```

The tagged release workflow publishes multi-platform binaries and `checksums.txt`, which you can feed into the Homebrew formula renderer.
The generated formula should be committed to `https://github.com/itamaker/homebrew-tap`.

.PHONY: run docs

run:
	go run main.go

# Regenerate the API Reference section of README.md from the annotations in main.go.
docs:
	go run ./tools/gendocs

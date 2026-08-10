# MOVIES API

## DEV NOTE: Check if sqlite3 driver works normally
`main.go` contains examples of how to use SQLite with go (from https://github.com/mattn/go-sqlite3/blob/v1.14.49/_example/simple/simple.go). Run it with `go run ./...`

[brief description]

## Project overview

...

## Setup and installation instructions

...

### Git hooks

This project ships a `pre-commit` hook (in `.githooks/`) that runs `gofmt` and `go vet` on staged Go files. Install it once after cloning:

    make init

To bypass the hook for a single commit:

    git commit -m '<your_message>' --no-verify

## Usage guide

...

## Additional features

...

## Bonus functionality

...

# MOVIES API

[TODO: brief description]

## Project overview

...

## Setup and installation instructions

...

### Git hooks

This project ships a `pre-commit` hook, located in `.githooks/`, that runs `gofmt` and `go vet` on staged Go files. Install it once after cloning the repo:

    make hook

To bypass the hook for a single commit:

    git commit -m '<your_message>' --no-verify

## Usage guide

Start the server on `localhost:8080`:

    make run

The database file `movies.db` is created in the project root directory.

### Choosing the port

Pass `-port` to use a custom port of your choise:

    go run ./... -port 8765

The default is `8080`. The value must be between 1025 and 65535, since ports below 1024 need root on Linux and reserved for system use.

### Loading the example data

To clean the database tables and fill them with example data before the server starts, run

    make init

This inserts 40 actors, 12 genres and 21 movies from `db/seed.sql`.

> **Warning:** `-init`/`-i` first deletes every row in `Actor`, `Genre`, `Movie`, `movie_genres`, `movie_actors` and `sqlite_sequence`. Any data in those tables will be lost. Run server without this flag to keep your data.

## Additional features

...

## Bonus functionality

...

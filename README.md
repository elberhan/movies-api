# MOVIES API

A RESTful API for managing a movie database, built with **Go**, raw **SQL**, and **SQLite**. It lets a film society digitize its records and manage **movies**, their **genres**, and the **actors** who star in them — including the many-to-many relationships between them.

## Table of contents

- [Features](#features)
- [Tech stack](#tech-stack)
- [Project structure](#project-structure)
- [Getting started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Running the server](#running-the-server)
  - [Seeding example data](#seeding-example-data)
  - [Choosing the port](#choosing-the-port)
- [API reference](#api-reference)
  - [Genres](#genres)
  - [Movies](#movies)
  - [Actors](#actors)
  - [Common conventions](#common-conventions)
- [Error handling](#error-handling)
- [Database schema](#database-schema)
- [Architecture](#architecture)
- [Testing](#testing)
- [Git hooks](#git-hooks)

---

## Features

- Full **CRUD** (create, read, update, delete) for **genres**, **movies**, and **actors**
- **Many-to-many relationships**: a movie can have many genres and many actors; a genre or actor can belong to many movies
- **Filtering**: movies by genre, release year, or actor
- **Search**: case-insensitive, partial-match search of movies by title
- **Pagination** on every list endpoint via `page` and `size` query parameters
- **Partial updates** with `PATCH` (only the fields provided in the body are changed)
- **Safe & forced deletion**: deleting a resource that has relationships is blocked by default, or allowed with `?force=true`
- **Input validation** at the service layer (required fields, length limits, date formats, year/duration bounds)
- **Structured error handling** with consistent JSON error responses and proper HTTP status codes
- **Immutable IDs** enforced at the database level by triggers

---

## Tech stack

| Component | Choice |
|---|---|
| Language | Go (`go 1.26.5`) |
| HTTP router | Standard library `net/http` + `http.ServeMux` (Go 1.22+ method patterns) |
| Database | SQLite (embedded, file-based) |
| Driver | `github.com/mattn/go-sqlite3` |
| Data access | Raw SQL via the standard `database/sql` package |

---

## Project structure

```
movies-api/
├── cmd/api/            entry point: parses flags, opens DB, starts the HTTP server
│   └── main.go
├── db/                 database layer
│   ├── db.go           opens the connection, runs the schema, seeds the DB
│   ├── schema.sql      CREATE TABLE / INDEX / TRIGGER statements
│   ├── seed.sql        example data (INSERT statements)
│   └── reset.sql       clears tables + resets auto-increment counters
├── models/             shared structs & sentinel errors passed between layers
│   ├── genre.go        Genre
│   ├── movie.go        Movie, MovieDetail, MovieFilter
│   ├── actor.go        Actor
│   ├── page.go         Page (pagination response)
│   └── errors.go       ErrNotFound, ErrInvalidInput, ErrHasRelations
├── repository/         data-access layer (raw SQL only)
│   ├── genre.go
│   ├── movie.go
│   ├── actor.go
│   └── actor_test.go
├── service/            business-logic layer (validation, rules, orchestration)
│   ├── genre.go
│   ├── movie.go
│   └── actor.go
├── handler/            HTTP layer (decode, dispatch, respond)
│   ├── routes.go       central wiring of all routes
│   ├── genre.go
│   ├── movie.go
│   ├── actor.go
│   ├── json.go         JSON encode/decode helpers
│   ├── errors.go       maps errors → HTTP status codes
│   ├── params.go       URL/path parameter helpers
│   └── actor_test.go   handler-layer tests
├── postman/            Postman collection (if present)
├── Makefile            common dev commands
└── go.mod / go.sum
```

**Layer flow for every request:**

```
HTTP request
  → handler   (decodes body, parses params, sends response)
  → service   (validates input, applies business rules)
  → repository (runs raw SQL against the database)
  → SQLite
```

---

## Getting started

### Prerequisites

- **Go** 1.26 or newer installed on your machine
- (Optional) **Postman** or any HTTP client (e.g. `curl`) to exercise the API

### Installation

```bash
git clone git@github.com:Eyob-L/movies-api.git
cd movies-api
```

Install the pre-commit hook once (recommended — see [Git hooks](#git-hooks)):

```bash
make hook
```

### Running the server

```bash
make run
# or
go run ./...
```

On startup the server:

1. Opens (or creates) the `movies.db` file in the project root
2. Enables SQLite foreign keys
3. Runs `db/schema.sql` to create all tables, indexes, and triggers
4. Starts listening

You should see:

```
server listening on http://localhost:8080
```

The database file `movies.db` is created in the project root directory.

### Seeding example data

To wipe the tables and load the bundled example data (40 actors, 12 genres, 21 movies with relationships), start the server with the init flag:

```bash
make init
# or
go run ./... -init
# or with the shorthand
go run ./... -i
```

> **Warning:** `-init` / `-i` first **deletes every row** in `Actor`, `Genre`, `Movie`, `movie_genres`, `movie_actors`, and resets the auto-increment counters (`sqlite_sequence`). Any data you have added will be lost. Run the server **without** this flag to keep your data.

The reset and insert happen inside a single transaction, so if anything fails, the database is left exactly as it was.

### Choosing the port

The server defaults to port `8080`. Choose a custom port with `-port`:

```bash
go run ./... -port 8765
```

The port must be between `1025` and `65535`. Ports below 1024 require root on Linux and are reserved for system services.

---

## API reference

All endpoints live under the `/api` prefix and exchange JSON. List endpoints return a **pagination envelope** (see [Common conventions](#common-conventions)).

The base URL is `http://localhost:8080`.

### Genres

#### `POST /api/genres` — Create a genre
Request body:

```json
{ "name": "Action" }
```

Response `201 Created`:

```json
{ "id": 1, "name": "Action" }
```

#### `GET /api/genres?page=0&size=10` — List genres
Returns a paginated list of all genres, ordered by name.

#### `GET /api/genres/{id}` — Get one genre
Response `200 OK`:

```json
{ "id": 1, "name": "Action" }
```

`404 Not Found` if the genre does not exist.

#### `PATCH /api/genres/{id}` — Update (rename) a genre
Request body (only `name` is updatable for a genre):

```json
{ "name": "Action & Adventure" }
```

#### `DELETE /api/genres/{id}` — Delete a genre
- Without `force` (default): blocked if the genre still has movies.
  `400` with e.g. `{"error":"invalid input: resource has associated records: cannot delete genre 'Sci-Fi' because it has 30 associated movies"}`.
- With `?force=true`: removes all movie links, then deletes the genre. `204 No Content`.

---

### Movies

Movies are created together with the IDs of their **genres** and **actors**.

#### `POST /api/movies` — Create a movie
Request body:

```json
{
  "title": "Inception",
  "releaseYear": 2010,
  "duration": 8880,
  "genres": [1, 9, 11],
  "actors": [5, 12, 8]
}
```

- `duration` is in **seconds**.
- At least one genre and one actor is required.
- Response `201 Created` returns the `MovieDetail` (movie plus its full genre and actor lists).

#### `GET /api/movies` — List / filter movies
Returns a paginated list of movies. Optional mutually-combinable filters (each narrows the result):

| Query param | Meaning |
|---|---|
| `genre` | only movies in the genre with this id |
| `year` | only movies released in this year |
| `actor` | only movies starring the actor with this id |
| `title` | partial, case-insensitive match on the title |

Examples:

```text
GET /api/movies?genre=1
GET /api/movies?year=1999
GET /api/movies?actor=39
GET /api/movies?genre=1&year=1999
```

> A filter that matches nothing returns an **empty page** (`200`), not a `404` — an empty result is not a missing resource.

#### `GET /api/movies/{id}` — Get one movie with its details
Response `200 OK` — a `MovieDetail`:

```json
{
  "id": 21,
  "title": "The Matrix",
  "releaseYear": 1999,
  "duration": 8160,
  "genres": [{ "id": 1, "name": "Action" }],
  "actors": [{ "id": 39, "name": "Keanu Reeves", "birthDate": "1964-09-02" }]
}
```

#### `GET /api/movies/{id}/actors` — Get a movie's cast
Paginated list of actors starring in the movie (`200`). `404` if the movie does not exist.

#### `GET /api/movies/{id}/genres` — Get a movie's genres
Paginated list of genres the movie belongs to (`200`). `404` if the movie does not exist.

#### `GET /api/movies/search?title=matrix` — Search movies by title
Partial, case-insensitive match. `title` is required; a missing/empty `title` returns `400`.

```text
GET /api/movies/search?title=matrix
```

#### `PATCH /api/movies/{id}` — Partially update a movie
Only the fields present in the body are changed. Provide `title`, `releaseYear`, and/or `duration` as scalar values, and `genres`/`actors` as arrays of IDs to replace the whole relationship set:

```json
{ "title": "Inception (2010)", "genres": [1, 9] }
```

`200 OK` returns the updated `MovieDetail`.

#### `DELETE /api/movies/{id}` — Delete a movie
- Without `force`: blocked if the movie has genres or actors. `400` with a descriptive message.
- With `?force=true`: drops the relationships first, then deletes the movie. `204 No Content`.

---

### Actors

#### `POST /api/actors` — Create an actor
Request body:

```json
{ "name": "Keanu Reeves", "birthDate": "1964-09-02" }
```

`birthDate` must be in **ISO 8601** format (`YYYY-MM-DD`). Response `201 Created`.

#### `GET /api/actors?name=&page=0&size=10` — List / filter actors
Paginated list of all actors. Optionally filter by partial, case-insensitive `name`:

```text
GET /api/actors
GET /api/actors?name=keanu
```

#### `GET /api/actors/{id}` — Get one actor
`200 OK`:

```json
{ "id": 39, "name": "Keanu Reeves", "birthDate": "1964-09-02" }
```

#### `PATCH /api/actors/{id}` — Partially update an actor
Only present fields are changed (`name` and/or `birthDate`):

```json
{ "birthDate": "1964-09-02" }
```

#### `DELETE /api/actors/{id}` — Delete an actor
- Without `force`: blocked if the actor is cast in any movie. `400` with a descriptive message.
- With `?force=true`: removes the movie links, then deletes the actor. `204 No Content`.

---

### Common conventions

**Pagination envelope** — every list endpoint returns this shape:

```json
{
  "content": [ ... ],
  "page": 0,
  "size": 10,
  "totalElements": 21,
  "totalPages": 3
}
```

- `page` (zero-based) defaults to `0`; `size` defaults to `10` and is capped at `100`.
- Invalid `page`/`size` values are rejected with `400`.

**HTTP status codes**

| Code | Meaning |
|---|---|
| `200 OK` | successful read / update |
| `201 Created` | successful create |
| `204 No Content` | successful delete (empty body) |
| `400 Bad Request` | malformed input or blocked deletion (has relationships) |
| `404 Not Found` | resource does not exist |
| `405 Method Not Allowed` | method not supported for the path (watch out for trailing slashes) |
| `500 Internal Server Error` | unexpected server error |

**Naming conventions** — used consistently across the codebase:

| Layer | Convention | Example |
|---|---|---|
| Go struct fields | PascalCase | `ReleaseYear`, `BirthDate` |
| JSON tags / query params | camelCase | `"releaseYear"`, `"birthDate"`, `?genreId=` |
| SQL entity tables | PascalCase singular | `Movie`, `Genre`, `Actor` |
| SQL join tables | snake_case plural | `movie_genres`, `movie_actors` |
| SQL columns | camelCase | `releaseYear`, `birthDate` |
| SQL foreign keys | snake_case | `movie_id`, `genre_id`, `actor_id` |

---

## Error handling

Errors use a small set of **sentinel values** defined in `models/errors.go`, raised and wrapped by the service/repository layers, and mapped to HTTP responses by `handler/errors.go`:

| Sentinel | HTTP status |
|---|---|
| `ErrNotFound` | `404 Not Found` |
| `ErrInvalidInput` | `400 Bad Request` |
| `ErrHasRelations` | `400 Bad Request` |
| anything else | `500 Internal Server Error` |

Every error response is JSON in the same shape, e.g.:

```json
{ "error": "resource not found" }
```

Unexpected internal errors are logged on the server but a **generic** message is sent to the client (implementation details are never exposed).

---

## Database schema

Defined in `db/schema.sql`. Highlights:

- **Entities** — `Actor`, `Movie`, `Genre`, all with `id INTEGER PRIMARY KEY AUTOINCREMENT` and `STRICT` typing.
- **Join tables** — `movie_genres` and `movie_actors` implement the many-to-many relationships, with composite primary keys and `ON DELETE RESTRICT` foreign keys (which is why deletions need `?force=true`).
- **Constraints** — non-empty names/titles (`length(trim(...)) > 0`), release year between `1888–2100`, positive durations, `birthDate` validated as a real date (`birthDate IS date(birthDate)`), unique genre names (`COLLATE NOCASE`).
- **Indexes** — for fast lookups by genre, actor, release year, actor name, and movie title.
- **Triggers** — make `id` columns **immutable** (updating an `id` aborts the transaction).

---

## Architecture

The application follows a strict **layered** separation of concerns:

- **`handler`** — HTTP concerns only: decoding request bodies, reading path/query parameters, building responses and status codes. It never touches the database.
- **`service`** — business logic: input validation, pagination normalization, safe/forced deletion rules, and orchestrating repository calls. It does not write SQL.
- **`repository`** — raw SQL data access via `database/sql`, using `?` placeholders to prevent SQL injection. It is the only layer that queries the database.
- **`models`** — the shared structs and sentinel errors passed between layers.

**Dependency injection** — each layer receives its dependency through a constructor (e.g. `NewGenreService(repo)`). The `handler/routes.go` `RegisterRoutes` function is the single wiring point that builds the chain for every entity:

```
db (from db.Open)
  → repository.NewXRepository(db)
  → service.NewXService(repo)
  → handler.NewXHandler(service)
  → xHandler.RegisterXRoutes(mux)
```

This keeps each layer independently testable and swappable (e.g. switching databases only touches the repository layer).

---

## Testing

Unit tests are included for the repository and handler layers (see `repository/actor_test.go` and `handler/actor_test.go`). Run all tests:

```bash
make test
# or
go test -v ./...
```

Build the whole project:

```bash
make build
# or
go build ./...
```

---

## Git hooks

The repository ships a `pre-commit` hook (in `.githooks/`) that runs `gofmt` and `go vet` on staged Go files. Install it once:

```bash
make hook
```

To bypass the hook for a single commit:

```bash
git commit -m '<your_message>' --no-verify
```

---

## License

See the [LICENSE](LICENSE) file in the repository root.

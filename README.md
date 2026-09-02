# MOVIES API

A RESTful API for managing a movie database — **movies**, their **genres**, and the **actors** who star in them — built with **Go**, raw **SQL**, and **SQLite**. Implements full CRUD for each entity plus the many-to-many relationships between them.

## Features

- CRUD for genres, movies, and actors
- Many-to-many: a movie has many genres/actors; a genre/actor belongs to many movies
- Filtering: `GET /api/movies?genre=&year=&actor=`
- Search: case-insensitive partial match on title
- Pagination via `page` & `size` on every list endpoint
- Partial updates with `PATCH`
- Safe deletion (blocked when has relationships) or forced with `?force=true`
- Input validation and consistent JSON error handling
- Immutable IDs enforced by DB triggers

## Setup and installation instructions

**Prerequisites:**

- Go 1.26+
- gcc compiler present within your PATH (for go-sqlite3)

**Installation:**

```bash
git clone git@gitea.kood.tech:nikolailetshev/movies-api.git
cd movies-api
make hook        # install pre-commit hook (once, for dev team)
make run         # or: go run ./...
```

The server starts on `http://localhost:8080` and creates `movies.db` in the project root.

**Load sample data** (40 actors, 12 genres, 21 movies): `make init` or `go run ./... -init`.

> Warning: `-init` wipes all existing rows before seeding!

**Custom port:** `go run ./... -port 8765` (must be 1025–65535).

## API reference (usage guide)

All endpoints use the `/api` prefix and JSON. Endpoints returning many items (for example GET /api/movies) are wrappen in a pagination:

```json
{ "content": [ ... ], "page": 0, "size": 10, "totalElements": 21, "totalPages": 3 }
```

`page` is zero-based (default `0`); `size` defaults to `10` (max `100`).

### Genres

| Method & path             | Description                                                                                                              | Response codes      |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------ | ------------------- |
| `POST /api/genres`        | Adds a new genre. Send `{ "name": "Action" }`.                                                                           | `201`, `400`        |
| `GET /api/genres`         | Gives one page of all genres.                                                                                            | `200`, `400`        |
| `GET /api/genres/{id}`    | Gives one genre by its id.                                                                                               | `200`, `404`        |
| `PATCH /api/genres/{id}`  | Changes the name of one genre.                                                                                           | `200`, `400`, `404` |
| `DELETE /api/genres/{id}` | Removes one genre. The request fails if the genre has movies. Add `?force=true` to remove the genre and its movie links. | `204`, `400`, `404` |

### Movies

Created together with genre and actor IDs (`duration` in seconds; minimum of 1 genre and 1 actor required):

```json
{
  "title": "Inception",
  "releaseYear": 2010,
  "duration": 8880,
  "genres": [1, 9, 11],
  "actors": [5, 12, 8]
}
```

| Method & path                         | Description                                                                                                                  | Response codes      |
| ------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- | ------------------- |
| `POST /api/movies`                    | Adds a new movie and links it to the given genres and actors. Gives back the movie with its full genre and actor lists.      | `201`, `400`        |
| `GET /api/movies`                     | Gives one page of movies. Use `genre`, `year`, `actor` and `title` to make the list smaller.                                 | `200`, `400`        |
| `GET /api/movies/{id}`                | Gives one movie with its full genre and actor lists.                                                                         | `200`, `404`        |
| `GET /api/movies/{id}/actors`         | Gives one page of the actors in that movie.                                                                                  | `200`, `400`, `404` |
| `GET /api/movies/{id}/genres`         | Gives one page of the genres of that movie.                                                                                  | `200`, `400`, `404` |
| `GET /api/movies/search?title=matrix` | Finds the movies whose title contains the text. Letter case is not important. The `title` parameter is necessary.            | `200`, `400`        |
| `PATCH /api/movies/{id}`              | Changes only the fields that you send. A `genres` or `actors` list replaces the old links.                                   | `200`, `400`, `404` |
| `DELETE /api/movies/{id}`             | Removes one movie. The request fails if the movie has genres or actors. Add `?force=true` to remove the movie and its links. | `204`, `400`, `404` |

### Actors

`birthDate` uses ISO 8601 (`YYYY-MM-DD`).

| Method & path             | Description                                                                                                                  | Response codes      |
| ------------------------- | ---------------------------------------------------------------------------------------------------------------------------- | ------------------- |
| `POST /api/actors`        | Adds a new actor. Send `{ "name": "Wang Yibo", "birthDate": "1997-08-05" }`.                                                 | `201`, `400`        |
| `GET /api/actors`         | Gives one page of all actors. Use `?name=` to find actors by a part of the name. Letter case is not important.               | `200`, `400`        |
| `GET /api/actors/{id}`    | Gives one actor by their id.                                                                                                 | `200`, `404`        |
| `PATCH /api/actors/{id}`  | Changes the `name` or the `birthDate` of one actor.                                                                          | `200`, `400`, `404` |
| `DELETE /api/actors/{id}` | Removes one actor. The request fails if the actor is in movies. Add `?force=true` to remove the actor and their movie links. | `204`, `400`, `404` |

Any endpoint can answer `500` if the server fails. A bad `page` or `size` value gives `400` on every list endpoint.

## Error handling

Errors use sentinel values in `models/errors.go`, mapped to HTTP statuses in `handler/errors.go`. Every error is a JSON object:

```json
{ "error": "resource not found" }
```

| Sentinel                              | HTTP                        |
| ------------------------------------- | --------------------------- |
| `ErrNotFound`                         | `404 Not Found`             |
| `ErrInvalidInput` / `ErrHasRelations` | `400 Bad Request`           |
| `ErrConflict`                         | `409 Conflict`              |
| anything else                         | `500 Internal Server Error` |

**Duplicate names:** genre names are unique (`name` is `UNIQUE COLLATE NOCASE` in `db/schema.sql`). Creating a genre whose name already exists — in any letter case — returns `409 Conflict` with a message such as `genre 'Action' already exists`, rather than a confusing `500`. The conflict is detected from the database's `UNIQUE` violation rather than a check-then-insert, so it stays correct even under concurrent requests.

Movie titles and actor names are intentionally **not** unique. Genres form a small controlled vocabulary where a duplicate is always a mistake, while real movie databases legitimately contain same-titled remakes and actors who happen to share a name. Identity is the `id` primary key, which is what every relationship uses; the title/name is display data, not identity.


## Architecture

Strict three-layer separation, wired together in `handler/routes.go`:

```
db → repository (raw SQL) → service (validation & rules) → handler (HTTP) → mux
```

- **handler**: HTTP only — decode body, read params, respond.
- **service**: validation, pagination, deletion rules, orchestration.
- **repository**: raw SQL via `database/sql` (`?` placeholders prevent injection).
- **models**: shared structs and sentinel errors.

## Testing

```bash
make test        # or: go test -v ./...
make build       # or: go build ./...
```

Unit tests cover the repository and handler layers (`repository/actor_test.go`, `handler/actor_test.go`).

# Bonus features

- Makefile used to simplify life of a developer. You can run multiple commands using plain English verbs (see Makefile)
- git hook used to format and vet staged go files (improves code quality and formatting, wouldn't allow you commit suspicious code)
- Implemented case-insensitive search functionality with partial match for movies endpoint
- Pagination implemented for every endpoint that returns multiple entities
- Implemented thorough error handling and input validation
- User can set custom port when launching the server: `go run ./... -port 8090`
- User can filter out movie listing with query parameters, which **order does not matter**: `http://localhost:8080/api/movies?genre=8&year=2020&page=0` and `http://localhost:8080/api/movies?page=0&year=2020&genre=8` returns the same thing.
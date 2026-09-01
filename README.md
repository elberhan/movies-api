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

**Prerequisites:** Go 1.26+.

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

| Method & path | Description |
|---|---|
| `POST /api/genres` | Create `{ "name": "Action" }` → `201` |
| `GET /api/genres` | List (paginated) → `200` |
| `GET /api/genres/{id}` | Get one → `200` / `404` |
| `PATCH /api/genres/{id}` | Rename → `200` |
| `DELETE /api/genres/{id}` | Delete → `204`. Blocked (`400`) if it has movies unless `?force=true` |

### Movies

Created together with genre and actor IDs (`duration` in seconds; minimum of 1 genre and 1 actor required):

```json
{ "title": "Inception", "releaseYear": 2010, "duration": 8880, "genres": [1, 9, 11], "actors": [5, 12, 8] }
```

| Method & path | Description |
|---|---|
| `POST /api/movies` | Create → `201` (returns movie + full genre/actor lists) |
| `GET /api/movies` | List/filter by `genre`, `year`, `actor`, `title` → `200` |
| `GET /api/movies/{id}` | Get one with its genres & actors → `200` / `404` |
| `GET /api/movies/{id}/actors` | Movie's cast (paginated) → `200` |
| `GET /api/movies/{id}/genres` | Movie's genres (paginated) → `200` |
| `GET /api/movies/search?title=matrix` | Partial, case-insensitive title search → `200` |
| `PATCH /api/movies/{id}` | Update fields; pass `genres`/`actors` arrays to replace links → `200` |
| `DELETE /api/movies/{id}` | Delete → `204`. Blocked (`400`) if linked unless `?force=true` |

### Actors

`birthDate` uses ISO 8601 (`YYYY-MM-DD`).

| Method & path | Description |
|---|---|
| `POST /api/actors` | Create `{ "name": "Wang Yibo", "birthDate": "1997-08-05" }` → `201` |
| `GET /api/actors` | List (paginated); filter by `?name=` (partial, case-insensitive) → `200` |
| `GET /api/actors/{id}` | Get one → `200` / `404` |
| `PATCH /api/actors/{id}` | Update `name` / `birthDate` → `200` |
| `DELETE /api/actors/{id}` | Delete → `204`. Blocked (`400`) if cast in movies unless `?force=true` |

## Error handling

Errors use sentinel values in `models/errors.go`, mapped to HTTP statuses in `handler/errors.go`. Every error is a JSON object:

```json
{ "error": "resource not found" }
```

| Sentinel | HTTP |
|---|---|
| `ErrNotFound` | `404 Not Found` |
| `ErrInvalidInput` / `ErrHasRelations` | `400 Bad Request` |
| anything else | `500 Internal Server Error` |

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

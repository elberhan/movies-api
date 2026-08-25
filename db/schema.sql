-- actors table
CREATE TABLE IF NOT EXISTS Actor (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL CHECK (length(trim(name)) > 0),
    birthDate TEXT NOT NULL CHECK (birthDate IS date(birthDate))
) STRICT;


-- movies table
CREATE TABLE IF NOT EXISTS Movie (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL CHECK (length(trim(title)) > 0),
    releaseYear INTEGER NOT NULL CHECK (releaseYear BETWEEN 1888 AND 2100),
    -- duration is in seconds
    duration INTEGER NOT NULL CHECK (duration > 0)
) STRICT;


-- genres table
CREATE TABLE IF NOT EXISTS Genre (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL COLLATE NOCASE UNIQUE CHECK (length(trim(name)) > 0)
) STRICT;


-- joint table for Movie & Genre
CREATE TABLE IF NOT EXISTS MovieGenre (
    movie_id INTEGER NOT NULL,
    genre_id INTEGER NOT NULL,
    PRIMARY KEY (movie_id, genre_id),
    FOREIGN KEY (movie_id) REFERENCES Movie(id) ON DELETE RESTRICT,
    FOREIGN KEY (genre_id) REFERENCES Genre(id) ON DELETE RESTRICT
) STRICT;


-- joint table for Movie & Actor
CREATE TABLE IF NOT EXISTS MovieActor (
    movie_id INTEGER NOT NULL,
    actor_id INTEGER NOT NULL,
    PRIMARY KEY (movie_id, actor_id),
    FOREIGN KEY (movie_id) REFERENCES Movie(id) ON DELETE RESTRICT,
    FOREIGN KEY (actor_id) REFERENCES Actor(id) ON DELETE RESTRICT
) STRICT;


-- indexes. should significantly speed up the lookup for a huge databases,
-- for example with 1m records, but may slow down writing.

-- useful to filter movie by specific genre
CREATE INDEX IF NOT EXISTS idx_moviegenre_genre ON MovieGenre(genre_id);

-- useful to search all the movies with specific actor
CREATE INDEX IF NOT EXISTS idx_movieactor_actor ON MovieActor(actor_id);

-- for finding movies released in specific year
CREATE INDEX IF NOT EXISTS idx_movie_releaseYear ON Movie(releaseYear);

-- actors sorted by name, case-insensitive check
CREATE INDEX IF NOT EXISTS idx_actor_name ON Actor(name COLLATE NOCASE);

-- movies sorted by title, case-insensitive check
CREATE INDEX IF NOT EXISTS idx_movie_title ON Movie(title COLLATE NOCASE);

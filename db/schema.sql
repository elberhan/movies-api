-- create actors table below


-- create movies table below
CREATE TABLE IF NOT EXISTS Movie (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    releaseYear INTEGER NOT NULL CHECK (releaseYear BETWEEN 1888 AND 2100),
    duration INTEGER NOT NULL CHECK (duration > 0)
);

-- create genres table below
CREATE TABLE IF NOT EXISTS Genre (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);


-- joint table for Genre & Movie
CREATE TABLE IF NOT EXISTS movie_genres (
    movie_id INTEGER NOT NULL,
    genre_id INTEGER NOT NULL,
    PRIMARY KEY (movie_id, genre_id),
    FOREIGN KEY (movie_id) REFERENCES Movie(id) ON DELETE CASCADE,
    FOREIGN KEY (genre_id) REFERENCES Genre(id) ON DELETE CASCADE
);

-- joint table for Movie & Actor


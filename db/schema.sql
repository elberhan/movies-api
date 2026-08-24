-- actors table
CREATE TABLE IF NOT EXISTS Actor (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    birthDate TEXT NOT NULL
        CHECK (birthDate IS date(birthDate))
) STRICT;


-- movies table
CREATE TABLE IF NOT EXISTS Movie (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    releaseYear INTEGER NOT NULL CHECK (releaseYear BETWEEN 1888 AND 2100),
    duration INTEGER NOT NULL CHECK (duration > 0)
) STRICT;


-- genres table
CREATE TABLE IF NOT EXISTS Genre (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
) STRICT;


-- joint table for Genre & Movie
CREATE TABLE IF NOT EXISTS movie_genres (
    movie_id INTEGER NOT NULL,
    genre_id INTEGER NOT NULL,
    PRIMARY KEY (movie_id, genre_id),
    FOREIGN KEY (movie_id) REFERENCES Movie(id) ON DELETE CASCADE,
    FOREIGN KEY (genre_id) REFERENCES Genre(id) ON DELETE CASCADE
) STRICT;


-- joint table for Movie & Actor
CREATE TABLE IF NOT EXISTS movie_actors (
    movie_id INTEGER NOT NULL,
    actor_id INTEGER NOT NULL,
    PRIMARY KEY (movie_id, actor_id),
    FOREIGN KEY (movie_id) REFERENCES Movie(id) ON DELETE CASCADE,
    FOREIGN KEY (actor_id) REFERENCES Actor(id) ON DELETE CASCADE
) STRICT;

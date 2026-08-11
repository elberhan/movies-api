-- create actors table below


-- create movies table below
CREATE TABLE IF NOT EXISTS Movie (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    releaseYear INTEGER NOT NULL CHECK (releaseYear BETWEEN 1888 AND 2100),
    duration INTEGER NOT NULL CHECK (duration > 0)
);

-- create genres table below



-- joint table for Genre & Movie


-- joint table for Movie & Actor


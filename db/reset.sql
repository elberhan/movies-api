-- delete joint tables (connections) first to be able to delete our main tables
DELETE FROM movie_genres;
DELETE FROM movie_actors;
DELETE FROM Movie;
DELETE FROM Genre;
DELETE FROM Actor;

-- Reset AUTOINCREMENT counters so new ids start from 1 again. FOR ANY TABLE!
DELETE FROM sqlite_sequence;

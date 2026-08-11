package db

// TODO: open SQLite connection, create tables, and
// stuff like Ping, PRAGMA foreign_keys = ON, etc.

import (
	_ "embed"
)

//go:embed schema.sql
var schema string

//go:embed seed.sql
var seed string

var _ = schema // for debug
var _ = seed   // for debug

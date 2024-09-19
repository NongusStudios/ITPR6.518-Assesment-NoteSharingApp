package main

// PostgreSQl configuration if not passed as env variables
const (
	dbHost     = "localhost" //127.0.0.1
	dbPort     = 5432
	dbUser     = "postgres"
	dbPassword = "pp"
	dbName     = "notedb"
	dbFileLock = "dbImported"
)

// Positional command line args
const (
	argBindport = 1
)

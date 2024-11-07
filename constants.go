package main

// PostgreSQl configuration if not passed as env variables
const (
	dbHost     = "localhost" //127.0.0.1
	dbPort     = 5432
	dbUser     = "postgres"
	dbPassword = "pp"
	dbName     = "postgres"
	dbFileLock = "dbImported"
)

// Positional command line args
const (
	argBindport = 1
)

// Note Flags
const (
	NoteFlagNote = iota
	NoteFlagInProgress
	NoteFlagCompleted
	NoteFlagCancelled
	NoteFlagDelegated
	NoteFlagMax
)

// Global Constants
const (
	UsernameMaxLength = 255
	PasswordMaxLength = 255
	NoteNameMaxLength = 255
)

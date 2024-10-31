# Documentation

This file documents the use and design decisions of this project.

## Dependencies

- [Gorilla - mux](https://github.com/gorilla/mux)
  - HTTP router.
- [icza - session](https://github.com/icza/session)
  - Creating user sessions
- [Jackc - pgx](https://github.com/jackc/pgx)
  - PostgreSQL Driver
- [lib - pq](https://github.com/lib/pq)
  - Used for the Array types that can be scanned from the db
- [x - crypto](https://golang.org/x/crypto)
  - Used to encrypt user passwords

## Database Design

### ERD
![erd]()
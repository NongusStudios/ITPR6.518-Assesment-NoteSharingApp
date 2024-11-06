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

## Design Philosphy

When building this application I approached it with a develop quickly,
fail quicker approach. In practice this meant writing required
functionality as simply as possible working out any bugs or issues with it
and then reiterating upon it until I was satisfied with its functionality and robustness.

## Database Design

### ERD

![erd.png](erd.png)

### Placeholders

I encountered a problem when checking if a table already had an entry I wanted to insert. If the table contained zero entries a different error
would be returned instead of `sql.ErrNoRows`, stopping data from being inserted because I was looking for this error to insert that entry.
\
I solved this issue by inserting a placeholder entry when creating the
tables. For users I added an empty user by the name `__placeholder__user__`
and I added a welcome note shared globally for the notes table.
I did not have to worry about the user_settings table as I just inserted a entry when a user was added.\
While this approach works it is probably quite naive and a better solution
most certainly exists.

### Avoiding SQL Injection

I have minimized SQL injection risks by taking the following precautionary steps:

- When a user creates an account spaces are not allowed in the username or password
- SQL statements are prepared and parameterized, apposed to concatenation or using fmt

## UX/UI

The UI I have developed for the website is purposely bare bones and simple. This allowed me to
focus on the backend implementation and rapidly develop a UI around it. In the first stages
of development it was purely HTML with no CSS, and the CSS that is in the final version is
nothing fancy (gets the job done).

### Login / Register

Any errors with input data will appear below the login button.

![login](login.PNG)

Once registered the user will be taken back to the login page. Error messages will appear in the same
place as the login page.

![register](register.PNG)

### Dashboard

Contains your notes and those shared with you.
![dashboard.png](dashboard.PNG)

### Create a Note

![create-modal](create-modal.PNG)

### Edit a Note

Fields update to the notes content and options when you change the note.
![edit-modal](edit-modal.PNG)

### Delete a Note

![delete-modal](delete-modal.PNG)

### User Settings

![settings-modal](settings-modal.PNG)

## Testing

While I haven't setup unit testing I have manually tested the functionality
of the app, the results are in the following tables.

### Authorization

| Action | Expected | Result |
| ------ | -------- | ------ |
| Register a new account | A new entry added to users of same name and password | Same as Expected |
| Register an existing account | Data is rejected and user prompted with warning | Same as Expected |
| Register a new account with an invalid name/password | Data is rejected and user prompted with warning | Same as expected |
| Login with an existing account and correct password | Taken to dashboard logged in as that user | Same as expected |
| Login with an existing account and wrong password | Rejected and user prompted with warning | Same as expected |
| Login with an account that doesn't exist | Rejected and user prompted with warning | Same as expected |

### Dashboard

| Action | Expected | Result |
| ------ | -------- | ------ |

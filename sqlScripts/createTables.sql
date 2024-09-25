DROP TABLE IF EXISTS "users";
CREATE TABLE "users" (
    user_id SERIAL PRIMARY KEY NOT NULL,
    username VARCHAR(255) NOT NULL, 
    pass VARCHAR(255) NOT NULL
);

-- When trying to register a user with nothing else in the table an error would return not equal to sql.ErrNoRows
-- This made adding the very first user impossible, so rather than accounting for this in code
-- I am using the genius (probably naive) approach of just inserting a placeholder user
INSERT INTO users(username, pass)
VALUES('__placeholder__user__', '');

DROP TABLE IF EXISTS "notes";
CREATE TABLE "notes" (
    note_id SERIAL PRIMARY KEY NOT NULL,
    noteName VARCHAR(255) NOT NULL,
    noteDate DATE,
    content TEXT
);

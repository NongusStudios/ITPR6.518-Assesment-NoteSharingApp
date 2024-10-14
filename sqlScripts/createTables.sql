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
    note_owner VARCHAR(255) NOT NULL,
    note_share VARCHAR(512) NOT NULL, -- Comma seperated list of usernames, If set to global it is visible to all users
    note_name VARCHAR(255) NOT NULL,
    note_date DATE,
    note_content TEXT
);

INSERT INTO notes(note_owner, note_share, note_name, note_date, note_content)
VALUES(' ', 'global', 'Welcome', NOW(), 'Welcome to Enterprise Note Sharer enjoy your notes');

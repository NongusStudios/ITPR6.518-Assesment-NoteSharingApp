DROP TABLE IF EXISTS "user";
CREATE TABLE "user" (
    id SERIAL PRIMARY KEY NOT NULL,
    username VARCHAR(255) NOT NULL, 
    password VARCHAR(255) NOT NULL
);

DROP TABLE IF EXISTS "note";
CREATE TABLE "note" (
    id SERIAL PRIMARY KEY NOT NULL,
    noteName VARCHAR(255) NOT NULL,
    noteDate DATE,
    content TEXT
);
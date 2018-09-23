DROP TABLE IF EXISTS notes;
CREATE TABLE notes (
    id serial PRIMARY KEY,
	Title    varchar (25) NOT NULL,
	Content  varchar (1024) NOT NULL
);
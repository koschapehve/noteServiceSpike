package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Indirect usage as sql driver
	"log"
	"os"
	"strconv"
)

type Configuration struct {
	Host     string
	Port     int
	User     string
	Password string
	DbName   string
	MaxOpenConns int
	MaxIdleConns int
}

type NoteServicePostgres struct {
	db *sqlx.DB
}

func noteServicePostgres() NoteService {
	conf := readConfig()
	conectionString := fmt.Sprintf("host = %s port = %d user = %s password = %s dbname = %s ", conf.Host, conf.Port, conf.User, conf.Password, conf.DbName)

	newDb, err := sqlx.Open("postgres", conectionString)

	if err != nil {
		panic(err)
	} else {
		log.Println("Postgres connection established")
	}
	newDb.SetMaxOpenConns(conf.MaxOpenConns)
	newDb.SetMaxIdleConns(conf.MaxIdleConns)
	err = newDb.Ping()
	if err != nil {
		panic(err)
	}
	return NoteServicePostgres{newDb}
}

func readConfig() Configuration {
	file, err := os.Open("src/noteservice/postgres_config.json")
	if err != nil {
		panic(err)
	}
	conf := Configuration{}
	err = json.NewDecoder(file).Decode(&conf)
	if err != nil {
		panic(err)
	}
	return conf
}

func (s NoteServicePostgres) ReadNoteById(id string) (note Note, err error) {
	const sql = `SELECT * FROM notes WHERE id=$1`
	row := s.db.QueryRow(sql, id)
	err = row.Scan(&note.Id, &note.Title, &note.Content)
	return
}

func (s NoteServicePostgres) ReadNoteList(size, page int) (noteList []Note, err error) {
	const sql = `SELECT * FROM notes  ORDER BY Id ASC LIMIT $1 OFFSET $2`
	rows, err := s.db.Query(sql, size, page*size)
	if err != nil {
		return
	}
	defer rows.Close()

	noteList = make([]Note, 0, size)
	for rows.Next() {
		note := Note{}
		err := rows.Scan(&note.Id, &note.Title, &note.Content)
		if err != nil {
			return nil, err
		}
		noteList = append(noteList, note)
	}
	return
}

func (s NoteServicePostgres) AddNote(note Note) (id int, err error) {
	// workaround fo missing  Postgres support for res.LastInsertId()
	const sql = "INSERT INTO notes ( title, content ) VALUES ( :title, :content ) RETURNING id"
	rows, err := s.db.NamedQuery(sql, note)

	if err != nil {
		return
	}
	defer rows.Close()
	rows.Next()
	err = rows.Scan(&id)
	return
}

func (s NoteServicePostgres) RemoveNote(id int) (err error) {
	const sql = `DELETE FROM notes WHERE id=$1`
	res, err := s.db.Exec(sql, id)
	if err != nil {
		return
	}
	deleted, err := res.RowsAffected()
	if err != nil {
		return
	}

	if deleted != 1 {
		delStr := strconv.FormatInt(deleted, 10)
		return errors.New(delStr + " entries deleted")
	}
	return nil
}

func (s NoteServicePostgres) UpdateNote(note Note) error {
	const sql = "UPDATE  notes set title =  :title, content = :content  where id = :id "
	res, err := s.db.NamedExec(sql, note)

	if err != nil {
		return err
	}
	updateCount, err := res.RowsAffected()
	if err != nil {
		return err
	}

	updateCountStr := strconv.FormatInt(updateCount, 10)
	if updateCount != 1 {
		return fmt.Errorf("%s entries updated", updateCountStr)
	}
	return nil
}
func (s NoteServicePostgres) Close() error {
	return s.db.Close()
}

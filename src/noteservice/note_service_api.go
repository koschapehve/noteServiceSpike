package main

type Note struct {
	Id      int    `json:"id,omitempty" db:"id"`
	Title   string `json:"title,omitempty" db:"title"`
	Content string `json:"content,omitempty" db:"content"`
}

type NoteService interface {
	ReadNoteList(size, page int) (noteList []Note, err error)
	ReadNoteById(id string) (note Note, err error)
	AddNote(note Note) (id int, err error)
	RemoveNote(id int) (err error)
	UpdateNote(note Note) error
	Close() error
}

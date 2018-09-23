package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

const helpMsg = "Availbale services \n" +
	"noteService/get/[Note_Id]\n" +
	"noteService/list/[List_Size]\n" +
	"noteService/list/[List_Size]/[List_Offset]\n" +
	"noteService/create/\n" +
	" Post with json note as body\n" +
	"noteService/update/\n" +
	" Post with json note as body\n" +
	"noteService/delete/[Note_Id]"

func main() {
	handler := NoteServiceHttphandler{noteServicePostgres()}
	handler.startHandler()
	defer handler.noteService.Close()
}

type NoteServiceHttphandler struct {
	noteService NoteService
}

func (h NoteServiceHttphandler) startHandler() {
	router := mux.NewRouter()
	router.HandleFunc("/noteService", printHelp).Methods("GET")
	router.Handle("/noteService/list/{size}", errorHandler(h.listNotes)).Methods("GET")
	router.Handle("/noteService/list/{size}/{page}", errorHandler(h.listNotes)).Methods("GET")
	router.Handle("/noteService/get/{id}", errorHandler(h.getNote)).Methods("GET")
	router.Handle("/noteService/create", errorHandler(h.createNote)).Methods("POST")
	router.Handle("/noteService/update", errorHandler(h.modifyNote)).Methods("POST")
	router.Handle("/noteService/delete/{id}", errorHandler(h.deleteNote)).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))

}

type errorHandler func(http.ResponseWriter, *http.Request) error

func (fn errorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (h NoteServiceHttphandler) listNotes(w http.ResponseWriter, r *http.Request) error {
	size, err := strconv.Atoi(mux.Vars(r)["size"])
	if err != nil {
		return err
	}
	var page int = 0
	if len(mux.Vars(r)["page"]) > 0 {
		page, err = strconv.Atoi(mux.Vars(r)["page"])
		if err != nil {
			return err
		}
	}

	noteList, err := h.noteService.ReadNoteList(size, page)
	if err != nil {
		return err
	}
	return sendBackNoteAsResponse(w, noteList...)
}

func (h NoteServiceHttphandler) getNote(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id := vars["id"]

	note, err := h.noteService.ReadNoteById(id)
	if err != nil {
		return err
	}
	return sendBackNoteAsResponse(w, note)
}

func (h NoteServiceHttphandler) deleteNote(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	idArg := vars["id"]
	if len(idArg) <= 0 {
		return errors.New("Invalid note id")
	}
	id, err := strconv.Atoi(idArg)
	if err != nil {
		return err
	}
	err = h.noteService.RemoveNote(id)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "Entry with id %d deleted", id)
	return nil
}

func (h NoteServiceHttphandler) createNote(w http.ResponseWriter, r *http.Request) error {
	var note Note
	err := json.NewDecoder(r.Body).Decode(&note)
	if err != nil {
		return err
	}

	id, err := h.noteService.AddNote(note)
	if err != nil {
		return err
	}
	note.Id = id
	return sendBackNoteAsResponse(w, note)
}

func (h NoteServiceHttphandler) modifyNote(w http.ResponseWriter, r *http.Request) error {
	var note Note
	err := json.NewDecoder(r.Body).Decode(&note)
	if err != nil {
		return err
	}

	err = h.noteService.UpdateNote(note)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Entry with id %d updated", note.Id)
	return nil
}

func printHelp(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, helpMsg)
}

func sendBackNoteAsResponse(w http.ResponseWriter, notes ...Note) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(notes)
}

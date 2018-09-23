package main

import (
	"errors"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPrintHelp(t *testing.T) {
	// Given
	writer := httptest.NewRecorder()

	// When
	printHelp(writer, nil)

	// Then
	res := writer.Result()
	body, _ := ioutil.ReadAll(res.Body)

	assert := assert.New(t)
	assert.NotNil(res)
	assert.Equal(200, res.StatusCode)
	assert.Equal("text/plain; charset=utf-8", res.Header.Get("Content-Type"))
	assert.Equal(helpMsg, string(body))
}

type MockedNoteService struct {
	mock.Mock
}

func (m *MockedNoteService) ReadNoteList(size, page int) (noteList []Note, err error) {
	args := m.Called(size, page)
	return args.Get(0).([]Note), args.Error(1)
}
func (m *MockedNoteService) ReadNoteById(id string) (note Note, err error) {
	args := m.Called(id)
	return args.Get(0).(Note), args.Error(1)
}
func (m *MockedNoteService) AddNote(note Note) (id int, err error) {
	args := m.Called(note)
	return args.Int(0), args.Error(1)
}
func (m *MockedNoteService) RemoveNote(id int) (err error) {
	args := m.Called(id)
	return args.Error(0)
}
func (m *MockedNoteService) UpdateNote(note Note) (err error) {
	args := m.Called(note)
	return args.Error(0)
}
func (m *MockedNoteService) Close() (err error) {
	args := m.Called()
	return args.Error(0)
}

func TestGetNote_Succcess(t *testing.T) {
	// Given
	service := new(MockedNoteService)
	handler := NoteServiceHttphandler{service}
	service.On("ReadNoteById", "20").Return(Note{20, "mocked title", "mocked content"}, nil)

	writer := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/any/url", nil)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{"id": "20"})

	// When
	err = handler.getNote(writer, req)

	// Then
	assert := assert.New(t)
	assert.Nil(err)
	res := writer.Result()
	body, _ := ioutil.ReadAll(res.Body)

	assert.NotNil(res)
	assert.Equal(200, res.StatusCode)
	assert.Equal("application/json", res.Header.Get("Content-Type"))
	assert.Equal("[{\"id\":20,\"title\":\"mocked title\",\"content\":\"mocked content\"}]\n", string(body))
}

func TestGetNote_Fail(t *testing.T) {
	// Given
	service := new(MockedNoteService)
	handler := NoteServiceHttphandler{service}
	expectedError := errors.New("Read failed")
	service.On("ReadNoteById", "20").Return(Note{20, "mocked title", "mocked content"}, expectedError)

	req, err := http.NewRequest("GET", "/any/url", nil)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{"id": "20"})

	// When
	err = handler.getNote(nil, req)

	// Then
	assert.New(t).Equal(expectedError, err)
}

func TestReadNoteList_Success(t *testing.T) {
	// Given
	service := new(MockedNoteService)
	handler := NoteServiceHttphandler{service}
	service.On("ReadNoteList", 4, 1).Return([]Note{
		Note{20, "mocked title", "mocked content"},
		Note{21, "another mocked title", "another mocked content"}}, nil)

	writer := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/any/url", nil)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{
		"size": "4",
		"page": "1"})

	// When
	err = handler.listNotes(writer, req)

	// Then
	assert := assert.New(t)
	assert.Nil(err)
	res := writer.Result()
	body, _ := ioutil.ReadAll(res.Body)

	assert.NotNil(res)
	assert.Equal(200, res.StatusCode)
	assert.Equal("application/json", res.Header.Get("Content-Type"))
	assert.Equal("[{\"id\":20,\"title\":\"mocked title\",\"content\":\"mocked content\"}"+
		",{\"id\":21,\"title\":\"another mocked title\",\"content\":\"another mocked content\"}]\n", string(body))
}

func TestReadNoteList_Success_missingPage(t *testing.T) {
	// Given
	service := new(MockedNoteService)
	handler := NoteServiceHttphandler{service}
	service.On("ReadNoteList", 4, 0).Return([]Note{
		Note{20, "mocked title", "mocked content"},
		Note{21, "another mocked title", "another mocked content"}}, nil)

	writer := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/any/url", nil)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{
		"size": "4"})

	// When
	err = handler.listNotes(writer, req)

	// Then
	assert := assert.New(t)
	assert.Nil(err)
	res := writer.Result()
	body, _ := ioutil.ReadAll(res.Body)

	assert.NotNil(res)
	assert.Equal(200, res.StatusCode)
	assert.Equal("application/json", res.Header.Get("Content-Type"))
	assert.Equal("[{\"id\":20,\"title\":\"mocked title\",\"content\":\"mocked content\"}"+
		",{\"id\":21,\"title\":\"another mocked title\",\"content\":\"another mocked content\"}]\n", string(body))
}

func TestReadNoteList_Fail_read(t *testing.T) {
	// Given
	service := new(MockedNoteService)
	handler := NoteServiceHttphandler{service}
	expectedError := errors.New("Read list failed")
	service.On("ReadNoteList", 4, 1).Return([]Note{
		Note{20, "mocked title", "mocked content"},
		Note{21, "another mocked title", "another mocked content"}}, expectedError)

	req, err := http.NewRequest("GET", "/any/url", nil)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{
		"size": "4",
		"page": "1"})

	// When
	err = handler.listNotes(nil, req)

	// Then
	assert.New(t).Equal(expectedError, err)
}
func TestReadNoteList_Fail_missingSizeParam(t *testing.T) {
	// Given

	handler := NoteServiceHttphandler{nil}
	req, err := http.NewRequest("GET", "/any/url", nil)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{"page": "4"})

	// When
	err = handler.listNotes(nil, req)

	// Then
	assert.New(t).Equal("strconv.Atoi: parsing \"\": invalid syntax", err.Error())
}

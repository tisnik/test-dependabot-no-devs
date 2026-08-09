package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const Port = ":8080"

const MaxRows = 99
const MaxColumns = 'Z'

type Storage interface {
}

type MemoryStorage struct {
}

func NewStorage() Storage {
	return MemoryStorage{}
}

type Server interface {
	Serve(port uint)
}

// Implementace HTTPServeru
type ServerImpl struct {
	storage Storage
}

func NewServer(storage Storage) Server {
	return ServerImpl{
		storage: storage,
	}
}

func (s ServerImpl) Serve(port uint) {
	log.Printf("Starting server on port %d", port)

	http.HandleFunc("/", s.mainEndpoint)

	// REST API endpoints
	http.HandleFunc("GET /cell/{id}", s.returnCell)

	// start the server
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func (s ServerImpl) mainEndpoint(writer http.ResponseWriter, request *http.Request) {
	PageHeader := `
		<!doctype html>
		<html>
		    <head>
		        <title>ε</title>
		        <!--script src="fengari-web.js"></script>-->
		        <!--<script src="canvas.lua" type="text/lua">-->
		    </head>
		    <body>
	`

	PageFooter := `
		    </body>
		</html>
	`
	io.WriteString(writer, PageHeader)
	io.WriteString(writer, "<table>\n")
	io.WriteString(writer, "<tr><th>&nbsp;</th>")
	for column := 'A'; column <= MaxColumns; column++ {
		io.WriteString(writer, fmt.Sprintf("<th>%c</th>", column))

	}
	io.WriteString(writer, "</tr>\n")
	for row := 1; row <= MaxRows; row++ {
		io.WriteString(writer, fmt.Sprintf("<tr><th>%d</th>", row))
		for column := 'A'; column < MaxColumns; column++ {
			io.WriteString(writer, fmt.Sprintf("<td></td>"))
		}
		io.WriteString(writer, "</tr>\n")
	}
	io.WriteString(writer, "</table>\n")
	io.WriteString(writer, PageFooter)
}

func (s ServerImpl) returnCell(writer http.ResponseWriter, r *http.Request) {
	ID := r.PathValue("id")
	log.Printf("Cell ID provided: %s", ID)

	/*
		ID, err := strconv.Atoi(IDs)
		if err != nil {
			writer.Header().Set("Content-Type", "text/plain")
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

			user, found := s.storage.ReadUser(ID)
			if !found {
				writer.Header().Set("Content-Type", "text/plain")
				writer.WriteHeader(http.StatusNotFound)
				return
			}*/
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(ID)
}

func main() {
	storage := NewStorage()
	server := NewServer(storage)
	server.Serve(8080)
}

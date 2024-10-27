package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

type flags struct {
	add   string
	del   int
	list  int
	purge bool
}

func listNotes(db *sql.DB) {
	stmt, err := db.Prepare("SELECT id, note FROM notes")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var note string
		err = rows.Scan(&id, &note)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(id, note)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
}

func listNoteById(noteId int, db *sql.DB) {
	stmt, err := db.Prepare("SELECT id, note FROM notes WHERE id = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	var id int
	var note string
	err = stmt.QueryRow(noteId).Scan(&id, &note)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(id, note)
}

func addNote(content string, db *sql.DB) {
	stmt, err := db.Prepare("INSERT INTO notes(id, note) values(NULL, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(content)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully added note")
}

func delNotes() {
	os.Remove("./sticky.db")
}

func delNote(noteId int, db *sql.DB) {
	stmt, err := db.Prepare("DELETE FROM notes WHERE id = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(noteId)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully deleted note #", noteId)
}

func main() {
	f := new(flags)

	flag.StringVar(&f.add, "add", "", "add note")
	flag.IntVar(&f.del, "del", 0, "delete note by id")
	flag.IntVar(&f.list, "list", 0, "list one note by id, list all notes for id=0")
	flag.BoolVar(&f.purge, "purge", false, "delete all notes")
	flag.Parse()

	db, err := sql.Open("sqlite3", "./sticky.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	stmt := `
	CREATE table IF NOT EXISTS notes (
		id integer NOT NULL PRIMARY KEY,
		note TEXT
	);
	`
	_, err = db.Exec(stmt)
	if err != nil {
		log.Printf("%q: %s\n", err, stmt)
		return
	}

	switch {
	case f.list != 0:
		listNoteById(f.list, db)
	case f.add != "":
		addNote(f.add, db)
	case f.del != 0:
		delNote(f.del, db)
	case f.purge != false:
		delNotes()
	default:
		listNotes(db)
	}
}

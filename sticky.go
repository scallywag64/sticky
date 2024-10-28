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

type Note struct {
	VirtualID int
	ID        int
	Content   string
}

const dbFilePath = "./sticky.db"

func initDb(path string) *sql.DB {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	stmt := `
	CREATE table IF NOT EXISTS notes (
		id integer NOT NULL PRIMARY KEY,
		note TEXT
	);
	`
	_, err = db.Exec(stmt)
	if err != nil {
		log.Printf("%q: %s\n", err, stmt)
		return nil
	}

	return db
}

func listNotes(db *sql.DB) {
	stmt, err := db.Prepare(`
		SELECT
			ROW_NUMBER() OVER (ORDER BY id) AS virtual_id,
			note
		FROM notes;
	`)
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
		var virtualId int
		var note string
		err = rows.Scan(&virtualId, &note)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(virtualId, note)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
}

func listNoteById(noteId int, db *sql.DB) {
	stmt, err := db.Prepare(`
		WITH ordered_notes AS (
			SELECT
				ROW_NUMBER() OVER (ORDER by id) AS virtual_id,
				note
			FROM notes
		)
		SELECT *
		FROM ordered_notes
		WHERE virtual_id = ?
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	var virtualId int
	var note string
	err = stmt.QueryRow(noteId).Scan(&virtualId, &note)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(virtualId, note)
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
	str := "\x1b[31mThis operation will delete your entire notes database.\x1b[0m\n" +
		"Type \"y\" to proceed, type anything else to cancel.\n\x1b[34m> \x1b[0m"
	fmt.Print(str)

	var answer string
	_, err := fmt.Scan(&answer)
	if err != nil {
		fmt.Println("Error reading input:", err)
		return
	}

	if answer == "y" {
		os.Remove("./sticky.db")
		fmt.Println("\x1b[33mSticky notes database deleted.\x1b[0m")
	} else {
		fmt.Println("\x1b[32mSticky notes database preserved.\x1b[0m")
	}
}

func delNote(noteId int, db *sql.DB) {
	stmt, err := db.Prepare(`
		WITH ordered_notes AS (
			SELECT
				id,
				ROW_NUMBER() OVER (ORDER BY id) as virtual_id
			FROM notes
		)
		DELETE FROM notes
		WHERE id = (SELECT id FROM ordered_notes WHERE virtual_id = ?);
	`)
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

	db := initDb(dbFilePath)
	if db == nil {
		log.Fatal("Failed to initialize the database")
	}
	defer db.Close()

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

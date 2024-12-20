package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type flags struct {
	add   string
	get   int
	list  bool
	del   int
	purge bool
}

type Note struct {
	VirtualID int
	ID        int
	Content   string
}

const (
	red    = "\x1b[31m"
	green  = "\x1b[32m"
	yellow = "\x1b[33m"
	blue   = "\x1b[34m"
	reset  = "\x1b[0m"
)

func getDbPath() string {
	dbPath := "./sticky.db"

	if os.Getenv("STICKY_ENV") != "dev" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}

		dbPath = filepath.Join(homeDir, ".local", "share", "sticky", "sticky.db")
		if err := os.MkdirAll(filepath.Dir(dbPath), os.ModePerm); err != nil {
			log.Fatalf("Error creating directory: %v", err)
		}
	}

	return dbPath
}

func initDb() *sql.DB {
	dbPath := getDbPath()

	_, err := os.Stat(dbPath)
	databaseExists := !os.IsNotExist(err)

	db, err := sql.Open("sqlite3", dbPath)
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

	if !databaseExists {

		wd, err := os.Getwd()
		if err != nil {
			log.Println(err)
		}
		fmt.Println(blue + "Created 'sticky.db' database at: " + wd + reset)
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

func getNoteById(noteId int, db *sql.DB) {
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
	str := red + "This operation will delete your entire notes database.\n" + reset +
		"Type \"y\" to proceed, type anything else to cancel.\n" +
		blue + "> " + reset
	fmt.Print(str)

	var answer string
	_, err := fmt.Scan(&answer)
	if err != nil {
		fmt.Println("Error reading input:", err)
		return
	}

	if answer == "y" {
		dbPath := getDbPath()
		os.Remove(dbPath)
		fmt.Println(yellow + "Sticky notes database deleted." + reset)
	} else {
		fmt.Println(green + "Sticky notes database preserved." + reset)
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

	fmt.Printf("Successfully deleted note #%d\n", noteId)
}

func main() {
	f := new(flags)
	flag.StringVar(&f.add, "add", "", "add note")
	flag.IntVar(&f.get, "get", 0, "get note by id")
	flag.BoolVar(&f.list, "list", false, "list all notes")
	flag.IntVar(&f.del, "del", 0, "delete note by id")
	flag.BoolVar(&f.purge, "purge", false, "delete notes database")
	flag.Parse()

	db := initDb()
	if db == nil {
		log.Fatal("Failed to initialize the database")
	}
	defer db.Close()

	switch {
	case f.add != "":
		addNote(f.add, db)
	case f.get != 0:
		getNoteById(f.get, db)
	case f.list:
		listNotes(db)
	case f.del != 0:
		delNote(f.del, db)
	case f.purge:
		delNotes()
	default:
		listNotes(db)
	}
}

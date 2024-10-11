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

func main() {
	f := new(flags)

	flag.StringVar(&f.add, "add", "", "add note")
	flag.IntVar(&f.del, "del", -1, "delete note by index")
	flag.IntVar(&f.list, "list", -1, "list one note by index, list all notes for index=0")
	flag.BoolVar(&f.purge, "purge", false, "purge database for debugging")
	flag.Parse()

	if f.purge {
		os.Remove("./sticky.db")
	}

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

	count := 0
	if f.add != "" {
		count++
	}
	if f.del > 0 {
		count++
	}
	if f.list >= 0 {
		count++
	}

	if count > 1 {
		fmt.Println("Error: only one flag is allowed at a time.")
		flag.Usage()
		os.Exit(1)
	}

	if len(f.add) > 0 {
		stmt, err := db.Prepare("INSERT INTO notes(id, note) values(NULL, ?)")
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()

		_, err = stmt.Exec(f.add)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Successfully added note")
	} else if f.del > 0 {
		stmt, err := db.Prepare("DELETE FROM notes WHERE id = ?")
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()

		_, err = stmt.Exec(f.del)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Successfully deleted note #", f.del)
	} else if f.list > 0 {
		stmt, err := db.Prepare("SELECT id, note FROM notes WHERE id = ?")
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()

		var id int
		var note string
		err = stmt.QueryRow(f.list).Scan(&id, &note)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(id, note)
	} else {
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
}

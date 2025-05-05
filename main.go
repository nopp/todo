package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var err error
	db, err = sql.Open("sqlite3", "./data/todo.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	createTables()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/add-task", addTaskHandler)
	http.HandleFunc("/delete-task", deleteTaskHandler)
	http.HandleFunc("/edit-task", editTaskHandler)
	http.HandleFunc("/update-task-status", updateTaskStatusHandler)
	http.HandleFunc("/categories", categoryPageHandler)
	http.HandleFunc("/add-category", addCategoryHandler)
	http.HandleFunc("/delete-category", deleteCategoryHandler)
	http.HandleFunc("/edit-category", editCategoryHandler)
	http.HandleFunc("/backlog", backlogHandler)
	http.HandleFunc("/done", doneHandler)
	http.HandleFunc("/add-task-page", addTaskPageHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server starting on :8080")
	log.Println("Logging is working!")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

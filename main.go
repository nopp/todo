package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type Category struct {
	Name string `json:"name"`
}

var db *sql.DB

func main() {
	// Configure logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting TODO application")

	var err error
	db, err = sql.Open("sqlite3", "./data/todo.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	createTables()

	loadData()

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
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func createTables() {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS categories (
			name TEXT PRIMARY KEY
		);
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT,
			category TEXT,
			description TEXT,
			status TEXT,
			created_at DATETIME,
			FOREIGN KEY(category) REFERENCES categories(name)
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}
}

func loadData() {
	log.Println("Loading data")
	rows, err := db.Query("SELECT id, title, category, description, status, created_at FROM tasks")
	if err != nil {
		log.Printf("Error querying tasks: %v", err)
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		var createdAt string
		if err := rows.Scan(&t.ID, &t.Title, &t.Category, &t.Description, &t.Status, &createdAt); err != nil {
			continue
		}
		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		tasks = append(tasks, t)
	}
	log.Printf("Loaded %d tasks", len(tasks))

	rows, err = db.Query("SELECT name FROM categories")
	if err != nil {
		log.Printf("Error querying categories: %v", err)
		return
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.Name); err != nil {
			continue
		}
		categories = append(categories, c)
	}
	log.Printf("Loaded %d categories", len(categories))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	tasks, err := getAllTasks()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	categories, err := getAllCategories()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	var filteredTasks []Task
	if query != "" {
		for _, task := range tasks {
			if task.Status == "backlog" || task.Status == "done" {
				continue
			}
			if strings.Contains(strings.ToLower(task.Title), strings.ToLower(query)) ||
				strings.Contains(strings.ToLower(task.Description), strings.ToLower(query)) ||
				strings.Contains(strings.ToLower(task.Category), strings.ToLower(query)) {
				filteredTasks = append(filteredTasks, task)
			}
		}
	} else {
		for _, task := range tasks {
			if task.Status != "backlog" && task.Status != "done" {
				filteredTasks = append(filteredTasks, task)
			}
		}
	}

	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/index.html"))
	tmpl.ExecuteTemplate(w, "index", struct {
		Tasks      []Task
		Categories []Category
		Query      string
	}{filteredTasks, categories, query})
}

func addTaskPageHandler(w http.ResponseWriter, r *http.Request) {
	categories, err := getAllCategories()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/add-task.html"))
	tmpl.ExecuteTemplate(w, "add-task", struct {
		Categories []Category
	}{categories})
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		category := r.FormValue("category")
		description := r.FormValue("description")
		status := r.FormValue("status")
		if status == "" {
			status = "todo"
		}
		_, err := db.Exec(
			"INSERT INTO tasks (title, category, description, status, created_at) VALUES (?, ?, ?, ?, ?)",
			title, category, description, status, time.Now().Format(time.RFC3339),
		)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	_, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func editTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		idStr := r.FormValue("id")
		id, _ := strconv.Atoi(idStr)
		title := r.FormValue("title")
		category := r.FormValue("category")
		description := r.FormValue("description")
		status := r.FormValue("status")

		_, err := db.Exec(
			"UPDATE tasks SET title = ?, category = ?, description = ?, status = ? WHERE id = ?",
			title, category, description, status, id,
		)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func categoryPageHandler(w http.ResponseWriter, r *http.Request) {
	categories, err := getAllCategories()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/categories.html"))
	tmpl.ExecuteTemplate(w, "categories", categories)
}

func addCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		_, err := db.Exec("INSERT OR IGNORE INTO categories (name) VALUES (?)", name)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}
	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/categories"
	}
	http.Redirect(w, r, referer, http.StatusSeeOther)
}

func deleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	_, err := db.Exec("DELETE FROM categories WHERE name = ?", name)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/categories", http.StatusSeeOther)
}

func updateTaskStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		idStr := r.FormValue("id")
		id, _ := strconv.Atoi(idStr)
		status := r.FormValue("status")
		_, err := db.Exec("UPDATE tasks SET status = ? WHERE id = ?", status, id)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func editCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		oldName := r.FormValue("old_name")
		newName := r.FormValue("new_name")
		if oldName != "" && newName != "" {
			_, err := db.Exec("UPDATE categories SET name = ? WHERE name = ?", newName, oldName)
			if err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
			_, err = db.Exec("UPDATE tasks SET category = ? WHERE category = ?", newName, oldName)
			if err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
		}
	}
	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/categories"
	}
	http.Redirect(w, r, referer, http.StatusSeeOther)
}

func backlogHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	tasks, err := getAllTasks()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	categories, err := getAllCategories()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	var backlogTasks []Task
	if query != "" {
		for _, task := range tasks {
			if task.Status != "backlog" {
				continue
			}
			if strings.Contains(strings.ToLower(task.Title), strings.ToLower(query)) ||
				strings.Contains(strings.ToLower(task.Description), strings.ToLower(query)) ||
				strings.Contains(strings.ToLower(task.Category), strings.ToLower(query)) {
				backlogTasks = append(backlogTasks, task)
			}
		}
	} else {
		for _, task := range tasks {
			if task.Status == "backlog" {
				backlogTasks = append(backlogTasks, task)
			}
		}
	}

	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/backlog.html"))
	tmpl.ExecuteTemplate(w, "backlog", struct {
		Tasks      []Task
		Categories []Category
		Query      string
	}{backlogTasks, categories, query})
}

func doneHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	tasks, err := getAllTasks()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	categories, err := getAllCategories()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	var doneTasks []Task
	if query != "" {
		for _, task := range tasks {
			if task.Status != "done" {
				continue
			}
			if strings.Contains(strings.ToLower(task.Title), strings.ToLower(query)) ||
				strings.Contains(strings.ToLower(task.Description), strings.ToLower(query)) ||
				strings.Contains(strings.ToLower(task.Category), strings.ToLower(query)) {
				doneTasks = append(doneTasks, task)
			}
		}
	} else {
		for _, task := range tasks {
			if task.Status == "done" {
				doneTasks = append(doneTasks, task)
			}
		}
	}

	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/done.html"))
	tmpl.ExecuteTemplate(w, "done", struct {
		Tasks      []Task
		Categories []Category
		Query      string
	}{doneTasks, categories, query})
}

func getAllCategories() ([]Category, error) {
	rows, err := db.Query("SELECT name FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var categories []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.Name); err != nil {
			continue
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func getAllTasks() ([]Task, error) {
	rows, err := db.Query("SELECT id, title, category, description, status, created_at FROM tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []Task
	for rows.Next() {
		var t Task
		var createdAt string
		if err := rows.Scan(&t.ID, &t.Title, &t.Category, &t.Description, &t.Status, &createdAt); err != nil {
			continue
		}
		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		tasks = append(tasks, t)
	}
	return tasks, nil
}

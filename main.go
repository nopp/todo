package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
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

var (
	tasks      = []Task{}
	categories = []Category{}
	taskIDSeq  = 1
	mutex      sync.Mutex
	db         *sql.DB
)

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

	for rows.Next() {
		var t Task
		var createdAt string
		if err := rows.Scan(&t.ID, &t.Title, &t.Category, &t.Description, &t.Status, &createdAt); err != nil {
			log.Printf("Error scanning task: %v", err)
			continue
		}
		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		tasks = append(tasks, t)
		if t.ID >= taskIDSeq {
			taskIDSeq = t.ID + 1
		}
	}
	log.Printf("Loaded %d tasks", len(tasks))

	rows, err = db.Query("SELECT name FROM categories")
	if err != nil {
		log.Printf("Error querying categories: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.Name); err != nil {
			log.Printf("Error scanning category: %v", err)
			continue
		}
		categories = append(categories, c)
	}
	log.Printf("Loaded %d categories", len(categories))
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
			return nil, err
		}
		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	var filteredTasks []Task
	if query != "" {
		for _, task := range tasks {
			// Don't show backlog or done tasks on the main page
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
		// Filter out backlog and done tasks
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

		// Default to todo if no status is provided
		if status == "" {
			status = "todo"
		}

		mutex.Lock()
		tasks = append(tasks, Task{
			ID:          taskIDSeq,
			Title:       title,
			Category:    category,
			Description: description,
			Status:      status,
			CreatedAt:   time.Now(),
		})
		taskIDSeq++
		mutex.Unlock()
	}

	// Redirect to homepage
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	mutex.Lock()
	for i, task := range tasks {
		if id == task.ID {
			tasks = append(tasks[:i], tasks[i+1:]...)
			break
		}
	}
	mutex.Unlock()
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

		mutex.Lock()
		for i := range tasks {
			if id == tasks[i].ID {
				tasks[i].Title = title
				tasks[i].Category = category
				tasks[i].Description = description
				if status != "" {
					tasks[i].Status = status
				}
				break
			}
		}
		mutex.Unlock()
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func categoryPageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/categories.html"))
	tmpl.ExecuteTemplate(w, "categories", categories)
}

func addCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		mutex.Lock()
		categories = append(categories, Category{Name: name})
		mutex.Unlock()
	}

	// Get the referer to redirect back to the page where the request came from
	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/categories"
	}

	http.Redirect(w, r, referer, http.StatusSeeOther)
}

func deleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	mutex.Lock()
	for i, c := range categories {
		if c.Name == name {
			categories = append(categories[:i], categories[i+1:]...)
			break
		}
	}
	mutex.Unlock()
	http.Redirect(w, r, "/categories", http.StatusSeeOther)
}

func updateTaskStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		idStr := r.FormValue("id")
		id, _ := strconv.Atoi(idStr)
		status := r.FormValue("status")

		mutex.Lock()
		for i := range tasks {
			if id == tasks[i].ID {
				tasks[i].Status = status
				break
			}
		}
		mutex.Unlock()
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func editCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		oldName := r.FormValue("old_name")
		newName := r.FormValue("new_name")

		if oldName != "" && newName != "" {
			mutex.Lock()
			// Update the category name
			for i := range categories {
				if categories[i].Name == oldName {
					categories[i].Name = newName

					// Also update all tasks that use this category
					for j := range tasks {
						if tasks[j].Category == oldName {
							tasks[j].Category = newName
						}
					}
					break
				}
			}
			mutex.Unlock()
		}
	}

	// Get the referer to redirect back to the page where the request came from
	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/categories"
	}

	http.Redirect(w, r, referer, http.StatusSeeOther)
}

func backlogHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	var backlogTasks []Task
	if query != "" {
		for _, task := range tasks {
			// Only include backlog tasks
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
		// Get only backlog tasks
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

	var doneTasks []Task
	if query != "" {
		for _, task := range tasks {
			// Only include done tasks
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
		// Get only done tasks
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

func insertTask(t Task) error {
	_, err := db.Exec(
		"INSERT INTO tasks (title, category, description, status, created_at) VALUES (?, ?, ?, ?, ?)",
		t.Title, t.Category, t.Description, t.Status, t.CreatedAt.Format(time.RFC3339),
	)
	return err
}

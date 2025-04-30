package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
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
)

func main() {
	// Configure logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting TODO application")

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
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func saveData() {
	mutex.Lock()
	defer mutex.Unlock()
	taskBytes, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		log.Printf("Error marshalling tasks: %v", err)
		return
	}
	categoryBytes, err := json.MarshalIndent(categories, "", "  ")
	if err != nil {
		log.Printf("Error marshalling categories: %v", err)
		return
	}

	if err := os.WriteFile("data/tasks.json", taskBytes, 0644); err != nil {
		log.Printf("Error writing tasks file: %v", err)
	}
	if err := os.WriteFile("data/categories.json", categoryBytes, 0644); err != nil {
		log.Printf("Error writing categories file: %v", err)
	}
	log.Println("Data saved successfully")
}

func loadData() {
	log.Println("Loading data")
	if taskBytes, err := os.ReadFile("data/tasks.json"); err == nil {
		if err := json.Unmarshal(taskBytes, &tasks); err != nil {
			log.Printf("Error unmarshalling tasks: %v", err)
		} else {
			for _, t := range tasks {
				if t.ID >= taskIDSeq {
					taskIDSeq = t.ID + 1
				}
			}
			log.Printf("Loaded %d tasks", len(tasks))
		}
	} else {
		log.Printf("Tasks file not found: %v", err)
		// Create data directory if it doesn't exist
		if err := os.MkdirAll("data", 0755); err != nil {
			log.Printf("Error creating data directory: %v", err)
		}
	}

	if categoryBytes, err := os.ReadFile("data/categories.json"); err == nil {
		if err := json.Unmarshal(categoryBytes, &categories); err != nil {
			log.Printf("Error unmarshalling categories: %v", err)
		} else {
			log.Printf("Loaded %d categories", len(categories))
		}
	} else {
		log.Printf("Categories file not found: %v", err)
		// Create data directory if it doesn't exist
		if err := os.MkdirAll("data", 0755); err != nil {
			log.Printf("Error creating data directory: %v", err)
		}
	}
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
		saveData()
	}

	// Get the referer to redirect back to the page where the request came from
	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/"
	}

	http.Redirect(w, r, referer, http.StatusSeeOther)
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
	saveData()
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
		saveData()
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
		saveData()
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
	saveData()
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
		saveData()
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
			saveData()
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

package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// INDEX: Show only tasks that are NOT "backlog" or "done"
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
	for _, task := range tasks {
		if task.Status == "backlog" || task.Status == "done" {
			continue
		}
		if query == "" ||
			strings.Contains(strings.ToLower(task.Title), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(task.Description), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(task.Category), strings.ToLower(query)) {
			filteredTasks = append(filteredTasks, task)
		}
	}

	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/index.html"))
	data := struct {
		Tasks      []Task
		Categories []Category
		Query      string
	}{
		Tasks:      filteredTasks,
		Categories: categories,
		Query:      query,
	}
	err = tmpl.ExecuteTemplate(w, "index", data)
	if err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// BACKLOG: Show only "backlog" tasks
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
	for _, task := range tasks {
		if task.Status != "backlog" {
			continue
		}
		if query == "" ||
			strings.Contains(strings.ToLower(task.Title), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(task.Description), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(task.Category), strings.ToLower(query)) {
			backlogTasks = append(backlogTasks, task)
		}
	}

	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/backlog.html"))
	data := struct {
		Tasks      []Task
		Categories []Category
		Query      string
	}{
		Tasks:      backlogTasks,
		Categories: categories,
		Query:      query,
	}
	err = tmpl.ExecuteTemplate(w, "backlog", data)
	if err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// DONE: Show only "done" tasks
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
	for _, task := range tasks {
		if task.Status != "done" {
			continue
		}
		if query == "" ||
			strings.Contains(strings.ToLower(task.Title), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(task.Description), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(task.Category), strings.ToLower(query)) {
			doneTasks = append(doneTasks, task)
		}
	}

	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/done.html"))
	data := struct {
		Tasks      []Task
		Categories []Category
		Query      string
	}{
		Tasks:      doneTasks,
		Categories: categories,
		Query:      query,
	}
	err = tmpl.ExecuteTemplate(w, "done", data)
	if err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// ADD TASK PAGE: Show form with categories
func addTaskPageHandler(w http.ResponseWriter, r *http.Request) {
	categories, err := getAllCategories()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/add-task.html"))
	data := struct {
		Categories []Category
	}{
		Categories: categories,
	}
	err = tmpl.ExecuteTemplate(w, "add-task", data)
	if err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// ADD TASK: Insert new task into DB
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		t := Task{
			Title:       r.FormValue("title"),
			Category:    r.FormValue("category"),
			Description: r.FormValue("description"),
			Status:      r.FormValue("status"),
			CreatedAt:   time.Now(),
		}
		if t.Status == "" {
			t.Status = "todo"
		}
		if err := insertTask(t); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		log.Printf("Task created: %+v", t)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// DELETE TASK
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	if err := deleteTaskByID(id); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	log.Printf("Task deleted: id=%d", id)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// EDIT TASK
func editTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		idStr := r.FormValue("id")
		id, _ := strconv.Atoi(idStr)
		t := Task{
			ID:          id,
			Title:       r.FormValue("title"),
			Category:    r.FormValue("category"),
			Description: r.FormValue("description"),
			Status:      r.FormValue("status"),
		}
		if err := updateTask(t); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// UPDATE TASK STATUS
func updateTaskStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		idStr := r.FormValue("id")
		id, _ := strconv.Atoi(idStr)
		status := r.FormValue("status")
		t, err := getTaskByID(id)
		if err != nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		t.Status = status
		if err := updateTask(*t); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// CATEGORIES PAGE
func categoryPageHandler(w http.ResponseWriter, r *http.Request) {
	categories, err := getAllCategories()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/categories.html"))
	err = tmpl.ExecuteTemplate(w, "categories", categories)
	if err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// ADD CATEGORY
func addCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		if err := insertCategory(name); err != nil {
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

// DELETE CATEGORY
func deleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if err := deleteCategory(name); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/categories", http.StatusSeeOther)
}

// EDIT CATEGORY
func editCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		oldName := r.FormValue("old_name")
		newName := r.FormValue("new_name")
		if oldName != "" && newName != "" {
			if err := updateCategoryName(oldName, newName); err != nil {
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

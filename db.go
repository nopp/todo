package main

import (
	"database/sql"
	"time"
)

var db *sql.DB

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
		panic(err)
	}
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

func getTaskByID(id int) (*Task, error) {
	row := db.QueryRow("SELECT id, title, category, description, status, created_at FROM tasks WHERE id = ?", id)
	var t Task
	var createdAt string
	err := row.Scan(&t.ID, &t.Title, &t.Category, &t.Description, &t.Status, &createdAt)
	if err != nil {
		return nil, err
	}
	t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &t, nil
}

func insertTask(t Task) error {
	_, err := db.Exec(
		"INSERT INTO tasks (title, category, description, status, created_at) VALUES (?, ?, ?, ?, ?)",
		t.Title, t.Category, t.Description, t.Status, t.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func updateTask(t Task) error {
	_, err := db.Exec(
		"UPDATE tasks SET title = ?, category = ?, description = ?, status = ? WHERE id = ?",
		t.Title, t.Category, t.Description, t.Status, t.ID,
	)
	return err
}

func deleteTaskByID(id int) error {
	_, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}

func insertCategory(name string) error {
	_, err := db.Exec("INSERT OR IGNORE INTO categories (name) VALUES (?)", name)
	return err
}

func deleteCategory(name string) error {
	_, err := db.Exec("DELETE FROM categories WHERE name = ?", name)
	return err
}

func updateCategoryName(oldName, newName string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec("UPDATE categories SET name = ? WHERE name = ?", newName, oldName)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec("UPDATE tasks SET category = ? WHERE category = ?", newName, oldName)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

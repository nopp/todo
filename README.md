# TODO

## Overview

A feature-rich Todo application built with Go and HTML/CSS, designed to help manage tasks efficiently. The application provides an intuitive interface for creating, organizing, and tracking tasks with support for categories, statuses, and different views.

## Features

- **Task Management**: Create, edit, delete, and update task status
- **Categories**: Organize tasks by custom categories
- **Multiple Views**: 
  - Main view for active tasks
  - Backlog view for tasks on hold
  - Done view for completed tasks
- **Search Functionality**: Search across all task fields
- **Responsive Design**: Clean, modern UI built with Material Design principles

## Technology Stack

- **Backend**: Go (Golang)
- **Frontend**: HTML, CSS, Material Design Lite
- **Data Storage**: SQLite database (`data/todo.db`)

## Project Structure

```
todo/
├── data/                  # Data storage directory
│   └── todo.db            # SQLite database file (persisted in PVC if using Kubernetes)
├── templates/             # HTML templates
│   ├── layout.html        # Main layout template
│   ├── index.html         # Home page template
│   ├── backlog.html       # Backlog view template
│   ├── done.html          # Done tasks view template
│   └── categories.html    # Categories management template
├── static/                # Static assets
│   └── style.css          # Application styles
├── k8s/                   # Kubernetes deployment files
│   ├── deployment.yaml    # Kubernetes deployment
│   ├── service.yaml       # Kubernetes service
│   ├── pvc.yaml           # Persistent Volume Claim
├── main.go                # Application entry point and server logic
├── handlers.go            # HTTP handler functions
├── db.go                  # Database connection and helper functions
└── models.go              # Data models (Task, Category)
```

Yes, the database is SQLite, because this system need to be very simple.

## Models

### Task

```go
type Task struct {
    ID          int
    Title       string
    Category    string
    Description string
    Status      string
    CreatedAt   time.Time
}
```
- Represents a to-do item with all its details.

### Category

```go
type Category struct {
    Name string
}
```
- Represents a group or label for tasks.

## Helper Functions

All database operations are handled by helper functions in `db.go`, such as:
- `getAllTasks()`, `getAllCategories()`
- `getTaskByID(id int)`
- `insertTask(t Task)`, `updateTask(t Task)`, `deleteTaskByID(id int)`
- `insertCategory(name string)`, `deleteCategory(name string)`, `updateCategoryName(oldName, newName string)`

These helpers keep your code clean and DRY.

## Logging

The application includes comprehensive logging for:
- Application startup
- Server initialization
- Error handling (template errors, database errors)
- CRUD operations (task/category creation, update, deletion)

You can see logs in your terminal when running the app locally, or via `docker logs`/`kubectl logs` in containerized environments.

## Running the Application

1. **Install Go and SQLite3 driver**

   Make sure you have Go installed. The SQLite3 driver is included as a dependency.

2. **Install dependencies**

   ```sh
   go mod tidy
   ```

3. **Build and run the application**

   ```sh
   go run .
   ```

   The server will start on [http://localhost:8080](http://localhost:8080).

4. **Database**

   - The application will automatically create the SQLite database at `data/todo.db` and initialize the required tables on first run.
   - All tasks and categories are now stored in this SQLite database.

## Kubernetes Deployment

Kubernetes manifests are provided in the `k8s/` directory for deploying the application in a containerized environment. Persistent storage is configured for the SQLite database:

- The `PersistentVolumeClaim` in `pvc.yaml` ensures that the `data/todo.db` file is stored on persistent storage.
- The `deployment.yaml` mounts this PVC at `/app/data` inside the container, so your database is always persisted, even if the pod is restarted or rescheduled.

## Application Features

### Task Status Workflow

The application supports different task statuses:
- **Backlog**: Tasks that are planned but not actively being worked on
- **Todo**: Tasks that are ready to be worked on
- **Doing**: Tasks currently in progress
- **Done**: Completed tasks

### Task Management

- Create tasks with title, category, description, and status
- Edit existing tasks
- Update task status
- Delete tasks
- Filter tasks by different views

### Category Management

- Create custom categories
- Edit category names
- Delete categories
- Categorize tasks for better organization

## Screenshots

![Main View](img/screenshot.png)

---

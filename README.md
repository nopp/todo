# TODO App

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
- **Data Storage**: JSON files (tasks.json, categories.json)

## Project Structure

```
todo_app_material/
├── data/                  # Data storage directory
│   ├── tasks.json         # Tasks data
│   └── categories.json    # Categories data
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
│   └── configmap.yaml     # ConfigMap for environment variables
└── main.go                # Application entry point and server logic
```

## Running the Application

### Local Development

1. Make sure you have Go installed (1.16+ recommended)
2. Clone the repository
3. Run the application:
   ```
   go run main.go
   ```
4. Access the application at `http://localhost:8080`

### Using Docker

```bash
# Build the Docker image
docker build -t todo-app:latest .

# Run the container
docker run -p 8080:8080 -v $(pwd)/data:/app/data todo-app:latest
```

### Kubernetes Deployment

The application can be deployed to Kubernetes using the provided manifests:

```bash
# Apply Kubernetes resources
kubectl apply -f k8s/pvc.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

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

## Logging

The application includes comprehensive logging for:
- Application startup
- Server initialization
- Data operations (load/save)
- Error handling

## Screenshots

![Main View](img/screenshot.png)

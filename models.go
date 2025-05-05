package main

import "time"

type Task struct {
	ID          int
	Title       string
	Category    string
	Description string
	Status      string
	CreatedAt   time.Time
}

type Category struct {
	Name string
}

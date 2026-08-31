package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Done      bool   `json:"done"`
	CreatedAt time.Time  `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}

func loadTasks(filename string) ([]Task, error) {
	data,err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err){
			return []Task{},nil
		}
		return nil,err
	}
	var tasks []Task
	if err := json.Unmarshal(data,&tasks); err != nil{
		return nil,err
	}
	return tasks,nil
}

func saveTasks(filename string,tasks []Task ) error{
	data,err := json.MarshalIndent(tasks,"", "  ")
	if err != nil{
		return err
	}
	return os.WriteFile(filename,data,0644)
}

func addTask(task Task,taskList []Task) []Task {
	taskList = append(taskList,task)
	return taskList
}

func listTasks(taskList []Task) {
	if len(taskList) == 0 {
		fmt.Println("No tasks found. Add one with: go run main.go add \"Task description\"")
		return
	}

	// Format headers with fixed widths
	fmt.Printf("%-4s %-8s %-25s %-18s %-18s\n", "ID", "STATUS", "TITLE", "CREATED AT", "COMPLETED AT")
	fmt.Println("----------------------------------------------------------------------------------")

	for _, t := range taskList {
		status := "[ ]"
		completedStr := "-"

		if t.Done {
			status = "[x]"
			if t.CompletedAt != nil {
				completedStr = t.CompletedAt.Format("02 Jan 15:04")
			}
		}

		createdStr := t.CreatedAt.Format("02 Jan 15:04")

		fmt.Printf("%-4d %-8s %-25s %-18s %-18s\n", t.ID, status, t.Title, createdStr, completedStr)
	}
}

func toggleTask(taskList []Task,ID int) error{
	found := false
	now := time.Now()
	for i:= range taskList{
		if taskList[i].ID == ID{
			taskList[i].Done = !taskList[i].Done
			if taskList[i].Done {
				taskList[i].CompletedAt = &now
			}else{
				taskList[i].CompletedAt = nil
			}
			found = true
			break
		}
	}
	if !found{
		return errors.New("invalid taskID")
	}
	return nil
}
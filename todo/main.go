package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)


func main(){
	filename := "tasks.json"
	taskList ,err := loadTasks(filename)
	if err != nil{
		fmt.Println("Error: ",err)
		return
	}
	
	if len(os.Args) < 2{
		fmt.Println("Usage: task-tracker <add|list|toggle> [arguments]")
		os.Exit(1)
	}
	command := os.Args[1]
	switch command {
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Error: Task description required. Example: add \"Buy milk\"")
			os.Exit(1)
		}
		description := os.Args[2]
		fmt.Println("Adding task:", description)
		newTask := Task{
			ID : len(taskList) + 1,
			Title: description,
			Done : false,
			CreatedAt: time.Now(),
		}
		taskList = addTask(newTask,taskList)
	case "list":
		listTasks(taskList)
		return

	case "toggle":
		if len(os.Args) < 3 {
			fmt.Println("Error: Task ID required. Example: toggle 1")
			os.Exit(1)
		}
		id, err := strconv.Atoi(os.Args[2])
				if err != nil {
					fmt.Println("Error: Task ID must be a valid number")
					os.Exit(1)
				}
		
				if err := toggleTask(taskList, id); err != nil {
					fmt.Println("Error:", err)
					return
				}
				fmt.Printf("Toggled status for task ID: %d\n", id)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
	if err := saveTasks(filename, taskList); err != nil {
			fmt.Println("Error saving tasks:", err)
			return
		}
}




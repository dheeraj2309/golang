# TODO CLI
This is a basic todo CLI app, to understand the basic working of go language
So what it will do (all through CLI)
* We can add the task to the list
* We can mark the tast completed
* We can list the tasks according to their order and percentage of completed

So most basic question is how do you even define a task
We will create a simple struct to store the info abt the task
``` go
type Task struct {
	ID        int    `json:"id"` <- this is a struct tag, working as metadata or extra info for the libs to understand how to handle this field
	Title     string `json:"title"`
	Done      bool   `json:"done"`
	CreatedAt time.Time  `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}
```
## What is serialization?
The data living in the RAM is not directly storable , it is collection of scattered pointers pointing to data in memory
Now the data stored in file is continous,we have to make that complex pointers object as single storable entity
In RAM:
``` 
RAM (Volatile, complex graph of pointers):
Memory Address 0x004F: [Struct Header]
Memory Address 0x0050: ID = 1
Memory Address 0x0054: Pointer -> points to Memory Address 0x99A0 ("Buy Milk")
Memory Address 0x0058: Completed = false
```
This whole entity is converted to json format :
```json
{"id":1,"title":"Buy milk","completed":false}
```
**Also pointers dont remain same they change everytime program runs**

### Flow
* We will keep a json file where all tasks will be stored
* We will load it once in the slide of taskList
* Now all the command Line arguments will be parsed and based on the commands run the corresponding functions
* Final slice will be saved in the same file after exit command

To achieve these following functions are created:
* `loadTasks`
* `saveTasks`
* `addTask`
* `listTasks`
* `toggleTask`

## How to run
* Clone the repo :
```bash
git clone https://github.com/dheeraj2309/todo.git
cd todo
go mod init todo
go run . <arg>
```
Args :
* Creats a new task with the description
```go
go run . add <task description>
```
* list all the tasks
```go
go run . list 
```
* Mark/Unmark the task with entered ID
```go
go run . toggle <task ID>
```
Example :
```go
go run . list
ID   STATUS   TITLE                     CREATED AT         COMPLETED AT
----------------------------------------------------------------------------------
1    [x]      New task                  29 Aug 13:22       29 Aug 13:23
2    [ ]      wake up early             29 Aug 13:26       -
3    [ ]      do the assignments        29 Aug 13:26       -
4    [ ]      finish the project        29 Aug 13:26       -
```


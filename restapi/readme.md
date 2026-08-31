# HTTP Server
In this project , I'm creating a very simple web server that serves the following requests:
* `GET /tasks` -> Fetch all tasks (Returns 200 OK + JSON array)
* `POST /tasks` -> Create a new task (Returns 201 Created + the created JSON object)
* `GET /tasks/{id}` -> Fetch one specific task (Returns 200 OK or 404 Not Found)
* `PATCH /tasks/{id}/toggle` -> Flip the done status and update completion timestamp (Returns 200 OK or 404 Not Found)
* `DELETE /tasks/{id}` -> Remove a task from memory (Returns 204 No Content or 404 Not Found)

## How exactly Go handles HTTP?

```markdown
[ Incoming HTTP Request ]
           │
           ▼
     [ ServeMux ]  <-- The Router (inspects Method: GET/POST and Path: /tasks)
           │
           ├─ (Spawns a new Goroutine / lightweight thread per request)
           ▼
     [ Handler Function ]
      ├── Reads from:  http.Request  (Headers, URL Params, JSON Body)
      └── Writes to:   http.ResponseWriter (Status Codes, Headers, Output JSON)
           │
           ▼
   [ In-Memory Store ] (Slice of Task Structs + Mutex Lock)
```
The Server scales and many users can hit the requests , for each request go creates a separate goroutine running parallely.
This parallel routines try to access the same memory , can cause race conditions causing crash of the system

### Buffering vs Streaming
* **Buffering** -> Basically the method where we first load the complete file into the RAM(here memory is allocated), then using this to create another object to serialize it(here more memory is allocated)
* **Streaming** -> On other hand,a small space is reserved for the incoming TCP packet data and the parser pick up that data and save them on the disk or append in the required data structure , massively saving the space and time
```markdown
Buffering (io.ReadAll + json.Unmarshal):
Network Socket ──────► [ Huge 50MB Buffer in RAM ] ──────► [ Your Struct ]
                             ▲
                       (Wastes lots of RAM)

Streaming (json.NewDecoder):
Network Socket ──[ 4KB chunk ]──► (Decoder) ──────► [ Your Struct ]
                       ▲
            (Reused tiny 4KB window)
```
## Flow
* Requests hit the server
* validate the incoming data
* Call the corresponding function to serve the request being demanded
* return the response

## Two Schemas 
* **Data Transfer Object** ->When for the client is decided what it is allwoed to send to server 
* **Database/Storage Model** -> Another schema is for db to save the data (the entities)<br>Since I'm not using a database to store these tasks , i am using in memory data  
**Now here i can chose two strategies :**
* I can declare global slice and every gorountine works on it , but since it is global it can be modified by any other function in package, also would have to explicitly manage mutex to work on global slice, otherwise race conditions can happen
* Other method is that i wrap the slice in another struct , encapsulating the lock with it 
```go
type TaskStore struct {
    mu     sync.Mutex  // Lock for thread-safety
    tasks  []Task      // The actual storage
    nextID int         // ID generator <- Auto incrementor key for TaskID
}
```
* This works similiarly as in memory database object protected by lock
* Also i can now create methods for it

## HTTP Header 
```http
HTTP/1.1 201 Created                 <-- Status Line (from w.WriteHeader)
Content-Type: application/json       <-- Header (from w.Header().Set)
Date: Sun, 30 Aug 2026 01:01:00 GMT  <-- Header (auto-added by Go)
                                     <-- Blank line (Separates headers from body)
{"id":1,"title":"Buy milk",...}      <-- BODY (written by json.NewEncoder(w).Encode)
```
* **I'm logging all the server requests in json format in log file**
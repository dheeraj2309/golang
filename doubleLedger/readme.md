# Database Consistency
One of the backend principles contains database consistency, how exactly the data is stored, retrieved and managed when many users are trying to access the same data, the data needs to maintain the ACID properties and it shall not be corrupted
To learn this principle I'm making a Double Entry Ledger for transactions. (What?)
Yes, as too complex  as it sounds, I'll break it down in my words

But why exactly we need this big scary thing?

## Single Entry Systems
Basically these are the systems where your transaction event is simply recorded as a single line denoting your expenses or credit
```
[1] Dheeraj bought icecream  -200
[2] Dheeraj received Salary  +1000
```
Now assume in the database , there is a row denoting Dheeraj's account  
Two autopay requests come at the same time , updating the balance of his account  
But there is a problem :  
* T=0ms: Thread A (Netflx) executes `SELECT`. It sees Dheeraj has ₹1000.  
* T=1ms: Thread B (Spotify) executes `SELECT`. It also sees Dheeraj has ₹1000.  
* T=2ms: Thread A calculates 1000 - 200 = 800.  
* T=3ms: Thread B calculates 1000 - 300 = 700.  
* T=4ms: Thread A executes UPDATE. Dheeraj's balance in the DB becomes ₹800.  
* T=5ms: Thread B executes UPDATE. It overwrites the row. Dheeraj's balance becomes ₹700.  
> Thread A and Thread B are two requests by netflix and spotify autopay  
> This is called a Read-Modify-Write Race Condition.

To tackle this row level locking is the solution, we lock the row with exclusive lock of write for Thread A and the thread B is forced to asleep unless the Thread A commits or rollbacks  
But this introduces another problem :)

## Deadlocks
```
(Holds Lock)                  (Waits For)
   +----------------> Amit's Row <----------------+
   |                                              |
   |                                              |
[ Thread A ]                                 [ Thread B ]
   |                                              |
   |                                              |
   +---------------> Priya's Row <----------------+
      (Waits For)                   (Holds Lock)
```
Code naturally locks the sender first, deducts the money, then locks the receiver, and adds the money.
* T=0ms: Amit transfers ₹100 to Priya. Thread A locks Amit's row.  
* T=0ms: Priya transfers ₹50 to Amit. Thread B locks Priya's row.  
* T=1ms: Thread A tries to lock Priya's row to give her the ₹100. It is blocked by Thread B. Thread A goes to sleep.  
* T=1ms: Thread B tries to lock Amit's row to give him the ₹50. It is blocked by Thread A. Thread B goes to sleep.  

Thread A is holding Amit and waiting for Priya. Thread B is holding Priya and waiting for Amit. Neither can ever progress.  

> Eventually, PostgreSQL's internal deadlock detector kicks in (usually after 1 second), realizes what is happening, and forcefully kills one of the transactions to break the tie, throwing an ugly error back to your Go server

**Solution** : 
If the problem is caused by threads crossing paths (Thread A goes Left-to-Right, Thread B goes Right-to-Left), the solution is to force every thread to walk in the exact same direction, regardless of who is sending the money.  

In your Go code, before you open the database transaction, you sort the accounts you need to lock. You can sort them alphabetically by name, or numerically by their UUIDs.  

**Rule: Always lock the smaller ID first, then the larger ID.**

Let's trace the exact same scenario where Amit (ID: 100) and Priya (ID: 200) send money to each other at the exact same millisecond:
* T=0ms: Thread A (Amit sending to Priya) sorts the IDs. It decides to lock 100 first, then 200.  
* T=0ms: Thread B (Priya sending to Amit) sorts the IDs. It decides to lock 100 first, then 200.  
* T=1ms: Both threads try to lock ID 100 (Amit). PostgreSQL gives the lock to Thread A. Thread B goes to sleep.  
* T=2ms: Thread A proceeds to lock ID 200 (Priya). It succeeds instantly because Thread B is asleep and hasn't locked her yet.  
* T=3ms: Thread A finishes the math and executes `COMMIT`. Both rows are unlocked.  
* T=4ms: Thread B wakes up, locks ID 100, then locks ID 200, finishes its math, and commits.  

By simply sorting two strings in Go before touching the database, you completely eliminate deadlocks from your system.  

Now look at where we stand with our single-entry system. You have to write row-level locks on every query, sort UUIDs in application memory, and manually calculate balances in Go. And if you want to charge a 2% platform fee on this transfer, you now have to sort three UUIDs, lock three rows, and write three UPDATE statements perfectly without missing a decimal point.

Can you see how single-entry UPDATE mutations become an unmaintainable nightmare as a financial system grows?

Imagine it is 8:00 PM on a Saturday. One million users are ordering food on Zomato.
If you use a single-entry system, every single one of those one million threads must acquire a lock on Zomato's main revenue account to add the delivery fee.  

Because of the `SELECT` ... `FOR UPDATE` lock, PostgreSQL forces those one million concurrent threads into a single-file line.
Thread 2 cannot give Zomato money until Thread 1 is done. 
Your database grinds to a complete halt, connections time out, and users see a spinning loading wheel on the checkout page. 

If the database crashes right then, the history is gone. You have no idea whose money actually made it to Zomato.

**But wait , what about *Write-Ahead loggers*, don't they already keep the logs of all transactions and are used to restore the consistent state of the database?**

Yes it is a log of transactions but it is for system level recovery and consistency , It cannot be queried for a user , it store transactions ids without binding the imporatant metadata like sender,receiver important for business 

## Project Structure
```
doubleLedger/
├── cmd/
│   └── api/
│       └── main.go       # The absolute entry point. Only wires things up.
├── internal/
│   ├── db/               # Database connection and SQL execution logic
│   ├── handler/          # HTTP API endpoints (request/response parsing)
│   └── service/          # Core business logic (the ledger rules)
├── migrations/           # Your SQL files go here
├── go.mod
└── go.sum
```

## Ledger Database Schema

To build this, we need three tables.
1. `accounts`: To store the UUIDs of the users.
2. `transactions`: The envelope that groups our ledger lines together.
3. `ledger_entries`: The actual append-only lines (the debits and credits).

```sql
-- 1. The Entities
CREATE TABLE accounts (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    cached_balance BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. The Envelope
CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    reference VARCHAR(255) UNIQUE, -- Used for Idempotency later
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. The Immutable Log
CREATE TABLE ledger_entries (
    id UUID PRIMARY KEY,
    transaction_id UUID REFERENCES transactions(id),
    account_id UUID REFERENCES accounts(id),
    amount BIGINT NOT NULL, -- Positive for Credit, Negative for Debit
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```
> **Why no balance field in account?**  
> *Because with double ledger systems balance is not a field, it is a calculated state, for large number of users the balance is cached*

>**What is UUID?**  
>
>UUID stands for Universally Unique Identifier. Instead of relying on PostgreSQL to generate sequential IDs (like 1, 2, 3), a UUID is a massive 128-bit random alphanumeric string (e.g., 550e8400-e29b-41d4-a716-446655440000).
>We use UUIDs in backend systems for two massive reasons:
>* `Pre-generation`: Your Go code can generate the ID locally before it even talks to the database. This is required for Idempotency 
>* `Distributed scaling`: If your startup grows and you split your database across three different servers, auto-incrementing IDs will collide (Server A creates ID 100, Server B creates ID 100). A UUID is so mathematically random that you can generate trillions of them across multiple servers without ever getting a duplicate.  

>**Why no `from_account` and `to_account` in `ledger_entries`?**

If you add from_account and to_account to a transaction, you are hardcoding the assumption that every financial movement only involves exactly two parties.  
What happens when Amit orders a ₹500 meal on Swiggy, and Swiggy charges a ₹50 delivery fee and a ₹10 platform fee?  
If you are locked into from and to, you have to create three separate fake transactions.  
By structuring the ledger_entries table with just an account_id and an amount, a single transaction can have limitless "legs."  
For that Swiggy order, under a single transaction_id, you just append four rows:  
Row 1: [Amit's Wallet, -560]  
Row 2: [Restaurant Escrow, +500]  
Row 3: [Driver Escrow, +50]  
Row 4: [Swiggy Revenue, +10]  

The database constraint simply verifies that -560 + 500 + 50 + 10 = 0. This structure allows you to model infinitely complex financial splits (taxes, commissions, multi-vendor payouts) without altering your schema. 

>**What is Idempotency?**  

Imagine Amit is on a train with a spotty 4G connection. He taps "Pay". The request hits your Go server, the database deducts ₹500, but right before your server sends the "Success" message back, Amit's train enters a tunnel. 

Amit's app times out. He thinks the payment failed, so he taps "Pay" again.  

If your system is dumb, it processes the new request and deducts another ₹500.  

**Idempotency is the architectural guarantee that making the same request multiple times produces the exact same result as making it once.**  

To achieve this, Amit's mobile app generates a UUID when he hits the checkout screen (this is the reference field in our transactions table).  
* **Attempt 1**: Go server receives `Ref: A1B2`, checks the database, sees it is new, and processes the transfer.  
* **Attempt 2**: Go server receives `Ref: A1B2`, asks PostgreSQL to insert it. Because we added a `UNIQUE` constraint on the reference column, PostgreSQL violently rejects it. Your Go server catches that specific error, realizes Amit is retrying, and simply returns "Payment Successful" without touching his balance again.
Now we had to test the concurreny and will hit many API requests at once, doing this on free cloud tier would not help us to experiment as free tiers have API rate limits  
We use **Podman** to intialize a container,on same device so that we can hit as many requets as we want(bounded by laptop hardware)

## What exactly Podman do?
* It creats a child process of the terminal session where it was executed, it standalone process with assigned PCB and system resources , what makes it a container is cgroups and namespaces , which restricts it usage of the resoruces and tricks the application into thinking that it is running on a new fresh device
* We assign a port in the container to list for the database requests
>When you run a command with a port mapping (like -p 5432:5432), you are telling your laptop's OS:
>*Open port 5432 on my actual laptop. If any requests arrive here, instantly forward them through the virtual network bridge directly into the >container's isolated network on its port 5432.*  
* You aren't sending requests to a "podURL"—you are just talking to localhost:5432 on your laptop, and the OS handles routing it into the container's straightjacket.

### Why a connection string to talk to container?
it acts as an instruction manual for your application (like a Python or Node.js script) so it knows exactly what to write inside those packets before sending them.
Here is what the connection string provides that raw packets lack:
* `Authentication (The ID Card)`: A database will not accept anonymous network traffic. The connection string packages your username and password so the very first packets sent prove you have the right to be there.
* `Context (The Specific Room)`: A single PostgreSQL container can host hundreds of entirely different databases. The connection string specifies exactly which one you want (e.g., `dbname=ecommerce_db`), so the server knows where to route your query.
* `The Language (The Wire Protocol)`: You cannot just send plain text like "Give me the user table" over the network. Every database has its own highly specific binary language (*PostgreSQL uses the Postgres Wire Protocol*). By specifying `postgresql://` at the start of the string, you are telling your application's database driver to translate your SQL commands into that exact binary format.  

Without the connection string, you are just blindly knocking on a door. With it, you are handing over a formatted, authenticated request that the database actually understands.
```http
postgresql://user:pass@localhost:5432/ecommerce_db
```
The credentials we set are for the admin of the database of the postgresql, so they are authenticated before touching the database  

## What is a Trigger?  

A trigger is an automated booby trap or a "bouncer" that lives permanently inside the database. You configure it to "listen" for specific events—usually an `INSERT`, `UPDATE`, or `DELETE` on a specific table.  
When that event happens, the database automatically pauses the operation, executes a custom function, and decides whether to let the operation proceed, alter it, or violently reject it.  
The industry standard is to use triggers strictly for data defense. You use them as advanced constraints. The Go application should still handle the business logic, but the trigger sits at the bottom like a concrete floor, ensuring that if the Go code makes a mathematical mistake, the database rejects the corrupted data.

```sql
-- the function
CREATE OR REPLACE FUNCTION check_balance() 
RETURNS TRIGGER AS $$
DECLARE
    total BIGINT;
BEGIN
    -- 'NEW' is a magic variable containing the row you are currently inserting
    SELECT SUM(amount) INTO total 
    FROM ledger_entries 
    WHERE transaction_id = NEW.transaction_id;

    IF total != 0 THEN
        -- This acts like a 'panic' in Go. It kills the whole SQL transaction.
        RAISE EXCEPTION 'Transaction % does not balance.', NEW.transaction_id;
    END IF;

    -- If the math is fine, let the row be inserted.
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- the trigger
CREATE CONSTRAINT TRIGGER ensure_balance
AFTER INSERT ON ledger_entries
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION check_balance();
-- By making it "Deferred", you are telling the database bouncer: 
-- "Hey, I'm about to do a bunch of inserts. Keep your eyes closed until I yell COMMIT. Then you can check my math."
```

## Transfer Alogrithm  
* **The Deterministic Sort**: We take the Sender ID and Receiver ID and sort them alphabetically. This guarantees that if Amit pays Priya, and Priya pays Amit at the exact same millisecond, both Goroutines lock the accounts in the exact same order (Amit first, then Priya). Zero deadlocks.
* **The Safety Net (defer Rollback)**: The absolute first line after opening a transaction is setting up a deferred rollback. If the Go code panics, crashes, or returns an error at any point, the transaction instantly aborts.
* **The Pessimistic Locks**: We execute `SELECT balance ... FOR UPDATE` on the sorted accounts. The Goroutines form an orderly queue at the database level.
* **The Balance Check**: We verify the sender has enough funds. (We calculate the balance from the ledger, or in a hybrid system, read a cached balance).
* **The Append & Commit**: We insert the envelope into transactions, insert the two lines into `ledger_entries`, and call `COMMIT`. The database runs the deferred zero-sum trigger, and if it passes, the money moves permanently.


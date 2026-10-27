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


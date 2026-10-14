### Terminologies
before we began with this project there are some important terms to know to understand things precisely and with clarity
* Concurrency vs Parallelism: Concurrency is about managing many tasks at once, yes one task might had to wait until other is finished, where parallelism is what we try to achieve using concurrency(it is not parallel in real, we try to make the system look like that it is doing many tasks at once). Parallelism is doing true parallel work , when two chefs are cooking at same time , this is true parallel, but when only one chef is cutting onions and stirring the dish as well it is concurrency
* Context Switching: As told in concurrency ,we try to to manage many tasks at once, this switching from one task to another is what happens in CPU , known as context switching, basically the task which CPU was performing is pasued , stored in program counter register , the task need to be performed its instructions are loaded to CPU and new task is carried, in this whole process the context for the CPU changed for which it was performing the task



## Concurrent Log/File Processing
Ok think of the context ,as there is a very large file that you must read to deduce some metrics or generate report or even narrow down the bug, bi
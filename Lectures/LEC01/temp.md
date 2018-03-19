What is a distributed system?
Why distributed?

in mapreduce
What will likely limit the performance?
How does detailed design reduce effect of slow network?
How do they get good load balance?
What about fault tolerance?
Details of worker crash recovery?
Other failures/problems?
  * What if the master gives two workers the same Map() task?
  * What if the master gives two workers the same Reduce() task?
  * What if a single worker is very slow -- a "straggler"?
  * What if a worker computes incorrect output, due to broken h/w or s/w?
  * What if the master crashes?
For what applications *doesn't* MapReduce work well?
How might a real-world web company use MapReduce?
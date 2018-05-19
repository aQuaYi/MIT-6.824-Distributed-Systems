# Lecture 17: Eventual Consistency, Bayou

## 预习

[Managing Update Conflicts in Bayou,a Weakly Connected Replicated Storage Syste](bayou-conflicts.pdf)

## FAQ

1. A lot of Bayou's design is driven by the desire to support disconnected operation. Is that still important today?
1. Doesn't widely-available wireless Internet mean everyone is connected all the time?
1. Bayou supports direct synchronization of one device to another, e.g. over Bluetooth or infrared, without going through the Internet or a server. Is that important?
1. Does anyone use Bayou today? If not, why are we reading this paper?
1. Has the idea of applications performing conflict resolution been used in other distributed systems?
1. Do companies like Dropbox use protocols similar to Bayou?
1. What does it mean for data to be weakly consistent?
1. Is eventual consistency the best you can do if you want to support disconnected operation?
1. It seems like writing dependency checks and merge procedures for a variety of operations could be a tough interface for programmers to handle. Is there anything I'm missing there?
1. Is the primary replica a single point of failure?
1. How do dependency checks detect Write-Write conflicts? The paper says "Such conflicts can be detected by having the dependency check query the current values of any data items being updated and ensure that they have not changed from the values they had at the time the Write was submitted", but I don't quite understand in this case what the expected result of the dependency check is.
1. When are dependency checks called?
1. If two clients make conflicting calendar reservations on partitioned servers, do the dependency checks get called when those two servers communicate?
1. It looks like the logic in the dependency check would take place when you're first inserting a write operation, but you wouldn't find any conflicts from partitioned servers.
1. What are anti-entropy sessions?
1. What is an epidemic algorithm?
1. Why are Write exchange sessions called anti-entropy sessions?
1. In order to know if writes are stabilized, does a server have to contact all other servers?
1. How much time could it take for a Write to reach all servers?
1. In what case is automatic resolution not possible? Does it only depend on the application, or is it the case that for any application, it's possible for automatic resolution to fail?
1. What are examples of good (quick convergence) and not-so-good anti-entropy policies?
1. I don't understand why "tentative deletion may result in a tuple that appears in the committed view but not in the full view." (very beginning of page 8)
1. Bayou introduces a lot of new ideas, but it's not clear which ideas are most important for performance.
1. What kind of information does the Undo Log contain? (e.g. does it contain a snapshot of changed files from a Write, or the reverse operation?) Or is this more of an implementation detail?
1. How is a particular server designated as the primary?
1. What if communication fails in the middle of an anti-entropy session?
1. Does Bayou cache?
1. What are the session guarantees mentioned by the paper?

## 课堂

[讲义](l-bayou.txt)

[FAQ 答案](bayou-faq.txt)

## [作业](6.824 Spring 2018 Paper Questions.html)
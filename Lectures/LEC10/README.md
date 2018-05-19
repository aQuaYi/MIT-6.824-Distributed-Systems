# Lecture 10: Distributed Transactions

## 预习

Read [6.033 book Chapter 9](atomicity_open_5_0.pdf) , just [9.1.5, 9.1.6](9.1.5-9.1.6.pdf), [9.5.2, 9.5.3](9.5.2-9.5.3.pdf), [9.6.3](9.6.3.pdf)

## FAQ

1. How does this material fit into 6.824?
1. Why is it so important for transactions to be atomic?
1. Could one use Raft instead of two-phase commit?
1. In two-phase commit, why would a worker send an abort message, rather than a PREPARED message?
1. Can two-phase locking generate deadlock?
1. Why does it matter whether locks are held until after a transaction commits or aborts?
1. What is the point of the two-phase locking rule that says a transaction isn't allowed to acquire any locks after the first time that it releases a lock?
1. Does two-phase commit solve the dilemma of the two generals described in the reading's Section 9.6.4?
1. Are the locks exclusive, or can they allow multiple readers to have simultaneous access?
1. How should one decide between pessimistic and optimistic concurrency control?
1. What should two-phase commit workers do if the transaction coordinator crashes?
1. Why don't people use three-phase commit, which allows workers to commit or abort even if the coordinator crashes?

## 课堂

[讲义](l-2pc.txt)

[FAQ 答案](chapter9-faq.txt)

## 作业

[6.033 Book](https://ocw.mit.edu/resources/res-6-004-principles-of-computer-system-design-an-introduction-spring-2009/online-textbook/). Read just these parts of Chapter 9: 9.1.5, 9.1.6, 9.5.2, 9.5.3, 9.6.3. The last two sections (on two-phase locking and distributed two-phase commit) are the most important. The Question: describe a situation where Two-Phase Locking yields higher performance than Simple Locking.
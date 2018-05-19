# Lecture 8: Zookeeper Case Study

## 课前阅读

[zookeeper](zookeeper.pdf)

## FAQ

1. Why are only update requests A-linearizable?
1. How does linearizability differ from serializability?
1. What is pipelining?
1. What about Zookeeper's use case makes wait-free better than locking?
1. What does wait-free mean?
1. What is the reason for implementing 'fuzzy snapshots'? How can state changes be idempotent?
1. How does ZooKeeper choose leaders?
1. How does Zookeeper's performance compare to other systems such as Paxos?
1. How does the ordering guarantee solve the race conditions in Section 2.3?
1. How big is the ZooKeeper database? It seems like the server must have a lot of memory.
1. What's a universal object?
1. How does a client know when to leave a barrier (top of page 7)?
1. Is it possible to add more servers into an existing ZooKeeper without taking the service down for a period of time?
1. How are watches implemented in the client library?

## 课堂讲义

[讲义](l-zookeeper.txt)

[FAQ](zookeeper-faq.txt)

## 作业

One use of [Zookeeper](zookeeper.pdf) is a fault-tolerant lock service (see the section "Simple locks" on page 6). Why isn't possible that two clients can acquire the same lock? In particular, how does Zookeeper decide if a client has failed and it can give the lock to some other client?

## LAB 3

[LAB 3 说明](6.824 Lab 3_ Fault-tolerant Key_Value Service.html)
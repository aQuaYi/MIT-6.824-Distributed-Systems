# Lecture 5: Raft (1)

## 课前阅读

[In Search of an Understandable Consensus Algorithm (Extended Version)]( ../LEC03/raft-extended.pdf), to end of Section 5

## FAQ

1. Does Raft sacrifice anything for simplicity?
1. Is raft used in real-world software, or do companies generally roll their own flavor of Paxos (or use a different consensus protocol)?
1. What is Paxos? In what sense is Raft simpler?
1. How long had Paxos existed before the authors created Raft? How widespread is Raft's usage in production now?
1. How does Raft's performance compare to Paxos in real-world applications?
1. Why are we learning/implementing Raft instead of Paxos?
1. Are there systems like Raft that can survive and continue to operate when only a minority of the cluster is active?
1. In Raft, the service which is being replicated is not available to the clients during an election process. In practice how much of a problem does this cause?
1. Are there other consensus systems that don't have leader-election pauses?
1. How are Raft and VMware FT related?
1. Why can't a malicious person take over a Raft server, or forge incorrect Raft messages?
1. The paper mentions that Raft works under all non-Byzantine conditions. What are Byzantine conditions and why could they make Raft fail?
1. In Figure 1, what does the interface between client and server look like?
1. What if a client sends a request to a leader, the the leader crashes before sending the client request to all followers, and the new leader doesn't have the request in its log? Won't that cause the client request to be lost?
1. If there's a network partition, can Raft end up with two leaders and split brain?
1. Suppose a new leader is elected while the network is partitioned, but the old leader is in a different partition. How will the old leader know to stop committing new entries?
1. When some servers have failed, does "majority" refer to a majority of the live servers, or a majority of all servers (even the dead ones)?
1. What if the election timeout is too short? Will that cause Raft to malfunction?
1. Why randomize election timeouts?
1. Can a candidate declare itself the leader as soon as it receives votes from a majority, and not bother waiting for further RequestVote replies?
1. Can a leader in Raft ever stop being a leader except by crashing?
1. When are followers' log entries sent to their state machines?
1. Should the leader wait for replies to AppendEntries RPCs?
1. What happens if more than half of the servers die?
1. Why is the Raft log 1-indexed?
1. When network partition happens, wouldn't client requests in minority partitions be lost?
1. Is the argument in 5.4.3 a complete proof?

## 课堂讲义

[讲义](l-raft.txt)

[FAQ ANSWERS](raft-faq.txt)

## 作业

Suppose we have the scenario shown in the Raft paper's Figure 7: a cluster of seven servers, with the log contents shown. The first server crashes (the one at the top of the figure), and cannot be contacted. A leader election ensues. For each of the servers marked (a), (d), and (f), could that server be elected? If yes, which servers would vote for it? If no, what specific Raft mechanism(s) would prevent it from being elected?
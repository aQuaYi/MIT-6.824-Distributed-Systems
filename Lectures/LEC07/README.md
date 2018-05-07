# Lecture 7: Raft (3) -- Snapshots, Linearizability, Duplicate Detection

## 课前阅读

[spinnaker](spinnaker.pdf)，全文

## FAQ

1. What is timeline consistency?
1. When there is only 1 node up in the cohort, the paper says it鈥檚 still timeline consistency; how is that possible?
1. What are the trade-offs of the different consistency levels?
1. How does Spinnaker implement timeline reads, as opposed to consistent reads?
1. Why are Spinnaker's timeline reads faster than its consistent reads, in Figure 8? 
1. Why do Spinnaker's consistent reads have much higher performance than Cassandra's quorum reads, in Figure 8?
1. What is the CAP theorem about?
1. Where does Spinnaker sit in the CAP scheme?
1. Could Spinnaker use Raft as the replication protocol rather than Paxos?
1. The paper mentions Paxos hadn't previously been used for database replication. What was it used for?
1. What is the reason for logical truncation?
1. What exactly is the key range (i.e. 'r') defined in the paper for leader election?
1. Is there a more detailed description of Spinnaker's replication protocol somewhere?
1. How does Spinnaker's leader election ensure there is at most one leader?
1. Does Spinnaker have something corresponding to Raft's terms?
1. Section 9.1 says that a Spinnaker leader replies to a consistent read without consulting the followers. How does the leader ensure that it is still the leader, so that it doesn't reply to a consistent read with stale data?
1. Would it be possible to replace the use of Zookeeper with a Raft-like leader election protocol?
1. What is the difference between a forced and a non-forced log write?
1. Step 6 of Figure 7 seems to say that the candidate with the longest long gets to be the next leader. But in Raft we saw that this rule doesn't work, and that Raft has to use the more elaborate Election Restriction. Why can Spinnaker safely use longest log?

## 课堂讲义

[讲义](l-spinnaker.txt)

## 作业

Please read the paper's Appendices. In Spinnaker a leader to responds to a client request after the leader and one follower have written a log record for the request on persistent storage. Why is this sufficient to guarantee strong consistency even after the leader or the one follower fail?

(This paper relies on Zookeeper, which we will read later.)
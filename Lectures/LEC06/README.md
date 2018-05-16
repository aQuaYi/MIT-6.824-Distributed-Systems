# Lecture 6: Raft (2)

## 课前阅读

[In Search of an Understandable Consensus Algorithm (Extended Version)]( ../LEC03/raft-extended.pdf), Section 6 to end

## FAQ

1. What are some uses of Raft besides GFS master replication?
1. When raft receives a read request does it still commit a no-op?
1. The paper states that no log writes are required on a read, but then immediately goes on to introduce committing a no-op as a technique to get the committed. Is this a contradiction or are no-ops not considered log 'writes'?
1. I find the line about the leader needing to commit a no-op entry in order to know which entries are committed pretty confusing. Why does it need to do this?
1. How does using the heartbeat mechanism to provide leases (for read-only) operations work, and why does this require timing for safety (e.g. bounded clock skew)?
1. What exactly do the C_old and C_new variables in section 6 (and figure 11) represent? Are they the leader in each configuration?
1. When transitioning from cluster C_old to cluster C_new, how can we create a hybrid cluster C_{old,new}? I don't really understand what that means. Isn't it following either the network configuration of C_old or of C_new? What if the two networks disagreed on a connection?
1. I'm confused about Figure 11 in the paper. I'm unsure about how exactly the transition from 'C_old' to 'C_old,new' to 'C_new' goes. Why is there the issue of the cluster leader not being a part of the new configuration, where the leader steps down once it has committed the 'C_new' log entry? (The second issue mentioned in Section 6)
1. About cluster configuration: During the configuration change time, if we have to stop receiving requests from the clients, then what's the point of having this automated configuration step? Doesn't it suffice to just 1) stop receiving requests 2) change the configurations 3) restart the system and continue?
1. The last two paragraphs of section 6 discuss removed servers interfering with the cluster by trying to get elected even though they're been removed from the configuration. Wouldn't a simpler solution be to require servers to be shut down when they leave the configuration? It seems that leaving the cluster implies that a server can't send or receive RPCs to the rest of the cluster anymore, but the paper doesn't assume that. Why not? Why can't you assume that the servers will shut down right away?
1. How common is it to get a majority from both the old and new configurations when doing things like elections and entry commitment, if it's uncommon, how badly would this affect performance?
1. and how important is the decision to have both majorities?
1. Just to be clear, the process of having new members join as non-voting entities isn't to speed up the process of replicating the log, but rather to influence the election process? How does this increase availability? These servers that need to catch up are not going to be available regardless, right?
1. If the cluster leader does not have the new configuration, why doesn't it just remove itself from majority while committing C_new, and then when done return to being leader? Is there a need for a new election process?
1. How does the non-voting membership status work in the configuration change portion of Raft. Does that server state only last during the changeover (i.e. while c_new not committed) or do servers only get full voting privileges after being fully "caught up"? If so, at what point are they considered "caught up"?
1. When exactly does joint consensus begin, and when does it end? Does joint consensus begin at commit time of "C_{o,n}"?
1. Can the configuration log entry be overwritten by a subsequent leader (assuming that the log entry has not been committed)?
1. How can the "C_{o,n}" log entry ever be committed? It seems like it must be replicated to a majority of "old" servers (as well as the "new" servers), but the append of "C_{o,n}" immediately transitions the old server to new, right?
1. When snapshots are created, is the data and state used the one for the client application? If it's the client's data then is this something that the client itself would need to support in addition to the modifications mentioned in the raft paper?
1. The paper says that "if the follower receives a snapshot that describves a prefix of its log, then log entries covered by the snapshot are deleted but entries following the snapshot are retained". This means that we could potentially be deleting operations on the state machinne.
1. It seems that snapshots are useful when they are a lot smaller than applying the sequence of updates (e.g., frequent updates to a few keys). What happens when a snapshot is as big as the sum of its updates (e.g., each update inserts a new unique key)? Are there any cost savings from doing snapshots at all in this case?
1. Also wouldn't a InstallSnapshot incur heavy bandwidth costs?
1. Is there a concern that writing the snapshot can take longer than the eleciton timeout because of the amount of data that needs to be appended to the log?
1. Under what circumstances would a follower receive a snapshot that is a prefix of its own log?
1. Additionally, if the follower receives a snapshot that is a prefix of its log, and then replaces the entries in its log up to that point, the entries after that point are ones that the leader is not aware of, right?
1. Will those entries ever get committed?
1. How does the processing of InstallSnapshot RPC handle reordering, when the check at step 6 references log entries that have been compacted? Specifically, shouldn't Figure 13 include: 1.5: If lastIncludedIndex < commitIndex, return immediately.  or alternatively 1.5: If there is already a snapshot and lastIncludedIndex < currentSnapshot.lastIncludedIndex, return immediately.
1. What happens when the leader sends me an InstallSnapshot command that is for a prefix of my log, but I've already undergone log compaction and my snapshot is ahead? Is it safe to assume that my snapshot that is further forward subsumes the smaller snapshot?
1. How do leaders decide which servers are lagging and need to be sent a snapshot to install?
1. Is there a concern that writing the snapshot can take longer than the eleciton timeout because of the amount of data that needs to be appended to the log?
1. In actual practical use of raft, how often are snapshots sent?
1. Is InstallSnapshot atomic? If a server crashes after partially installing a snapshot, and the leader re-sends the InstallSnapshot RPC, is this idempotent like RequestVote and AppendEntries RPCs?
1. Why is an offset needed to index into the data[] of an InstallSNapshot RPC, is there data not related to the snapshot? Or does it overlap previous/future chunks of the same snapshot? Thanks!
1. How does copy-on-write help with the performance issue of creating snapshots?
1. What data compression schemes, such as VIZ, ZIP, Huffman encoding, etc. are most efficient for Raft snapshotting?
1. Quick clarification: Does adding an entry to the log count as an executed operation?
1. According to the paper, a server disregards RequestVoteRPCs when they think a current leader exists, but then the moment they think a current leader doesn't exist, I thought they try to start their own election. So in what case would they actually cast a vote for another server?
1. For the second question, I'm still confused: what does the paper mean when it says a server should disregard a RequestVoteRPC when it thinks a current leader exists at the end of Section 6? In what case would a server think a current leader doesn't exist but hasn't started its own election? Is it if the server thinks it hasn't yet gotten a heartbeat from the server but before its election timeout?
1. What has been the impact of Raft, from the perspective of academic researchers in the field? Is it considered significant, inspiring, non-incremental work? Or is it more of "okay, this seems like a natural progression, and is a bit easier to teach, so let's teach this?"
1. The paper states that there are a fair amount of implementations of Raft out in the wild. Have there been any improvement suggestions that would make sense to include in a revised version of the algorithm?

## 课堂

[讲义](l-raft2.txt)

[FAQ 解答](raft2-faq.txt)

## 作业

Could a received InstallSnapshot RPC cause the state machine to go backwards in time? That is, could step 8 in Figure 13 cause the state machine to be reset so that it reflects fewer executed operations? If yes, explain how this could happen. If no, explain why it can't happen.
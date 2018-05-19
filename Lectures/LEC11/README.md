# Lecture 11: FaRM

## FAQ

1. What are some systems that currently uses FaRM?
1. Why did Microsoft Research release this? Same goes for the research branches of Google, Facebook, Yahoo, etc. Why do they always reveal the design of these new systems? It's definitely better for the advancement of tech, but it seems like (at least in the short term) it is in their best interest to keep these designs secret.
1. How does FaRM compare to Thor in terms of performance? It seems easier to understand.
1. Does FaRM really signal the end of necessary compromises in consistency/availability in distributed systems?
1. This paper does not really focus on the negatives of FaRM. What are some of the biggest cons of using FaRM?
1. How does RDMA differ from RPC? What is one-sided RDMA?
1. It seems like the performance of FaRM mainly comes from the hardware. What is the remarking points of the design of FaRM other than harware parts?
1. The paper states that FaRM exploits recent hardware trends, in particular the prolification of UPS to make DRAM memory non-volatile. Have any of the systems we have read about exploited this trend since they were created? For example, do modern implementations of Raft no longer have to worry about persisting non-volatile state because they can be equipped with a UPS?
1. The paper mentions the usage of hardware, such as DRAM and Li-ion batteries for better recovery times and storage. Does the usage of FaRM without the hardware optimizations still provide any benefit over past transactional systems?
1. I noticed that the non-volatility of this system relies on the fact that DRAM is being copied to SSD in the case of power outages. Why did they specify SSD and not allow copying to a normal hard disk? Were they just trying to use the fastest persistent storage device, and therefore did not talk about hard disks for simplicity’s sake?
1. If FaRM is supposed to be for low-latency high-throughput storage, why bother with in-memory solutions instead of regular persistent storage (because the main trade-offs are exactly that -- presistent storage is low-latency high-throughput)?
1. I'm confused about how RMDA in FaRM paper works or how this is crucial to its design. Is it correct that what the paper gives is a design optimized for CPU usage (since they mentioned their design is mostly CPU-bound)? So this would work even if we use SSD instead of in-memory access?
1. What is the distinction between primaries, backups, and configuration managers in FaRM? Why are there three roles?
1. Is FaRM useful or efficient over a range of scales? The authors describe a system that runs $12+/GB for a PETABYTE of storage, not to mention the overhead of 2000 systems and network infrastructure to support all that DRAM+SSD+UPS. This just seems silly-huge and expensive for what they modestly describe as "sufficient to hold the data sets of many interesting applications". How do you even test and verify something like that?
1. 'm a little confused about the difference between the strict serializability and the fact that Farm does not ensure atomicity across reads ( even if committed transactions are serializable ).
1. By bypassing the Kernel, how does FaRM ensure that the read done by RDMA is consistent? What happens if read is done in the middle of a transaction?
1. How do truncations work? When can an a log entry be removed? If one entry is removed by a truncate call, are all previous entries also removed?
1. I'm confused when is it possible to abort in the COMMIT-BACKUP stage. Is this only due to hardware failures?
1. Since this is an optimistic protocol, does it suffer when demand for a small number of resources is high? It seems like this scheme would perform poorly if it had to effectively backtrack every time a transaction had to be aborted.
1. Aren't locks a major hamstring to performance in this system?
1. Figure 7 shows significant decrease in performance when the number of operations crosses 120M . Is it because of the optimistic concurrency protocol, so after some threshold too many transactions get aborted?
1. If yes, then what are some ways a) to make the performance graph flatter? b) to avoid being in 120M+ range?

## 上课

[讲义](l-farm.txt)

[FAQ 答案](farm-faq.txt)

## 课堂作业

[No compromises: distributed transactions with consistency, availability, and performance](farm-2015_cropped.pdf): Suppose there are two FaRM transactions that both increment the same object. They start at the same time and see the same initial value for the object. One transaction completely finishes committing (see Section 4 and Figure 4). Then the second transaction starts to commit. There are no failures. What is the evidence that FaRM will use to realize that it must abort the second transaction? At what point in the Section 4 / Figure 4 protocol will FaRM realize that it must abort?
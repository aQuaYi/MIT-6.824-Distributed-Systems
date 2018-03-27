# 课程 03

## 课前思考

阅读论文 [GFS(2003)](gfs.pdf)，可以思考以下问题:

1. Why is atomic record append at-least-once, rather than exactly once?
1. How does an application know what sections of a chunk consist of padding and duplicate records?
1. How can clients find their data given that atomic record append writes it at an unpredictable offset in the file?
1. The paper mentions reference counts -- what are they?
1. If an application uses the standard POSIX file APIs, would it need to be modified in order to use GFS?
1. How does GFS determine the location of the nearest replica?
1. Does Google still use GFS?
1. Won't the master be a performance bottleneck?
1. How acceptable is it that GFS trades correctness for performance and simplicity?
1. What if the master fails?

问题的答案在[这里](gfs-faq.txt.md)

在阅读[讲义](l-gfs-short.txt.md)请先思考以下问题：

- Why are we reading this paper?
- What is consistency?
- "Ideal" consistency model
- Challenges to achieving ideal consistency
- GFS goals:
- High-level design / Reads / Writes
- Record append
- Housekeeping
- Failures
- Does GFS achieve "ideal" consistency?
- Authors claims weak consistency is not a big problems for apps
- Performance (Figure 3)

## Lab 02

[Lab 02 的上机说明](6.824 Lab 2: Raft.html)

### Lab 02 准备工作

- 阅读 [extended Raft paper](raft-extended.pdf)
- 阅读 [LEC 05](../LEC05/README.md) 和 [LEC 06](../LEC06/README.md)
- 阅读 <http://thesecretlivesofdata.com/raft/>
- 阅读 [Students' Guide to Raft](https://thesquareplanet.com/blog/students-guide-to-raft/)
- 阅读 [Raft Locking Advice](raft-locking.txt.md)
- 阅读 [Raft Structure Advice](raft-structure.txt.md)
- 选读 [Paxos Replicated State Machines as the Basis of a High-Performance Data Store](Bolosky.pdf)

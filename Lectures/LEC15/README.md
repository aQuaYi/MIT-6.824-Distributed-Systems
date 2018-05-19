# Lecture 15: Frangipani

## 预习

[Frangipani:A ScalableDistributedFileSystem](thekkath-frangipani.pdf)

## FAQ

1. Why are we reading this paper?
1. How does Frangipani differ from GFS?
1. Why do the Petal servers in the Frangipani system have a block interface? Why not have file servers (like AFS), that know about things like directories and files?
1. Can a workstation running Frangipani break security?
1. What is Digital Equipment?
1. What does the comment "File creation takes longer..." in Section 9.2 mean?
1. Frangipani is over 20 years old now; what's the state-of-the-art in distributed file systems?
1. The paper says Frangipani only does crash recovery for its own file system meta-data (i-nodes, directories, free bitmaps), but not for users' file content. What does that mean, and why is it OK?
1. What's the difference between the log stored in the Frangipani workstation and the log stored on Petal?
1. How does Petal take efficient snapshots of the large virtual disk that it represents?
1. The paper says that Frangipani doesn't immediately send new log entries to Petal. What happens if a workstation crashes after a system call completes, but before it send the corresponding log entry to Petal?
1. What does it mean to stripe a file? Is this similar to sharding?
1. What is the "false sharing" problem mentioned in the paper?

## 课堂

[讲义](l-frangipani.txt)

[FAQ 答案](frangipani-faq.txt)

## [作业](6.824 Spring 2018 Paper Questions.html)
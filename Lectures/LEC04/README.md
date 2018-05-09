# Lecture 4: Primary/Backup Replication

## 课前阅读

[The Design of a for Practical System Practical Virtual System Machines for Fault-Tolerant Fault-Tolerant Virtual Machines](vm-ft.pdf)

## FAQ

1. The introduction says that it is more difficult to ensure deterministic execution on physical servers than on VMs. Why is this the case?
1. What is a hypervisor?
1. Both GFS and VMware FT provide fault tolerance. How should we think about when one or the other is better?
1. How do Section 3.4's bounce buffers help avoid races?
1. What is "an atomic test-and-set operation on the shared storage"?
1. How much performance is lost by following the Output Rule?
1. What if the application calls a random number generator? Won't that yield different results on primary and backup and cause the executions to diverge?
1. How were the creators certain that they captured all possible forms of non-determinism?
1. What happens if the primary fails just after it sends output to the external world?
1. Section 3.4 talks about disk I/Os that are outstanding on the primary when a failure happens; it says "Instead, we re-issue the pending I/Os during the go-live process of the backup VM." Where are the pending I/Os located/stored, and how far back does the re-issuing need to go?
1. How secure is this system?
1. Is it reasonable to address only the fail-stop failures? What are other type of failures?

## 课堂讲义

[讲义](l-vm-ft.txt)

[FAQ ANSWERS](vm-ft-faq.txt)

## 作业

How does [VM FT](vm-ft.pdf) handle network partitions? That is, is it possible that if the primary and the backup end up in different network partitions that the backup will become a primary too and the system will run with two primaries?
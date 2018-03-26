GFS FAQ
Q: Why is atomic record append at-least-once, rather than exactly
once?

It is difficult to make the append exactly once, because a primary
would then need to keep state to perform duplicate detection. That
state must be replicated across servers so that if the primary fails,
this information isn't lost. You will implement exactly once in lab
3, but with more complicated protocols that GFS uses.

Q: How does an application know what sections of a chunk consist of
padding and duplicate records?

A: To detect padding, applications can put a predictable magic number
at the start of a valid record, or include a checksum that will likely
only be valid if the record is valid. The application can detect
duplicates by including unique IDs in records. Then, if it reads a
record that has the same ID as an earlier record, it knows that they
are duplicates of each other. GFS provides a library for applications
that handles these cases.

Q: How can clients find their data given that atomic record append
writes it at an unpredictable offset in the file?

A: Append (and GFS in general) is mostly intended for applications
that read entire files. Such applications will look for every record
(see the previous question), so they don't need to know the record
locations in advance. For example, the file might contain the set of
link URLs encountered by a set of concurrent web crawlers. The
file offset of any given URL doesn't matter much; readers just want to
be able to read the entire set of URLs.

Q: The paper mentions reference counts -- what are they?

A: They are part of the implementation of copy-on-write for snapshots.
When GFS creates a snapshot, it doesn't copy the chunks, but instead
increases the reference counter of each chunk. This makes creating a
snapshot inexpensive. If a client writes a chunk and the master
notices the reference count is greater than one, the master first
makes a copy so that the client can update the copy (instead of the
chunk that is part of the snapshot). You can view this as delaying the
copy until it is absolutely necessary. The hope is that not all chunks
will be modified and one can avoid making some copies.

Q: If an application uses the standard POSIX file APIs, would it need
to be modified in order to use GFS?

A: Yes, but GFS isn't intended for existing applications. It is
designed for newly-written applications, such as MapReduce programs.

Q: How does GFS determine the location of the nearest replica?

A: The paper hints that GFS does this based on the IP addresses of the
servers storing the available replicas. In 2003, Google must have
assigned IP addresses in such a way that if two IP addresses are close
to each other in IP address space, then they are also close together
in the machine room.

Q: Does Google still use GFS?

A: GFS is still in use by Google and is the backend of other storage
systems such as BigTable. GFS's design has doubtless been adjusted
over the years since workloads have become larger and technology has
changed, but I don't know the details. HDFS is a public-domain clone
of GFS's design, which is used by many companies.

Q: Won't the master be a performance bottleneck?

A: It certainly has that potential, and the GFS designers took trouble
to avoid this problem. For example, the master keeps its state in
memory so that it can respond quickly. The evaluation indicates that
for large file/reads (the workload GFS is targeting), the master is
not a bottleneck. For small file operations or directory operations,
the master can keep up (see 6.2.4).

Q: How acceptable is it that GFS trades correctness for performance
and simplicity?

A: This a recurring theme in distributed systems. Strong consistency
usually requires protocols that are complex and require chit-chat
between machines (as we will see in the next few lectures). By
exploiting ways that specific application classes can tolerate relaxed
consistency, one can design systems that have good performance and
sufficient consistency. For example, GFS optimizes for MapReduce
applications, which need high read performance for large files and are
OK with having holes in files, records showing up several times, and
inconsistent reads. On the other hand, GFS would not be good for
storing account balances at a bank.

Q: What if the master fails?

A: There are replica masters with a full copy of the master state; an
unspecified mechanism switches to one of the replicas if the current
master fails (Section 5.1.3). It's possible a human has to intervene
to designate the new master. At any rate, there is almost certainly a
single point of failure lurking here that would in theory prevent
automatic recovery from master failure. We will see in later lectures
how you could make a fault-tolerant master using Raft.

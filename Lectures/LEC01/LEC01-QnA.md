# 课堂问题

## 常规问题

What is a distributed system?

```text
1. 多个单机系统
2. 通过网络连接器起来进行通信
3. 相互协作共同完成一项任务
```

Why distributed?

```text
1. 单机系统处理能力存在上限
```

## MapReduce 相关问题

What will likely limit the performance?

```text
网络带宽
```

How does detailed design reduce effect of slow network?

```text
3.4 locality
利用 GFS 把输入数据保存在处理数据的计算机集群上。数据被分为了多个 64MB 的数据包，在不同的及其上保存多个副本。MR master 在分配任务的时候，会尽可能地让 worker 与其需要处理的数据在同一个局域网中，以便读。所以，大部分时候，MR 数据只会在局域网内部传递，不会消耗带宽。
```

How do they get good load balance?

What about fault tolerance?

```text
1. worker A 失效：规定时间内，A 没有响应 master 的应答
  1.1 A 所有已经完成的 map 任务都会被重新安排到其他 worker 执行 → map 任务的结果,仅保存在 A 上，A 失效后，执行 reduce 任务的 worker 无法读取结果，所以，由 B 重新执行 map 任务，任务结果会保存在 B 上。→ master 通知所有正在执行 reduce 任务的 worker，以前放在 A 上的结果，现在在 B 上了。
  1.2 A 正在执行执行的 map 或 reduce 任务，会被重新安排到其他 worker 执行。
  1.3 A 所有已经完成的 reduce 任务 不会 被重新执行 → reduce 任务的结果保存在 GFS 上，而非 A 上。所以 不 需要重新执行。
2. master 失效：规定时间内，master 没有响应。
  2.1 通知 client 任务失败。
3. 失效存在的语义表示
  3.1 原子操作
    3.1.1 第一个被完成的 map 任务，才会被记录在 master 中
    3.1.2 reduce 任务的结果重命名必须是原子操作，才能保证最终结果是有同一个 reduce任务完成的。
```

Details of worker crash recovery?

Other failures/problems?

What if the master gives two workers the same Map() task?
What if the master gives two workers the same Reduce() task?
What if a single worker is very slow -- a "straggler"?
What if a worker computes incorrect output, due to broken h/w or s/w?
What if the master crashes?

For what applications *doesn't* MapReduce work well?

How might a real-world web company use MapReduce?
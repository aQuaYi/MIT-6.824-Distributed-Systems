# Introduction 简介

## 预习材料

- [MapReduce](mapreduce.pdf)

## 课堂讲义

在阅读 [LEC01 讲义原文](l01.txt.md) 前，请先思考以下问题：

### 普遍问题

- What is a distributed system?
- Why distributed?

### MapReduce 相关问题

- What will likely limit the performance?
- How does detailed design reduce effect of slow network?
- How do they get good load balance?
- What about fault tolerance?
- Details of worker crash recovery?
- Other failures/problems?
  - What if the master gives two workers the same Map() task?
  - What if the master gives two workers the same Reduce() task?
  - What if a single worker is very slow -- a "straggler"?
  - What if a worker computes incorrect output, due to broken h/w or s/w?
  - What if the master crashes?
- For what applications *doesn't* MapReduce work well?
- How might a real-world web company use MapReduce?

## Lab

[Lab 1: MapReduce](../../src/mapreduce) 的 [上机说明](6.824-Lab1-MapReduce.html)
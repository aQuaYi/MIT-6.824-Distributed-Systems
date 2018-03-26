# 课程 02

## 预习

- 完成[Go 语言之旅](https://tour.golang.org/),[中文版](https://tour.go-zh.org/)
- 阅读[常见问题清单](tour-faq.txt.md)
- 了解 [Paper Question](6.824 Spring 2018 Paper Questions.html) 规则
- 了解 [Go rpc 标准库](https://golang.org/pkg/net/rpc/)，[中文版](https://go-zh.org/pkg/net/rpc/)

## 课堂讲义

阅读[课堂讲义](l-rpc.txt.md)，请先思考以下问题：

- Most commonly-asked question: Why Go?
- What is Threads?
- Why threads?
- How many threads in a program?
- Threading challenges?
- What is a crawler?
- Crawler challenges?
- When to use sharing and locks, versus channels?
- RPC problem: what to do about failures?
- What does a failure look like to the client RPC library?
- What if an at-most-once server crashes and re-starts?

## 课后作业

我为 [crawler](crawler/crawler.go) 和 [kv](kv/kv.go) 添加了注释，[这个压缩包](LEC02-source-code.zip) 中保存了原始的代码。
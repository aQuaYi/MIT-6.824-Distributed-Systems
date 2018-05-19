# Lecture 20: Dynamo

## 预习

[Dynamo: Amazon’s Highly Available Key-value Store](dynamo.pdf)

## FAQ

Q: What's a typical number of nodes in a Dynamo instance? Large enough that vector clocks will become impractical large? 
Q: How can deleted items resurface in a shopping cart (Section 4.4)? 
Q: How does Dynamo recover from permanent failure -- what is anti-entropy using Merkle trees? 
Q: What are virtual nodes? 
Q: Will Dynamo's use of a DHT prevent it from scaling? 
Q: What's a gossip-based protocol?

## 课堂

[讲义](l-dynamo.txt)

[FAQ 答案](dynamo-faq.txt)

## [作业](6.824 Spring 2018 Paper Questions.html)
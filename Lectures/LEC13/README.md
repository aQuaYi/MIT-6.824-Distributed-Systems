# Lecture 13: Naiad Case Study

## 预习

[Naiad: A Timely Dataflow System](naiad_cropped.pdf)

## FAQ

1. Is Naiad/timely dataflow in use in production environments anywhere? If not, and what would be barriers to adoption?
1. The paper says that Naiad performs better than Spark. In what cases would one prefer which one over the other?
1. What was the reasoning for implmenting Naiad in C#? It seems like this language choice introduced lots of pain, and performance would be better in C/C++.
1. I'm confused by the OnRecv/OnNotify/SendBy/NotifyAt API. Does the user need to tell the system when to notify another vertex?
1. What do the authors mean when they state systems like Spark "requires centralized modifications to the dataflow graph, which introduce substantial overhead"? Why does Spark perform worse than Naiad for iterative computations?

## 课堂

[讲义](l-naiad.txt)

[FAQ 答案](naiad-faq.txt)

## [作业](6.824 Spring 2018 Paper Questions.html)
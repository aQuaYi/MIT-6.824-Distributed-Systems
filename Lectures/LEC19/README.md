# Lecture 19: P2P, DHTs, and Chord

## 预习

[Chord:A Scalable Peer-to-peerLookupServiceforInternetApplications](stoica-chord.pdf) 和 [Trackerless Bittorrent](bep_0005.rst_post.html)

## FAQ

1. Is hashing across machines a good way to get load balanced sharding? Why not explicitly divide up the key space so it's evenly split?
1. Does BitTorrent use Chord?
1. If you want to add fault-tolerance to a Chord-based system should you replicate each Chord node using Raft?
1. What if Chord DHT nodes are malicious?
1. Is Chord used anywhere in practice?
1. Could the performance be improved if the nodes knew more about network locality?
1. Is it possible to design a DHT in which lookups take less than log(N) hops?
1. Does consistent hashing of keys still guarantee load balanced nodes if keys are not evenly distributed?
1. In the case of concurrent joins and failures, Chord pauses when a get fails to find the key it was looking for. If there's constant activity, how can Chord distinguish between the system not being stable and the key not actually existing?
1. If I introduce a malicious peer in Chord that keeps returning wrong values or inexistent addresses how disruptive can it be to the whole DHT? How does Bittorrent deal with particular issue?
1. Why isn’t there a danger of improper load balancing if some keys are simply used more than others?

## 课堂

[讲义](l-dht.txt)

[FAQ 答案](chord-faq.txt)

## [作业](6.824 Spring 2018 Paper Questions.html)
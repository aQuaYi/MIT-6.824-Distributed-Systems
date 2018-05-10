package mutualexclusion

import (
	"math/rand"
	"time"
)

type request struct {
	time    int // request 的时间
	process int // request 的 process
}

type process struct {
	me               int
	clock            *clock
	resource         *resource
	peers            []chan message
	requestQueue     []request
	sentTime         []int         // 给各个 process 中发送的消息，所携带的最后时间
	receiveTime      []int         // 从各个 process 中收到的消息，所携带的最后时间
	minReceiveTime   int           // lastReceiveTime 中的最小值
	toCheckRule5Chan chan struct{} // 每次收到 message 后，都靠这个 chan 来通知检查此 process 是否已经满足 rule 5，以便决定是否占有 resource
	isTaken          bool          // process 正在占用资源
	takenCounterDown int           // 每个 preocess 想要占用 resource 的次数
}

func newProcess(me, takenTimes int, r *resource, peers []chan message) *process {
	rq := make([]request, 1, len(peers)*takenTimes*2)
	rq[0] = request{
		time:    -1,
		process: 0,
	}
	p := &process{
		me:               me,
		clock:            newClock(),
		resource:         r,
		peers:            peers,
		requestQueue:     rq,
		sentTime:         make([]int, len(peers)),
		receiveTime:      make([]int, len(peers)),
		minReceiveTime:   0,
		toCheckRule5Chan: make(chan struct{}, 1),
		takenCounterDown: takenTimes,
	}

	return p
}

func (p *process) requestLoop() {
	for {
		if p.isCounterDown() {
			wg.Done()
			return
		}

		if p.isTaken {
			go p.release()
		} else {
			go p.request()
		}

		timeout := time.Duration(100+rand.Intn(900)) * time.Millisecond
		time.Sleep(timeout)
	}
}

func (p *process) request() {

	return
}

func (p *process) release() {

	return
}

func (p *process) isCounterDown() bool {
	return p.takenCounterDown == 0 && !p.isTaken
}

func (p *process) messageLoop() {
	msgChan := p.peers[p.me]
	for {
		msg := <-msgChan
		switch msg.msgType {
		case requestResource:
		case releaseResource:
		case acknowledgment:
		}
	}
}

package mutualexclusion

type message struct {
	msgType msgType

	time        int // 发送 message 时， process.clock 的时间
	senderID    int // message 发送方的 ID
	requestTime int // releaseResource 需要填写以前的 requestTime
}

type msgType int

// 枚举了 message 的所有类型
const (
	// REQUEST_RESOURCE 请求资源
	requestResource msgType = iota
	releaseResource
	acknowledgment
)

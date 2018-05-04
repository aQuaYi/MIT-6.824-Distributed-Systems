package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
)

const (
	// OK 表示服务器存在 key
	OK = "OK"
	// ErrNoKey 表示服务器存在 key
	ErrNoKey = "ErrNoKey"
)

//
// main
//

func main() {
	// 启动服务器
	server()

	// 在服务器的 kv.data 中添加 ("subject", "6.824")
	put("subject", "6.824")
	fmt.Printf("Put(subject, 6.824) done\n")

	// 从服务器中，读取 "subject" 对应的值
	fmt.Printf("get(\"subject\") -> \"%s\"\n", get("subject"))

	// 从服务器中，读取 不存在的 key 对应的值
	fmt.Printf("get(\"NoExist\") -> \"%s\"\n", get("NoExist"))

}

//
// Server
//

// KV 是键值对
// 其实例会在 rpc 中注册，供客户端远程访问
type KV struct {
	mu   sync.Mutex
	data map[string]string
}

// 启动服务器
func server() {
	// 创建 *KV 对象
	kv := new(KV)
	kv.data = map[string]string{}

	// 生成 rpc 服务器
	rpcs := rpc.NewServer()
	// 把 kv 对象注册入 rpc 服务器
	rpcs.Register(kv)
	// 开启 rpcs 的监听功能
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}

	// 启动监听处理 goroutine
	go func() {
		for {
			// l 作为 *TCPListener
			// 会等待被访问，
			// 被访问时会生成一个 Conn
			conn, err := l.Accept()
			if err != nil {
				// 生成的 conn 有问题，放弃这个链接
				break
			}
			// 使用 rpcs.ServeConn 处理刚刚生成的 conn
			go rpcs.ServeConn(conn)
		}

		l.Close()
	}()

}

// Err 定义错误类型
type Err string

// GetArgs : Get 命令输入参数
type GetArgs struct {
	Key string
}

// GetReply : Get 命令返回结果的结构体
type GetReply struct {
	Err   Err
	Value string
}

// Get 是 *KV 为 rpc 提供的获取 key 对应的值的方法
func (kv *KV) Get(args *GetArgs, reply *GetReply) error {
	// 获取之前，先锁定
	kv.mu.Lock()
	defer kv.mu.Unlock()

	// 从 kv.data 中获取值
	val, ok := kv.data[args.Key]
	if ok {
		reply.Err = OK
		reply.Value = val
	} else {
		reply.Err = ErrNoKey
		reply.Value = ""
	}
	return nil
}

// PutArgs : Put 输入命令参数
type PutArgs struct {
	Key   string
	Value string
}

// PutReply : Put 命令返回结果的结构体
type PutReply struct {
	Err Err
}

// Put 是 *KV 为 rpc 提供的放入 (args.key, args.val) 的方法
func (kv *KV) Put(args *PutArgs, reply *PutReply) error {
	// 放入之前，先锁定
	kv.mu.Lock()
	defer kv.mu.Unlock()

	// 在 kv.data 中添加值
	kv.data[args.Key] = args.Value
	reply.Err = OK
	return nil
}

//
// Client
// 客户端
//

// 连接上服务器，返回一个 *rpc.Client
func connect() *rpc.Client {
	client, err := rpc.Dial("tcp", ":1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	return client
}

// 从服务器获取 key 对应的值
func get(key string) string {
	// 先连上服务器
	client := connect()
	// 构建 Get 的输入参数
	args := GetArgs{key}
	// 构建 Get 的应答
	reply := GetReply{}
	// 远程访问
	err := client.Call("KV.Get", &args, &reply)
	if err != nil {
		log.Fatal("error:", err)
	}
	// 关闭客户端
	client.Close()
	// 返回结果
	return reply.Value
}

// 往服务器存放数据 (key, val)
func put(key string, val string) {
	// 连接上服务器
	client := connect()
	// 构建 Put 的输入参数
	args := PutArgs{"subject", "6.824"}
	// 构建 Put 的应答参数
	reply := PutReply{}
	// 远程访问服务器
	err := client.Call("KV.Put", &args, &reply)
	if err != nil {
		log.Fatal("error:", err)
	}
	// 关闭服务器
	client.Close()
}

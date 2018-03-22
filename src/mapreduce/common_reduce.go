package mapreduce

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"sort"
)

func doReduce(
	jobName string, // the name of the whole MapReduce job
	reduceTask int, // which reduce task this is
	outFile string, // write the output here
	nMap int, // the number of map tasks that were run ("M" in the paper)
	reduceF func(key string, values []string) string,
) {

	// doReduce manages one reduce task: it should read the intermediate
	// files for the task, sort the intermediate key/value pairs by key,
	// call the user-defined reduce function (reduceF) for each key, and
	// write reduceF's output to disk.

	// You'll need to read one intermediate file from each map task;
	// reduceName(jobName, m, reduceTask) yields the file
	// name from map task m.

	// Your doMap() encoded the key/value pairs in the intermediate
	// files, so you will need to decode them. If you used JSON, you can
	// read and decode by creating a decoder and repeatedly calling
	// .Decode(&kv) on it until it returns an error.

	// You may find the first example in the golang sort package
	// documentation useful.

	// reduceF() is the application's reduce function. You should
	// call it once per distinct key, with a slice of all the values
	// for that key. reduceF() returns the reduced value for that key.

	// You should write the reduce output as JSON encoded KeyValue
	// objects to the file named outFile. We require you to use JSON
	// because that is what the merger than combines the output
	// from all the reduce tasks expects. There is nothing special about
	// JSON -- it is just the marshalling format we chose to use. Your
	// output code will look something like this:

	// enc := json.NewEncoder(file)
	// for key := ... {
	// 	enc.Encode(KeyValue{key, reduceF(...)})
	// }
	// file.Close()

	// 从 m*r 个 Map 任务生成的中间文件中，读取此 Reduce 任务需要的 m 个文件

	// Reduce 任务需要读取 Map 任务生成的 intermediate file 作为输入
	// kvs 是存在这些输入的容器
	kvs := make([]*KeyValue, 0, 2048)
	// 每个 Map 任务，都为此 Reduce 任务生成了一个 intermediate file
	// 所以，此 Reduce 任务总共需要读取 nMap 个文件
	for m := 0; m < nMap; m++ {
		// 打开需要读取的文件
		filename := reduceName(jobName, m, reduceTask)
		file, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}

		// 创建流式解码器
		dec := json.NewDecoder(file)
		for {
			var kv KeyValue
			// dec 读取一行，并解码到 kv 中
			err := dec.Decode(&kv)
			if err != nil {
				if err == io.EOF {
					// 读取到末尾
					break
				}
				log.Fatal(err)
			}
			// 把新获取的值保存到 kvs 中
			kvs = append(kvs, &kv)
		}

		file.Close()
	}

	// kvs 按照 key 的升序排列
	sort.Slice(kvs, func(i int, j int) bool {
		return kvs[i].Key < kvs[j].Key
	})

	// 集合所有的 values
	values := make([]string, len(kvs))
	for i := range values {
		values[i] = kvs[i].Value
	}

	// 创建输出文件
	file, err := os.Create(outFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 创建流式编码器，每次被编码的内容，都会输出到 file 的一个新行
	enc := json.NewEncoder(file)

	// 对每个 key 都运用 reduceF
	for i := range kvs {
		key := kvs[i].Key
		err := enc.Encode(KeyValue{key, reduceF(key, values)})
		if err != nil {
			log.Fatal(err)
		}
	}
}

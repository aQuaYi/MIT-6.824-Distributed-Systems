package mapreduce

import (
	"fmt"
	"sync"
)

//
// schedule() starts and waits for all tasks in the given phase (mapPhase
// or reducePhase). the mapFiles argument holds the names of the files that
// are the inputs to the map phase, one per map task. nReduce is the
// number of reduce tasks. the registerChan argument yields a stream
// of registered workers; each item is the worker's RPC address,
// suitable for passing to call(). registerChan will yield all
// existing registered workers (if any) and new ones as they register.
//
func schedule(jobName string, mapFiles []string, nReduce int, phase jobPhase, registerChan chan string) {
	var ntasks int
	var nOther int // number of inputs (for reduce) or outputs (for map)
	switch phase {
	case mapPhase:
		ntasks = len(mapFiles)
		nOther = nReduce
	case reducePhase:
		ntasks = nReduce
		nOther = len(mapFiles)
	}

	fmt.Printf("Schedule: %d %v tasks (%d I/Os)\n", ntasks, phase, nOther)

	// All ntasks tasks have to be scheduled on workers. Once all tasks
	// have completed successfully, schedule() should return.
	//
	// TODO: Your code here (Part III, Part IV).
	//
	// 要等待所有的任务完成后，才能结束这个函数，所以，添加 wg
	var wg sync.WaitGroup
	wg.Add(ntasks)

	for i := 0; i < ntasks; i++ {
		// 从 registerChan 获取服务器的地址
		srv := <-registerChan

		fmt.Printf("开始 %v 阶段的第 %d 个任务: %s\n", phase, i, srv)

		// 为将要执行的任务准备相关参数
		args := DoTaskArgs{
			JobName:       jobName,
			Phase:         phase,
			TaskNumber:    i,
			NumOtherPhase: nOther,
		}
		if phase == mapPhase {
			// 只有 mapPhase 才需要设置 .File 属性
			args.File = mapFiles[i]
		}

		go func() {
			// 把任务发送给 srv
			call(srv, "Worker.DoTask", args, nil)
			// 任务结束
			wg.Done()
			// 任务结束后，srv 闲置，srv 进入 registerChan 等待下一次分配任务
			registerChan <- srv
		}()
	}

	wg.Wait()

	fmt.Printf("Schedule: %v done\n", phase)
}

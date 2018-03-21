package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	// 打开文件，创建buffered reader
	file, err := os.Open("mrtmp.test-0-0")
	if err != nil {
		log.Fatal(err)
	}
	br := bufio.NewReader(file)
	res := make([]KeyValue, 12)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		fmt.Println(string(a))
	}

	fmt.Println("Done~~~")
}

type KeyValue struct {
	Key   string
	Value string
}

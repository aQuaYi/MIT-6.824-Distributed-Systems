package main

import (
	"fmt"
)

func main() {
	c := make(chan struct{})
	c2 := make(chan struct{})
	a := 1
	go func() {
		<-c
		a = 2
		c2 <- struct{}{}
	}()
	a = 3
	c <- struct{}{}
	<-c2
	fmt.Println("a is ", a)
}

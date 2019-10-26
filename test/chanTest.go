package main

import (
	"fmt"
	"strconv"
	"time"
)

var ch = make(chan string)
var chSize = make(chan int)


func testFunction(num int)  {
	for {
		ch <- "this is number:"+strconv.Itoa(num)
		time.Sleep(100)
	}
}
func Producer()  {
	for i:=0;i<150000;i++ {
		go testFunction(i)
	}
}

func Consumer()  {
	go func() {
		for {
			str := <- ch
			fmt.Println("Get string for chan:",str)
		}
	}()

}

func GetChanSize()  {
	go func() {
		for {
			chSize <- len(ch)
			time.Sleep(10000)
		}
	}()
}

func PrintChanSize()  {
	size := <- chSize
	for {
		fmt.Println("Channel size is:",size)
		time.Sleep(10000)
	}
}

func main()  {
	Producer()
	Consumer()
	GetChanSize()
	PrintChanSize()
}


package main

import (
	"./pkg"
	"encoding/json"
	"fmt"
	"log"
)

var ch = make(chan string)

func activeCheck () {
	var data pkg.MajorData

	var values []pkg.MinorData
	var value pkg.MinorData
	value = pkg.MinorData{
		Host:  "host_test",
		Key:   "key_test[\"{$URL}\",\"github\",\"{$HOST}\",\"space_use\"]",
		Value: 99.87,
		Clock: 1566481943,
	}
	values = append(values, value)
	data = pkg.MajorData{
		Request: "agent data",
		Data: values ,
	}
	jsonBytes,err := json.Marshal(data)
	if err != nil {
		log.Println(err)
	}
	jsonString := string(jsonBytes)
	if err != nil {
		fmt.Println(err)
	}
	data2 := []byte(jsonString)
	res,err := pkg.DataSender(data2)
	if err != nil {
		fmt.Println("error",err)
	}
	ch <- res
}

func main() {
	go activeCheck()
	res := <-ch
	fmt.Println(res)
}

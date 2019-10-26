/*******************************************************************************
* FileName:  logTest.go
* Author: Victor
* Date: 2019/08/23 21:09
* Description:
* Project: zabbix_agent
*******************************************************************************/
package main


import (
	"fmt"
	"strconv"
	"strings"
)

type severity int32
const (
	infoLog severity = iota
	warningLog
	errorLog
	fatalLog
	severityNum = 4
)
var severityByName = []string{
	infoLog: "INFO",
	warningLog:"WARNING",
	errorLog:"ERROR",
	fatalLog:"FATAL",}

func (s *severity) Get()interface{} {
	return *s
}

func (s *severity) Set(val string) error{
	var threholds severity
	if v,ok:= severityByName1(val);ok{
		threholds=v
	}else{
		if v,err:=strconv.Atoi(val);err !=nil {
			return err
		}else {
			threholds=severity(v)}
	}
	fmt.Println(threholds)
	return nil
}

func severityByName1(s string) (severity,bool){
	s=strings.ToUpper(s)
	for i ,v :=range severityByName {
		if v == s {
			return severity(i),true
		}
	}
	return 0,false
} //func  severity01(s severity) (severity,bool){//// for i ,_ := range severityByName{//    if severity(i) == s {//       return severity(i) ,true//    }// }//// return 0 ,false////}
func main(){
	var logLevel severity =2
	fmt.Println(logLevel.Get())
	if err:=logLevel.Set("10");err != nil{
		fmt.Println(err)
	}else{
		fmt.Println("set Ok")
	}
}


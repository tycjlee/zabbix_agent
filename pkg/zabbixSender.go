package pkg

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
)


func DataSender(data []byte) (string,error){
	conf := Config()
	address := conf.Server.Ip+":"+strconv.Itoa(conf.Server.Port)
	fmt.Println(address)
	// ,conf.Server_ip+":"+string(conf.Server_port))
	conn,err := net.Dial("tcp",address)
	if err != nil {
		return "连接失败",err
	}
	zbxHeader := []byte("ZBXD\x01")
	zbxHeaderLength := len(zbxHeader)+8
	dataLength := len(data)
	msgArray := make([]byte, zbxHeaderLength+dataLength)[:0]
	msgArray = append(msgArray, zbxHeader...)
	byteBuff := make([]byte, 8)
	binary.LittleEndian.PutUint64(byteBuff, uint64(dataLength))
	msgArray = append(msgArray, byteBuff...)
	msgArray = append(msgArray, data...)
	_,err = conn.Write(msgArray)
	if err != nil {
		log.Println("发送数据失败",err)
		return "发送数据失败",err
	}
	defer func(){
		err = conn.Close()
		if err != nil {
			return
		}
	}()
	var buf bytes.Buffer
	_,err = io.Copy(&buf,conn)
	if err != nil {
		log.Println("error data",err)
		return "error data",err
	}
	if string(buf.Bytes()[:5]) != "ZBXD\x01" {
		log.Println("zabbix server: 无效响应数据头",err)
		return "zabbix server: 无效响应数据头",err
	}
	return string(buf.Bytes())[13:],nil
	/* 方法一：这个方法有问题
	buf := make([]byte, 0, 4096)
	dataLen := 0
	for {
		n, err := conn.Read(buf[dataLen:])
		if n > 0 {
			dataLen += n
		}
		if err != nil {
			if err != io.EOF {
				fmt.Println("数据接收完毕")
			}
			break
		}
	}*/

	/* 方法三：使用ioutil.ReadALL
	response, err := ioutil.ReadAll(conn)
	if string(response[:5]) != "ZBXD\x01" {
		return "zabbix server: 无效响应数据头",err
	}
	return string(response)[13:],nil
	 */
}
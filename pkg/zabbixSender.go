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
		return "connect error:",err
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
		log.Println("send data error:",err)
		return "send data error:",err
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
		log.Println("zabbix server:Invalid data header",err)
		return "zabbix server:Invalid data header",err
	}
	return string(buf.Bytes())[13:],nil


	/* used: ioutil.ReadALL
	response, err := ioutil.ReadAll(conn)
	if string(response[:5]) != "ZBXD\x01" {
		return "zabbix server:Invalid data header",err
	}
	return string(response)[13:],nil
	 */
}

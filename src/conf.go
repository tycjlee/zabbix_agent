/*******************************************************************************
* FileName:  conf.go
* Author: Victor
* Date: 2019/08/24 09:39
* Description:
* Project: zabbix_agent
*******************************************************************************/
package fastlog

import (
	"github.com/BurntSushi/toml"
	"log"
)


func Config() *tomlConfig {
	var config *tomlConfig
	filePath := "D:/Development/GolangProject/zabbix_agent/bin/conf/conf.tml"
	_,err := toml.DecodeFile(filePath,&config)
	if err != nil {
		log.Panic(err)
	}
	return config
}
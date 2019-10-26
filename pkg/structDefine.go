package pkg

type MajorData struct {
	Request string `json:"request"`
	Data []MinorData `json:"data"`
}

type MinorData struct {
	Host string `json:"host"`
	Key string `json:"key"`
	Value interface{} `json:"value"`
	Clock int32 `json:"clock"`
}

type DiscoveryData struct {
	Data []interface{} `json:"data"`
}

type ActiveCheckData struct {
	Request string `json:"request"`
	Host string `json:"host"`
} 

type ResData struct {
	Response string `json:"response"`
	Info string `json:"info"`
}

type tomlConfig struct {
	Server server `toml:"server"`
	Agent agent `toml:"agent"`
}

type server struct {
	Ip string
	Port int `toml:"port"`
	Version string `toml:"version"`
}

type agent struct {
	Port int `toml:"port"`
	LogLevel string `toml:"loglevel"`
	Logfile string `toml:"logfile"`
}

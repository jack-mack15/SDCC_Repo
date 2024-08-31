package node

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	P          float64
	Def_RTT    int
	Sd_ip      string
	Sd_port    int
	MaxNum     int
	gossipType int
}

func ReadConfigFile() (Config, int) {
	var conf Config

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	file, err := os.Open(cwd + "\\node\\node_config.txt")
	if err != nil {
		fmt.Println("ReadConfigFile()--> errore nell'apertura del file:", err)
		return conf, 0
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
		}
	}(file)

	data := make(map[string]string)

	//leggo riga per riga, ometto i commenti e aggiungo elementi alla map
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			fmt.Println("ReadConfigFile()--> formato della linea non valido:", line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		data[key] = value
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("ReadConfigFile()--> 3rrore durante la lettura del file:", err)
	}

	conf.P, err = strconv.ParseFloat(data["p"], 64)
	if err != nil {
		fmt.Println("ReadConfigFile()--> errore nella conversione:", err)
		return conf, 0
	}

	conf.Def_RTT, err = strconv.Atoi(data["def_RTT"])
	if err != nil {
		fmt.Println("ReadConfigFile()--> errore nella conversione:", err)
		return conf, 0
	}
	conf.Sd_ip = data["sd_ip"]
	conf.Sd_port, err = strconv.Atoi(data["sd_port"])
	if err != nil {
		fmt.Println("ReadConfigFile()--> errore nella conversione:", err)
		return conf, 0
	}
	conf.MaxNum, err = strconv.Atoi(data["num"])
	if err != nil {
		fmt.Println("ReadConfigFile()--> errore nella conversione:", err)
		return conf, 0
	}
	conf.gossipType, err = strconv.Atoi(data["gossip_type"])
	if err != nil {
		fmt.Println("ReadConfigFile()--> errore nella conversione gossipType:", err)
		return conf, 0
	}

	check := checkParameters(&conf)

	if !check {
		return conf, 0
	} else {
		return conf, 1
	}
}

func checkParameters(conf *Config) bool {

	//check gossiptype
	if (*conf).gossipType != 1 && (*conf).gossipType != 2 {
		fmt.Println("config file error: gossipType must be equal to 1 or 2")
		return false
	}

	//check p
	if (*conf).P < 0 || (*conf).P > 1 {
		fmt.Println("config file error: parameter P must be between 0 and 1")
		return false
	}

	//check def_rtt
	if (*conf).Def_RTT < 0 || (*conf).Def_RTT > 1000 {
		fmt.Println("config file error: parameter Def_RTT must be between 0 and 1000")
		return false
	}

	//check maxNum
	if (*conf).MaxNum < 0 {
		fmt.Println("config file error: MaxNum must be greater or equal to -1")
		return false
	}

	//check address
	addrToCheck := (*conf).Sd_ip
	count := strings.Count(addrToCheck, ".")
	if count != 3 {
		fmt.Println("config file error: address of service registry is incorrect")
		return false
	}
	count++
	parts := strings.SplitN(addrToCheck, ".", count)
	count--
	for i := 0; i < count; i++ {
		elem, err := strconv.Atoi(parts[i])
		if err != nil {
			fmt.Println("checkParameters()--> errore nella conversione address:", err)
			return false
		}
		//controllo molto grossolano
		if elem > 256 || elem < 0 {
			fmt.Println("config file error: please insert a correct address")
			return false
		}
	}

	//check port number
	port := (*conf).Sd_port
	if port != 8080 {
		fmt.Println("config file error: please use port 8080")
		return false
	}

	return true
}

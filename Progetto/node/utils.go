package node

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type config struct {
	P       float64
	Def_RTT int
	Sd_ip   string
	Sd_port int
	MaxNum  int
}

func ReadConfigFile() (config, int) {
	var conf config

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	file, err := os.Open(cwd + "\\node\\node_config.txt")
	if err != nil {
		fmt.Println("errore nell'apertura del file:", err)
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
			fmt.Println("formato della linea non valido:", line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		data[key] = value
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("3rrore durante la lettura del file:", err)
	}

	conf.P, err = strconv.ParseFloat(data["p"], 64)
	if err != nil {
		fmt.Println("errore nella conversione:", err)
		return conf, 0
	}

	conf.Def_RTT, err = strconv.Atoi(data["def_RTT"])
	if err != nil {
		fmt.Println("errore nella conversione:", err)
		return conf, 0
	}
	conf.Sd_ip = data["sd_ip"]
	conf.Sd_port, err = strconv.Atoi(data["sd_port"])
	if err != nil {
		fmt.Println("errore nella conversione:", err)
		return conf, 0
	}
	conf.MaxNum, err = strconv.Atoi(data["num"])
	if err != nil {
		fmt.Println("errore nella conversione:", err)
		return conf, 0
	}

	return conf, 1
}

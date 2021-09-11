package main

import (
	"github.com/pelletier/go-toml/v2"
	"log"
	"os"
)

const filePath = "config.toml"

type conf struct {
	DbPath string
	Port   int
}

func RetrieveConf() *conf {
	var config conf
	buf, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Read config error\nerr:\t%e\n", err)
	}
	err = toml.Unmarshal(buf, &config)
	if err != nil {
		log.Fatalf("config format error\nerr:\t%e\n", err)
	}
	return &config
}

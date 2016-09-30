package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/adiclepcea/mailsender/service"
)

func main() {
	if _, err := os.Stat("config.json"); os.IsNotExist(err) {
		log.Fatal("Config file not found\n")
	}
	confString, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Error reading config: %s\n", err.Error())
	}
	serv, err := service.NewMailSenderService(string(confString))
	if err != nil {
		log.Fatalf("Error creating service from config: %s\n", err.Error())
	}
	serv.Run()
}

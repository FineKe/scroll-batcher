package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"scroll-batch/core"
	"syscall"
)

func main() {

	config, err := loadConfig()
	if err != nil {
		log.Fatalln("load config failed: ", err.Error())
	}

	scroller := core.NewScroller(config)
	scroller.StartServer()

	sigs := make(chan os.Signal, 1)
	done := make(chan struct{}, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.Println("receive a sig: ", sig)
		scroller.ShutDown()
		done <- struct{}{}
	}()

	<-done
}

func loadConfig() (core.Config, error) {
	bytes, err := ioutil.ReadFile("config.json")
	if err != nil {
		return core.Config{}, err
	}

	config := core.Config{}
	json.Unmarshal(bytes, &config)

	return config, nil
}

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	. "restapi/cmd/internal/server"
	"time"
)

func main() {
	s := GetServer()
	go func() {
		log.Println("Listening to ", Port)
		err := s.ListenAndServe()
		if err != nil {
			fmt.Println(err)
			return
		}
	}()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs)
	sig := <-sigs
	log.Println("Quitting after signal:", sig)
	time.Sleep(time.Second * 2)
	s.Shutdown(nil)
}

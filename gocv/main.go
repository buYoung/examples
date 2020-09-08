package main

import (
	"./global"
	"log"
	"time"
)

func main() {
	global.Cam1 = &global.Camerainfo{
		Name:   "first",
		Number: 0,
		Debug:  true,
	}

	err := global.Cam1.Init()
	if err != nil {
		log.Println(err)
	}
	global.Cam1.Read()

	timeer := time.Now()
	for {
		if time.Since(timeer).Milliseconds() >= 2000 {
			log.Println(global.Cam1.Encodestr) // output base64 image code
			timeer = time.Now()
		}
	}
}

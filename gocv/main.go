package main

import (
	"./global"
	"context"
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

	ctx, cancle := context.WithTimeout(context.Background(), time.Second)
	defer cancle()
	savedone := false
	for {
		if !savedone {
			select {
			case <-ctx.Done():
				global.Cam1.VideoWrite()
				time.Sleep(time.Second)
				global.Cam1.Videowriterstate <- 1
				time.Sleep(time.Second * 10)
				global.Cam1.Videowriterstate <- 2
				savedone = true
			}
		}

	}
}

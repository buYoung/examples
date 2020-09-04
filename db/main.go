package main

import (
	"./global"
	"log"
)

func main() {
	log.Println("start")

	dbi := &global.DB{
		Db:   nil,
		Conn: nil,
	}

	err := dbi.Setup("mysql", "", true)
	if err != nil {
		log.Println("오류1", err)
	} else {
		err = dbi.SetConnection()
		if err != nil {
			log.Println("오류2", err)
		} else {
			result := dbi.Query("SELECT ?", 1)
			log.Println(result)
			result = dbi.Query("SELECT ? + ?", 1, 2)
			log.Println(result)
			result = dbi.Query("SELECT now()")
			log.Println(result)
		}
	}

}

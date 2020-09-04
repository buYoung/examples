package main

import (
	"./global"
	"log"
	"os"
	"path"
)

func main() {
	zip := global.Zips{
		Level: global.BestCompression,
	}
	_ = os.Chdir("compress") // compress폴더 안에서 실행할경우 제거해야함. examples 폴더에서 실행시 필요한 항목

	err := zip.Makezip(path.Join("./zip", "test.zip"), []string{path.Join("./zip", "README.md"), path.Join("./zip", "새폴더")})
	if err != nil {
		log.Println(err)
	} else {
		log.Println("압축완료")
	}

	err = zip.Decompresszipfile(path.Join("./zip", "test.zip"), path.Join("./zip", "unzip"), true)
	if err != nil {
		log.Println(err)
	}
}

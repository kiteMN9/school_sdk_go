package check_code

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
	"time"
)

func SaveImgStream(data []byte, path, msg string) {
	if len(data) == 0 {
		return
	}
	img, err := decodeImage(data)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	SaveImg(img, path, msg)
}

func SaveImg(img image.Image, path, msg string) {
	if msg == "" {
		msg = fmt.Sprint(time.Now().UnixMilli())
	}
	if !IsExist(path) {
		err := os.Mkdir(path, 0777)
		if err != nil {
			return
		}
	}
	finalPath := path + "/" + msg + ".png"
	// 检查文件是否存在
	//if _, err := os.Stat(finalPath); err == nil {
	//	//fmt.Printf("文件 %s 存在\n", finalPath)
	//	finalPath = path + "/" + msg + fmt.Sprint(time.Now().UnixMilli()) + ".png"
	//}
	outFile, err1 := os.Create(finalPath)
	if err1 != nil {
		log.Println(err1)
		return
	}
	err2 := png.Encode(outFile, img)
	if err2 != nil {
		return
	}
}

func IsExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}

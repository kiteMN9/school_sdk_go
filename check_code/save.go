package check_code

import (
	"os"
)

func SaveImg(data *[]byte, msg string) {

}

func IsExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}

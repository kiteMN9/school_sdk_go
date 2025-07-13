package utils

import (
	"fmt"
	"testing"
)

// func Test_identy(t *testing.T) {
func Test_Read_excel(t *testing.T) {
	var classList []string
	var teacherList []string
	var typeList []string
	classList, teacherList, typeList = ReadExcel()
	fmt.Println(classList, teacherList, typeList)
}

package utils

import (
	"fmt"
	"testing"
)

func Test_Read_excel(t *testing.T) {
	var classList []string
	var teacherList []string
	var typeList []string
	classList, teacherList, typeList = ReadExcel("want.xlsx")
	fmt.Println(classList, teacherList, typeList)
}

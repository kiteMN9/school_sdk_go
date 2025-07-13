package utils

import (
	"fmt"
	"log"

	"github.com/xuri/excelize/v2"
)

func ReadExcel() ([]string, []string, []string) {

	var f *excelize.File
	var err error
	for {
		f, err = excelize.OpenFile("want.xlsx")
		if err != nil {

			writeExcel()
			continue
		}
		err := f.Close()
		if err != nil {
			return nil, nil, nil
		}
		break
	}
	return *readExcel(f, "class"), *readExcel(f, "teacher"), *readExcel(f, "type")
}

func readExcel(f *excelize.File, sheetName string) *[]string {
	var dataList []string

	sheetMap := f.GetSheetMap()
	if len(sheetMap) == 0 {
		fmt.Println("Excel文件中没有工作表")
		return &dataList
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		log.Println("读取行数据失败:", err)
		return &dataList
	}

	for rowIdx, row := range rows {
		if rowIdx == 0 {
			continue
		}
		if len(row) > 0 {
			firstColumn := row[0]

			dataList = append(dataList, firstColumn)
		} else {

		}
	}

	return &dataList
}

func writeExcel() {

	f := excelize.NewFile()

	sheetClass := "class"

	index, _ := f.NewSheet(sheetClass)
	f.SetActiveSheet(index)

	if err := f.SetColWidth(sheetClass, "A", "A", 37); err != nil {
		fmt.Println("设置A列宽度失败:", err)
		return
	}
	sheetTeacher := "teacher"
	_, err := f.NewSheet(sheetTeacher)
	if err != nil {
		return
	}

	if err := f.SetColWidth(sheetTeacher, "A", "A", 25); err != nil {
		fmt.Println("设置A列宽度失败:", err)
		return
	}
	sheetType := "type"
	_, err = f.NewSheet(sheetType)
	if err != nil {
		return
	}

	if err := f.SetColWidth(sheetType, "A", "A", 29); err != nil {
		fmt.Println("设置A列宽度失败:", err)
		return
	}

	if err := f.DeleteSheet("Sheet1"); err != nil {
		fmt.Println("删除Sheet1失败:", err)
		return
	}
	headers := []string{"教学班名称"}
	data := [][]interface{}{
		{"足球-防止匹配-第一志愿"},
		{"蓝球-防止匹配-第二志愿"},
	}
	writeExcelData(f, sheetClass, headers, data)

	headers = []string{"上课教师"}
	data = [][]interface{}{
		{"超星尔雅"},
	}
	writeExcelData(f, sheetTeacher, headers, data)

	headers = []string{"课程类型"}
	data = [][]interface{}{
		{"体育"},
		{"艺术类"},
		{"人文类"},
	}
	writeExcelData(f, sheetType, headers, data)

	if err := f.SaveAs("want.xlsx"); err != nil {
		fmt.Println("保存文件失败:", err)
		return
	}

}

func writeExcelData(f *excelize.File, sheetName string, headers []string, data [][]interface{}) {

	for colIdx, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		err := f.SetCellValue(sheetName, cell, header)
		if err != nil {
			return
		}
	}

	for rowIdx, rowData := range data {
		for colIdx, cellData := range rowData {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			err := f.SetCellValue(sheetName, cell, cellData)
			if err != nil {
				return
			}
		}
	}
}

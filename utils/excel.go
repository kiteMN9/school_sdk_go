package utils

import (
	"fmt"
	"log"
	"strings"

	"github.com/xuri/excelize/v2"
)

func ReadExcel() ([]string, []string, []string) {
	// var class_list []string
	// var teacher_list []string
	// var type_list []string
	// 打开Excel文件
	var f *excelize.File
	var err error
	for {
		f, err = excelize.OpenFile("want.xlsx")
		if err != nil {
			// fmt.Println("打开文件失败:", err)
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
	// 获取第一个工作表的名称
	sheetMap := f.GetSheetMap()
	if len(sheetMap) == 0 {
		fmt.Println("Excel文件中没有工作表")
		return &dataList
	}
	// firstSheetIndex := 1 // 工作表索引从1开始
	// sheetName := sheetMap[firstSheetIndex]

	// 读取所有行数据
	rows, err := f.GetRows(sheetName)
	if err != nil {
		log.Println("读取行数据失败:", err)
		return &dataList
	}

	// 遍历行，跳过首行并提取第一列
	for rowIdx, row := range rows {
		if rowIdx == 0 { // 跳过标题行
			continue
		}
		if len(row) > 0 {
			firstColumn := strings.TrimSpace(row[0])
			// fmt.Printf("第%d行第一列: %s\n", rowIdx+1, firstColumn)
			if firstColumn == "" {
				continue
			}
			dataList = append(dataList, firstColumn)
		} else {
			//fmt.Printf("第%d行无数据\n", rowIdx+1)
		}
	}
	// fmt.Println(data_list)
	return &dataList
}

func writeExcel() {
	// 创建一个新的Excel文件
	f := excelize.NewFile()

	// 默认会创建一个名为"Sheet1"的工作表，可以直接使用它
	sheetClass := "class"
	// 或者创建自定义名称的工作表：
	index, _ := f.NewSheet(sheetClass)
	f.SetActiveSheet(index) // 打开Excel时默认显示此工作表
	// f.NewSheet("class")
	// 设置列宽度
	// 1. 设置单个列宽（A列宽度为15）// 37 25 29
	if err := f.SetColWidth(sheetClass, "A", "A", 37); err != nil {
		fmt.Println("设置A列宽度失败:", err)
		return
	}
	sheetTeacher := "teacher"
	_, err := f.NewSheet(sheetTeacher)
	if err != nil {
		return
	}
	// 设置列宽度
	// 1. 设置单个列宽（A列宽度为15）// 37 25 29
	if err := f.SetColWidth(sheetTeacher, "A", "A", 25); err != nil {
		fmt.Println("设置A列宽度失败:", err)
		return
	}
	sheetType := "type"
	_, err = f.NewSheet(sheetType)
	if err != nil {
		return
	}
	// 设置列宽度
	// 1. 设置单个列宽（A列宽度为15）// 37 25 29
	if err := f.SetColWidth(sheetType, "A", "A", 29); err != nil {
		fmt.Println("设置A列宽度失败:", err)
		return
	}

	// 3. 删除默认的Sheet1
	if err := f.DeleteSheet("Sheet1"); err != nil {
		fmt.Println("删除Sheet1失败:", err)
		return
	}
	headers := []string{"教学班名称"}
	data := [][]interface{}{
		{"防止匹配-第一志愿-如足球"},
		{"蓝球-防止匹配-第二志愿"},
		{"第三志愿"},
		{"采用包含匹配策略"},
		{"'ABC' 包含 'B'"},
		{"可以理解为关键字"},
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

	// 保存文件
	if err := f.SaveAs("want.xlsx"); err != nil {
		fmt.Println("保存文件失败:", err)
		return
	}

	// fmt.Println("Excel文件已成功创建: want.xlsx")
}

func writeExcelData(f *excelize.File, sheetName string, headers []string, data [][]interface{}) {
	// 写入表头（第一行）

	for colIdx, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1) // 列从1开始，行从1开始
		err := f.SetCellValue(sheetName, cell, header)
		if err != nil {
			return
		}
	}

	// 写入数据行（示例数据）

	for rowIdx, rowData := range data {
		for colIdx, cellData := range rowData {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2) // 数据从第二行开始
			err := f.SetCellValue(sheetName, cell, cellData)
			if err != nil {
				return
			}
		}
	}
}

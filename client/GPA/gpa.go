package GPA

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func GPA(htmlText string) {
	if htmlText == "" {
		return
	}
	// 加载HTML文档
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlText))
	if err != nil {
		fmt.Println("Error loading HTML:", err)
		log.Println("Error loading HTML:", err)
		return
	}

	// 提取学生姓名
	name := doc.Find("font[style='font-weight: bold']").First().Text()
	nameL := regexp.MustCompile(`(.+?)同学`).FindStringSubmatch(name)
	if len(nameL) < 2 {
		return
	}
	name = nameL[1]
	name = strings.TrimSpace(strings.ReplaceAll(name, "&nbsp;", ""))

	// 提取统计时间
	timeText := doc.Find("font[style='font-weight: bold']").First().Text()
	//statTime := regexp.MustCompile(`(\d{4}-\d{2}-\d{2} \d{2}:\s*\d{2}:\s*\d{2})`).FindString(timeText)
	statTime := regexp.MustCompile(`(\d{4}-\d{2}-\d{2} \d{1,2}\s*:\s*\d{1,2}\s*:\s*\d{1,2})`).FindString(timeText)

	// 提取GPA
	gpa := doc.Find("font[size='2px'][style='color: red;']").First().Text()
	gpa = strings.TrimSpace(gpa)

	// 提取课程数据
	courseText := doc.Find("#alertBox").Text()
	courseText = strings.ReplaceAll(courseText, "\n", "")
	courseText = strings.ReplaceAll(courseText, "&nbsp;", " ")

	// 使用正则表达式提取所有数值
	re := regexp.MustCompile(`(\d+)`)
	matches := re.FindAllString(courseText, -1)
	// [总课程, 通过, 未通过, 未修, 在读, 计划外通过, 计划外未通过]
	if len(matches) < 7 {
		log.Println("Failed to extract course numbers")
		fmt.Println("Failed")
		return
	}

	// 输出结果
	fmt.Printf("学生姓名: %s\n", name)
	fmt.Printf("统计时间: %s\n", statTime)
	fmt.Printf("平均学分绩点(GPA): %s\n", gpa)
	fmt.Printf("计划总课程: %s 门\n", matches[8])
	fmt.Printf("已通过: %s 门\n", matches[9])
	fmt.Printf("未通过: %s 门\n", matches[10])
	fmt.Printf("未修: %s 门\n", matches[11])
	fmt.Printf("在读: %s 门\n", matches[12])
	fmt.Printf("计划外通过: %s 门\n", matches[13])
	fmt.Printf("计划外未通过: %s 门\n", matches[14])
}

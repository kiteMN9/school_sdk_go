package GPA

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

func GPA(htmlBody io.Reader) {
	if htmlBody == nil {
		fmt.Println("empty htmlBody")
		return
	}
	err := gpa(htmlBody)
	if err != nil {
		fmt.Println(err)
	}
}

func extractStats(text string) {
	text = strings.ReplaceAll(text, "\n", "")
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	// 将不间断空格替换为普通空格以便匹配
	text = strings.ReplaceAll(text, "\u00a0", " ")
	nameL := regexp.MustCompile(`(.+?)同学`).FindStringSubmatch(text)
	if len(nameL) < 2 {
		fmt.Println("未找到有效信息")
		return
	}
	name := nameL[1]
	name = strings.TrimSpace(strings.ReplaceAll(name, "&nbsp;", ""))
	fmt.Printf("学生姓名: %s\n", name)

	reTime := regexp.MustCompile(`(\d{4}-\d{2}-\d{2} \d{1,2}\s*:\s*\d{1,2}\s*:\s*\d{1,2})`).FindString(text)
	fmt.Printf("统计时间: %s\n", reTime)

	reGPA := regexp.MustCompile(`平均学分绩点.*?（GPA）[：:]\s*([\d.]+)`)

	if m := reGPA.FindStringSubmatch(text); len(m) > 1 {
		fmt.Printf("平均学分绩点(GPA): %s\n", m[1])
	}

	re := regexp.MustCompile(`(\d+)`)
	matches := re.FindAllString(text, -1)
	// 预期顺序: [总课程, 通过, 未通过, 未修, 在读, 计划外通过, 计划外未通过]
	if len(matches) < 7 {
		log.Println("Failed to extract course numbers")
		fmt.Println("Failed")
		return
	}

	fmt.Printf("计划总课程: %s 门\n", matches[8])
	fmt.Printf("已通过: %s 门\n", matches[9])
	fmt.Printf("未通过: %s 门\n", matches[10])
	fmt.Printf("未修: %s 门\n", matches[11])
	fmt.Printf("在读: %s 门\n", matches[12])
	fmt.Printf("计划外通过: %s 门\n", matches[13])
	fmt.Printf("计划外未通过: %s 门\n", matches[14])
}

func gpa(htmlBody io.Reader) error {
	tokenizer := html.NewTokenizer(htmlBody)
	var inAlert bool
	var alertText strings.Builder

	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			if tokenizer.Err() != io.EOF {
				return tokenizer.Err()
			}
			fmt.Println(io.EOF)
			break // 正常结束
		}
		switch tt {
		case html.StartTagToken:
			token := tokenizer.Token()
			if token.Data == "div" {
				for _, attr := range token.Attr {
					if attr.Key == "id" && attr.Val == "alertBox" {
						inAlert = true
						break
					}
				}
			}
		case html.EndTagToken:
			token := tokenizer.Token()
			if token.Data == "div" && inAlert {
				inAlert = false
				// 已收集完，立即结束解析
				break
			}
		case html.TextToken:
			if inAlert {
				alertText.Write(bytes.TrimSpace(tokenizer.Text()))
			}
		}
		// 收集完成后跳出外层循环
		if !inAlert && alertText.Len() > 0 {
			break
		}
	}

	extractStats(alertText.String())
	return nil
}

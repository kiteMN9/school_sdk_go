package utils

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func UserIsLogin(account, html string) bool {

	accountPattern := fmt.Sprintf(`value="%s"`, regexp.QuoteMeta(account))
	re1 := regexp.MustCompile(accountPattern)
	if re1.MatchString(html) {
		return true
	}

	re2 := regexp.MustCompile(`id="tips"`)
	if !re2.MatchString(html) {
		return true
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return false
	}

	errMsg := strings.TrimSpace(doc.Find("#tips").Text())

	if errMsg == "" {
		return false
	}

	if strings.Contains(errMsg, "验证码") {
		log.Println(errMsg)
		return false
	}

	fmt.Println("\r" + errMsg)
	log.Println(errMsg)
	return false
}

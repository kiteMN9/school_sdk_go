package utils

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
)

func UserInputWithSigInt(info string) (string, error) {
	input := ""
	prompt := &survey.Input{
		Message: info,
	}
	err := survey.AskOne(prompt, &input)

	if err != nil {
		// 处理所有平台的中断信号
		if errors.Is(err, terminal.InterruptErr) || errors.Is(err, io.EOF) {
			return "", err
		}

		// 通用错误处理：回退到标准输入
		fmt.Print("\r" + info)
		var result string
		_, scanErr := fmt.Scanln(&result)
		if scanErr != nil {
			// 处理扫描错误（包括无输入的情况）
			if errors.Is(scanErr, io.EOF) {
				return "", errors.New("input interrupted or scanErr")
			}
			return "", nil
		}
		return strings.TrimSpace(result), nil
	}
	return strings.TrimSpace(input), nil
}

package utils

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"io"
	"strings"
)

func UserInputWithSigInt(info string) (string, error) {
	input := ""
	prompt := &survey.Input{
		Message: "\r" + info,
	}

	err := survey.AskOne(prompt, &input, survey.WithIcons(func(icons *survey.IconSet) {
		icons.Question.Text = ""
		icons.Question.Format = ""
	}))

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
				return "", errors.New("input interrupted")
			}
			return "", scanErr
		}
		return strings.TrimSpace(result), nil
	}
	return strings.TrimSpace(input), nil
}

package utils

import (
	"fmt"
	//"strings"

	"school_sdk/utils/color" // 替换为你的模块路径
)

// TestTerminalColors 测试终端色彩支持情况
func TestTerminalColors() {
	testTextStyles()
	test16Colors()
	test256Colors()
	testTrueColors()
}

func testTextStyles() {
	fmt.Println("\n=== 文本样式 ===")

	styles := []struct {
		name string
		code string
	}{
		{"Bold", color.Bold},
		{"Dim", color.Dim},
		{"Italic", color.Italic},
		{"Underline", color.Underline},
		{"Blink", color.Blink},
		{"Reverse", color.Reverse},
		{"Hidden", color.Hidden},
		{"Strikethrough", color.Strikethrough},
	}

	for _, style := range styles {
		fmt.Printf("%s%s%s ", style.code, style.name, color.Reset)
	}
	fmt.Printf("\n\n")
}

func test16Colors() {
	fmt.Println("=== 16色前景色 ===")

	foregrounds := []struct {
		name string
		code string
	}{
		{"Black", color.Black},
		{"Red", color.Red},
		{"Green", color.Green},
		{"Yellow", color.Yellow},
		{"Blue", color.Blue},
		{"Magenta", color.Magenta},
		{"Cyan", color.Cyan},
		{"White", color.White},
		{"BrightBlack", color.BrightBlack},
		{"BrightRed", color.BrightRed},
		{"BrightGreen", color.BrightGreen},
		{"BrightYellow", color.BrightYellow},
		{"BrightBlue", color.BrightBlue},
		{"BrightMagenta", color.BrightMagenta},
		{"BrightCyan", color.BrightCyan},
		{"BrightWhite", color.BrightWhite},
	}

	for i, fg := range foregrounds {
		fmt.Printf("%s%-15s%s", fg.code, fg.name, color.Reset)
		if (i+1)%4 == 0 {
			fmt.Println()
		}
	}
	fmt.Println()

	fmt.Println("=== 16色背景色 ===")

	backgrounds := []struct {
		name string
		code string
	}{
		{"BGBlack", color.BGBlack},
		{"BGRed", color.BGRed},
		{"BGGreen", color.BGGreen},
		{"BGYellow", color.BGYellow},
		{"BGBlue", color.BGBlue},
		{"BGMagenta", color.BGMagenta},
		{"BGCyan", color.BGCyan},
		{"BGWhite", color.BGWhite},
		{"BGBrightBlack", color.BGBrightBlack},
		{"BGBrightRed", color.BGBrightRed},
		{"BGBrightGreen", color.BGBrightGreen},
		{"BGBrightYellow", color.BGBrightYellow},
		{"BGBrightBlue", color.BGBrightBlue},
		{"BGBrightMagenta", color.BGBrightMagenta},
		{"BGBrightCyan", color.BGBrightCyan},
		{"BGBrightWhite", color.BGBrightWhite},
	}

	for i, bg := range backgrounds {
		fmt.Printf("%s%s%-16s%s", bg.code, color.White, bg.name, color.Reset)
		if (i+1)%4 == 0 {
			fmt.Println()
		}
	}
	fmt.Println()
}

func test256Colors() {
	fmt.Println("=== 256色支持 ===")

	// 系统颜色 (0-15)
	fmt.Println("系统颜色:")
	for i := 0; i < 16; i++ {
		fmt.Printf("%s  %s", color.Background256(i), color.Reset)
	}
	fmt.Printf("\n\n")

	// 色彩立方 (16-231)
	fmt.Println("色彩立方 (6x6x6):")
	for i := 16; i < 232; i++ {
		fmt.Printf("%s %s", color.Background256(i), color.Reset)
		if (i-15)%36 == 0 {
			fmt.Println()
		}
	}
	fmt.Println()

	// 灰度 (232-255)
	fmt.Println("灰度:")
	for i := 232; i < 256; i++ {
		fmt.Printf("%s %s", color.Background256(i), color.Reset)
	}
	fmt.Printf("\n\n")
}

func testTrueColors() {
	fmt.Println("=== 24位真彩色 (RGB) ===")

	// 红色渐变
	fmt.Println("红色渐变:")
	for i := 0; i < 256; i += 16 {
		fmt.Printf("%s %s", color.BackgroundRGB(i, 0, 0), color.Reset)
	}

	// 绿色渐变
	fmt.Println("\n\n绿色渐变:")
	for i := 0; i < 256; i += 16 {
		fmt.Printf("%s %s", color.BackgroundRGB(0, i, 0), color.Reset)
	}

	// 蓝色渐变
	fmt.Println("\n\n蓝色渐变:")
	for i := 0; i < 256; i += 16 {
		fmt.Printf("%s %s", color.BackgroundRGB(0, 0, i), color.Reset)
	}

	// RGB色谱
	fmt.Println("\n\nRGB色谱:")
	for g := 0; g < 256; g += 64 {
		for r := 0; r < 256; r += 8 {
			b := 255 - r
			fmt.Printf("%s %s", color.BackgroundRGB(r, g, b), color.Reset)
		}
		fmt.Println()
	}

	fmt.Println("\n测试完成！")
}

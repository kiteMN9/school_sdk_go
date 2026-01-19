package color

import "fmt"

// ANSI 转义序列
const (
	Reset = "\033[0m"

	// 文本样式
	Bold          = "\033[1m"
	Dim           = "\033[2m"
	Italic        = "\033[3m"
	Underline     = "\033[4m"
	Blink         = "\033[5m"
	Reverse       = "\033[7m"
	Hidden        = "\033[8m"
	Strikethrough = "\033[9m"

	// 16色前景色
	Black         = "\033[30m"
	Red           = "\033[31m"
	Green         = "\033[32m"
	Yellow        = "\033[33m"
	Blue          = "\033[34m"
	Magenta       = "\033[35m"
	Cyan          = "\033[36m"
	White         = "\033[37m"
	BrightBlack   = "\033[90m"
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"

	// 16色背景色
	BGBlack         = "\033[40m"
	BGRed           = "\033[41m"
	BGGreen         = "\033[42m"
	BGYellow        = "\033[43m"
	BGBlue          = "\033[44m"
	BGMagenta       = "\033[45m"
	BGCyan          = "\033[46m"
	BGWhite         = "\033[47m"
	BGBrightBlack   = "\033[100m"
	BGBrightRed     = "\033[101m"
	BGBrightGreen   = "\033[102m"
	BGBrightYellow  = "\033[103m"
	BGBrightBlue    = "\033[104m"
	BGBrightMagenta = "\033[105m"
	BGBrightCyan    = "\033[106m"
	BGBrightWhite   = "\033[107m"
)

// 生成256色前景色
func Foreground256(color int) string {
	return fmt.Sprintf("\033[38;5;%dm", color)
}

// 生成256色背景色
func Background256(color int) string {
	return fmt.Sprintf("\033[48;5;%dm", color)
}

// 生成真彩色前景色
func ForegroundRGB(r, g, b int) string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
}

// 生成真彩色背景色
func BackgroundRGB(r, g, b int) string {
	return fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b)
}

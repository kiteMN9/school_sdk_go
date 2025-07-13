package utils

import "testing"

func Test_Exit(t *testing.T) {
	Exit()
	select {} // 保持主goroutine运行
}

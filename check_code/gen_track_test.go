package check_code

import (
	"encoding/json"
	"testing"
)

func Test_GetTrack(t *testing.T) {
	// 生成轨迹数据
	trackData := GetTrack(210, 480)

	// 直接序列化
	jsonData, err := json.Marshal(trackData)
	if err != nil {
		panic(err)
	}
	println(string(jsonData))

}

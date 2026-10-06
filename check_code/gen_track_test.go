package check_code

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GetTrack(t *testing.T) {
	// 生成轨迹数据
	trackData := GetTrack(210, 480)

	// 直接序列化
	jsonData, err := json.Marshal(trackData)
	require.NoError(t, err)
	println(string(jsonData))

}

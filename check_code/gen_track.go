package check_code

import (
	"encoding/json"
)

type TrackPoint struct {
	X int `json:"x"`
	Y int `json:"y"`
	T int `json:"t"`
}

func GetTrack(distance int, y int) []TrackPoint {
	track := make([]TrackPoint, 0)
	// 不开源
	return track
}

func GetTrackString(x, y int) string {
	trackData := GetTrack(x, y)
	if trackData == nil {
		return ""
	}

	// 直接序列化
	jsonData, err := json.Marshal(trackData)
	if err != nil {
		panic(err)
	}
	strTrack := string(jsonData)
	// println(str_track)
	return strTrack
}

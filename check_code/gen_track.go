package check_code

import (
	"encoding/json"
	"math"
	"math/rand"
	"time"
)

// TrackPoint 带JSON标签的结构体定义
type TrackPoint struct {
	X int `json:"x"`
	Y int `json:"y"`
	T int `json:"t"`
}

func GetTrack(distance int, y int) []TrackPoint {
	if distance < 5 {
		return nil
	}
	start := rand.Intn(1030-950+1) + 950
	y1 := 0
	current := 0.0
	var track []TrackPoint
	count := 0
	t1 := 0.0
	jitter := rand.Float64()
	startTime := float64(time.Now().UnixNano()) / 1e9 // 转换为秒级时间戳

	for current < float64(distance+3) && t1 < 1.04 {
		var currentVal float64
		if math.Abs(current-float64(distance)) > 15 {
			currentVal = (3*math.Pow(t1, 2) - 2*math.Pow(t1, 3)) * float64(distance) * 1.008
			currentVal += math.Sin(jitter+t1*2) * 8
		} else {
			currentVal = (3*math.Pow(t1, 2) - 2*math.Pow(t1, 3)) * float64(distance) * 1.008
			currentVal += math.Sin(jitter+t1*2) * 2.5
		}

		current = currentVal
		if y1 < 5 {
			y1 += rand.Intn(2)
		} else {
			y1 += rand.Intn(2) - 1
		}

		track = append(track, TrackPoint{
			X: start + int(current),
			Y: y + y1,
			T: int((startTime + t1/2.43) * 1000),
		})
		count++
		t1 += 0.086
	}

	for t1 < 3.3 {
		move := math.Sin(t1)*9 + math.Sin(t1*2)*1.4
		current1 := current + move

		if y1 < 5 {
			y1 += rand.Intn(2)
		} else {
			y1 += rand.Intn(2) - 1
		}

		track = append(track, TrackPoint{
			X: start + int(current1),
			Y: y + y1,
			T: int((startTime + t1/2.43) * 1000),
		})
		count++
		t1 += 0.2
	}

	return track
}

//func GetTrackJSON(distance int, y int) ([]byte, error) {
//	track := GetTrack(distance, y)
//	return json.MarshalIndent(track, "", "  ")
//}

func GetTrackByte(x, y int) []byte {
	trackData := GetTrack(x, y)
	if trackData == nil {
		return nil
	}

	// 直接序列化
	jsonData, err := json.Marshal(trackData)
	if err != nil {
		panic(err)
	}
	//strTrack := string(jsonData)
	// println(str_track)
	return jsonData
}

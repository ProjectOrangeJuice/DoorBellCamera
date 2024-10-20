package main

type settings struct {
	Name               string
	Connection         string
	FPS                int
	MinCount           int
	Motion             bool
	Blur               int
	Debug              bool
	BufferBefore       int
	BufferAfter        int
	NoMoveRefreshCount int
	Zones              []zone
}
type zone struct {
	X1          int
	Y1          int
	X2          int
	Y2          int
	Threshold   int
	BoxJump     int
	SmallIgnore int
	Area        int
}

func genTestSetting() settings {
	z := zone{10, 10, 500, 500, 20, 400, 2, 50}
	zo := make([]zone, 1)
	zo[0] = z
	s := settings{"test", "", 5, 3, true, 21, true, 5, 5, 3, zo}
	return s
}

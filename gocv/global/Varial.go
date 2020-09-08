package global

import (
	"gocv.io/x/gocv"
	"time"
)

type Camerainfo struct {
	Name      string //
	Number    int64
	Debug     bool
	camera    *gocv.VideoCapture
	window    *gocv.Window
	eWindow   *gocv.Window
	mat       gocv.Mat
	encodemat *gocv.Mat
	Encodestr string
	state     chan bool
}

var (
	oldnow time.Time
)

var (
	Cam1 *Camerainfo
)

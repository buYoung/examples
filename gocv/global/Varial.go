package global

import (
	"gocv.io/x/gocv"
	"time"
)

type Camerainfo struct {
	Name             string //
	Number           int64
	Debug            bool
	camera           *gocv.VideoCapture
	videowriter      *gocv.VideoWriter
	window           *gocv.Window
	eWindow          *gocv.Window
	mat              gocv.Mat
	encodemat        *gocv.Mat
	Videowritermat   *gocv.Mat
	Encodestr        string
	FPS              int
	state            chan bool
	Videowriterstate chan uint8
}

var (
	oldnow time.Time
)

var (
	Cam1 *Camerainfo
)

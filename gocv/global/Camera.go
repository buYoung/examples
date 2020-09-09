package global

import "C"
import (
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"gocv.io/x/gocv"
	"log"
	"strings"
	"time"
)

func (c *Camerainfo) Init() (err error) {
	defer func() {
		s := recover()
		if s != nil {
			err = errors.Errorf("Camera Init Error : %v", s)
		}
	}()
	var errd error
	c.camera, errd = gocv.VideoCaptureDevice(int(c.Number))
	if errd != nil {
		return errd
	}
	c.FPS = 20
	c.Videowriterstate = make(chan uint8, 1)
	c.camera.Set(gocv.VideoCaptureFOURCC, c.camera.ToCodec("MJPG"))
	c.camera.Set(gocv.VideoCaptureFrameWidth, 800)
	c.camera.Set(gocv.VideoCaptureFrameHeight, 600)
	c.camera.Set(gocv.VideoCaptureFPS, 10)
	c.mat = gocv.NewMat()
	return nil
}

func (c *Camerainfo) Read() {
	go func() {
		defer func() {
			s := recover()
			if s != nil {
				log.Printf("Camera Read Error : %v", s)
			}
		}()
		oldnow = time.Now()
		if c.Debug { // init에서 설정할경우 출력이 되지않음.
			c.window = gocv.NewWindow("test")
			c.eWindow = gocv.NewWindow("encode window")
		}
		for {
			checkifcamera := fmt.Sprintf("%#v", c.camera)
			if !strings.Contains(checkifcamera, "nil") {
				if ok := c.camera.Read(&c.mat); ok {
					c.encodemat = &c.mat
					c.Videowritermat = &c.mat
					c.encodeimg()
					if c.Debug {
						c.window.IMShow(c.mat)
						if c.window.WaitKey(1) == 27 {
							break
						}
					}
				} else {
					log.Println("Camera Read Fail")
					return
				}
			} else {
				log.Println("Camera is nil")
				return
			}
		}
	}()
}

func (c *Camerainfo) encodeimg() {
	go func() {
		var buf []byte
		var err error
		defer func() {
			s := recover()
			if s != nil {
				log.Printf("encode error camera : %s  detail : %v", c.Name, s)
			}
		}()
		buf, err = gocv.IMEncodeWithParams(gocv.JPEGFileExt, *c.encodemat, []int{gocv.IMWriteJpegQuality, 70})
		if err != nil {
			log.Printf("encode Err %v", err)
		} else {
			c.Encodestr = base64.StdEncoding.EncodeToString(buf)
			//P(fmt.Sprintf("base64 encode result : %v", len(imgeb64)))
		}
		buf = nil
		if c.Debug {
			c.eWindow.IMShow(c.mat)
			if c.eWindow.WaitKey(1) == 27 {
				return
			}
		}
	}()
}
func (c *Camerainfo) VideoWrite() {
	go func() {
		var err error
		defer func() {
			s := recover()
			if s != nil {
				log.Printf("video writer error camera : %s  detail : %v", c.Name, s)
			}
		}()
		c.videowriter, err = gocv.VideoWriterFile("./test.avi", "MJPG", float64(c.FPS), int(c.mat.Cols()), int(c.mat.Rows()), true)
		if err != nil {
			log.Println("video writer", err)
		}
		log.Println("make video file")
		var state uint8
		var stats = false

		for {
			select {
			case state = <-c.Videowriterstate:
				switch state {
				case 1:
					log.Println("get 1")
					stats = true
				case 2:
					err = c.videowriter.Close()
					if err != nil {
						log.Println("video writer close err", err)
					} else {
						log.Println("saved complate")
						state = 3
					}
				}
			default:
			}
			st := fmt.Sprintf("%#v", c.videowriter)

			if state == 3 {
				if !strings.Contains(st, "nil") {
					c.videowriter.Close()
				}
				break
			}
			if stats {
				if strings.Contains(st, "nil") || (*c.Videowritermat).Empty() {
					continue
				}
				err = c.videowriter.Write(*c.Videowritermat)
				if err != nil {
					log.Println("video write fail")
				}
				log.Println("write video")

			}
			time.Sleep(time.Millisecond * time.Duration(1000/c.FPS))
		}
	}()
}

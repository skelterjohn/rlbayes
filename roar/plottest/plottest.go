package main

import (
	"time"
	"exp/draw"
	"exp/draw/x11"
	"os"
	"fmt"
	"image"
	"github.com/skelterjohn/rlbayes/roar"
	"gomatrix.googlecode.com/hg/matrix"
	"gonicetrace.googlecode.com/hg/nicetrace"
)

const (
	spanX	= 10
	spanY	= 10
)

var (
	block	chan bool
	window	draw.Window
	running	bool
	rpost	*roar.Posterior
	colors	= []image.RGBAColor{image.RGBAColor{255, 0, 0, 255}, image.RGBAColor{0, 255, 0, 255}, image.RGBAColor{0, 0, 255, 255}, image.RGBAColor{0, 255, 255, 255}, image.RGBAColor{255, 0, 255, 255}, image.RGBAColor{255, 255, 0, 255}}
)

func clearScreen() {
	scr := window.Screen()
	min, max := scr.Bounds().Min, scr.Bounds().Max
	for x := min.X; x < max.X; x++ {
		for y := min.Y; y < max.Y; y++ {
			scr.Set(x, y, image.RGBAColor{0, 0, 0, 255})
		}
	}
}
func drawField() {
	scr := window.Screen()
	min, max := scr.Bounds().Min, scr.Bounds().Max
	for dx := min.X; dx < max.X; dx += 3 {
		for dy := min.Y; dy < max.Y; dy += 3 {
			rx := float64(dx) / spanX
			ry := float64(dy) / spanY
			x := matrix.MakeDenseMatrix([]float64{rx, ry}, 2, 1)
			clusterIndex := rpost.BestCluster(x)
			var clusterColor image.RGBAColor
			if clusterIndex == -1 {
				continue
			} else {
				clusterColor = colors[clusterIndex%len(colors)]
			}
			window.Screen().Set(dx, dy, clusterColor)
		}
	}
}
func drawAssignments() {
	for t, x := range rpost.X {
		dx, dy := int(x.Get(0, 0)*spanX), int(x.Get(1, 0)*spanY)
		clusterIndex := rpost.C.Get(t)
		groupIndex := rpost.G.Get(clusterIndex)
		var clusterColor image.RGBAColor
		if clusterIndex == -1 {
			clusterColor = image.RGBAColor{100, 100, 100, 255}
		} else {
			clusterColor = colors[clusterIndex%len(colors)]
		}
		var groupColor image.RGBAColor
		if groupIndex == -1 {
			groupColor = image.RGBAColor{100, 100, 100, 255}
		} else {
			groupColor = colors[groupIndex%len(colors)]
		}
		for _, ox := range []int{-2, -1, 0, 1, 2} {
			for _, oy := range []int{-2, -1, 0, 1, 2} {
				window.Screen().Set(dx+ox, dy+oy, groupColor)
			}
		}
		for _, ox := range []int{-1, 0, 1} {
			for _, oy := range []int{-1, 0, 1} {
				window.Screen().Set(dx+ox, dy+oy, clusterColor)
			}
		}
	}
}
func click(x, y float64) {
	x_t := matrix.MakeDenseMatrix([]float64{x, y}, 2, 1)
	rpost.Insert(x_t)
}
func rclick(x float64) {
	x1 := matrix.MakeDenseMatrix([]float64{x}, 1, 1)
	x2 := rpost.ConditionalSample(x1)
	click(x1.Get(0, 0), x2.Get(0, 0))
}
func inferenceLoop() {
	defer nicetrace.Print()
	for i := 0; running; i++ {
		const cycle = 100
		counter := i % cycle
		temperature := 2 * (float64(cycle-counter) / cycle)
		rpost.SweepC(temperature)
		rpost.SweepG(temperature)
	}
}
func drawLoop() {
	defer nicetrace.Print()
	for i := 0; running; i++ {
		const cycle = 100
		counter := i % cycle
		drawAssignments()
		window.FlushImage()
		if true {
			time.Sleep(0.5e8)
		}
		if counter == 0 {
		}
	}
}
func processEvents(events <-chan interface{}) {
	for event := range events {
		switch e := event.(type) {
		case draw.MouseEvent:
			if e.Buttons == 1 {
				clickX, clickY := float64(e.Loc.X)/spanX, float64(e.Loc.Y)/spanY
				click(clickX, clickY)
			} else if e.Buttons == 4 {
				clickX := float64(e.Loc.X) / spanX
				rclick(clickX)
			}
		}
	}
}
func main() {
	defer nicetrace.Print()
	block = make(chan bool, 1)
	var err os.Error
	window, err = x11.NewWindow()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	clearScreen()
	cfg := roar.PosteriorCFGDefault()
	cfg.Partition = 1
	cfg.M = 1
	rpost = roar.New(cfg)
	running = true
	go drawLoop()
	go inferenceLoop()
	processEvents(window.EventChan())
	running = false
}

package main

import (
	"log"
	"time"

	"github.com/faiface/pixel/pixelgl"
)

const (
	winWidth  = 1400
	winHeight = 1200

	timeToDrop = 1.0

	columns = 7
	rows    = 6
	target  = 4
	players = 2

	lost = -999999999
	won  = 999999999

	empty = -1

	stateWaitingForTurn = 0
	stateFalling        = 1
	stateFinished       = 2
)

func main() {
	pixelgl.Run(func() {
		win, err := newConnect4Window(winWidth, winHeight)
		if err != nil {
			log.Fatal(err)
		}

		g := newGame()
		state := stateWaitingForTurn
		var startTime time.Time
		needsRender := true
		for !win.Closed() {
			if needsRender {
				win.Render(g, state, startTime)
			}
			needsRender = true
			switch state {
			case stateWaitingForTurn:
				if g.Turn == 0 {
					colClicked := win.JustClickedCol()
					if colClicked == -1 {
						needsRender = false
					} else {
						ns := g.MakeMove(colClicked)
						if ns != nil {
							g = ns
							state = stateFalling
							startTime = time.Now()
						}
					}
				} else {
					ns := g.AutoChooseMove(4)
					if ns == nil {
						log.Fatal("computer can't find a move")
					}
					g = ns
					state = stateFalling
					startTime = time.Now()
				}
			case stateFalling:
				if time.Now().Sub(startTime).Seconds() >= timeToDrop {
					if g.Finished {
						state = stateFinished
						startTime = time.Now()
					} else {
						state = stateWaitingForTurn
					}
				}
			case stateFinished:
				if win.JustClickedCol() != -1 {
					g = newGame()
					state = stateWaitingForTurn
				}
			}
			win.Update()
		}
	})
}

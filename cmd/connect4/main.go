package main

import (
	"log"
	"time"

	"github.com/faiface/pixel/pixelgl"
)

const (
	winWidth  = 1400
	winHeight = 1200

	timeToDrop  = 1.0
	idleTimeout = 120.0

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

type stateHistory struct {
	states []*gameState
	sp     int
}

func (sh *stateHistory) Current() *gameState {
	return sh.states[sh.sp]
}

func (sh *stateHistory) Back() bool {
	if sh.sp > 0 {
		sh.sp--
		return true
	}
	return false
}

func (sh *stateHistory) Forward() bool {
	if sh.sp < (len(sh.states) - 1) {
		sh.sp++
		return true
	}
	return false
}

func (sh *stateHistory) Add(g *gameState) {
	if len(sh.states) > (sh.sp + 1) {
		sh.states = sh.states[0 : sh.sp+1]
	}
	sh.states = append(sh.states, g)
	sh.sp = len(sh.states) - 1
}

func main() {
	pixelgl.Run(func() {
		win, err := newConnect4Window(winWidth, winHeight)
		if err != nil {
			log.Fatal(err)
		}

		states := &stateHistory{}
		states.Add(newGame())
		state := stateWaitingForTurn
		var startTime time.Time
		idleTTL := time.Now().Add(idleTimeout * time.Second)
		needsRender := true
		for !win.Closed() {
			if time.Now().After(idleTTL) {
				log.Fatal("idle timeout, see ya later")
			}
			if win.JustClickedLeftArrow() {
				if states.Back() {
					states.Back() // go back again behind computer
					state = stateWaitingForTurn
					needsRender = true
				}
			} else if win.JustClickedRightArrow() {
				if states.Forward() {
					states.Forward()
					state = stateWaitingForTurn
					needsRender = true
				}
			}

			if needsRender {
				win.Render(states.Current(), state, startTime)
			}
			needsRender = true
			switch state {
			case stateWaitingForTurn:
				if states.Current().Turn == 0 {
					colClicked := win.JustClickedCol()
					if colClicked == -1 {
						needsRender = false
					} else {
						ns := states.Current().MakeMove(colClicked)
						if ns != nil {
							if !ns.Finished {
								ns.StartAutoChooseMove(4)
							}
							states.Add(ns)
							state = stateFalling
							startTime = time.Now()
							idleTTL = time.Now().Add(idleTimeout * time.Second)
						}
					}
				} else {
					ns := states.Current().WaitAutoChooseMove()
					if ns == nil {
						log.Fatal("computer can't find a move")
					}
					states.Add(ns)
					state = stateFalling
					startTime = time.Now()
				}
			case stateFalling:
				if time.Now().Sub(startTime).Seconds() >= timeToDrop {
					if states.Current().Finished {
						state = stateFinished
						startTime = time.Now()
					} else {
						state = stateWaitingForTurn
					}
				}
			case stateFinished:
				if win.JustClickedCol() != -1 {
					states = &stateHistory{}
					states.Add(newGame())
					state = stateWaitingForTurn
				}
			}
			win.Update()
		}
	})
}

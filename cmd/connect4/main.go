package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/faiface/pixel/pixelgl"
)

const (
	winWidth  = 1400
	winHeight = 1200

	defaultTimeToDrop = 1.0
	idleTimeout       = 120.0

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

	depthToSearch = 4
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
		rand.Seed(time.Now().UnixNano())
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
		inAutoPilot := false
		timeToDrop := defaultTimeToDrop
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
			} else if win.JustClickedAutopilot() {
				inAutoPilot = !inAutoPilot
			} else if win.JustClickedFaster() {
				if timeToDrop < defaultTimeToDrop {
					timeToDrop = defaultTimeToDrop
				} else {
					timeToDrop = defaultTimeToDrop / 10.0
				}
			}

			if needsRender {
				win.Render(states.Current(), state, startTime, timeToDrop)
			}
			needsRender = true
			switch state {
			case stateWaitingForTurn:
				if states.Current().Turn == 0 && !inAutoPilot {
					colClicked := win.JustClickedCol()
					if colClicked == -1 {
						needsRender = false
					} else {
						ns := states.Current().MakeMove(colClicked)
						if ns != nil {
							if !ns.Finished {
								ns.StartAutoChooseMove(depthToSearch)
							}
							states.Add(ns)
							state = stateFalling
							startTime = time.Now()
							idleTTL = time.Now().Add(idleTimeout * time.Second)
						}
					}
				} else {
					ns := states.Current().WaitAutoChooseMove(depthToSearch)
					if ns == nil {
						log.Fatal("computer can't find a move")
					}
					if !ns.Finished && inAutoPilot {
						ns.StartAutoChooseMove(depthToSearch)
					}
					states.Add(ns)
					state = stateFalling
					startTime = time.Now()
				}
			case stateFalling:
				if time.Now().Sub(startTime).Seconds() >= timeToDrop {
					if states.Current().Finished {
						state = stateFinished
						timeToDrop = defaultTimeToDrop
						inAutoPilot = false
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

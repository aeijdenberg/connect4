package main

import (
	"math"
)

var (
	allPossibles = genPossibilities()
)

func toStateIdx(c, r int) int {
	return c*rows + r
}

func toColRow(cr int) (int, int) {
	return (cr / rows), (cr % rows)
}

func getIt(sofar [][target]int, c, r, dx, dy int) [][target]int {
	var rv [target]int
	for i := 0; i < target; i++ {
		actualC := c + (i * dx)
		actualR := r + (i * dy)
		if actualC < 0 || actualC >= columns || actualR < 0 || actualR >= rows {
			return sofar
		}
		rv[i] = toStateIdx(actualC, actualR)
	}
	return append(sofar, rv)
}

func genPossibilities() [][target]int {
	var possibilities [][target]int
	for c := 0; c < columns; c++ {
		for r := 0; r < rows; r++ {
			// Horizontal
			possibilities = getIt(possibilities, c, r, 1, 0)
			// Vertical
			possibilities = getIt(possibilities, c, r, 0, 1)
			// Up and to the right
			possibilities = getIt(possibilities, c, r, 1, 1)
			// The other way
			possibilities = getIt(possibilities, c, r, 1, -1)
		}
	}
	return possibilities
}

type gameState struct {
	State  [columns * rows]int
	Height [columns]int
	Turn   int

	// Populated by MakeMove
	Finished    bool
	WinningMove int

	LastMoveCol  int
	LastMoveRow  int
	LastMoveTurn int
}

func newGame() *gameState {
	rv := &gameState{}
	for i := 0; i < columns*rows; i++ {
		rv.State[i] = empty
	}
	return rv
}

func (s *gameState) MakeMove(c int) *gameState {
	if c < 0 || c >= columns || s.Height[c] == rows || s.Finished {
		return nil
	}
	rv := &gameState{
		State:        s.State,
		Height:       s.Height,
		Turn:         (s.Turn + 1) % players,
		LastMoveTurn: s.Turn,
		LastMoveCol:  c,
		LastMoveRow:  s.Height[c],
	}
	rv.State[toStateIdx(c, s.Height[c])] = s.Turn
	rv.Height[c] = s.Height[c] + 1
	ss, wm := rv.score(s.Turn)
	if ss == won {
		rv.Finished = true
		rv.WinningMove = wm
	} else if rv.Height[c] == rows {
		full := true
		for x := 0; x < columns && full; x++ {
			if rv.Height[x] != rows {
				full = false
			}
		}
		if full {
			rv.Finished = true
		}
	}
	return rv
}

func (s *gameState) AutoChooseMove(depth int) *gameState {
	_, rv := s.minimax(depth, s.Turn)
	return rv
}

// Returns score gamestate
func (s *gameState) minimax(depth, origPlayer int) (int, *gameState) {
	if depth == 0 || s.Finished {
		ss, _ := s.score(origPlayer)
		return ss, nil
	}
	var bestValue int
	if s.Turn == origPlayer {
		bestValue = math.MinInt64
	} else {
		bestValue = math.MaxInt64
	}

	var bs *gameState
	for c := 0; c < columns; c++ {
		ns := s.MakeMove(c)
		if ns != nil {
			v, _ := ns.minimax(depth-1, origPlayer)
			if (s.Turn == origPlayer && v > bestValue) || (s.Turn != origPlayer && v < bestValue) {
				bestValue = v
				bs = ns
			}
		}
	}
	return bestValue, bs
}

// Return score, and (if won) index of winning move in possibilities
func (s *gameState) score(ourPlayer int) (int, int) {
	rv := 0
	for pi, p := range allPossibles {
		for pl := 0; pl < players; pl++ {
			cnt := 0
			for _, cr := range p {
				sv := s.State[cr]
				if sv == pl {
					cnt++
				} else if sv != empty {
					cnt = 0
					break
				}
			}
			if cnt == target {
				if pl == ourPlayer {
					return won, pi
				}
				return lost, -1
			}
			if pl == ourPlayer {
				rv += (cnt * cnt)
			}
		}
	}
	return rv, -1
}

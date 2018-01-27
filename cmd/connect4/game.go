package main

import (
	"image"
	"image/color"
	"log"
	"math"
	"time"

	"golang.org/x/image/colornames"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

const (
	WinWidth  = 1400
	WinHeight = 1200

	TimeToDrop = 1.0

	Columns = 7
	Rows    = 6
	Target  = 4
	Players = 2

	Lost = -999999999
	Won  = 999999999

	Empty = -1

	StateIdle     = 0
	StateFalling  = 1
	StateFinished = 2
)

type circle struct {
	p image.Point
	r int
	c color.Color
}

func (c *circle) ColorModel() color.Model {
	return color.RGBAModel
}

func (c *circle) Bounds() image.Rectangle {
	return image.Rect(c.p.X-c.r, c.p.Y-c.r, c.p.X+c.r, c.p.Y+c.r)
}

func (c *circle) At(x, y int) color.Color {
	xx, yy, rr := float64(x-c.p.X)+0.5, float64(y-c.p.Y)+0.5, float64(c.r)
	if xx*xx+yy*yy < rr*rr {
		return c.c
	}
	return color.Alpha{0}
}

type board struct {
	cWidth float64
	r      float64
	color  color.Color
}

func (c *board) ColorModel() color.Model {
	return color.RGBAModel
}

func (c *board) Bounds() image.Rectangle {
	return image.Rect(0, 0, WinWidth, WinHeight)
}

func (c *board) At(x, y int) color.Color {
	xx := float64(x) - (c.cWidth * (0.5 + float64(x/int(c.cWidth)))) + 0.5
	yy := float64(y) - (c.cWidth * (0.5 + float64(y/int(c.cWidth)))) + 0.5
	if xx*xx+yy*yy >= c.r*c.r {
		return c.color
	}
	return color.RGBA{0, 0, 0, 0}
}

var (
	AllPossibles = genPossibilities()
)

func toStateIdx(c, r int) int {
	return c*Rows + r
}

func toColRow(cr int) (int, int) {
	return (cr / Rows), (cr % Rows)
}

func getIt(sofar [][Target]int, c, r, dx, dy int) [][Target]int {
	var rv [Target]int
	for i := 0; i < Target; i++ {
		actualC := c + (i * dx)
		actualR := r + (i * dy)
		if actualC < 0 || actualC >= Columns || actualR < 0 || actualR >= Rows {
			return sofar
		}
		rv[i] = toStateIdx(actualC, actualR)
	}
	return append(sofar, rv)
}

func genPossibilities() [][Target]int {
	var possibilities [][Target]int
	for c := 0; c < Columns; c++ {
		for r := 0; r < Rows; r++ {
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

type GameState struct {
	State       [Columns * Rows]int
	Height      [Columns]int
	Turn        int
	Finished    bool
	WinningMove int
}

func NewGame() *GameState {
	rv := &GameState{}
	for i := 0; i < Columns*Rows; i++ {
		rv.State[i] = Empty
	}
	return rv
}

// Returns col, row, player and state (which includes this)
func (s *GameState) MakeMove(c int) (int, int, int, *GameState) {
	if c < 0 || c >= Columns || s.Height[c] == Rows {
		return 0, 0, 0, nil
	}
	rv := &GameState{
		State:  s.State,
		Height: s.Height,
	}
	rv.State[toStateIdx(c, s.Height[c])] = s.Turn
	rv.Height[c] = s.Height[c] + 1
	rv.Turn = (s.Turn + 1) % Players
	ss, wm := rv.Score(s.Turn)
	if ss == Won {
		rv.Finished = true
		rv.WinningMove = wm
	} else if rv.Height[c] == Rows {
		full := true
		for x := 0; x < Columns && full; x++ {
			if rv.Height[x] != Rows {
				full = false
			}
		}
		if full {
			rv.Finished = true
		}
	}
	return c, s.Height[c], s.Turn, rv
}

// score, col, row, turn
func (s *GameState) Minimax(depth, origPlayer int) (int, int, int, int, *GameState) {
	if depth == 0 || s.Finished {
		ss, _ := s.Score(origPlayer)
		return ss, 0, 0, 0, nil
	}
	var bestValue int
	if s.Turn == origPlayer {
		bestValue = math.MinInt64
	} else {
		bestValue = math.MaxInt64
	}

	var bCol, bRow, bTurn int
	var bs *GameState
	for c := 0; c < Columns; c++ {
		cc, rr, tt, ns := s.MakeMove(c)
		if ns != nil {
			v, _, _, _, _ := ns.Minimax(depth-1, origPlayer)
			if (s.Turn == origPlayer && v > bestValue) || (s.Turn != origPlayer && v < bestValue) {
				bestValue = v
				bCol, bRow, bTurn, bs = cc, rr, tt, ns
			}
		}
	}
	return bestValue, bCol, bRow, bTurn, bs
}

// Return score, and (if won) index of winning move in possibilities
func (s *GameState) Score(ourPlayer int) (int, int) {
	rv := 0
	for pi, p := range AllPossibles {
		for pl := 0; pl < Players; pl++ {
			cnt := 0
			for _, cr := range p {
				sv := s.State[cr]
				if sv == pl {
					cnt++
				} else if sv != Empty {
					cnt = 0
					break
				}
			}
			if cnt == Target {
				if pl == ourPlayer {
					return Won, pi
				}
				return Lost, -1
			}
			if pl == ourPlayer {
				rv += (cnt * cnt)
			}
		}
	}
	return rv, -1
}

func reallyRun() error {
	win, err := pixelgl.NewWindow(pixelgl.WindowConfig{
		Title:  "Connect4",
		Bounds: pixel.R(0, 0, WinWidth, WinHeight),
		VSync:  true,
	})
	if err != nil {
		return err
	}
	cWidth := float64(WinWidth) / Columns
	rad := cWidth / 2.5

	bp := pixel.PictureDataFromImage(&board{
		color:  color.RGBA{102, 204, 255, 255},
		cWidth: cWidth,
		r:      rad,
	})
	board := pixel.NewSprite(bp, bp.Bounds())
	boardM := pixel.IM
	boardM = boardM.Moved(pixel.Vec{X: (WinWidth / 2), Y: (WinHeight / 2)})

	p1 := pixel.PictureDataFromImage(&circle{
		p: image.Point{X: int(rad), Y: int(rad)},
		r: int(rad),
		c: color.RGBA{255, 252, 51, 255},
	})
	p2 := pixel.PictureDataFromImage(&circle{
		p: image.Point{X: int(rad), Y: int(rad)},
		r: int(rad),
		c: color.RGBA{249, 60, 60, 255},
	})

	players := []*pixel.Sprite{
		pixel.NewSprite(p1, p1.Bounds()),
		pixel.NewSprite(p2, p2.Bounds()),
	}

	g := NewGame()
	state := StateIdle
	var destCol, destRow, destTurn int
	var ns *GameState
	var startTime, ttl time.Time
	dirty := true

	for !win.Closed() {
		drawAtEnd := false
		if dirty {
			win.Clear(colornames.Whitesmoke)

			for r := 0; r < Rows; r++ {
				for c := 0; c < Columns; c++ {
					s := g.State[toStateIdx(c, r)]
					if s != Empty {
						if !(state == 1 && c == destCol && r == destRow) {
							mat := pixel.IM
							mat = mat.Moved(pixel.Vec{X: (float64(c) + 0.5) * cWidth, Y: (float64(r) + 0.5) * cWidth})
							players[s].Draw(win, mat)
						}
					}
				}
			}
			drawAtEnd = true
			dirty = false
		}

		switch state {
		case StateIdle:
			if g.Turn == 0 {
				if win.JustPressed(pixelgl.MouseButtonLeft) {
					colClicked := int(win.MousePosition().X / cWidth)
					destCol, destRow, destTurn, ns = g.MakeMove(colClicked)
					if ns != nil {
						g = ns
						state = StateFalling
						startTime = time.Now()
					}
				}
			} else {
				_, destCol, destRow, destTurn, ns = g.Minimax(4, g.Turn)
				if ns == nil {
					log.Fatal("Computer can't find a move.")
					break
				}
				g = ns
				state = StateFalling
				startTime = time.Now()
			}
		case StateFalling:
			n := time.Now()

			startPos := float64(Rows + 1)
			endPos := float64(destRow) + 0.5
			delta := endPos - startPos

			percent := n.Sub(startTime).Seconds() / TimeToDrop

			mat := pixel.IM
			mat = mat.Moved(pixel.Vec{X: (float64(destCol) + 0.5) * cWidth, Y: (startPos + (delta * percent)) * cWidth})

			players[destTurn].Draw(win, mat)
			dirty = true

			if percent >= 0.99 {
				if g.Finished {
					state = StateFinished
					startTime = time.Now()
					ttl = startTime.Add(time.Minute)
				} else {
					state = StateIdle
				}
			}

		case StateFinished:
			if win.JustPressed(pixelgl.MouseButtonLeft) {
				g = NewGame()
				state = StateIdle
			} else {
				board.Draw(win, boardM)
				drawAtEnd = false
				n := time.Now()
				if n.After(ttl) {
					return nil
				}
				percent := n.Sub(startTime).Seconds() / TimeToDrop
				for _, cr := range AllPossibles[g.WinningMove] {
					c, r := toColRow(cr)

					mat := pixel.IM
					mat = mat.Scaled(pixel.Vec{}, 1.0+((math.Sin(percent*2.0*math.Pi)+1.0)/5.0))
					mat = mat.Moved(pixel.Vec{X: (float64(c) + 0.5) * cWidth, Y: (float64(r) + 0.5) * cWidth})

					players[g.State[cr]].Draw(win, mat)
				}
			}
			dirty = true
		}
		if drawAtEnd {
			board.Draw(win, boardM)
		}
		win.Update()
	}
	return nil
}

func main() {
	pixelgl.Run(func() {
		err := reallyRun()
		if err != nil {
			log.Fatal(err)
		}
	})
}

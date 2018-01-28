package main

import (
	"image"
	"image/color"
	"math"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
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
	return image.Rect(0, 0, winWidth, winHeight)
}

func (c *board) At(x, y int) color.Color {
	xx := float64(x) - (c.cWidth * (0.5 + float64(x/int(c.cWidth)))) + 0.5
	yy := float64(y) - (c.cWidth * (0.5 + float64(y/int(c.cWidth)))) + 0.5
	if xx*xx+yy*yy >= c.r*c.r {
		return c.color
	}
	return color.RGBA{0, 0, 0, 0}
}

type connect4Window struct {
	win     *pixelgl.Window
	board   *pixel.Sprite
	players []*pixel.Sprite
	boardM  pixel.Matrix
	cWidth  float64
}

func (w *connect4Window) Closed() bool {
	return w.win.Closed()
}

func (w *connect4Window) Update() {
	w.win.Update()
}

func (w *connect4Window) JustClickedAutopilot() bool {
	return w.win.JustPressed(pixelgl.KeyA)
}

func (w *connect4Window) JustClickedFaster() bool {
	return w.win.JustPressed(pixelgl.KeyF)
}

func (w *connect4Window) JustClickedLeftArrow() bool {
	return w.win.JustPressed(pixelgl.KeyLeft)
}

func (w *connect4Window) JustClickedRightArrow() bool {
	return w.win.JustPressed(pixelgl.KeyRight)
}

func (w *connect4Window) JustClickedCol() int {
	if w.win.JustPressed(pixelgl.MouseButtonLeft) {
		return int(w.win.MousePosition().X / w.cWidth)
	}
	return -1
}

func (w *connect4Window) Render(g *gameState, state int, startTime time.Time, timeToDrop float64) {
	w.win.Clear(colornames.Whitesmoke)

	for r := 0; r < rows; r++ {
		for c := 0; c < columns; c++ {
			s := g.State[toStateIdx(c, r)]
			if s != empty {
				mat := pixel.IM
				yPos := float64(r) + 0.5
				if state == stateFalling && c == g.LastMoveCol && r == g.LastMoveRow {
					yPos = (float64(rows+1) + (((float64(g.LastMoveRow) + 0.5) - float64(rows+1)) * (time.Now().Sub(startTime).Seconds() / timeToDrop)))
				}
				mat = mat.Moved(pixel.Vec{X: (float64(c) + 0.5) * w.cWidth, Y: yPos * w.cWidth})
				w.players[s].Draw(w.win, mat)
			}
		}
	}

	w.board.Draw(w.win, w.boardM)
	if state == stateFinished {
		percent := time.Now().Sub(startTime).Seconds() / timeToDrop
		for _, cr := range allPossibles[g.WinningMove] {
			c, r := toColRow(cr)

			mat := pixel.IM
			mat = mat.Scaled(pixel.Vec{}, 1.0+((math.Sin(percent*2.0*math.Pi)+1.0)/5.0))
			mat = mat.Moved(pixel.Vec{X: (float64(c) + 0.5) * w.cWidth, Y: (float64(r) + 0.5) * w.cWidth})

			w.players[g.State[cr]].Draw(w.win, mat)
		}

	}
}

func newConnect4Window(width, height int) (*connect4Window, error) {
	win, err := pixelgl.NewWindow(pixelgl.WindowConfig{
		Title:  "Connect4",
		Bounds: pixel.R(0, 0, float64(width), float64(height)),
		VSync:  true,
	})
	if err != nil {
		return nil, err
	}
	cWidth := float64(winWidth) / columns
	rad := cWidth / 2.5

	bp := pixel.PictureDataFromImage(&board{
		color:  color.RGBA{102, 204, 255, 255},
		cWidth: cWidth,
		r:      rad,
	})
	board := pixel.NewSprite(bp, bp.Bounds())
	boardM := pixel.IM
	boardM = boardM.Moved(pixel.Vec{X: (winWidth / 2), Y: (winHeight / 2)})

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

	return &connect4Window{
		win:     win,
		players: players,
		board:   board,
		boardM:  boardM,
		cWidth:  cWidth,
	}, nil
}

package main

import (
	"errors"
	"image"
	"image/color"
	"log"
	"math"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
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

	stateIdle     = 0
	stateFalling  = 1
	stateFinished = 2
)

func main() {
	pixelgl.Run(func() {
		err := func() error {
			win, err := pixelgl.NewWindow(pixelgl.WindowConfig{
				Title:  "Connect4",
				Bounds: pixel.R(0, 0, winWidth, winHeight),
				VSync:  true,
			})
			if err != nil {
				return err
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

			g := newGame()
			state := stateIdle
			var startTime, ttl time.Time
			dirty := true

			for !win.Closed() {
				drawAtEnd := false
				if dirty {
					win.Clear(colornames.Whitesmoke)

					for r := 0; r < rows; r++ {
						for c := 0; c < columns; c++ {
							s := g.State[toStateIdx(c, r)]
							if s != empty {
								if !(state == stateFalling && c == g.LastMoveCol && r == g.LastMoveRow) {
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
				case stateIdle:
					if g.Turn == 0 {
						if win.JustPressed(pixelgl.MouseButtonLeft) {
							colClicked := int(win.MousePosition().X / cWidth)
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
							return errors.New("computer can't find a move")
						}
						g = ns
						state = stateFalling
						startTime = time.Now()
					}
				case stateFalling:
					n := time.Now()

					startPos := float64(rows + 1)
					endPos := float64(g.LastMoveRow) + 0.5
					delta := endPos - startPos

					percent := n.Sub(startTime).Seconds() / timeToDrop

					mat := pixel.IM
					mat = mat.Moved(pixel.Vec{X: (float64(g.LastMoveCol) + 0.5) * cWidth, Y: (startPos + (delta * percent)) * cWidth})

					players[g.LastMoveTurn].Draw(win, mat)
					dirty = true

					if percent >= 0.99 {
						if g.Finished {
							state = stateFinished
							startTime = time.Now()
							ttl = startTime.Add(time.Minute)
						} else {
							state = stateIdle
						}
					}

				case stateFinished:
					if win.JustPressed(pixelgl.MouseButtonLeft) {
						g = newGame()
						state = stateIdle
					} else {
						board.Draw(win, boardM)
						drawAtEnd = false
						n := time.Now()
						if n.After(ttl) {
							return nil
						}
						percent := n.Sub(startTime).Seconds() / timeToDrop
						for _, cr := range allPossibles[g.WinningMove] {
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
		}()
		if err != nil {
			log.Fatal(err)
		}
	})
}

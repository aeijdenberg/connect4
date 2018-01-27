package main

import (
	"image"
	"image/color"
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

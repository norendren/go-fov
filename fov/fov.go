package fov

import (
	"fmt"
	"math"
)

type GridMap interface {
	Index(x, y int) (int, int)
	InBounds(x, y int) bool
	IsOpaque(x, y int) bool
}

type gridSet map[string]struct{}

type MapFOV struct {
	Visible gridSet
}

func (m *MapFOV) IsVisible(x, y int) bool {
	if _, ok := m.Visible[fmt.Sprintf("%d%d", x, y)]; !ok {
		return false
	}
	return true
}

func (m *MapFOV) ComputeFOV(grid GridMap, px, py, r int) {
	m.Visible = make(map[string]struct{})
	m.Visible[fmt.Sprintf("%d%d", px, py)] = struct{}{}
	for i := 1; i <= 8; i++ {
		m.fov(grid, px, py, 1, 0, 1, i, r)
	}
}

func (m *MapFOV) fov(grid GridMap, px, py, dist int, lowSlope, highSlope float64, oct, rad int) {
	if dist > rad {
		return
	}
	low := math.Floor(lowSlope*float64(dist) + 0.5)
	high := math.Floor(highSlope*float64(dist) + 0.5)
	inGap := false

	for height := low; height <= high; height++ {
		mapx, mapy := distHeightXY(px, py, dist, int(height), oct)
		if grid.InBounds(mapx, mapy) && distTo(px, py, mapx, mapy) < rad {
			m.Visible[fmt.Sprintf("%d%d", mapx, mapy)] = struct{}{}
		}

		if grid.InBounds(mapx, mapy) && !grid.IsOpaque(mapx, mapy) {
			if inGap {
				m.fov(grid, px, py, dist+1, lowSlope, (height-0.5)/float64(dist), oct, rad)
			}
			lowSlope = (height + 0.5) / float64(dist)
			inGap = false
		} else {
			inGap = true
			if height == high {
				m.fov(grid, px, py, dist+1, lowSlope, highSlope, oct, rad)
			}
		}
	}
}

func distHeightXY(px, py, d, h, oct int) (int, int) {
	if oct&0x1 > 0 {
		d = -d
	}
	if oct&0x2 > 0 {
		h = -h
	}
	if oct&0x4 > 0 {
		return px + h, py + d
	}
	return px + d, py + h
}

func distTo(x1, y1, x2, y2 int) int {
	vx := math.Pow(float64(x1-x2), 2)
	vy := math.Pow(float64(y1-y2), 2)
	return int(math.Sqrt(vx + vy))
}

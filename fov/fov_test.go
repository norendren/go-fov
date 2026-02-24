package fov

import "testing"

// mockGrid implements GridMap for testing using a 2D bool slice (true = opaque/wall).
type mockGrid struct {
	width, height int
	walls         [][]bool
}

func newOpenGrid(w, h int) *mockGrid {
	walls := make([][]bool, h)
	for i := range walls {
		walls[i] = make([]bool, w)
	}
	return &mockGrid{width: w, height: h, walls: walls}
}

func (g *mockGrid) InBounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < g.width && y < g.height
}

func (g *mockGrid) IsOpaque(x, y int) bool {
	if !g.InBounds(x, y) {
		return false
	}
	return g.walls[y][x]
}

func TestPlayerTileAlwaysVisible(t *testing.T) {
	v := New()
	v.Compute(newOpenGrid(10, 10), 5, 5, 3)
	if !v.IsVisible(5, 5) {
		t.Error("player's own tile should always be visible")
	}
}

func TestOpenMapTilesVisibleWithinRadius(t *testing.T) {
	grid := newOpenGrid(20, 20)
	px, py, radius := 10, 10, 6
	v := New()
	v.Compute(grid, px, py, radius)

	cardinals := [][2]int{
		{px + 1, py}, {px - 1, py}, {px, py + 1}, {px, py - 1},
		{px + 3, py}, {px, py + 3},
		{px + 2, py + 2}, {px - 2, py + 2},
	}
	for _, pos := range cardinals {
		if !v.IsVisible(pos[0], pos[1]) {
			t.Errorf("tile (%d,%d) should be visible in open map within radius %d", pos[0], pos[1], radius)
		}
	}
}

func TestTilesOutsideRadiusNotVisible(t *testing.T) {
	grid := newOpenGrid(30, 30)
	v := New()
	v.Compute(grid, 15, 15, 4)

	far := [][2]int{{0, 0}, {29, 29}, {0, 29}, {29, 0}}
	for _, pos := range far {
		if v.IsVisible(pos[0], pos[1]) {
			t.Errorf("tile (%d,%d) should not be visible beyond radius", pos[0], pos[1])
		}
	}
}

func TestZeroRadiusOnlyPlayerVisible(t *testing.T) {
	v := New()
	v.Compute(newOpenGrid(10, 10), 5, 5, 0)

	if !v.IsVisible(5, 5) {
		t.Error("player tile should be visible with radius 0")
	}
	if v.IsVisible(6, 5) {
		t.Error("adjacent tile should not be visible with radius 0")
	}
}

func TestOpaqueWallBlocksTilesBehindit(t *testing.T) {
	grid := newOpenGrid(10, 10)
	// Wall directly north of player (lower y = north in grid coords)
	grid.walls[3][5] = true // wall at (5,3), player at (5,5)

	v := New()
	v.Compute(grid, 5, 5, 6)

	if !v.IsVisible(5, 4) {
		t.Error("tile just south of wall should be visible")
	}
	if !v.IsVisible(5, 3) {
		t.Error("wall tile itself should be visible")
	}
	if v.IsVisible(5, 2) {
		t.Error("tile directly behind wall should not be visible")
	}
	if v.IsVisible(5, 1) {
		t.Error("tile two behind wall should not be visible")
	}
	// Tiles adjacent to the shadow should still be lit
	if !v.IsVisible(4, 2) {
		t.Error("tile beside the shadow should be visible")
	}
}

func TestComputeResetsVisibleSet(t *testing.T) {
	grid := newOpenGrid(30, 30)
	v := New()

	v.Compute(grid, 2, 2, 3)
	if !v.IsVisible(2, 2) {
		t.Fatal("setup: (2,2) should be visible after first Compute")
	}

	// Recompute far away — (2,2) must no longer be visible
	v.Compute(grid, 27, 27, 3)
	if v.IsVisible(2, 2) {
		t.Error("old player position should not be visible after recompute")
	}
	if !v.IsVisible(27, 27) {
		t.Error("new player position should be visible after recompute")
	}
}

// Reading from a nil map in Go is safe; this just confirms IsVisible doesn't panic.
func TestIsVisibleBeforeCompute(t *testing.T) {
	v := New()
	if v.IsVisible(0, 0) {
		t.Error("nothing should be visible before Compute is called")
	}
}

func TestPlayerAtMapBoundary(t *testing.T) {
	grid := newOpenGrid(10, 10)
	v := New()
	v.Compute(grid, 0, 0, 3) // corner — many neighbours are out of bounds
	if !v.IsVisible(0, 0) {
		t.Error("player tile at boundary should be visible")
	}
}

func TestDistTo(t *testing.T) {
	tests := []struct {
		x1, y1, x2, y2, want int
	}{
		{0, 0, 0, 0, 0},
		{0, 0, 3, 4, 5}, // 3-4-5 right triangle
		{0, 0, 0, 5, 5}, // vertical
		{0, 0, 5, 0, 5}, // horizontal
	}
	for _, tt := range tests {
		if got := distTo(tt.x1, tt.y1, tt.x2, tt.y2); got != tt.want {
			t.Errorf("distTo(%d,%d,%d,%d) = %d, want %d", tt.x1, tt.y1, tt.x2, tt.y2, got, tt.want)
		}
	}
}

// distHeightXY maps (dist, height, octant) → grid coords relative to (px, py).
// Octant bits: bit0 negates d, bit1 negates h, bit2 transposes (swaps d and h role).
func TestDistHeightXY(t *testing.T) {
	px, py, d, h := 5, 5, 2, 1
	tests := []struct {
		oct        int
		wantX, wantY int
	}{
		{1, 3, 6}, // d=-2, h=1  → px+d, py+h = (3,6)
		{2, 7, 4}, // d=2,  h=-1 → px+d, py+h = (7,4)
		{3, 3, 4}, // d=-2, h=-1 → px+d, py+h = (3,4)
		{4, 6, 7}, // d=2,  h=1  → px+h, py+d = (6,7)
		{5, 6, 3}, // d=-2, h=1  → px+h, py+d = (6,3)
		{6, 4, 7}, // d=2,  h=-1 → px+h, py+d = (4,7)
		{7, 4, 3}, // d=-2, h=-1 → px+h, py+d = (4,3)
		{8, 7, 6}, // no flags   → px+d, py+h = (7,6)
	}
	for _, tt := range tests {
		gotX, gotY := distHeightXY(px, py, d, h, tt.oct)
		if gotX != tt.wantX || gotY != tt.wantY {
			t.Errorf("distHeightXY(oct=%d) = (%d,%d), want (%d,%d)",
				tt.oct, gotX, gotY, tt.wantX, tt.wantY)
		}
	}
}

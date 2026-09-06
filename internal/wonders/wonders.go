package wonders

import "math"

type Block struct {
	X, Y, Z    int
	ID, Meta   int
}

type Wonder struct {
	Key       string
	Name      string
	Lat, Lng  float64
	Honorary  bool
	build     func() []Block
}

func (w Wonder) Footprint() []Block { return w.build() }

const (
	stone     = 1
	cobble    = 4
	sandstone = 24
	stonebr   = 98
	quartz    = 155
	clay      = 159
	wool      = 35
	glass     = 20
	gold      = 41
)

func filledCircle(cx, cz, r, y, id, meta int, out *[]Block) {
	for dx := -r; dx <= r; dx++ {
		for dz := -r; dz <= r; dz++ {
			if dx*dx+dz*dz <= r*r {
				*out = append(*out, Block{cx + dx, y, cz + dz, id, meta})
			}
		}
	}
}

func ring(cx, cz, r, y, id, meta int, out *[]Block) {
	for a := 0.0; a < 2*math.Pi; a += 0.14 {
		x := cx + int(math.Round(float64(r)*math.Cos(a)))
		z := cz + int(math.Round(float64(r)*math.Sin(a)))
		*out = append(*out, Block{x, y, z, id, meta})
	}
}

func box(x0, z0, x1, z1, y, id, meta int, out *[]Block) {
	for x := x0; x <= x1; x++ {
		for z := z0; z <= z1; z++ {
			*out = append(*out, Block{x, y, z, id, meta})
		}
	}
}

func steppedPyramid(cx, cz, base, height, id, meta int) []Block {
	var out []Block
	for level := 0; level < height; level++ {
		r := base - level
		if r < 0 {
			break
		}
		if r <= 1 {
			box(cx-r, cz-r, cx+r, cz+r, level, id, meta, &out)
			continue
		}
		for x := cx - r; x <= cx+r; x++ {
			out = append(out, Block{x, level, cz - r, id, meta})
			out = append(out, Block{x, level, cz + r, id, meta})
		}
		for z := cz - r + 1; z <= cz+r-1; z++ {
			out = append(out, Block{cx - r, level, z, id, meta})
			out = append(out, Block{cx + r, level, z, id, meta})
		}
	}
	return out
}

func colosseum() []Block {
	var out []Block
	for y := 0; y < 5; y++ {
		ring(0, 0, 11, y, stonebr, 0, &out)
		ring(0, 0, 10, y, stonebr, 0, &out)
		ring(0, 0, 6, y, cobble, 0, &out)
	}
	filledCircle(0, 0, 5, 0, stone, 0, &out)
	return out
}

func pyramid() []Block { return steppedPyramid(0, 0, 8, 9, sandstone, 0) }

func elCastillo() []Block {
	out := steppedPyramid(0, 0, 7, 8, stonebr, 0)
	for y := 9; y < 12; y++ {
		box(-2, -2, 2, 2, y, stonebr, 0, &out)
	}
	return out
}

func taj() []Block {
	var out []Block
	for x := -10; x <= 10; x++ {
		out = append(out, Block{x, 0, -10, quartz, 0}, Block{x, 0, 10, quartz, 0})
	}
	for z := -9; z <= 9; z++ {
		out = append(out, Block{-10, 0, z, quartz, 0}, Block{10, 0, z, quartz, 0})
	}
	for y := 1; y <= 6; y++ {
		r := 5 - (y-1)/2
		filledCircle(0, 0, r, y, quartz, 0, &out)
	}
	filledCircle(0, 0, 1, 7, gold, 0, &out)
	for _, c := range [][2]int{{-9, -9}, {9, -9}, {-9, 9}, {9, 9}} {
		for y := 1; y <= 9; y++ {
			out = append(out, Block{c[0], y, c[1], quartz, 0})
		}
	}
	return out
}

func redeemer() []Block {
	var out []Block
	out = append(out, steppedPyramid(0, 0, 3, 3, stone, 0)...)
	for y := 3; y <= 10; y++ {
		out = append(out, Block{0, y, 0, quartz, 0})
	}
	for x := -4; x <= 4; x++ {
		out = append(out, Block{x, 9, 0, quartz, 0})
	}
	out = append(out, Block{0, 11, 0, quartz, 0})
	return out
}

func greatWall() []Block {
	var out []Block
	for x := -16; x <= 16; x++ {
		h := 3
		if (x+16)%8 == 0 {
			h = 5
		}
		for y := 0; y < h; y++ {
			out = append(out, Block{x, y, 0, cobble, 0})
		}
		if h == 5 {
			out = append(out, Block{x, 3, -1, cobble, 0}, Block{x, 3, 1, cobble, 0})
		}
	}
	return out
}

func machuPicchu() []Block {
	var out []Block
	for t := 0; t < 6; t++ {
		box(-8+t, -8+t, 8-t, 8-t, t, stone, 0, &out)
	}
	for _, c := range [][2]int{{-4, -4}, {-4, 2}, {2, -4}, {2, 2}} {
		box(c[0], c[1], c[0]+2, c[1]+2, 6, cobble, 0, &out)
	}
	return out
}

func petra() []Block {
	var out []Block
	box(-7, 0, 7, 0, 0, sandstone, 0, &out)
	for y := 1; y <= 10; y++ {
		for x := -7; x <= 7; x++ {
			out = append(out, Block{x, y, 0, sandstone, 0})
		}
	}
	box(-2, 0, 2, 0, 1, 0, 0, &out)
	for y := 1; y <= 4; y++ {
		for x := -1; x <= 1; x++ {
			out = append(out, Block{x, y, 0, 0, 0})
		}
	}
	for _, x := range []int{-5, -2, 2, 5} {
		for y := 1; y <= 6; y++ {
			out = append(out, Block{x, y, 1, sandstone, 0})
		}
	}
	return out
}

var List = []Wonder{
	{"colosseum", "Colosseum", 41.8902, 12.4922, false, colosseum},
	{"tajmahal", "Taj Mahal", 27.1751, 78.0421, false, taj},
	{"christredeemer", "Christ the Redeemer", -22.9519, -43.2105, false, redeemer},
	{"machupicchu", "Machu Picchu", -13.1631, -72.5450, false, machuPicchu},
	{"chichenitza", "Chichen Itza", 20.6829, -88.5686, false, elCastillo},
	{"greatwall", "Great Wall of China", 40.4319, 116.5704, false, greatWall},
	{"petra", "Petra", 30.3285, 35.4444, false, petra},
	{"giza", "Great Pyramid of Giza", 29.9792, 31.1342, true, pyramid},
}

func ByKey(k string) (Wonder, bool) {
	for _, w := range List {
		if w.Key == k {
			return w, true
		}
	}
	return Wonder{}, false
}

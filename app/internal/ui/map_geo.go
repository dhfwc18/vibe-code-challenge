package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// regionDrawOrder defines the fixed draw / hit-test order for region polygons.
// Smaller / more specific regions appear later (drawn on top) and are
// preferred in hit-testing when a click lands on a shared edge.
var regionDrawOrder = []string{
	"northern_highlands",
	"eastern_lowlands",
	"northern_industrial",
	"pennine_corridor",
	"north_west_cities",
	"east_midlands",
	"west_midlands",
	"eastern_counties",
	"western_coast",
	"capital_region",
	"south_east",
	"south_west",
}

// regionPolygons defines normalized [0,1] polygon vertices for each of the
// 12 Taitan regions. (0,0) = top-left of the map canvas, (1,1) = bottom-right.
// Adjacent polygons share edges exactly; there are no gaps or overlaps.
// The outline approximates a UK-equivalent island with the north at the top.
var regionPolygons = map[string][][2]float32{
	// Northern Highlands: clean rectangle, top-left band.
	"northern_highlands": {
		{0.00, 0.00}, {0.64, 0.00}, {0.64, 0.38}, {0.00, 0.38},
	},
	// Eastern Lowlands: clean rectangle, full right strip from the top.
	"eastern_lowlands": {
		{0.64, 0.00}, {1.00, 0.00}, {1.00, 0.53}, {0.64, 0.53},
	},
	// Northern Industrial Belt: clean rectangle below N.High on the left.
	"northern_industrial": {
		{0.00, 0.38}, {0.36, 0.38}, {0.36, 0.53}, {0.00, 0.53},
	},
	// Pennine Corridor: center strip between NI Belt and Eastern Lowlands.
	"pennine_corridor": {
		{0.36, 0.38}, {0.64, 0.38}, {0.64, 0.53}, {0.36, 0.53},
	},
	// North West Cities: center-left, below NI Belt.
	"north_west_cities": {
		{0.00, 0.53}, {0.26, 0.53}, {0.26, 0.67}, {0.00, 0.67},
	},
	// West Midlands: center, below Pennine (left half).
	"west_midlands": {
		{0.26, 0.53}, {0.52, 0.53}, {0.52, 0.67}, {0.26, 0.67},
	},
	// East Midlands: center, below Pennine (right half).
	"east_midlands": {
		{0.52, 0.53}, {0.72, 0.53}, {0.72, 0.67}, {0.52, 0.67},
	},
	// Eastern Counties: right-side spur spanning two rows.
	"eastern_counties": {
		{0.72, 0.53}, {1.00, 0.53}, {1.00, 0.82},
		{0.88, 0.82}, {0.78, 0.79}, {0.72, 0.74}, {0.72, 0.67},
	},
	// Western Coast: simple west-side strip (Wales analogue).
	// Clean rectangle sharing x=0.20 with capital_region and y=0.84 with south_west.
	"western_coast": {
		{0.00, 0.67}, {0.20, 0.67}, {0.20, 0.84}, {0.00, 0.84},
	},
	// Capital Region: inner city, suburbs, outer commuter zone.
	// South boundary: (0.20,0.84)->(0.40,0.88)->(0.72,0.84) shared with SW/SE.
	// Right boundary: (0.72,0.74)->(0.78,0.79)->(0.88,0.82) shared with eastern_counties.
	"capital_region": {
		{0.20, 0.67}, {0.72, 0.67}, {0.72, 0.74},
		{0.78, 0.79}, {0.88, 0.82}, {0.72, 0.84},
		{0.40, 0.88}, {0.20, 0.84},
	},
	// South East: lower-right, butts against capital_region and eastern_counties.
	// Top boundary: (0.72,0.84)->(0.88,0.82)->(1.00,0.82) shared with cap/counties.
	"south_east": {
		{0.72, 0.84}, {0.88, 0.82}, {1.00, 0.82},
		{1.00, 1.00}, {0.40, 1.00}, {0.40, 0.88},
	},
	// South West: lower-left peninsula, shares y=0.84 with western_coast and
	// the diagonal (0.40,0.88) boundary with capital_region and south_east.
	"south_west": {
		{0.00, 0.84}, {0.20, 0.84}, {0.40, 0.88},
		{0.40, 1.00}, {0.00, 1.00},
	},
}

// regionShortName maps region IDs to abbreviated map labels.
var regionShortName = map[string]string{
	"northern_highlands":  "N.High",
	"eastern_lowlands":    "E.Low",
	"northern_industrial": "N.Ind",
	"pennine_corridor":    "Pen",
	"north_west_cities":   "NW",
	"west_midlands":       "W.Mid",
	"east_midlands":       "E.Mid",
	"eastern_counties":    "E.Co",
	"western_coast":       "W.Co",
	"capital_region":      "Cap",
	"south_east":          "SE",
	"south_west":          "SW",
}

// pointInPolygon reports whether (px, py) is inside the polygon using ray casting.
func pointInPolygon(px, py float32, pts [][2]float32) bool {
	n := len(pts)
	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		xi, yi := pts[i][0], pts[i][1]
		xj, yj := pts[j][0], pts[j][1]
		if ((yi > py) != (yj > py)) && px < (xj-xi)*(py-yi)/(yj-yi)+xi {
			inside = !inside
		}
		j = i
	}
	return inside
}

// hitTestMap returns the region ID whose polygon contains the normalized point
// (nx, ny) in [0,1]x[0,1]. Uses regionDrawOrder so the result is deterministic
// when a click lands on a shared edge between two regions.
func hitTestMap(nx, ny float32) string {
	// Iterate in reverse draw order: last-drawn (most specific) wins.
	for i := len(regionDrawOrder) - 1; i >= 0; i-- {
		id := regionDrawOrder[i]
		pts := regionPolygons[id]
		if pointInPolygon(nx, ny, pts) {
			return id
		}
	}
	return ""
}

// whitePixel is a lazily initialized 1x1 white image used as the DrawTriangles
// source texture. All colour comes from vertex ColorR/G/B/A.
var whitePixel *ebiten.Image

func getWhitePixel() *ebiten.Image {
	if whitePixel == nil {
		whitePixel = ebiten.NewImage(1, 1)
		whitePixel.Fill(color.White)
	}
	return whitePixel
}

// fillMapPolygon draws a filled polygon scaled to the map canvas at (ox,oy,sw,sh).
func fillMapPolygon(screen *ebiten.Image, pts [][2]float32, col color.RGBA, ox, oy, sw, sh float32) {
	if len(pts) < 3 {
		return
	}
	var p vector.Path
	p.MoveTo(pts[0][0]*sw+ox, pts[0][1]*sh+oy)
	for _, pt := range pts[1:] {
		p.LineTo(pt[0]*sw+ox, pt[1]*sh+oy)
	}
	p.Close()
	vs, is := p.AppendVerticesAndIndicesForFilling(nil, nil)
	r := float32(col.R) / 255
	g := float32(col.G) / 255
	b := float32(col.B) / 255
	for i := range vs {
		vs[i].ColorR = r
		vs[i].ColorG = g
		vs[i].ColorB = b
		vs[i].ColorA = 1
	}
	screen.DrawTriangles(vs, is, getWhitePixel(), &ebiten.DrawTrianglesOptions{AntiAlias: true})
}

// strokeMapPolygon draws a closed polygon outline (1-pixel stroke) scaled to
// the map canvas. Used to draw visible borders between adjacent regions.
func strokeMapPolygon(screen *ebiten.Image, pts [][2]float32, col color.RGBA, ox, oy, sw, sh float32) {
	if len(pts) < 2 {
		return
	}
	var p vector.Path
	p.MoveTo(pts[0][0]*sw+ox, pts[0][1]*sh+oy)
	for _, pt := range pts[1:] {
		p.LineTo(pt[0]*sw+ox, pt[1]*sh+oy)
	}
	p.Close()
	vs, is := p.AppendVerticesAndIndicesForStroke(nil, nil, &vector.StrokeOptions{
		Width:    1.5,
		LineJoin: vector.LineJoinMiter,
		LineCap:  vector.LineCapButt,
	})
	r := float32(col.R) / 255
	g := float32(col.G) / 255
	b := float32(col.B) / 255
	for i := range vs {
		vs[i].ColorR = r
		vs[i].ColorG = g
		vs[i].ColorB = b
		vs[i].ColorA = 1
	}
	screen.DrawTriangles(vs, is, getWhitePixel(), &ebiten.DrawTrianglesOptions{AntiAlias: false})
}

// polygonLabelPos returns the screen centroid (average of vertices) scaled to
// the map canvas at (ox,oy,sw,sh).
func polygonLabelPos(pts [][2]float32, ox, oy, sw, sh float32) (float32, float32) {
	var cx, cy float32
	for _, pt := range pts {
		cx += pt[0]
		cy += pt[1]
	}
	n := float32(len(pts))
	return cx/n*sw + ox, cy/n*sh + oy
}

// overlayColour maps a 0-100 value to a heat-map colour for the given overlay.
func overlayColour(ov mapOverlay, val float64) color.RGBA {
	t := val / 100.0
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	switch ov {
	case overlayFuelPoverty:
		// low = good (green), high = bad (red)
		return lerpRGBA(colour(0x27, 0xAE, 0x60), colour(0xC0, 0x39, 0x2B), t)
	case overlayPolitical:
		// 0 = left (blue), 100 = right (warm red)
		return lerpRGBA(colour(0x29, 0x80, 0xB9), colour(0xC0, 0x39, 0x2B), t)
	case overlayInsulation:
		// low = bad (red), high = good (green)
		return lerpRGBA(colour(0xC0, 0x39, 0x2B), colour(0x27, 0xAE, 0x60), t)
	}
	return ColourPanel
}

// lerpRGBA linearly interpolates between two RGBA colours.
func lerpRGBA(a, b color.RGBA, t float64) color.RGBA {
	lerp := func(x, y uint8) uint8 {
		return uint8(float64(x)*(1-t) + float64(y)*t)
	}
	return color.RGBA{R: lerp(a.R, b.R), G: lerp(a.G, b.G), B: lerp(a.B, b.B), A: 0xFF}
}

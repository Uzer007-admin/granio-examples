package widget

import "sort"

// collectionGeometry indexes sparse size overrides. Positions need logarithmic
// lookup in the override count, independent of the total number of data rows.
type collectionGeometry struct {
	count, overrides int
	base, padding    float32
	keys             []int
	deltas           []float64
	dirty            bool
}

func (g *collectionGeometry) prepare(values map[int]float32, count int, base, padding float32) {
	if !g.dirty && g.count == count && g.overrides == len(values) && g.base == base && g.padding == padding {
		return
	}
	g.count, g.overrides, g.base, g.padding, g.dirty = count, len(values), base, padding, false
	g.keys = g.keys[:0]
	for id := range values {
		if id >= 0 && id < count {
			g.keys = append(g.keys, id)
		}
	}
	sort.Ints(g.keys)
	g.deltas = g.deltas[:0]
	total := float64(0)
	for _, id := range g.keys {
		total += float64(values[id] - base)
		g.deltas = append(g.deltas, total)
	}
}

func (g *collectionGeometry) position(id int) float32 {
	position := float64(id) * float64(g.base+g.padding)
	if before := sort.SearchInts(g.keys, id); before > 0 {
		position += g.deltas[before-1]
	}
	return float32(position)
}

func (g *collectionGeometry) itemAt(position float32) int {
	if g.count <= 0 {
		return 0
	}
	index := sort.Search(g.count, func(id int) bool { return g.position(id+1) > position })
	if index >= g.count {
		return g.count - 1
	}
	return index
}

func (g *collectionGeometry) endAt(position float32) int {
	return sort.Search(g.count, func(id int) bool { return g.position(id) >= position })
}

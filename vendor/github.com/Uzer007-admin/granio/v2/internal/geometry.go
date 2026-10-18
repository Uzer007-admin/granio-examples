package internal

import "github.com/Uzer007-admin/granio/v2"

// MaxSizes returns the element-wise maximum of the two sizes. It is what
// granio.Size.Max does, minus boxing the argument into the Vector2 parameter,
// which heap-allocates in per-frame layout code.
func MaxSizes(a, b granio.Size) granio.Size {
	return granio.Size{Width: max(a.Width, b.Width), Height: max(a.Height, b.Height)}
}

// MinSizes returns the element-wise minimum of the two sizes, allocation-free
// like MaxSizes.
func MinSizes(a, b granio.Size) granio.Size {
	return granio.Size{Width: min(a.Width, b.Width), Height: min(a.Height, b.Height)}
}

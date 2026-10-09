package floats

import (
	"github.com/shogo82148/ints"
)

var optimized = true

var lowerHex = "0123456789abcdef"
var upperHex = "0123456789ABCDEF"

func nonzero16(x uint16) uint16 {
	if x != 0 {
		return 1
	}
	return 0
}

func nonzero32(x uint32) uint32 {
	if x != 0 {
		return 1
	}
	return 0
}

func nonzero64(x uint64) uint64 {
	if x != 0 {
		return 1
	}
	return 0
}

func nonzero128(x ints.Uint128) ints.Uint128 {
	if x[0]|x[1] != 0 {
		return ints.Uint128{0, 1}
	}
	return ints.Uint128{0, 0}
}

func nonzero256(x ints.Uint256) ints.Uint256 {
	y := x[0] | x[1] | x[2] | x[3]
	if y != 0 {
		return ints.Uint256{0, 0, 0, 1}
	}
	return ints.Uint256{0, 0, 0, 0}
}

func nonzero512(x ints.Uint512) ints.Uint512 {
	y := x[0] | x[1] | x[2] | x[3] | x[4] | x[5] | x[6] | x[7]
	if y != 0 {
		return ints.Uint512{0, 0, 0, 0, 0, 0, 0, 1}
	}
	return ints.Uint512{0, 0, 0, 0, 0, 0, 0, 0}
}

// squash512 squashes the bits of x to a single bit.
func squash512(x ints.Uint512) uint64 {
	y := x[0] | x[1] | x[2] | x[3] | x[4] | x[5] | x[6] | x[7]
	y |= y >> 32
	y |= y >> 16
	y |= y >> 8
	y |= y >> 4
	y |= y >> 2
	y |= y >> 1
	return y & 1
}

func shrcompress32(x uint32, n uint) uint32 {
	if n >= 32 {
		return nonzero32(x)
	}
	y := x >> n
	y |= nonzero32(x & ((1 << n) - 1))
	return y
}

func shrcompress64(x uint64, n uint) uint64 {
	if n >= 64 {
		return nonzero64(x)
	}
	y := x >> n
	y |= nonzero64(x & ((1 << n) - 1))
	return y
}

func shrcompress128(x ints.Uint128, n uint) ints.Uint128 {
	if n >= 128 {
		return nonzero128(x)
	}
	y := x.Rsh(n)
	// x.Lsh(128-n) is the bits shifted out.
	if !x.Lsh(128 - n).IsZero() {
		y[1] |= 1
	}
	return y
}

func shrcompress256(x ints.Uint256, n uint) ints.Uint256 {
	if n >= 256 {
		return nonzero256(x)
	}

	// shift by words, and then by bits.
	// It is faster than ints.Uint256.Rsh, which shifts in constant time.
	w := int(n / 64)
	b := n % 64
	var y ints.Uint256
	y[w] = x[0] >> b
	for i := w + 1; i < len(y); i++ {
		y[i] = x[i-w]>>b | x[i-w-1]<<(64-b)
	}

	// the bits shifted out
	sticky := x[len(x)-1-w] << (64 - b)
	for i := len(x) - w; i < len(x); i++ {
		sticky |= x[i]
	}
	y[len(y)-1] |= nonzero64(sticky)
	return y
}

func shrcompress512(x ints.Uint512, n uint) ints.Uint512 {
	if n >= 512 {
		return nonzero512(x)
	}

	// shift by words, and then by bits.
	// It is faster than ints.Uint512.Rsh, which shifts in constant time.
	w := int(n / 64)
	b := n % 64
	var y ints.Uint512
	y[w] = x[0] >> b
	for i := w + 1; i < len(y); i++ {
		y[i] = x[i-w]>>b | x[i-w-1]<<(64-b)
	}

	// the bits shifted out
	sticky := x[len(x)-1-w] << (64 - b)
	for i := len(x) - w; i < len(x); i++ {
		sticky |= x[i]
	}
	y[len(y)-1] |= nonzero64(sticky)
	return y
}

// lsh256 returns x << n.
// It is faster than ints.Uint256.Lsh, which shifts in constant time.
func lsh256(x ints.Uint256, n uint) ints.Uint256 {
	if n >= 256 {
		return ints.Uint256{}
	}
	w := int(n / 64)
	b := n % 64
	var y ints.Uint256
	last := len(y) - 1 - w
	for i := range last {
		y[i] = x[i+w]<<b | x[i+w+1]>>(64-b)
	}
	y[last] = x[len(x)-1] << b
	return y
}

// lsh256small returns x << n for n in [0, 64).
func lsh256small(x ints.Uint256, n uint) ints.Uint256 {
	return ints.Uint256{
		x[0]<<n | x[1]>>(64-n),
		x[1]<<n | x[2]>>(64-n),
		x[2]<<n | x[3]>>(64-n),
		x[3] << n,
	}
}

// rsh256 returns x >> n.
// It is faster than ints.Uint256.Rsh, which shifts in constant time.
func rsh256(x ints.Uint256, n uint) ints.Uint256 {
	if n >= 256 {
		return ints.Uint256{}
	}
	w := int(n / 64)
	b := n % 64
	var y ints.Uint256
	y[w] = x[0] >> b
	for i := w + 1; i < len(y); i++ {
		y[i] = x[i-w]>>b | x[i-w-1]<<(64-b)
	}
	return y
}

// lsh512 returns x << n.
// It is faster than ints.Uint512.Lsh, which shifts in constant time.
func lsh512(x ints.Uint512, n uint) ints.Uint512 {
	if n >= 512 {
		return ints.Uint512{}
	}
	w := int(n / 64)
	b := n % 64
	var y ints.Uint512
	last := len(y) - 1 - w
	for i := range last {
		y[i] = x[i+w]<<b | x[i+w+1]>>(64-b)
	}
	y[last] = x[len(x)-1] << b
	return y
}

// rsh512 returns x >> n.
// It is faster than ints.Uint512.Rsh, which shifts in constant time.
func rsh512(x ints.Uint512, n uint) ints.Uint512 {
	if n >= 512 {
		return ints.Uint512{}
	}
	w := int(n / 64)
	b := n % 64
	var y ints.Uint512
	y[w] = x[0] >> b
	for i := w + 1; i < len(y); i++ {
		y[i] = x[i-w]>>b | x[i-w-1]<<(64-b)
	}
	return y
}

func roundToNearestEven16(x uint16, shift uint) uint16 {
	mask := uint16(1)<<(shift-1) - 1
	x = (x + mask) + ((x >> shift) & 1)
	return x >> shift
}

func roundToNearestEven32(x uint32, shift uint) uint32 {
	mask := uint32(1)<<(shift-1) - 1
	x = (x + mask) + ((x >> shift) & 1)
	return x >> shift
}

// roundToNearestEven128 returns x >> shift rounded to nearest even.
// shift must be in the range [1, 128].
func roundToNearestEven128(x ints.Uint128, shift uint) ints.Uint128 {
	q := x.Rsh(shift)
	// r is the bits shifted out, aligned to the most significant bit.
	r := x.Lsh(128 - shift)
	const half = 1 << 63
	if r[0] > half || (r[0] == half && (r[1] != 0 || q[1]&1 != 0)) {
		q = q.Add(ints.Uint128{0, 1})
	}
	return q
}

// roundToNearestEven256 returns x >> shift rounded to nearest even.
// shift must be in the range [1, 256].
func roundToNearestEven256(x ints.Uint256, shift uint) ints.Uint256 {
	q := rsh256(x, shift)
	// r is the bits shifted out, aligned to the most significant bit.
	r := lsh256(x, 256-shift)
	const half = 1 << 63
	if r[0] > half || (r[0] == half && (r[1]|r[2]|r[3] != 0 || q[3]&1 != 0)) {
		q = q.Add(ints.Uint256{0, 0, 0, 1})
	}
	return q
}

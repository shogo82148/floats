package floats

import "github.com/shogo82148/ints"

func zero16(x uint16) uint16 {
	if x == 0 {
		return 1
	}
	return 0
}

func nonzero16(x uint16) uint16 {
	if x != 0 {
		return 1
	}
	return 0
}

func zero32(x uint32) uint32 {
	if x == 0 {
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

// squash64 squashes the bits of x to a single bit.
func squash64(x uint64) uint64 {
	x |= x >> 32
	x |= x >> 16
	x |= x >> 8
	x |= x >> 4
	x |= x >> 2
	x |= x >> 1
	return x & 1
}

// squash256 squashes the bits of x to a single bit.
func squash256(x ints.Uint256) uint64 {
	y := x[0] | x[1] | x[2] | x[3]
	y |= y >> 32
	y |= y >> 16
	y |= y >> 8
	y |= y >> 4
	y |= y >> 2
	y |= y >> 1
	return y & 1
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

func shrcompress16(x uint16, n uint) uint16 {
	if n >= 16 {
		return nonzero16(x)
	}
	y := x >> n
	y |= nonzero16(x & (1<<n - 1))
	return y
}

func shrcompress32(x uint32, n uint) uint32 {
	if n >= 32 {
		return nonzero32(x)
	}
	y := x >> n
	y |= nonzero32(x & ((1<<n) - 1))
	return y
}

func roundToNearestEven16(x uint16, shift uint) uint16 {
	mask := uint16(1)<<(shift-1) - 1
	x = (x + mask) + ((x >> shift) & 1)
	return x >> shift
}

func roundToNearestEven128(x ints.Uint128, shift uint) ints.Uint128 {
	one := ints.Uint128{0, 1}
	mask := one.Lsh(uint(shift - 1)).Sub(one)
	return x.Add(mask).Add(x.Rsh(uint(shift)).And(one))
}

func roundToNearestEven256(x ints.Uint256, shift uint) ints.Uint256 {
	one := ints.Uint256{0, 0, 0, 1}
	mask := one.Lsh(uint(shift - 1)).Sub(one)
	return x.Add(mask).Add(x.Rsh(uint(shift)).And(one))
}

func roundToNearestEven512(x ints.Uint512, shift uint) ints.Uint512 {
	one := ints.Uint512{0, 0, 0, 0, 0, 0, 0, 1}
	mask := one.Lsh(uint(shift - 1)).Sub(one)
	return x.Add(mask).Add(x.Rsh(uint(shift)).And(one))
}

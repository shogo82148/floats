package floats

// Float16 returns a itself.
func (a Float16) Float16() Float16 {
	return a
}

// Float32 converts a to a Float32.
func (a Float16) Float32() Float32 {
	return 0 // TODO: implement
}

// Float64 converts a to a Float64.
func (a Float16) Float64() Float64 {
	return 0 // TODO: implement
}

// Float128 converts a to a Float128.
func (a Float16) Float128() Float128 {
	return Float128{0, 0} // TODO: implement
}

// Float256 converts a to a Float256.
func (a Float16) Float256() Float256 {
	return Float256{0, 0, 0, 0} // TODO: implement
}

// Float16 converts a to a Float16.
func (a Float32) Float16() Float16 {
	return Float16(0) // TODO: implement
}

// Float32 returns a itself.
func (a Float32) Float32() Float32 {
	return a
}

// Float64 converts a to a Float64.
func (a Float32) Float64() Float64 {
	return Float64(a)
}

// Float128 converts a to a Float128.
func (a Float32) Float128() Float128 {
	return Float128{0, 0} // TODO: implement
}

// Float256 converts a to a Float256.
func (a Float32) Float256() Float256 {
	return Float256{0, 0, 0, 0} // TODO: implement
}

// Float16 converts a to a Float16.
func (a Float64) Float16() Float16 {
	return Float16(0) // TODO: implement
}

// Float32 converts a to a Float32.
func (a Float64) Float32() Float32 {
	return Float32(a)
}

// Float64 returns a itself.
func (a Float64) Float64() Float64 {
	return a
}

// Float128 converts a to a Float128.
func (a Float64) Float128() Float128 {
	return Float128{0, 0} // TODO: implement
}

// Float256 converts a to a Float256.
func (a Float64) Float256() Float256 {
	return Float256{0, 0, 0, 0} // TODO: implement
}

// Float16 converts a to a Float16.
func (a Float128) Float16() Float16 {
	return Float16(0) // TODO: implement
}

// Float32 converts a to a Float32.
func (a Float128) Float32() Float32 {
	return Float32(0) // TODO: implement
}

// Float64 converts a to a Float64.
func (a Float128) Float64() Float64 {
	return Float64(0) // TODO: implement
}

// Float128 returns a itself.
func (a Float128) Float128() Float128 {
	return a
}

// Float256 converts a to a Float256.
func (a Float128) Float256() Float256 {
	return Float256{0, 0, 0, 0} // TODO: implement
}

// Float16 converts a to a Float16.
func (a Float256) Float16() Float16 {
	return Float16(0) // TODO: implement
}

// Float32 converts a to a Float32.
func (a Float256) Float32() Float32 {
	return Float32(0) // TODO: implement
}

// Float64 converts a to a Float64.
func (a Float256) Float64() Float64 {
	return Float64(0) // TODO: implement
}

// Float128 converts a to a Float128.
func (a Float256) Float128() Float128 {
	return Float128{0, 0} // TODO: implement
}

// Float256 returns a itself.
func (a Float256) Float256() Float256 {
	return a
}

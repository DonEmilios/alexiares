package favicon

// murmur3_32 implements the 32-bit x86 variant of MurmurHash3 with the
// given seed. It has no external dependency and is verified against
// the algorithm's published test vectors in murmur3_test.go.
func murmur3_32(data []byte, seed uint32) int32 { //nolint:revive // matches the well-known algorithm name
	const (
		c1 = 0xcc9e2d51
		c2 = 0x1b873593
	)

	h := seed
	length := len(data)
	nBlocks := length / 4

	for i := 0; i < nBlocks; i++ {
		k := uint32(data[i*4]) | uint32(data[i*4+1])<<8 | uint32(data[i*4+2])<<16 | uint32(data[i*4+3])<<24
		k *= c1
		k = rotl32(k, 15)
		k *= c2

		h ^= k
		h = rotl32(h, 13)
		h = h*5 + 0xe6546b64
	}

	var k1 uint32
	tail := data[nBlocks*4:]
	switch len(tail) {
	case 3:
		k1 ^= uint32(tail[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint32(tail[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint32(tail[0])
		k1 *= c1
		k1 = rotl32(k1, 15)
		k1 *= c2
		h ^= k1
	}

	h ^= uint32(length)
	h = fmix32(h)

	return int32(h)
}

func rotl32(x uint32, r uint8) uint32 {
	return (x << r) | (x >> (32 - r))
}

func fmix32(h uint32) uint32 {
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16
	return h
}

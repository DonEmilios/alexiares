package fingerprint

import "hash/fnv"

// shingleSize is the n-gram width used to build SimHash features from
// a tag sequence. 3 balances sensitivity (too small overweights common
// short runs like "div div div") against robustness to small template
// edits (too large makes every page look unique).
const shingleSize = 3

// simHash computes a 64-bit SimHash over 3-tag shingles of tags. It is
// a locality-sensitive fingerprint: near-duplicate tag sequences
// produce hashes with a small Hamming distance, letting the
// correlation engine detect "same template, different content"
// clones that an exact hash would miss entirely.
//
// A tag sequence shorter than one shingle hashes to 0, since no
// shingles exist to vote on any bit.
func simHash(tags []string) uint64 {
	if len(tags) < shingleSize {
		return 0
	}

	var weight [64]int
	for i := 0; i+shingleSize <= len(tags); i++ {
		feature := shingleKey(tags[i : i+shingleSize])
		h := fnv.New64a()
		_, _ = h.Write([]byte(feature))
		sum := h.Sum64()

		for bit := 0; bit < 64; bit++ {
			if sum&(1<<uint(bit)) != 0 {
				weight[bit]++
			} else {
				weight[bit]--
			}
		}
	}

	var result uint64
	for bit := 0; bit < 64; bit++ {
		if weight[bit] > 0 {
			result |= 1 << uint(bit)
		}
	}
	return result
}

func shingleKey(tags []string) string {
	// A single separator byte not valid in an HTML tag name avoids
	// collisions like ["ab","c"] vs ["a","bc"].
	const sep = "\x00"
	total := 0
	for _, t := range tags {
		total += len(t) + 1
	}
	key := make([]byte, 0, total)
	for _, t := range tags {
		key = append(key, t...)
		key = append(key, sep...)
	}
	return string(key)
}

// HammingDistance64 counts the differing bits between two SimHash
// values, exported for the correlation engine's fuzzy-matching use.
func HammingDistance64(a, b uint64) int {
	x := a ^ b
	count := 0
	for x != 0 {
		count++
		x &= x - 1
	}
	return count
}

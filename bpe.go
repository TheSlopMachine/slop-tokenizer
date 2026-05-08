package tokenizer

const NO_RANK = int(^uint(0) >> 1) // MaxInt

// bytePairMerge performs the BPE merge algorithm on a byte slice
// This matches the TypeScript implementation exactly
func bytePairMerge(
	piece []byte,
	stringRanks map[string]int,
	binaryIndex [256][]BinaryToken,
) []int {
	if len(piece) == 0 {
		return []int{}
	}

	if len(piece) == 1 {
		// Single byte - look it up directly
		rank := getRankForSlice(piece, stringRanks, binaryIndex)
		if rank != NO_RANK {
			return []int{rank}
		}
		return []int{}
	}

	starts := []int{}
	ranks := []int{}

	// Helper function matching TypeScript's getRank
	getRank := func(startIndex int, pairStart, pairEnd int) int {
		// If pairStart/pairEnd are -1, use starts array
		if pairStart == -1 {
			if startIndex < 0 || startIndex >= len(starts) {
				return NO_RANK
			}
			pairStart = starts[startIndex]
		}
		if pairEnd == -1 {
			if startIndex+2 >= len(starts) {
				return NO_RANK
			}
			pairEnd = starts[startIndex+2]
		}
		
		if pairStart < 0 || pairEnd > len(piece) || pairStart >= pairEnd {
			return NO_RANK
		}
		return getRankForSlice(piece[pairStart:pairEnd], stringRanks, binaryIndex)
	}

	// Initialize starts and ranks
	for i := 0; i <= len(piece); i++ {
		starts = append(starts, i)
		if i < len(piece)-1 {
			ranks = append(ranks, getRank(i, i, i+2))
		} else {
			ranks = append(ranks, NO_RANK)
		}
	}

	// Main merge loop
	for len(starts) > 1 {
		minRank := NO_RANK
		minIdx := -1

		// Find minimum rank
		for i := 0; i < len(ranks)-1; i++ {
			if ranks[i] < minRank {
				minRank = ranks[i]
				minIdx = i
			}
		}

		if minRank == NO_RANK || minIdx == -1 {
			break
		}

		// Remove elements (merge the pair)
		starts = append(starts[:minIdx+1], starts[minIdx+2:]...)
		ranks = append(ranks[:minIdx], ranks[minIdx+1:]...)

		// Update ranks around the merge point
		if minIdx < len(ranks) {
			ranks[minIdx] = getRank(minIdx, -1, -1)
		}
		if minIdx > 0 {
			ranks[minIdx-1] = getRank(minIdx-1, -1, -1)
		}
	}

	// Build output
	output := make([]int, 0, len(starts)-1)
	for i := 0; i < len(starts)-1; i++ {
		pairStart := starts[i]
		pairEnd := starts[i+1]
		rank := getRankForSlice(piece[pairStart:pairEnd], stringRanks, binaryIndex)
		if rank != NO_RANK {
			output = append(output, rank)
		}
	}

	return output
}

// getRankForSlice looks up the rank for a byte slice
// Fast path: try UTF-8 string lookup
// Slow path: binary search in indexed bucket
func getRankForSlice(slice []byte, stringRanks map[string]int, binaryIndex [256][]BinaryToken) int {
	// Fast path: try UTF-8 string lookup
	if isValidUTF8(slice) {
		if rank, ok := stringRanks[string(slice)]; ok {
			return rank
		}
	}

	// Slow path: binary search in indexed bucket
	if len(slice) > 0 {
		bucket := binaryIndex[slice[0]]
		if bucket != nil {
			idx := binarySearch(bucket, slice)
			if idx != -1 {
				return bucket[idx].Rank
			}
		}
	}

	return NO_RANK
}

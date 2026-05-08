package tokenizer

// isValidUTF8 performs fast UTF-8 validation without try/catch overhead
// Based on the TypeScript implementation
func isValidUTF8(bytes []byte) bool {
	i := 0
	for i < len(bytes) {
		if i >= len(bytes) {
			return false
		}
		byte1 := bytes[i]
		var numBytes int
		var codePoint int

		// Determine the number of bytes in the current UTF-8 character
		if byte1 <= 0x7f {
			numBytes = 1
			codePoint = int(byte1)
		} else if (byte1 & 0xe0) == 0xc0 {
			numBytes = 2
			codePoint = int(byte1 & 0x1f)
			if byte1 <= 0xc1 {
				return false // Overlong encoding
			}
		} else if (byte1 & 0xf0) == 0xe0 {
			numBytes = 3
			codePoint = int(byte1 & 0x0f)
		} else if (byte1 & 0xf8) == 0xf0 {
			numBytes = 4
			codePoint = int(byte1 & 0x07)
			if byte1 > 0xf4 {
				return false // Code points above U+10FFFF
			}
		} else {
			return false
		}

		// Ensure there are enough continuation bytes
		if i+numBytes > len(bytes) {
			return false
		}

		// Process continuation bytes
		for j := 1; j < numBytes; j++ {
			b := bytes[i+j]
			if (b & 0xc0) != 0x80 {
				return false
			}
			codePoint = (codePoint << 6) | int(b&0x3f)
		}

		// Check for overlong encodings
		if numBytes == 2 && codePoint < 0x80 {
			return false
		}
		if numBytes == 3 && codePoint < 0x800 {
			return false
		}
		if numBytes == 4 && codePoint < 0x10000 {
			return false
		}

		// Check for surrogate halves (U+D800 to U+DFFF)
		if codePoint >= 0xd800 && codePoint <= 0xdfff {
			return false
		}

		// Check for code points above U+10FFFF
		if codePoint > 0x10ffff {
			return false
		}

		i += numBytes
	}
	return true
}

// tryBytesToString attempts to convert bytes to a UTF-8 string
// Returns the string and true if valid UTF-8, empty string and false otherwise
func tryBytesToString(bytes []byte) (string, bool) {
	if !isValidUTF8(bytes) {
		return "", false
	}
	return string(bytes), true
}

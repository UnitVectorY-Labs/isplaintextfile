package isplaintextfile

import (
	"errors"
	"io"
	"os"
	"unicode/utf8"
)

// isBufferPlaintext examines a slice of bytes and returns whether it appears to be valid plaintext.
func isBufferPlaintext(buffer []byte) bool {
	if !utf8.Valid(buffer) {
		return false
	}

	pos := 0
	for pos < len(buffer) {
		r, size := utf8.DecodeRune(buffer[pos:])

		// Check for control characters (except whitespace)
		if (r < 32 && r != '\n' && r != '\r' && r != '\t') || r == 0x7F {
			return false
		}

		pos += size
	}
	return true
}

// trimIncompleteUTF8 removes any trailing incomplete UTF-8 byte sequence from the buffer.
// This is used in preview mode where the read limit may split a multi-byte character.
func trimIncompleteUTF8(data []byte) []byte {
	if len(data) == 0 || utf8.Valid(data) {
		return data
	}
	// Try trimming 1 to 3 trailing bytes to find a valid UTF-8 boundary.
	// UTF-8 characters are at most 4 bytes, so an incomplete trailing
	// sequence is at most 3 bytes.
	for i := 1; i <= 3 && i <= len(data); i++ {
		if utf8.Valid(data[:len(data)-i]) {
			return data[:len(data)-i]
		}
	}
	return data
}

// isPlaintextFromReader reads from the given reader and checks if the content is valid plaintext.
// If preview is true, incomplete trailing UTF-8 sequences are trimmed before validation,
// since the read limit may have split a multi-byte character.
func isPlaintextFromReader(reader io.Reader, preview bool) (bool, error) {
	buffer := make([]byte, 0, 32*1024)
	tempBuf := make([]byte, 1024)

	for {
		n, err := reader.Read(tempBuf)
		if n > 0 {
			buffer = append(buffer, tempBuf[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, err
		}
	}

	if len(buffer) == 0 {
		return true, nil
	}

	if preview {
		buffer = trimIncompleteUTF8(buffer)
		if len(buffer) == 0 {
			return true, nil
		}
	}

	return isBufferPlaintext(buffer), nil
}

// Bytes checks if the provided byte slice is valid plaintext.
func Bytes(data []byte) (bool, error) {
	// In-memory data: no IO error is expected.
	return isBufferPlaintext(data), nil
}

// File opens the file at the given path and checks if its entire content is plaintext.
func File(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	return isPlaintextFromReader(file, false)
}

// FilePreview opens the file at the given path, reads up to maxKB kilobytes,
// and checks if that portion of the file is plaintext.
func FilePreview(path string, maxKB int) (bool, error) {
	if maxKB <= 0 {
		return false, errors.New("invalid length: maxKB must be greater than 0")
	}

	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Limit the reader to maxKB*1024 bytes.
	limitedReader := io.LimitReader(file, int64(maxKB*1024))
	return isPlaintextFromReader(limitedReader, true)
}

// Reader checks if the content provided by the io.Reader is plaintext.
func Reader(reader io.Reader) (bool, error) {
	return isPlaintextFromReader(reader, false)
}

// ReaderPreview checks if the content provided by the io.Reader is plaintext,
// reading only up to maxKB kilobytes from the reader.
func ReaderPreview(reader io.Reader, maxKB int) (bool, error) {
	if maxKB <= 0 {
		return false, errors.New("invalid length: maxKB must be greater than 0")
	}

	limitedReader := io.LimitReader(reader, int64(maxKB*1024))
	return isPlaintextFromReader(limitedReader, true)
}

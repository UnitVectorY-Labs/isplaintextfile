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
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			return false
		}

		pos += size
	}
	return true
}

// isPlaintextFromReader reads from the given reader and checks if the content is valid plaintext.
func isPlaintextFromReader(reader io.Reader) (bool, error) {
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
	return isBufferPlaintext(buffer), nil
}

// IsPlaintextBytes checks if the provided byte slice is valid plaintext.
func IsPlaintextBytes(data []byte) (bool, error) {
	// In-memory data: no IO error is expected.
	return isBufferPlaintext(data), nil
}

// IsPlaintextPath opens the file at the given path and checks if its entire content is plaintext.
func IsPlaintextPath(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	return isPlaintextFromReader(file)
}

// IsPlaintextPathPreview opens the file at the given path, reads up to maxKB kilobytes,
// and checks if that portion of the file is plaintext.
func IsPlaintextPathPreview(path string, maxKB int) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	if maxKB == 0 {
		return true, errors.New("invalid length: maxKB must be greater than 0")
	}

	// Limit the reader to maxKB*1024 bytes.
	limitedReader := io.LimitReader(file, int64(maxKB*1024))
	return isPlaintextFromReader(limitedReader)
}

// IsPlaintextReader checks if the content provided by the io.Reader is plaintext.
func IsPlaintextReader(reader io.Reader) (bool, error) {
	return isPlaintextFromReader(reader)
}

// IsPlaintextReaderPreview checks if the content provided by the io.Reader is plaintext,
// reading only up to maxKB kilobytes from the reader.
func IsPlaintextReaderPreview(reader io.Reader, maxKB int) (bool, error) {
	maxBytes := maxKB * 1024

	if maxKB == 0 {
		return true, errors.New("invalid length: maxKB must be greater than 0")
	}

	limitedReader := io.LimitReader(reader, int64(maxBytes))
	return isPlaintextFromReader(limitedReader)
}

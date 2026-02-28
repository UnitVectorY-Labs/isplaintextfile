package isplaintextfile

import (
	"bytes"
	"errors"
	"os"
	"testing"
)

func TestPlaintextMethods(t *testing.T) {
	// Each test case now specifies:
	// - content: the full file content as a byte slice.
	// - unlimitedExpected: the expected result when analyzing the entire content.
	// - limitedExpected: the expected result when analyzing only the first limitKB kilobytes.
	// - limitKB: the kilobyte limit for the limited‑length variants; if zero, we set it high enough to cover the content.
	tests := []struct {
		name              string
		content           []byte
		unlimitedExpected bool
		limitedExpected   bool
		limitKB           int // the limit in KB
	}{
		{
			name:              "plain ASCII text",
			content:           []byte("Hello, World!\n"),
			unlimitedExpected: true,
			limitedExpected:   true,
			limitKB:           0, // no explicit limit needed (content is small)
		},
		{
			name:              "text with emoji",
			content:           []byte("Hello 👋 World! 🌍\n"),
			unlimitedExpected: true,
			limitedExpected:   true,
			limitKB:           0,
		},
		{
			name:              "text with Chinese characters",
			content:           []byte("你好，世界！\n"),
			unlimitedExpected: true,
			limitedExpected:   true,
			limitKB:           0,
		},
		{
			name:              "binary content",
			content:           []byte{0x00, 0x01, 0x02, 0x03},
			unlimitedExpected: false,
			limitedExpected:   false,
			limitKB:           0,
		},
		{
			name:              "text with control characters",
			content:           []byte{'H', 'e', 'l', 'l', 'o', 0x07},
			unlimitedExpected: false,
			limitedExpected:   false,
			limitKB:           0,
		},
		{
			name:              "large text file",
			content:           bytes.Repeat([]byte("Large plain text content\n"), 1000), // ~23KB text
			unlimitedExpected: true,
			limitedExpected:   true,
			limitKB:           0,
		},
		{
			name: "partial binary after limit",
			content: func() []byte {
				// First 1024 bytes are valid text, then a few invalid (binary) bytes are appended.
				validPart := bytes.Repeat([]byte("A"), 1024) // 1 KB valid text
				binaryPart := []byte{0x00, 0x01, 0x02, 0x03}
				return append(validPart, binaryPart...)
			}(),
			unlimitedExpected: false, // full content includes binary → not plaintext
			limitedExpected:   true,  // limiting to 1 KB sees only valid text
			limitKB:           1,     // limit to 1 KB
		},
	}

	for _, tt := range tests {
		t.Run("Bytes_"+tt.name, func(t *testing.T) {
			// 1. Test the in-memory bytes variant (always processes the entire byte slice).
			res, err := Bytes(tt.content)
			if err != nil {
				t.Errorf("Bytes() error: %v", err)
			}
			if res != tt.unlimitedExpected {
				t.Errorf("Bytes() = %v, want %v", res, tt.unlimitedExpected)
			}
		})
	}

	for _, tt := range tests {
		t.Run("Reader_"+tt.name, func(t *testing.T) {
			// 2. Test the io.Reader variant (unlimited).
			reader := bytes.NewReader(tt.content)
			res, err := Reader(reader)
			if err != nil {
				t.Errorf("Reader() error: %v", err)
			}
			if res != tt.unlimitedExpected {
				t.Errorf("Reader() = %v, want %v", res, tt.unlimitedExpected)
			}
		})
	}

	for _, tt := range tests {
		t.Run("ReaderPreview_"+tt.name, func(t *testing.T) {
			// For tests where no explicit limit is set, use a limit high enough to cover the entire content.
			limitKB := tt.limitKB
			if limitKB == 0 {
				limitKB = (len(tt.content) + 1023) / 1024
			}

			// 3. Test the io.Reader-with-limit variant (limit specified in KB).
			reader := bytes.NewReader(tt.content)
			res, err := ReaderPreview(reader, limitKB)
			if err != nil {
				t.Errorf("ReaderPreview() error: %v", err)
			}
			if res != tt.limitedExpected {
				t.Errorf("ReaderPreview() = %v, want %v", res, tt.limitedExpected)
			}

			// Test the io.Reader with a limit that is larger than the content.
			reader = bytes.NewReader(tt.content)
			res, err = ReaderPreview(reader, len(tt.content)+10)
			if err != nil {
				t.Errorf("ReaderPreview() error: %v", err)
			}
			if res != tt.unlimitedExpected {
				t.Errorf("ReaderPreview() = %v, want %v", res, tt.unlimitedExpected)
			}

		})
	}

	for _, tt := range tests {
		t.Run("File_"+tt.name, func(t *testing.T) {
			// For file-based tests, create a temporary file with the test content.
			tmpfile, err := os.CreateTemp("", "plaintext_test")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write(tt.content); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatalf("Failed to close temp file: %v", err)
			}

			// 4. Test the file path variant (unlimited).
			res, err := File(tmpfile.Name())
			if err != nil {
				t.Errorf("File() error: %v", err)
			}
			if res != tt.unlimitedExpected {
				t.Errorf("File() = %v, want %v", res, tt.unlimitedExpected)
			}

		})
	}

	for _, tt := range tests {
		t.Run("FilePreview_"+tt.name, func(t *testing.T) {
			// For tests where no explicit limit is set, use a limit high enough to cover the entire content.
			limitKB := tt.limitKB
			if limitKB == 0 {
				limitKB = (len(tt.content) + 1023) / 1024
			}

			// For file-based tests, create a temporary file with the test content.
			tmpfile, err := os.CreateTemp("", "plaintext_test")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write(tt.content); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatalf("Failed to close temp file: %v", err)
			}

			// 5. Test the file path-with-limit variant (limit specified in KB).
			res, err := FilePreview(tmpfile.Name(), limitKB)
			if err != nil {
				t.Errorf("FilePreview() error: %v", err)
			}
			if res != tt.limitedExpected {
				t.Errorf("FilePreview() = %v, want %v", res, tt.limitedExpected)
			}

			// Test the path-with-limit variant with a limit that is larger than the content.
			res, err = FilePreview(tmpfile.Name(), len(tt.content)+10)
			if err != nil {
				t.Errorf("FilePreview() error: %v", err)
			}
			if res != tt.unlimitedExpected {
				t.Errorf("FilePreview() = %v, want %v", res, tt.unlimitedExpected)
			}
		})
	}
}

// TestDELCharacter verifies that the DEL character (0x7F) is detected as a control character.
func TestDELCharacter(t *testing.T) {
	data := []byte("Hello\x7FWorld")
	res, err := Bytes(data)
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}
	if res != false {
		t.Error("Bytes() should return false for content with DEL character (0x7F)")
	}
}

// TestAllControlCharacters tests that each ASCII control character (except allowed whitespace) is rejected.
func TestAllControlCharacters(t *testing.T) {
	for c := byte(0); c < 32; c++ {
		expected := c == '\n' || c == '\r' || c == '\t'
		data := []byte{'A', c, 'B'}
		res, err := Bytes(data)
		if err != nil {
			t.Fatalf("Bytes() error for control char 0x%02X: %v", c, err)
		}
		if res != expected {
			t.Errorf("Bytes() for control char 0x%02X = %v, want %v", c, res, expected)
		}
	}
	// Also test DEL (0x7F)
	data := []byte{'A', 0x7F, 'B'}
	res, err := Bytes(data)
	if err != nil {
		t.Fatalf("Bytes() error for DEL: %v", err)
	}
	if res != false {
		t.Error("Bytes() should return false for content with DEL character")
	}
}

// TestEmptyContent verifies that empty content is considered plaintext.
func TestEmptyContent(t *testing.T) {
	// Empty byte slice
	res, err := Bytes([]byte{})
	if err != nil {
		t.Fatalf("Bytes(empty) error: %v", err)
	}
	if res != true {
		t.Error("Bytes(empty) should return true")
	}

	// Nil byte slice
	res, err = Bytes(nil)
	if err != nil {
		t.Fatalf("Bytes(nil) error: %v", err)
	}
	if res != true {
		t.Error("Bytes(nil) should return true")
	}

	// Empty reader
	res, err = Reader(bytes.NewReader([]byte{}))
	if err != nil {
		t.Fatalf("Reader(empty) error: %v", err)
	}
	if res != true {
		t.Error("Reader(empty) should return true")
	}

	// Empty file
	tmpfile, err := os.CreateTemp("", "plaintext_test_empty")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	res, err = File(tmpfile.Name())
	if err != nil {
		t.Fatalf("File(empty) error: %v", err)
	}
	if res != true {
		t.Error("File(empty) should return true")
	}
}

// TestPreviewMaxKBZero verifies that maxKB=0 returns false with an error.
func TestPreviewMaxKBZero(t *testing.T) {
	data := []byte("Hello, World!")
	reader := bytes.NewReader(data)

	res, err := ReaderPreview(reader, 0)
	if err == nil {
		t.Fatal("ReaderPreview(maxKB=0) should return an error")
	}
	if res != false {
		t.Error("ReaderPreview(maxKB=0) should return false when error is returned")
	}

	// FilePreview with maxKB=0
	tmpfile, err := os.CreateTemp("", "plaintext_test_zero")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Write(data)
	tmpfile.Close()

	res, err = FilePreview(tmpfile.Name(), 0)
	if err == nil {
		t.Fatal("FilePreview(maxKB=0) should return an error")
	}
	if res != false {
		t.Error("FilePreview(maxKB=0) should return false when error is returned")
	}
}

// TestPreviewNegativeMaxKB verifies that negative maxKB returns false with an error.
// Previously negative maxKB caused binary content to appear as plaintext.
func TestPreviewNegativeMaxKB(t *testing.T) {
	data := []byte{0x00, 0x01, 0x02} // binary content

	res, err := ReaderPreview(bytes.NewReader(data), -1)
	if err == nil {
		t.Fatal("ReaderPreview(maxKB=-1) should return an error")
	}
	if res != false {
		t.Error("ReaderPreview(maxKB=-1) should return false, not true")
	}

	tmpfile, err := os.CreateTemp("", "plaintext_test_neg")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Write(data)
	tmpfile.Close()

	res, err = FilePreview(tmpfile.Name(), -5)
	if err == nil {
		t.Fatal("FilePreview(maxKB=-5) should return an error")
	}
	if res != false {
		t.Error("FilePreview(maxKB=-5) should return false, not true")
	}
}

// TestPreviewSplitsMultibyteUTF8 verifies that preview doesn't give a false negative
// when the read limit splits a multi-byte UTF-8 character at the boundary.
func TestPreviewSplitsMultibyteUTF8(t *testing.T) {
	// Build content where a 4-byte emoji sits exactly at the 1KB boundary.
	// Fill with 1021 ASCII bytes, then add a 4-byte emoji (🌍 = F0 9F 8C 8D).
	// Total: 1025 bytes. Preview of 1KB reads 1024 bytes, cutting the emoji.
	filler := bytes.Repeat([]byte("A"), 1021)
	emoji := []byte("🌍") // 4 bytes in UTF-8
	content := append(filler, emoji...)

	// Full content should be plaintext
	res, err := Bytes(content)
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}
	if res != true {
		t.Error("Bytes() should return true for valid text with emoji")
	}

	// Preview of 1KB should also be plaintext (not false negative due to split emoji)
	reader := bytes.NewReader(content)
	res, err = ReaderPreview(reader, 1)
	if err != nil {
		t.Fatalf("ReaderPreview() error: %v", err)
	}
	if res != true {
		t.Error("ReaderPreview() should return true even when preview splits a multi-byte UTF-8 character")
	}

	// Also test with FilePreview
	tmpfile, err := os.CreateTemp("", "plaintext_test_utf8split")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Write(content)
	tmpfile.Close()

	res, err = FilePreview(tmpfile.Name(), 1)
	if err != nil {
		t.Fatalf("FilePreview() error: %v", err)
	}
	if res != true {
		t.Error("FilePreview() should return true even when preview splits a multi-byte UTF-8 character")
	}
}

// TestPreviewSplitsTwoByteUTF8 tests splitting a 2-byte UTF-8 character at boundary.
func TestPreviewSplitsTwoByteUTF8(t *testing.T) {
	// é is U+00E9, encoded as C3 A9 (2 bytes)
	filler := bytes.Repeat([]byte("B"), 1023)
	twoByteChar := []byte("é") // 2 bytes
	content := append(filler, twoByteChar...)

	// Full content should be plaintext
	res, err := Bytes(content)
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}
	if res != true {
		t.Error("Bytes() should return true for text with 2-byte UTF-8 char")
	}

	// Preview of 1KB = 1024 bytes; content is 1025 bytes, boundary splits the é
	reader := bytes.NewReader(content)
	res, err = ReaderPreview(reader, 1)
	if err != nil {
		t.Fatalf("ReaderPreview() error: %v", err)
	}
	if res != true {
		t.Error("ReaderPreview() should return true even when preview splits a 2-byte UTF-8 character")
	}
}

// TestPreviewSplitsThreeByteUTF8 tests splitting a 3-byte UTF-8 character at boundary.
func TestPreviewSplitsThreeByteUTF8(t *testing.T) {
	// 你 is U+4F60, encoded as E4 BD A0 (3 bytes)
	filler := bytes.Repeat([]byte("C"), 1022)
	threeByteChar := []byte("你") // 3 bytes
	content := append(filler, threeByteChar...)

	res, err := Bytes(content)
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}
	if res != true {
		t.Error("Bytes() should return true for text with 3-byte UTF-8 char")
	}

	reader := bytes.NewReader(content)
	res, err = ReaderPreview(reader, 1)
	if err != nil {
		t.Fatalf("ReaderPreview() error: %v", err)
	}
	if res != true {
		t.Error("ReaderPreview() should return true even when preview splits a 3-byte UTF-8 character")
	}
}

// TestFileNotFound verifies error handling for nonexistent files.
func TestFileNotFound(t *testing.T) {
	res, err := File("/nonexistent/path/to/file.txt")
	if err == nil {
		t.Fatal("File() should return an error for nonexistent file")
	}
	if res != false {
		t.Error("File() should return false when error occurs")
	}

	res, err = FilePreview("/nonexistent/path/to/file.txt", 1)
	if err == nil {
		t.Fatal("FilePreview() should return an error for nonexistent file")
	}
	if res != false {
		t.Error("FilePreview() should return false when error occurs")
	}
}

// errReader is a helper that returns an error on Read.
type errReader struct{}

func (e *errReader) Read(p []byte) (int, error) {
	return 0, errors.New("simulated read error")
}

// TestReaderError verifies error handling when the io.Reader returns an error.
func TestReaderError(t *testing.T) {
	res, err := Reader(&errReader{})
	if err == nil {
		t.Fatal("Reader() should return an error for failing reader")
	}
	if res != false {
		t.Error("Reader() should return false when error occurs")
	}

	res, err = ReaderPreview(&errReader{}, 1)
	if err == nil {
		t.Fatal("ReaderPreview() should return an error for failing reader")
	}
	if res != false {
		t.Error("ReaderPreview() should return false when error occurs")
	}
}

// partialErrReader returns some valid data then errors.
type partialErrReader struct {
	data []byte
	read bool
}

func (p *partialErrReader) Read(buf []byte) (int, error) {
	if !p.read {
		p.read = true
		n := copy(buf, p.data)
		return n, nil
	}
	return 0, errors.New("simulated partial read error")
}

// TestReaderPartialError verifies error handling when reader returns data then errors.
func TestReaderPartialError(t *testing.T) {
	r := &partialErrReader{data: []byte("Hello")}
	res, err := Reader(r)
	if err == nil {
		t.Fatal("Reader() should propagate the read error")
	}
	if res != false {
		t.Error("Reader() should return false when error occurs")
	}
}

// TestInvalidUTF8 verifies that invalid UTF-8 byte sequences are detected as non-plaintext.
func TestInvalidUTF8(t *testing.T) {
	// Overlong encoding of '/' (invalid UTF-8)
	data := []byte{0xC0, 0xAF}
	res, err := Bytes(data)
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}
	if res != false {
		t.Error("Bytes() should return false for invalid UTF-8 (overlong encoding)")
	}

	// Invalid continuation byte
	data = []byte{0xE0, 0x80}
	res, err = Bytes(data)
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}
	if res != false {
		t.Error("Bytes() should return false for incomplete UTF-8 sequence")
	}

	// 0xFE and 0xFF are never valid in UTF-8
	data = []byte{0xFE, 0xFF}
	res, err = Bytes(data)
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}
	if res != false {
		t.Error("Bytes() should return false for 0xFE 0xFF bytes")
	}
}

// TestWhitespaceOnly verifies that whitespace-only content is plaintext.
func TestWhitespaceOnly(t *testing.T) {
	data := []byte("\t\n\r \t\n")
	res, err := Bytes(data)
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}
	if res != true {
		t.Error("Bytes() should return true for whitespace-only content")
	}
}

// TestFormFeed verifies that form feed (0x0C) is detected as a control character.
func TestFormFeed(t *testing.T) {
	data := []byte("Page 1\x0CPage 2")
	res, err := Bytes(data)
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}
	if res != false {
		t.Error("Bytes() should return false for content with form feed (0x0C)")
	}
}

// TestNullByte verifies that a single null byte is detected.
func TestNullByte(t *testing.T) {
	data := []byte("Hello\x00World")
	res, err := Bytes(data)
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}
	if res != false {
		t.Error("Bytes() should return false for content with null byte")
	}
}

// TestReplacementCharacter verifies that the Unicode replacement character (U+FFFD)
// is accepted as valid plaintext since it is a valid Unicode codepoint.
func TestReplacementCharacter(t *testing.T) {
	data := []byte("Hello \xef\xbf\xbd World") // U+FFFD in UTF-8
	res, err := Bytes(data)
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}
	if res != true {
		t.Error("Bytes() should return true for content with replacement character U+FFFD")
	}
}

// TestByteOrderMark verifies that BOM (U+FEFF) is accepted as valid plaintext.
func TestByteOrderMark(t *testing.T) {
	data := []byte("\xef\xbb\xbfHello, World!") // UTF-8 BOM + text
	res, err := Bytes(data)
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}
	if res != true {
		t.Error("Bytes() should return true for content with UTF-8 BOM")
	}
}

// TestMixedValidContent verifies that text mixing ASCII, multi-byte characters, and
// allowed whitespace is considered plaintext.
func TestMixedValidContent(t *testing.T) {
	data := []byte("Hello\t世界\n🌍\rEnd")
	res, err := Bytes(data)
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}
	if res != true {
		t.Error("Bytes() should return true for mixed valid content")
	}
}

// TestPreviewWithExactBoundary verifies preview when content is exactly the limit size.
func TestPreviewWithExactBoundary(t *testing.T) {
	// Exactly 1024 bytes of valid content
	data := bytes.Repeat([]byte("X"), 1024)
	reader := bytes.NewReader(data)
	res, err := ReaderPreview(reader, 1)
	if err != nil {
		t.Fatalf("ReaderPreview() error: %v", err)
	}
	if res != true {
		t.Error("ReaderPreview() should return true for exactly 1KB of valid content")
	}
}

// TestReaderPreviewWithLargeLimit verifies that a large preview limit works correctly.
func TestReaderPreviewWithLargeLimit(t *testing.T) {
	data := []byte("Small content")
	reader := bytes.NewReader(data)
	res, err := ReaderPreview(reader, 1000)
	if err != nil {
		t.Fatalf("ReaderPreview() error: %v", err)
	}
	if res != true {
		t.Error("ReaderPreview() should return true when limit exceeds content size")
	}
}

// TestSingleByte verifies behavior with single-byte content.
func TestSingleByte(t *testing.T) {
	// Single valid ASCII character
	res, err := Bytes([]byte("A"))
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}
	if res != true {
		t.Error("Bytes('A') should return true")
	}

	// Single null byte
	res, err = Bytes([]byte{0x00})
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}
	if res != false {
		t.Error("Bytes(0x00) should return false")
	}
}

// TestPreviewDoesNotTrimWhenFullContent verifies that non-preview (full read) does NOT
// trim incomplete UTF-8, correctly detecting it as invalid.
func TestPreviewDoesNotTrimWhenFullContent(t *testing.T) {
	// File with trailing incomplete UTF-8 (not caused by preview truncation)
	data := []byte("Hello")
	data = append(data, 0xC3) // First byte of a 2-byte UTF-8 sequence, missing second byte

	res, err := Bytes(data)
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}
	if res != false {
		t.Error("Bytes() should return false for content with trailing incomplete UTF-8")
	}

	reader := bytes.NewReader(data)
	res, err = Reader(reader)
	if err != nil {
		t.Fatalf("Reader() error: %v", err)
	}
	if res != false {
		t.Error("Reader() should return false for content with trailing incomplete UTF-8")
	}

	// File variant
	tmpfile, err := os.CreateTemp("", "plaintext_test_incomplete_utf8")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Write(data)
	tmpfile.Close()

	res, err = File(tmpfile.Name())
	if err != nil {
		t.Fatalf("File() error: %v", err)
	}
	if res != false {
		t.Error("File() should return false for file with trailing incomplete UTF-8")
	}
}

// TestEscapeCharacter verifies that ESC (0x1B) is detected as a control character.
func TestEscapeCharacter(t *testing.T) {
	data := []byte("Hello\x1B[31mRed\x1B[0m")
	res, err := Bytes(data)
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}
	if res != false {
		t.Error("Bytes() should return false for content with ANSI escape sequences")
	}
}

// TestBackspaceCharacter verifies that backspace (0x08) is detected as a control character.
func TestBackspaceCharacter(t *testing.T) {
	data := []byte("Hello\x08World")
	res, err := Bytes(data)
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}
	if res != false {
		t.Error("Bytes() should return false for content with backspace character")
	}
}

// TestFilePreviewLargeKB verifies FilePreview with a limit larger than the file.
func TestFilePreviewLargeKB(t *testing.T) {
	data := []byte("Small file content")
	tmpfile, err := os.CreateTemp("", "plaintext_test_largekb")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Write(data)
	tmpfile.Close()

	res, err := FilePreview(tmpfile.Name(), 100)
	if err != nil {
		t.Fatalf("FilePreview() error: %v", err)
	}
	if res != true {
		t.Error("FilePreview() should return true when limit exceeds file size")
	}
}

// TestPreviewEntirelyInvalidUTF8 covers the trimIncompleteUTF8 fallback path
// where the data has pervasive invalid UTF-8 that trimming cannot fix.
func TestPreviewEntirelyInvalidUTF8(t *testing.T) {
	// 4 bytes all invalid UTF-8 (continuation bytes without a start byte)
	data := []byte{0x80, 0x81, 0x82, 0x83}
	reader := bytes.NewReader(data)
	res, err := ReaderPreview(reader, 1)
	if err != nil {
		t.Fatalf("ReaderPreview() error: %v", err)
	}
	if res != false {
		t.Error("ReaderPreview() should return false for entirely invalid UTF-8 data")
	}
}

// TestPreviewOnlyPartialMultibyte covers the case where preview reads only
// an incomplete multi-byte character (buffer becomes empty after trimming).
func TestPreviewOnlyPartialMultibyte(t *testing.T) {
	// Content is a single 4-byte emoji, but preview limit is so small
	// that only the first 1-3 bytes are read, then trimming removes them all.
	// We need to carefully construct this: 3 bytes of a 4-byte char.
	// Use LimitReader to read exactly 3 bytes of a 4-byte emoji.
	emoji := []byte("🌍") // F0 9F 8C 8D (4 bytes)
	// Write the emoji to a temp file and preview just enough to get partial bytes
	tmpfile, err := os.CreateTemp("", "plaintext_test_partial_mb")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	// Write filler of exactly 0 bytes + the emoji = 4 bytes total
	tmpfile.Write(emoji)
	tmpfile.Close()

	// Use ReaderPreview with a very small content that's just the emoji
	reader := bytes.NewReader(emoji[:3]) // Read only first 3 bytes of 4-byte emoji
	// Can't use ReaderPreview directly since minKB is 1 (1024 bytes)
	// Instead test via internal function behavior through Bytes with specific preview path
	// Actually, let's just craft it: the emoji is 4 bytes, if we have only those 4 bytes
	// and preview 1KB, we get all 4 bytes (fits in 1KB), so no split.

	// Better approach: put 1023 bytes of valid text + a 4-byte emoji = 1027 bytes.
	// Preview 1KB = 1024 bytes reads 1023 valid + 1 byte of emoji → trimmed to 1023.
	// That's covered above. For the "empty after trim" case, we need content that is
	// entirely a partial multi-byte char.
	// Use content of just 2 bytes of a 3-byte char.
	partial := []byte{0xE4, 0xBD} // First 2 bytes of 你 (E4 BD A0)
	reader = bytes.NewReader(partial)
	res, err := ReaderPreview(reader, 1)
	if err != nil {
		t.Fatalf("ReaderPreview() error: %v", err)
	}
	if res != true {
		t.Error("ReaderPreview() should return true when trimming removes all content (only partial char)")
	}
}

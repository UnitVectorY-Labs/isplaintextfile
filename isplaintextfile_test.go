package isplaintextfile

import (
	"bytes"
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
			// For tests where no explicit limit is set, use a limit high enough to cover the entire content.
			limitKB := tt.limitKB
			if limitKB == 0 {
				limitKB = (len(tt.content) + 1023) / 1024
			}

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
			// For tests where no explicit limit is set, use a limit high enough to cover the entire content.
			limitKB := tt.limitKB
			if limitKB == 0 {
				limitKB = (len(tt.content) + 1023) / 1024
			}

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

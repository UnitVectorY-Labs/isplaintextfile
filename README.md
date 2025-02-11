# isplaintextfile

A lightweight Go module that determines whether a given file is plaintext by analyzing its content.

## Features

- **Multiple Input Types:**  
  Analyze plaintext from in-memory byte slices, file paths, or any `io.Reader`.
- **Preview Mode:**  
  Check only the first portion (specified in kilobytes) of a file or stream to determine if it’s plaintext.
- **No External Dependencies:**  
  Relies solely on the Go standard library.

## Usage

```go
import "github.com/UnitVectorY-Labs/isplaintextfile/filetype"
```

1. Checking In-Memory Data

Use `IsPlaintextBytes` to determine if a byte slice contains valid plaintext:

```go
data := []byte("Hello, World!\n")
isText, err := filetype.IsPlaintextBytes(data)
if err != nil {
    // Handle error.
}
if isText {
    // The byte slice is plaintext.
}
```

2. Checking a File by Path (Full Content)

To analyze the entire content of a file, use `IsPlaintextPath`:

```go
isText, err := filetype.IsPlaintextPath("example.txt")
if err != nil {
    // Handle error.
}
if isText {
    // The file is entirely plaintext.
}
```

3. Checking a File by Path (Preview Mode)

If you want to check only the first part of a file (e.g. the first N kilobytes), use `IsPlaintextPathPreview`:

```go
// Check only the first 2KB of the file.
isText, err := filetype.IsPlaintextPathPreview("example.txt", 2)
if err != nil {
    // Handle error.
}
if isText {
    // The first 2KB of the file is plaintext.
}
```

4. Checking Data from an io.Reader (Full Content)

For situations where the data comes from an io.Reader (such as a network stream), use `IsPlaintextReader`:

```go
// Assume 'reader' is an io.Reader.
isText, err := filetype.IsPlaintextReader(reader)
if err != nil {
    // Handle error.
}
if isText {
    // The stream content is plaintext.
}
```

5. Checking an io.Reader in Preview Mode

To analyze only the first portion of data from an io.Reader, use `IsPlaintextReaderPreview`:

```go
// Check only the first 1KB from the reader.
isText, err := filetype.IsPlaintextReaderPreview(reader, 1)
if err != nil {
    // Handle error.
}
if isText {
    // The preview (first 1KB) of the stream is plaintext.
}
```

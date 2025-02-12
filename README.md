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
import "github.com/UnitVectorY-Labs/isplaintextfile"
```

1. Checking In-Memory Data

Use `Bytes` to determine if a byte slice contains valid plaintext:

```go
data := []byte("Hello, World!\n")
isText, err := isplaintextfile.Bytes(data)
if err != nil {
    // Handle error.
}
if isText {
    // The byte slice is plaintext.
}
```

2. Checking a File by Path (Full Content)

To analyze the entire content of a file, use `File`:

```go
isText, err := isplaintextfile.File("example.txt")
if err != nil {
    // Handle error.
}
if isText {
    // The file is entirely plaintext.
}
```

3. Checking a File by Path (Preview Mode)

If you want to check only the first part of a file (e.g. the first N kilobytes), use `FilePreview`:

```go
// Check only the first 2KB of the file.
isText, err := isplaintextfile.FilePreview("example.txt", 2)
if err != nil {
    // Handle error.
}
if isText {
    // The first 2KB of the file is plaintext.
}
```

4. Checking Data from an io.Reader (Full Content)

For situations where the data comes from an io.Reader (such as a network stream), use `Reader`:

```go
// Assume 'reader' is an io.Reader.
isText, err := isplaintextfile.Reader(reader)
if err != nil {
    // Handle error.
}
if isText {
    // The stream content is plaintext.
}
```

5. Checking an io.Reader in Preview Mode

To analyze only the first portion of data from an io.Reader, use `ReaderPreview`:

```go
// Check only the first 1KB from the reader.
isText, err := isplaintextfile.ReaderPreview(reader, 1)
if err != nil {
    // Handle error.
}
if isText {
    // The preview (first 1KB) of the stream is plaintext.
}

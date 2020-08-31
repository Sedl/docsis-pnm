package tftp

import (
    "bytes"
    "errors"
    "fmt"
)

type SafeBuffer struct {
    bytes.Buffer
    BytesWritten int
    Limit int
}

func (b *SafeBuffer) Write(p []byte) (n int, err error) {
    b.BytesWritten += len(p)
    if b.BytesWritten > b.Limit {
        return 0, errors.New(fmt.Sprintf("write (%d bytes) exceeds maximum buffer size of %d bytes", b.BytesWritten, b.Limit))
    }
    return b.Buffer.Write(p)
}

func NewSafeBuffer(sizeLimit int) *SafeBuffer {
    return &SafeBuffer{Limit: sizeLimit}
}
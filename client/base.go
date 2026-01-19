package client

import (
	"io"
	"log"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
)

// Create Brotli decompress logic
func decompressBrotli(r io.ReadCloser) (io.ReadCloser, error) {
	br := &brotliReader{s: r, r: brotli.NewReader(r)}
	return br, nil
}

type brotliReader struct {
	s io.ReadCloser
	r *brotli.Reader
}

func (b *brotliReader) Read(p []byte) (n int, err error) {
	return b.r.Read(p)
}

func (b *brotliReader) Close() error {
	return b.s.Close()
}

// Create Zstandard decompress logic
func decompressZstd(r io.ReadCloser) (io.ReadCloser, error) {
	zr, err := zstd.NewReader(r, nil)
	if err != nil {
		return nil, err
	}
	z := &zstdReader{s: r, r: zr}
	return z, nil
}

type zstdReader struct {
	s io.ReadCloser
	r *zstd.Decoder
}

func (b *zstdReader) Read(p []byte) (n int, err error) {
	return b.r.Read(p)
}

func (b *zstdReader) Close() error {
	b.r.Close()
	return b.s.Close()
}

// 自定义日志记录器
type CustomLogger struct{}

// Implement the Log方法
// func (c *CustomLogger) Log(v ...interface{}) {
// 	log.Println(v...)
// }

// Debugf 实现了 resty.Logger 接口的 Debugf 方法
func (l *CustomLogger) Debugf(format string, v ...interface{}) {
	log.Printf("DEBUG: "+format, v...)
}

// Errorf 实现了 resty.Logger 接口的 Errorf 方法
func (l *CustomLogger) Errorf(format string, v ...interface{}) {
	log.Printf("ERROR: "+format, v...)
}
func (l *CustomLogger) Warnf(format string, v ...interface{}) {
	log.Printf("ERROR: "+format, v...)
}

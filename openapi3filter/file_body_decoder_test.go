package openapi3filter

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

type fileBodyReaderOnly struct{ io.Reader }

type fileBodyErrorReader struct{ err error }

func (r fileBodyErrorReader) Read([]byte) (int, error) { return 0, r.err }

type fileBodyWriterTo struct {
	*bytes.Reader
	err error
}

func (r fileBodyWriterTo) WriteTo(w io.Writer) (int64, error) {
	n, err := r.Reader.WriteTo(w)
	if err != nil {
		return n, err
	}
	return n, r.err
}

func (r fileBodyWriterTo) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if err == io.EOF && r.err != nil {
		return n, r.err
	}
	return n, err
}

func TestFileBodyDecoderBytes(t *testing.T) {
	t.Parallel()

	for _, data := range [][]byte{nil, {}, []byte("hello"), {0, 0xff, 0xfe, '\r', '\n', 0x80}} {
		for _, kind := range []string{"bytes", "reader-only", "one-byte", "data-and-eof", "writer-to"} {
			t.Run(fmt.Sprintf("%x/%s", data, kind), func(t *testing.T) {
				input := bytes.Clone(data)
				want := string(input)
				var body io.Reader = bytes.NewReader(input)
				switch kind {
				case "reader-only":
					body = fileBodyReaderOnly{body}
				case "one-byte":
					body = iotest.OneByteReader(body)
				case "data-and-eof":
					body = iotest.DataErrReader(body)
				case "writer-to":
					body = fileBodyWriterTo{Reader: bytes.NewReader(input)}
				}
				got, err := FileBodyDecoder(body, nil, nil, nil)
				require.NoError(t, err)
				require.Equal(t, want, got)
				for i := range input {
					input[i] = 'x'
				}
				require.Equal(t, want, got, "the result must not alias the reader's bytes")
			})
		}
	}
}

func TestFileBodyDecoderReaderPosition(t *testing.T) {
	t.Parallel()

	data := []byte("prefix\x00\xffsuffix")
	for _, offset := range []int64{0, 6, int64(len(data)), int64(len(data) + 10)} {
		t.Run(fmt.Sprint(offset), func(t *testing.T) {
			body := bytes.NewReader(data)
			_, err := body.Seek(offset, io.SeekStart)
			require.NoError(t, err)
			got, err := FileBodyDecoder(body, nil, nil, nil)
			require.NoError(t, err)
			want := ""
			if offset < int64(len(data)) {
				want = string(data[offset:])
			}
			require.Equal(t, want, got)
			require.Zero(t, body.Len())
			_, err = body.ReadByte()
			require.ErrorIs(t, err, io.EOF)
		})
	}
}

func TestFileBodyDecoderExhaustedRune(t *testing.T) {
	t.Parallel()

	body := bytes.NewReader([]byte("x"))
	_, _, err := body.ReadRune()
	require.NoError(t, err)
	got, err := FileBodyDecoder(body, nil, nil, nil)
	require.NoError(t, err)
	require.Equal(t, "", got)
	require.NoError(t, body.UnreadRune(), "an EOF read must not invalidate the preceding ReadRune")
}

func TestFileBodyDecoderReadError(t *testing.T) {
	t.Parallel()

	readErr := errors.New("read failed")
	for _, kind := range []string{"before-data", "after-data", "writer-to"} {
		t.Run(kind, func(t *testing.T) {
			var body io.Reader = fileBodyErrorReader{readErr}
			if kind == "after-data" {
				body = io.MultiReader(bytes.NewReader([]byte("partial")), body)
			} else if kind == "writer-to" {
				body = fileBodyWriterTo{Reader: bytes.NewReader([]byte("partial")), err: readErr}
			}
			got, err := FileBodyDecoder(body, nil, nil, nil)
			require.Same(t, readErr, err)
			require.Nil(t, got)
		})
	}
}

func TestFileBodyValidationAndReplay(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		body   []byte
		schema *openapi3.Schema
		valid  bool
	}{
		{"binary", []byte{0, 0xff, 0xfe}, openapi3.NewStringSchema().WithFormat("binary"), true},
		{"min-length", []byte("ab"), openapi3.NewStringSchema().WithMinLength(3), false},
		{"max-length", []byte("abcd"), openapi3.NewStringSchema().WithMaxLength(3), false},
		{"pattern-match", []byte("abc"), openapi3.NewStringSchema().WithPattern("^abc$"), true},
		{"pattern-mismatch", []byte("xyz"), openapi3.NewStringSchema().WithPattern("^abc$"), false},
		{"enum-match", []byte("abc"), openapi3.NewStringSchema().WithEnum("abc"), true},
		{"enum-mismatch", []byte("xyz"), openapi3.NewStringSchema().WithEnum("abc"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "http://example.invalid/upload", fileBodyReaderOnly{bytes.NewReader(tc.body)})
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/octet-stream")
			body := &openapi3.RequestBody{Content: openapi3.NewContentWithSchema(tc.schema, []string{"application/octet-stream"})}
			err = ValidateRequestBody(t.Context(), &RequestValidationInput{Request: req}, body)
			if tc.valid {
				require.NoError(t, err)
			} else {
				var schemaErr *openapi3.SchemaError
				require.ErrorAs(t, err, &schemaErr)
			}
			data, err := io.ReadAll(req.Body)
			require.NoError(t, err)
			require.Equal(t, tc.body, data)
			require.NoError(t, req.Body.Close())
			replay, err := req.GetBody()
			require.NoError(t, err)
			defer replay.Close()
			data, err = io.ReadAll(replay)
			require.NoError(t, err)
			require.Equal(t, tc.body, data)
			require.Equal(t, int64(len(tc.body)), req.ContentLength)
		})
	}
}

func TestFileBodyValidationCustomDecoder(t *testing.T) {
	const mediaType = "application/x-file-body-test"
	calls := 0
	RegisterBodyDecoder(mediaType, func(body io.Reader, _ http.Header, _ *openapi3.SchemaRef, _ EncodingFn) (any, error) {
		calls++
		data, err := io.ReadAll(body)
		require.NoError(t, err)
		require.Equal(t, "raw", string(data))
		return "decoded", nil
	})
	defer UnregisterBodyDecoder(mediaType)
	req, err := http.NewRequest(http.MethodPost, "http://example.invalid/upload", bytes.NewReader([]byte("raw")))
	require.NoError(t, err)
	req.Header.Set("Content-Type", mediaType)
	body := &openapi3.RequestBody{Content: openapi3.NewContentWithSchema(openapi3.NewStringSchema().WithEnum("decoded"), []string{mediaType})}
	require.NoError(t, ValidateRequestBody(t.Context(), &RequestValidationInput{Request: req}, body))
	require.Equal(t, 1, calls)
}

func BenchmarkFileBodyDecoder(b *testing.B) {
	for _, size := range []int{81, 1 << 20, 8 << 20} {
		data := bytes.Repeat([]byte("x"), size)
		for _, kind := range []string{"bytes", "reader-only"} {
			b.Run(fmt.Sprintf("%d/%s", size, kind), func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(size))
				for b.Loop() {
					var body io.Reader = bytes.NewReader(data)
					if kind == "reader-only" {
						body = fileBodyReaderOnly{body}
					}
					got, err := FileBodyDecoder(body, nil, nil, nil)
					if err != nil || len(got.(string)) != size {
						b.Fatalf("unexpected result: %v", err)
					}
				}
			})
		}
	}
}

func BenchmarkFileBodyValidation(b *testing.B) {
	for _, size := range []int{1 << 20, 8 << 20} {
		data := bytes.Repeat([]byte("x"), size)
		body := &openapi3.RequestBody{Content: openapi3.NewContentWithSchema(openapi3.NewStringSchema().WithFormat("binary"), []string{"application/octet-stream"})}
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(size))
			for b.Loop() {
				req, err := http.NewRequest(http.MethodPost, "http://example.invalid/upload", bytes.NewReader(data))
				if err != nil {
					b.Fatal(err)
				}
				req.Header.Set("Content-Type", "application/octet-stream")
				if err := ValidateRequestBody(b.Context(), &RequestValidationInput{Request: req}, body); err != nil {
					b.Fatal(err)
				}
				req.Body.Close()
			}
		})
	}
}

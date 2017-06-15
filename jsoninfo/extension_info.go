package jsoninfo

type ExtensionInfo struct {
	Summary    string
	EncodeFunc func(encoder *ObjectEncoder, value interface{}) error
	DecodeFunc func(decoder *ObjectDecoder, value interface{}) error
}

package tgapi

import "encoding/base64"

//go:generate go run golang.org/x/tools/cmd/stringer@latest -type=TextData
type TextData int

const (
	TextDataText TextData = iota
	TextDataAnimation
	TextDataAudio
	TextDataDocument
	TextDataPhoto
	TextDataSticker
	TextDataVideo
	TextDataVoice
)

func (i TextData) IsValid() bool {
	return i >= TextDataText && i <= TextDataVoice
}

type SerializableTextData struct {
	Type    TextData
	Message []byte
}

func (s SerializableTextData) String() string {
	if s.Type == TextDataText {
		return string(s.Message)
	}

	return s.Type.String() + ":" + s.ToBase64()
}

func (s SerializableTextData) ToBase64() string {
	return base64.StdEncoding.EncodeToString(s.Message)
}

package rrpc

import "rrpc/codec"

const MagicNumber = 0x0a0d0d0a

type Option struct {
	MagicNumber int        // MagicNumber marks this's a rrpc request
	CodecType   codec.Type // client may choose different Codec to encode body
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

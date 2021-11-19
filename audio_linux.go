package reisen

func bufferSize(maxBufferSize C.int) C.ulong {
	var byteSize C.ulong = 8
	return C.ulong(maxBufferSize) * byteSize
}

func channelLayout(audio *AudioStream) C.long {
	return C.long(audio.codecCtx.channel_layout)
}


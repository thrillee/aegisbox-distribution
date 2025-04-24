package smpphelper

import "time"

// EncodeConcatenatedUDH creates the UDH byte slice for multipart messages.
func EncodeConcatenatedUDH(totalSegments, sequenceNum uint8) []byte {
	refNum := uint8(time.Now().Unix() & 0xFF) // Simple example, might have collisions! Use a better reference generator.
	concat := udh.NewConcatMessage(udh.Ref8bit, totalSegments, sequenceNum)
	concat.SetReference(uint16(refNum)) // Assuming Ref8bit uses low byte of uint16
	return concat.Encode()
}

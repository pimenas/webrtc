package samplebuilder

import (
	"testing"
	"time"

	"github.com/pions/webrtc/pkg/media"
	"github.com/pions/webrtc/pkg/rtp"
	"github.com/stretchr/testify/assert"
)

type sampleBuilderTest struct {
	message    string
	packets    []*rtp.Packet
	samples    []*media.RTCSample
	bufferSize uint16
}

type fakeDepacketizer struct {
}

func (f *fakeDepacketizer) Unmarshal(packet *rtp.Packet) ([]byte, error) {
	return packet.Payload, nil
}

var testCases = []sampleBuilderTest{
	{
		message: "SampleBuilder shouldn't emit anything if only one RTP packet has been pushed",
		packets: []*rtp.Packet{
			{SequenceNumber: 5000, Timestamp: 5, Payload: []byte{0x01}},
		},
		samples:    []*media.RTCSample{},
		bufferSize: 50,
	},
	{
		message: "SampleBuilder should emit one packet, we had three packets with unique timestamps",
		packets: []*rtp.Packet{
			{SequenceNumber: 5000, Timestamp: 0x00050000, Payload: []byte{0x01}},
			{SequenceNumber: 5001, Timestamp: 0x00060000, Payload: []byte{0x02}},
			{SequenceNumber: 5002, Timestamp: 0x00070000, Payload: []byte{0x03}},
		},
		samples: []*media.RTCSample{
			{Data: []byte{0x02}, Duration: time.Second},
		},
		bufferSize: 50,
	},
	{
		message: "SampleBuilder should emit one packet, we had two packets but two with duplicate timestamps",
		packets: []*rtp.Packet{
			{SequenceNumber: 5000, Timestamp: 0x00050000, Payload: []byte{0x01}},
			{SequenceNumber: 5001, Timestamp: 0x00060000, Payload: []byte{0x02}},
			{SequenceNumber: 5002, Timestamp: 0x00060000, Payload: []byte{0x03}},
			{SequenceNumber: 5003, Timestamp: 0x00070000, Payload: []byte{0x04}},
		},
		samples: []*media.RTCSample{
			{Data: []byte{0x02, 0x03}, Duration: time.Second},
		},
		bufferSize: 50,
	},
	{
		message: "SampleBuilder shouldn't emit a packet because we have a gap before a valid one",
		packets: []*rtp.Packet{
			{SequenceNumber: 5000, Timestamp: 0x00050000, Payload: []byte{0x01}},
			{SequenceNumber: 5007, Timestamp: 0x00060000, Payload: []byte{0x02}},
			{SequenceNumber: 5008, Timestamp: 0x00070000, Payload: []byte{0x03}},
		},
		samples:    []*media.RTCSample{},
		bufferSize: 50,
	},
	{
		message: "SampleBuilder should emit multiple valid packets",
		packets: []*rtp.Packet{
			{SequenceNumber: 5000, Timestamp: 0x00010000, Payload: []byte{0x01}},
			{SequenceNumber: 5001, Timestamp: 0x00020000, Payload: []byte{0x02}},
			{SequenceNumber: 5002, Timestamp: 0x00030000, Payload: []byte{0x03}},
			{SequenceNumber: 5003, Timestamp: 0x00040000, Payload: []byte{0x04}},
			{SequenceNumber: 5004, Timestamp: 0x00050000, Payload: []byte{0x05}},
			{SequenceNumber: 5005, Timestamp: 0x00060000, Payload: []byte{0x06}},
		},
		samples: []*media.RTCSample{
			{Data: []byte{0x02}, Duration: time.Second},
			{Data: []byte{0x03}, Duration: time.Second},
			{Data: []byte{0x04}, Duration: time.Second},
			{Data: []byte{0x05}, Duration: time.Second},
		},
		bufferSize: 50,
	},
}

func TestSampleBuilder(t *testing.T) {
	assert := assert.New(t)

	for _, t := range testCases {
		s := New(t.bufferSize, &fakeDepacketizer{})
		samples := []*media.RTCSample{}

		for _, p := range t.packets {
			s.Push(p)
		}
		for sample := s.Pop(); sample != nil; sample = s.Pop() {
			samples = append(samples, sample)
		}

		assert.Equal(samples, t.samples, t.message)
	}
}

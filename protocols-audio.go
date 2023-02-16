package main

import (
	"github.com/gen2brain/malgo"
	"github.com/libp2p/go-libp2p/core/network"
)

var pReceivedSamples = make(chan []byte)
var RecvBuff = make(chan []byte, 500)
var audioChan = make(chan []byte)

var SampleRate int = 44100
var temp float32 = 2 //time in s to send

type StreamConfig struct {
	Format     malgo.FormatType
	Channels   int
	SampleRate int
}

func SendAudioHandler(rendezvous string) {
	for {
		data := <-audioChan
		WriteDataRend(data, "/audio/1.1.0", rendezvous)
	}

}

func ReceiveAudioHandler(stream network.Stream) {
	count := len(<-audioChan)
	reps := temp / float32(count) * float32(SampleRate) * 4
	length := (int(reps) * count)

	go readData(stream, uint16(length), func(buff []byte, stream network.Stream) {

		RecvBuff <- buff

	})

}

func initAudio(ctx *malgo.AllocatedContext) {

	var config StreamConfig
	config.Format = malgo.FormatS16
	config.Channels = 2
	config.SampleRate = SampleRate

	var captureChan = make(chan []byte)

	Capture(ctx, captureChan, config)
	go Playback(ctx, RecvBuff, config)

	go FrameChan(captureChan)

}
func recordAudio(ctx *malgo.AllocatedContext, rendezvous string, quitchan chan bool) {
	var config StreamConfig
	config.Format = malgo.FormatS16
	config.Channels = 2
	config.SampleRate = SampleRate
	var aux []byte
	var captureChan = make(chan []byte)
	Capture(ctx, captureChan, config)
	select {

	case data := <-audioChan:

		aux = append(aux, data...)

	case <-quitchan:

		WriteDataRend(aux, "/audio/1.1.0", rendezvous)
		break
	}

}
func quitAudio(ctx *malgo.AllocatedContext) {
	ctx.Free()
}

//write 20 ms of data into AudioChan variable a send audio data at 20 ms intervals to audio chan
func FrameChan(channel chan []byte) {

	var count int

	count = len(<-channel)
	reps := temp / float32(count) * float32(SampleRate) * 4

	for {
		var aux []byte

		for i := 0; i < int(reps); i++ {

			data := <-channel

			aux = append(aux, data...)
		}

		audioChan <- aux
	}

}

func stream(ctx *malgo.AllocatedContext, deviceConfig malgo.DeviceConfig, deviceCallbacks malgo.DeviceCallbacks) error {

	device, err := malgo.InitDevice(ctx.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		return err
	}

	err = device.Start()

	return err
}

// StreamConfig describes the parameters for an audio stream.
// Default values will pick the defaults of the default device.

func (config StreamConfig) asDeviceConfig(deviceType malgo.DeviceType) malgo.DeviceConfig {
	deviceConfig := malgo.DefaultDeviceConfig(deviceType)
	if config.Format != malgo.FormatUnknown {
		deviceConfig.Capture.Format = config.Format
		deviceConfig.Playback.Format = config.Format
	}
	if config.Channels != 0 {
		deviceConfig.Capture.Channels = uint32(config.Channels)
		deviceConfig.Playback.Channels = uint32(config.Channels)
	}
	if config.SampleRate != 0 {
		deviceConfig.SampleRate = uint32(config.SampleRate)
	}
	return deviceConfig
}

// Capture records incoming samples into the provided writer.
// The function initializes a capture device in the default context using
// provide stream configuration.
// Capturing will commence writing the samples to the writer until either the
// writer returns an error, or the context signals done.

func Capture(ctx *malgo.AllocatedContext, samples chan []byte, config StreamConfig) error {
	deviceConfig := config.asDeviceConfig(malgo.Capture)

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: func(outputSamples, inputSamples []byte, frameCount uint32) {

			samples <- inputSamples

		},
	}

	return stream(ctx, deviceConfig, deviceCallbacks)
}

// Playback streams samples from a reader to the sound device.
// The function initializes a playback device in the default context using
// provide stream configuration.
// Playback will commence playing the samples provided from the reader until either the
// reader returns an error, or the context signals done.
func Playback(ctx *malgo.AllocatedContext, samples chan []byte, config StreamConfig) error {
	deviceConfig := config.asDeviceConfig(malgo.Playback)

	sizeInBytes := uint32(malgo.SampleSizeInBytes(deviceConfig.Capture.Format))
	buffer := make([]byte, 0)
	var Samples uint32
	Samples = 0
	deviceCallbacks := malgo.DeviceCallbacks{
		Data: func(outputSamples, inputSamples []byte, frameCount uint32) {
			samplesToRead := frameCount * deviceConfig.Playback.Channels * sizeInBytes

			if int(Samples+samplesToRead) < len(buffer) {
				copy(outputSamples, buffer[Samples:Samples+samplesToRead])

				Samples += samplesToRead
			} else {
				Samples = 0

				buffer = <-samples

			}

		},
	}

	return stream(ctx, deviceConfig, deviceCallbacks)
}

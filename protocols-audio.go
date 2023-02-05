package main

import (
	"fmt"
	"os"

	"github.com/gen2brain/malgo"
	"github.com/libp2p/go-libp2p/core/network"
)

var pReceivedSamples = make(chan []byte)
var RecvBuff = make([]byte, 0)
var AudioChan = make(chan []byte)
var length int

func SendAudioHandler() {
	//go func() { AudioChan <- RecvBuff }()

	go WriteData(AudioChan, "/audio/1.1.0")

}

func ReceiveAudioHandler(stream network.Stream) {

	go readData(stream, uint16(length), func(buff []byte, stream network.Stream) {

		RecvBuff = buff

	})

}

func initAudio() *malgo.AllocatedContext {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("LOG <%v>\n", message)
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()
	return ctx
}
func initCaptureDevice(ctx *malgo.AllocatedContext) *malgo.Device {

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Duplex)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 2
	deviceConfig.SampleRate = 44100

	var capturedSampleCount uint32
	var pCapturedSamples = make([]byte, 0)
	sizeInBytes := uint32(malgo.SampleSizeInBytes(deviceConfig.Capture.Format))
	length = int(deviceConfig.SampleRate * deviceConfig.Capture.Channels * sizeInBytes / 10)
	onRecvFrames := func(pOutputSample, pInputSamples []byte, framecount uint32) {

		sampleCount := framecount * deviceConfig.Capture.Channels * sizeInBytes

		newCapturedSampleCount := capturedSampleCount + sampleCount

		pCapturedSamples = append(pCapturedSamples, pInputSamples...)
		//fmt.Println(capturedSampleCount)
		if uint32(len(pCapturedSamples)) > uint32(length) {

			//RecvBuff = pCapturedSamples
			fmt.Println(len(pCapturedSamples))
			AudioChan <- pCapturedSamples

			newCapturedSampleCount = 0
			pCapturedSamples = make([]byte, 0)

		}

		capturedSampleCount = newCapturedSampleCount

	}
	captureCallbacks := malgo.DeviceCallbacks{
		Data: onRecvFrames,
	}
	device, err := malgo.InitDevice(ctx.Context, deviceConfig, captureCallbacks)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return device
}
func startDevice(device *malgo.Device) {
	err := device.Start()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
func stopDevice(device *malgo.Device) {
	err := device.Start()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	device.Uninit()

}

func initPlaybackDevice(ctx *malgo.AllocatedContext) *malgo.Device {
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Duplex)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 2

	deviceConfig.Playback.Format = malgo.FormatS16
	deviceConfig.Playback.Channels = 2
	deviceConfig.SampleRate = 44100
	var Samples uint32
	Samples = 0
	sizeInBytes := uint32(malgo.SampleSizeInBytes(deviceConfig.Capture.Format))

	onSendFrames := func(pSample, nil []byte, framecount uint32) {

		samplesToRead := framecount * deviceConfig.Playback.Channels * sizeInBytes
		fmt.Println(int(Samples+samplesToRead), len(RecvBuff))
		if int(Samples+samplesToRead) < len(RecvBuff) {
			copy(pSample, RecvBuff[Samples:Samples+samplesToRead])
			fmt.Println("dsad")
			Samples += samplesToRead
		} else {
			Samples = 0
			RecvBuff = make([]byte, length)

		}

	}
	playbackCallbacks := malgo.DeviceCallbacks{
		Data: onSendFrames,
	}

	device, err := malgo.InitDevice(ctx.Context, deviceConfig, playbackCallbacks)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return device
}

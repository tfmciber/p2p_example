package main

import (
	"fmt"
	"os"

	"github.com/gordonklaus/portaudio"
)

const sampleRate = 44100
const seconds = 1

func record() {

	portaudio.Initialize()
	fmt.Print(portaudio.DefaultInputDevice())
	fmt.Print(portaudio.DefaultOutputDevice())

	fmt.Print(portaudio.Devices())
	os.Exit(2)

	defer portaudio.Terminate()
	buffer := make([]float32, sampleRate*seconds)
	stream, err := portaudio.OpenDefaultStream(1, 0, sampleRate, len(buffer), func(in []float32) {
		for i := range buffer {
			buffer[i] = in[i]
		}
	})

	if err != nil {
		panic(err)
	}
	stream.Start()
	defer stream.Close()

	play(buffer, sampleRate, seconds)
}

func play(buffer []float32, sampleRate float64, seconds float32) {

	portaudio.Initialize()
	defer portaudio.Terminate()

	stream, err := portaudio.OpenDefaultStream(0, 1, sampleRate, len(buffer), func(out float32) {
		for i := range buffer {
			out = buffer[i]
		}
	})

	if err != nil {
		panic(err)
	}
	stream.Start()
	defer stream.Close()

}

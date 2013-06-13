package main

import (
	"code.google.com/p/portaudio-go/portaudio"
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("Usage:\n \t./%s audio_file", os.Args[0])
		os.Exit(0)
	}
	inputfile := os.Args[1]
	framePerBuffer := 2048
	ff := NewFFmpegDecoder(inputfile)
	defer ff.Close()
	stream, err := portaudio.OpenDefaultStream(0, 2, 44100, framePerBuffer, ff)
	chk(err)
	defer stream.Close()
	chk(stream.Start())
	if err := ff.cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	chk(stream.Stop())
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}

package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"os/exec"
	"regexp"
)

type FFmpegDecoder struct {
	in         io.ReadCloser
	err        io.ReadCloser
	cmd        *exec.Cmd
	Filename   string
	sampleRate int64
	info       chan struct{}
}

func NewFFmpegDecoder(filename string) *FFmpegDecoder {
	cmd := exec.Command("ffmpeg", "-i", filename, "-f", "s16le", "-")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	ffdec := &FFmpegDecoder{in: stdout, err: stderr, cmd: cmd, Filename: filename}
        ffdec.info = make(chan struct{})
	go ffdec.getInfo()
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	return ffdec
}

func (f *FFmpegDecoder) Close() error {
	f.in.Close()
	return f.err.Close()
}

func (f *FFmpegDecoder) getInfo() {
	stderrReader := bufio.NewReader(f.err)
	noSuchFile := []byte("No such file")
	invalidData := []byte("Invalid data found")
	durationLine := []byte("Duration:")
	audioLine := []byte("Audio:")
	metaRegEx, err := regexp.Compile("(\\d+)\\shz,\\s([^,]+),")
	if err != nil {
		log.Fatal(err)
	}
	var line []byte
	for {
		l, _, err := stderrReader.ReadLine()
		if err != nil {
			log.Fatal("failed to read from stderr: ", err)
		}
		if bytes.Contains(l, noSuchFile) || bytes.Contains(l, invalidData) {
			log.Fatal("invalid file: ", string(l), f.Filename)
		} else if bytes.Contains(l, durationLine) {
			line = append(line, l...)
		} else if bytes.Contains(l, audioLine) {
			matches := metaRegEx.FindAll(line, 0)
			if matches != nil {
				if f.sampleRate, err = binary.ReadVarint(bytes.NewBuffer(matches[0])); err != nil {
					log.Fatal("failed to get sample rate: ", err)
				}
			}
			break
		}
	}
	close(f.info)
}

func (f *FFmpegDecoder) SampleRate() int64 {
	<-f.info
	return f.sampleRate
}

// test for portaudio
func (f *FFmpegDecoder) ProcessAudio(_, out [][]int16) {
	// int16 takes 2 bytes
	bufferSize := len(out[0]) * 4
	var pack = make([]byte, bufferSize)
	if _, err := f.in.Read(pack); err != nil {
		log.Fatal(err)
	}
	n := make([]int16, len(out[0])*2)
	for i := range n {
		var x int16
		buf := bytes.NewBuffer(pack[2*i : 2*(i+1)])
		binary.Read(buf, binary.LittleEndian, &x)
		n[i] = x
	}

	for i := range out[0] {
		out[0][i] = n[2*i]
		out[1][i] = n[2*i+1]
	}
}

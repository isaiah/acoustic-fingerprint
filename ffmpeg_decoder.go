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
	cmd        *exec.Cmd
	SampleRate int64
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
	stderrReader := bufio.NewReader(stderr)
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	ffdec := &FFmpegDecoder{in: stdout, cmd: cmd}
	noSuchFile := []byte("no such file")
	invalidData := []byte("invalid data found")
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
                        log.Fatal("invalid file: ", err)
		}
		line = append(line, l...)
	}
	matches := metaRegEx.FindAll(line, 0)
	if matches != nil {
		if ffdec.SampleRate, err = binary.ReadVarint(bytes.NewBuffer(matches[0])); err != nil {
			log.Fatal("failed to get sample rate: ", err)
		}
	}
	return ffdec
}

func (f *FFmpegDecoder) Close() error {
	return f.in.Close()
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

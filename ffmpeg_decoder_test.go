package main

import (
	"launchpad.net/gocheck"
	"testing"
)

func Test(t *testing.T) { gocheck.TestingT(t) }

type S struct {
	testData string
}

var _ = gocheck.Suite(&S{})

func (s *S) SetupSuite(c *gocheck.C) {
	s.testData = "test_data/test.mp3"
}

func (s *S) TestNewFFmpeg(c *gocheck.C) {
	ffDec := NewFFmpegDecoder(s.testData)
	defer ffDec.Close()
	c.Assert(ffDec.SampleRate, gocheck.Equals, 44100)
}

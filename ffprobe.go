package radiorenamer

import (
	"context"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

type ffprobe struct {
	*exec.Cmd
}

func newFfprobe(ctx context.Context) (*ffprobe, error) {
	cmdPath, err := exec.LookPath("ffprobe")
	if err != nil {
		return nil, err
	}

	return &ffprobe{exec.CommandContext(
		ctx,
		cmdPath,
	)}, nil
}

func (f *ffprobe) setDir(dir string) {
	f.Dir = dir
}

func (f *ffprobe) setArgs(args ...string) {
	f.Args = append(f.Args, args...)
}

func (f *ffprobe) setInput(input string) {
	f.setArgs(input)
}

func (f *ffprobe) run() (*[]byte, error) {
	output, err := f.Output()
	return &output, err
}

func (f *ffprobe) start() error {
	return f.Start()
}

func (f *ffprobe) wait() error {
	return f.Wait()
}

func (f *ffprobe) stdinPipe() (io.WriteCloser, error) {
	return f.StdinPipe()
}

func (f *ffprobe) stderrPipe() (io.ReadCloser, error) {
	return f.StderrPipe()
}

func Duration(ctx context.Context, input string) (float32, error) {
	f, err := newFfprobe(ctx)
	if err != nil {
		return 0, err
	}
	f.setInput(input)
	f.setArgs("-hide_banner", "-show_entries", "format=duration")
	// run ffprobe
	output, err := f.run()
	if err != nil {
		log.Fatal("failed to get duration.")
	}

	// parse output
	outStr := string(*output)
	record := strings.Split(outStr, "\n")

	dur := strings.Split(record[1], "=")
	var duration float64
	duration, err = strconv.ParseFloat(dur[1], 64)
	if err != nil {
		log.Fatal("failed to parse duration.")
	}

	return float32(duration), nil

}

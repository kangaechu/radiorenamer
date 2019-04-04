package radiorenamer

import (
	"context"
	"io"
	"os/exec"
)

type ffmpeg struct {
	*exec.Cmd
}

func newFfmpeg(ctx context.Context) (*ffmpeg, error) {
	cmdPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, err
	}

	return &ffmpeg{exec.CommandContext(
		ctx,
		cmdPath,
	)}, nil
}

func (f *ffmpeg) setDir(dir string) {
	f.Dir = dir
}

func (f *ffmpeg) setArgs(args ...string) {
	f.Args = append(f.Args, args...)
}

func (f *ffmpeg) setInput(input string) {
	f.setArgs("-i", input)
}

func (f *ffmpeg) run(output string) error {
	f.setArgs(output)
	return f.Run()
}

func (f *ffmpeg) start(output string) error {
	f.setArgs(output)
	return f.Start()
}

func (f *ffmpeg) wait() error {
	return f.Wait()
}

func (f *ffmpeg) stdinPipe() (io.WriteCloser, error) {
	return f.StdinPipe()
}

func (f *ffmpeg) stderrPipe() (io.ReadCloser, error) {
	return f.StderrPipe()
}

// ConvertM4AtoMP3 converts an m4a file to a mp3 file.
func ConvertM4AtoMP3(ctx context.Context, input, output string) error {
	f, err := newFfmpeg(ctx)
	if err != nil {
		return err
	}

	f.setInput(input)
	f.setArgs(
		"-c:a", "libmp3lame",
		"-ac", "2",
		"-q:a", "2",
		"-y", // overwrite the output file without asking
	)
	// TODO: Collect log
	return f.run(output)
}

// ConcatM4AFiles concatenate files of the same type.
func PutM4aTag(ctx context.Context, input, output string, metadata []string) error {
	f, err := newFfmpeg(ctx)
	if err != nil {
		return err
	}
	f.setInput(input)
	f.setArgs(metadata...)
	f.setArgs("-c", "copy")
	// TODO: Collect log
	return f.run(output)
}

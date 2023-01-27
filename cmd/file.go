package cmd

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

func readFileToChan(ctx context.Context, file *os.File, c chan string) {
	scanner := bufio.NewScanner(file)
	logrus.Info("starting to read addresses")
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			logrus.Info("stopping to read inputs")
			return
		default:
			c <- scanner.Text()
		}
	}
	logrus.Info("finished reading input file")
}

func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}

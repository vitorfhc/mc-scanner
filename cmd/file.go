package cmd

import (
	"bufio"
	"context"
	"os"

	"github.com/sirupsen/logrus"
)

func readFileToChan(ctx context.Context, file *os.File, c chan string) {
	scanner := bufio.NewScanner(file)
	logrus.Info("starting to read addresses")
	for scanner.Scan() {
		if _, ok := <-ctx.Done(); !ok {
			return
		}
		c <- scanner.Text()
	}
	logrus.Info("finished reading input file")
}

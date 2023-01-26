package main

import (
	"context"
	"fmt"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if len(ctx.Done()) > 0 {
		_, ok := <-ctx.Done()
		fmt.Println(ok)
	}
	cancel()
	if len(ctx.Done()) > 0 {
		_, ok := <-ctx.Done()
		fmt.Println(ok)
	}
}

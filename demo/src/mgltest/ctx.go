package main

import "context"

var (
	gCtx    context.Context
	gCancel context.CancelFunc
)

func init() {
	gCtx, gCancel = context.WithCancel(context.Background())
}

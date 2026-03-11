package controllers

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/CodeSyncr/nimbus/http"
)

var counterState struct {
	mu    sync.Mutex
	value int
	at    time.Time
}

// Counter controller for the counter demo.
type Counter struct{}

func (c *Counter) Index(ctx *http.Context) error {
	counterState.mu.Lock()
	v := counterState.value
	counterState.mu.Unlock()

	return ctx.View("apps/counter/index", map[string]any{
		"title": "Counter",
		"count": v,
	})
}

func (c *Counter) Increment(ctx *http.Context) error {
	fmt.Println("[counter] Increment called")
	counterState.mu.Lock()
	counterState.value++
	counterState.at = time.Now()
	counterState.mu.Unlock()

	ctx.Redirect(http.StatusFound, "/demos/counter")
	return nil
}

func (c *Counter) Decrement(ctx *http.Context) error {
	counterState.mu.Lock()
	counterState.value--
	counterState.at = time.Now()
	counterState.mu.Unlock()

	ctx.Redirect(http.StatusFound, "/demos/counter")
	return nil
}

func (c *Counter) Set(ctx *http.Context) error {
	_ = ctx.Request.ParseForm()
	n, _ := strconv.Atoi(ctx.Request.FormValue("count"))

	counterState.mu.Lock()
	counterState.value = n
	counterState.at = time.Now()
	counterState.mu.Unlock()

	ctx.Redirect(http.StatusFound, "/demos/counter")
	return nil
}

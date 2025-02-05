package control

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

type Control struct {
	mu       sync.Mutex
	paused   bool
	exitChan chan struct{}
}

func NewController() *Control {
	return &Control{
		exitChan: make(chan struct{}),
	}
}

func (c *Control) Start() {
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			char, _, _ := reader.ReadRune()
			switch char {
			case ' ':
				c.mu.Lock()
				c.paused = !c.paused
				status := "暂停"
				if !c.paused {
					status = "恢复"
				}
				fmt.Printf("\n[系统] %s数据刷新\n", status)
				c.mu.Unlock()
			case 'q':
				close(c.exitChan)
				return
			}
		}
	}()
}

func (c *Control) ShouldPause() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.paused
}

func (c *Control) WaitExit() <-chan struct{} {
	return c.exitChan
}

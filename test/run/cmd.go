package run

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type Cmd struct {
	cmd   *exec.Cmd
	Name  string
	Args  []string
	Dir   string
	wait  WaitWriter
	Waits []chan struct{}
	done  bool
	Save  bool
}

func (c Cmd) GetOut() []byte {
	return c.wait.Out
}

func (c *Cmd) Start() {
	c.cmd = exec.Command(c.Name, c.Args...)
	if c.Dir != "" {
		c.cmd.Dir = c.Dir
	}
	c.wait = WaitWriter{
		Writer: os.Stdout,
		Save:   c.Save,
	}
	go func() {
		defer c.Stop()
		c.cmd.Stdout = &c.wait
		c.cmd.Stderr = os.Stderr
		err := c.cmd.Run()
		if err != nil && !jerr.HasError(err, "signal: terminated") {
			log.Printf("error running process (%s %s); %v", c.Name, strings.Join(c.Args, " "), err)
		}
		c.done = true
		for _, wait := range c.Waits {
			wait <- struct{}{}
		}
	}()
}

func (c *Cmd) WaitUntil(text string, timeout time.Duration) {
	var wait = Wait{
		Text:  text,
		Found: make(chan struct{}),
	}
	c.wait.Waits = append(c.wait.Waits, wait)
	select {
	case <-wait.Found:
	case <-time.NewTimer(timeout).C:
	}
	c.cmd.Stdout = os.Stdout
	return
}

func (c *Cmd) Wait(timeout time.Duration) {
	if c.done {
		return
	}
	var wait = make(chan struct{})
	c.Waits = append(c.Waits, wait)
	if timeout > 0 {
		select {
		case <-wait:
		case <-time.NewTimer(timeout).C:
		}
	} else {
		<-wait
	}
}

func (c *Cmd) IsRunning() bool {
	return c.cmd != nil && !c.done
}

func (c *Cmd) Stop() error {
	if c.cmd != nil && c.cmd.Process.Pid > 0 {
		err := c.cmd.Process.Signal(syscall.SIGTERM)
		if err != nil && !jerr.HasError(err, "os: process already finished") {
			return fmt.Errorf("error terminating process; %w", err)
		}
		c.Wait(0)
	}
	return nil
}

type Wait struct {
	Text  string
	Found chan struct{}
}

type WaitWriter struct {
	Writer io.Writer
	Waits  []Wait
	Save   bool
	Out    []byte
}

func (w *WaitWriter) Write(p []byte) (n int, err error) {
	if w.Save {
		w.Out = append(w.Out, p...)
		return len(p), nil
	}
	text := string(p)
	for i := 0; i < len(w.Waits); i++ {
		if strings.Contains(text, w.Waits[i].Text) {
			var found = w.Waits[i].Found
			go func() {
				found <- struct{}{}
			}()
			w.Waits = append(w.Waits[:i], w.Waits[i+1:]...)
			i--
		}
	}
	return w.Writer.Write(p)
}

package main

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type BackgroundProcess struct {
	prefix   string
	line     []string
	lock     sync.Mutex
	cmd      *exec.Cmd
	callback BackgroundProcessCallback
}

type BackgroundProcessCallback interface {
	Line(text string)
}

type Waitable interface {
	Wait() error
}

func NewBackgroundProcess(prefix string, line []string, restarting bool, cb BackgroundProcessCallback) (bp *BackgroundProcess, err error) {
	bp = &BackgroundProcess{
		prefix:   prefix,
		line:     line,
		callback: cb,
	}

	if restarting {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGCHLD)

		go func() {
			for {
				select {
				case <-sigs:
					bp.lock.Lock()
					r := bp.cmd.Wait()
					log.Printf("SigChild %v (Restarting)", r)
					bp.lock.Unlock()
					bp.Start()
				case <-time.After(10 * time.Second):
					log.Printf("10 sec without SIGCHLD.")
				}
			}
		}()
	}

	return

}

func (bp *BackgroundProcess) prepare() error {
	return nil
}

func (bp *BackgroundProcess) Start() error {
	bp.lock.Lock()

	defer bp.lock.Unlock()

	c := exec.Command(bp.line[0], bp.line[1:]...)
	if false {
		c.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
			Pgid:    0,
		}
	}

	soReader, err := c.StdoutPipe()
	if err != nil {
		return err
	}
	seReader, err := c.StderrPipe()
	if err != nil {
		return err
	}

	soScanner := bufio.NewScanner(soReader)
	go func() {
		for soScanner.Scan() {
			if bp.callback != nil {
				bp.callback.Line(soScanner.Text())
			}
			log.Printf("%s%s\n", bp.prefix, soScanner.Text())
		}

	}()

	seScanner := bufio.NewScanner(seReader)
	go func() {
		for seScanner.Scan() {
			if bp.callback != nil {
				bp.callback.Line(seScanner.Text())
			}
			log.Printf("%s%s\n", bp.prefix, seScanner.Text())
		}
	}()

	bp.cmd = c

	err = bp.cmd.Start()
	if err != nil {
		return err
	}
	return nil
}

func (bp *BackgroundProcess) Wait() error {
	err := bp.cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

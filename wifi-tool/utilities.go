package main

import (
	"bufio"
	"log"
	"os/exec"
	"syscall"
)

type BackgroundProcess struct {
	cmd      *exec.Cmd
	callback *BackgroundProcessCallback
}

type BackgroundProcessCallback interface {
	Line(text string)
}

type Waitable interface {
	Wait() error
}

func NewBackgroundProcess(prefix string, line []string, cb BackgroundProcessCallback) (bp *BackgroundProcess, err error) {
	c := exec.Command(line[0], line[1:]...)
	if false {
		c.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
			Pgid:    0,
		}
	}

	soReader, err := c.StdoutPipe()
	if err != nil {
		return
	}
	seReader, err := c.StderrPipe()
	if err != nil {
		return
	}

	soScanner := bufio.NewScanner(soReader)
	go func() {
		for soScanner.Scan() {
			if cb != nil {
				cb.Line(soScanner.Text())
			}
			log.Printf("%s%s\n", prefix, soScanner.Text())
		}
	}()

	seScanner := bufio.NewScanner(seReader)
	go func() {
		for seScanner.Scan() {
			if cb != nil {
				cb.Line(seScanner.Text())
			}
			log.Printf("%s%s\n", prefix, seScanner.Text())
		}
	}()

	bp = &BackgroundProcess{
		cmd: c,
	}

	return

}

func (bp *BackgroundProcess) Start() error {
	err := bp.cmd.Start()
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

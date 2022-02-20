package cmdpipe

import (
	"io"
	"os/exec"
)

type LinuxCommand struct {
	Name string
	Args []string
}

type LinuxCommandPipe struct {
	r        io.Reader
	w        io.Writer
	commands []LinuxCommand
}

func NewLinuxCommandPipe(r io.Reader, w io.Writer, cmd ...LinuxCommand) *LinuxCommandPipe {
	return &LinuxCommandPipe{
		r:        r,
		w:        w,
		commands: cmd,
	}
}

func (p *LinuxCommandPipe) Execute() error {
	cmdList := make([]*exec.Cmd, len(p.commands))
	for i := range p.commands {
		cmdList[i] = exec.Command(p.commands[i].Name, p.commands[i].Args...)
		if i == 0 {
			cmdList[i].Stdin = p.r
		}
		if i == len(p.commands)-1 {
			cmdList[i].Stdout = p.w
		}
		if i > 0 {
			pipe, err := cmdList[i-1].StdoutPipe()
			if err != nil {
				return err
			}
			cmdList[i].Stdin = pipe
		}
	}

	for i := range cmdList {
		err := cmdList[i].Start()
		if err != nil {
			return err
		}
	}

	for i := range cmdList {
		err := cmdList[i].Wait()
		if err != nil {
			return err
		}
	}

	return nil
}

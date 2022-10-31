package cmd

import (
	"errors"
	"os/exec"
)

func execComand(cmd string, args ...string) error {
	if b, err := exec.Command(cmd, args...).CombinedOutput(); err != nil {
		return errors.New(string(b))
	}
	return nil
}

func goGet(mod string) error {
	return execComand("go", "get", "-u", mod)
}

func goInit(mod string) error {
	return execComand("go", "mod", "init", mod)
}

func gitInit() error {
	return execComand("git", "init")
}

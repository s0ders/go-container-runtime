package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"slices"
	"syscall"
)

func main() {
	var exitCode int
	var containerErr error
	
	args := os.Args[1:]

	switch args[0] {
	case "run":
		exitCode, containerErr = run(args[1:]...)
	case "child":
		child(args[0], args[1:]...)
	default:
		panic("unknown command")
	}


	if containerErr != nil {
		log.Fatalf("container ran into an error: %s", containerErr)
	} else {
		os.Exit(exitCode)
	}

}

func run(args ...string) (int, error) {

	args = slices.Insert(args, 0, "child")

	// Fork
	cmd := exec.Command("/proc/self/exe", args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWNS | syscall.CLONE_NEWPID,
	}

	exitCode := 0

	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return 0, fmt.Errorf("error while executing command: %s", err)
		}
	}

	return exitCode, nil
}

func child(command string, args ...string) (int, error) {

	if err := syscall.Sethostname([]byte("container")); err != nil {
		return -1, fmt.Errorf("hostname change failed: %s", err)
	}

	if err := syscall.Mount("proc", "proc", "proc", 0, ""); err != nil {
		return -1, fmt.Errorf("proc failed to mount: %s", err)
	}

	if err := syscall.Chroot("./alpine_roots"); err != nil {
		return -1, fmt.Errorf("chroot failed: %s", err)
	}

	syscall.Chdir("/")

	exitCode := 0

	cmd := exec.Command(command, args...)

	if err := cmd.Run(); err != nil {
		if exitStatus, ok := err.(*exec.ExitError); ok {
			exitCode = exitStatus.ExitCode()
		} else {
			return -1, fmt.Errorf("container failed: %s", err)
		}
	}

	return exitCode, nil
}

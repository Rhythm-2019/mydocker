package main

import (
    "log"
    "os"
    "os/exec"
    "syscall"
)

func main() {
    cmd := exec.Command("/bin/sh")
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWPID,
    }
    // cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(1), Gid: uint32(1)}
    cmd.Stderr = os.Stderr
    cmd.Stdout = os.Stdout
    cmd.Stdin = os.Stdin

    if err := cmd.Run(); err != nil {
        log.Fatal(err)
    }
    println("aaa")
}

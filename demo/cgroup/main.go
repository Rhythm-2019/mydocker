package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "os/exec"
    "path"
    "strconv"
    "syscall"
)

const SysCgroupMemoryHierarchuMount = "/sys/fs/cgroup/memory"

func main() {
    if os.Args[0] == "/proc/self/exe" {
        // Other process
        fmt.Printf("other process pid is %d\n", syscall.Getpid())

        cmd := exec.Command("/bin/sh", "-c", "stress --vm-bytes 200m --vm-keep -m 1")
        cmd.Stderr = os.Stderr
        cmd.Stdout = os.Stdout
        //cmd.Stdin = os.Stdin

        if err := cmd.Run(); err != nil {
            log.Fatal(err)
        }
        fmt.Printf("stress finish !\n")
        os.Exit(0)
    }

    // Main process
    cmd := exec.Command("/proc/self/exe")
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWNS | syscall.CLONE_NEWPID,
    }
    cmd.Stderr = os.Stderr
    cmd.Stdout = os.Stdout
    //cmd.Stdin = os.Stdin

    // limit resource
    if err := cmd.Run(); err != nil {
        log.Fatal(err)
    }
    os.Mkdir(path.Join(SysCgroupMemoryHierarchuMount, "testMemoryLimit"), 0755)
    ioutil.WriteFile(path.Join(SysCgroupMemoryHierarchuMount, "testMemoryLimit", "tasks"),
        []byte(strconv.Itoa(cmd.Process.Pid)), 0644)
    ioutil.WriteFile(path.Join(SysCgroupMemoryHierarchuMount, "testMemoryLimit", "memory_limit_in_bytes"),
        []byte("100m"), 0644)

    cmd.Process.Wait()
}

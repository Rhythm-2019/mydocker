package main

import (
	"fmt"
	"github.com/rhythm_2019/mydocker/container/cgroup"
	"github.com/rhythm_2019/mydocker/container/cgroup/limiter"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

var VERSAION = "<UNDEFINED>"

func main() {
	app := cli.NewApp()
	app.Name = "mydocker"
	app.Usage = "a simple container runtime implementation like docker"

	app.Commands = []*cli.Command{
		{
			Name:  "version",
			Usage: "show mydocker version",
			Action: func(context *cli.Context) error {
				fmt.Println(VERSAION)
				return nil
			},
		},
		{
			Name:  "init",
			Usage: "initial container",
			Action: func(context *cli.Context) error {
				// TODO init logic
				log.Debugf("init a process[%d]", syscall.Getpid())

				cmd := context.Args().Get(0)

				// mount proc
				executable, _ := os.Executable()
				mountDir := filepath.Join(filepath.Dir(executable), "proc")
				if _, err := os.Stat(mountDir); os.IsNotExist(err) {
					if err := os.Mkdir(mountDir, 0775); err != nil {
						log.Fatal("mkdir:", err)
					}
				}
				flags := syscall.MS_NOEXEC | syscall.MS_NODEV | syscall.MS_NOSUID
				//	syscall.MS_PRIVATE | syscall.MS_REC
				//flags := syscall.MS_PRIVATE
				if err := syscall.Mount("proc", mountDir, "proc", uintptr(flags), ""); err != nil {
					log.Fatal("mount:", err)
				}
				log.Infof("run a command %s", cmd)
				cmdAndArgs := strings.Split(cmd, " ")
				if err := syscall.Exec(cmdAndArgs[0], cmdAndArgs, os.Environ()); err != nil {
					log.Fatal("exec:", err)
				}
				return nil
			},
		},
		{
			Name:  "run",
			Usage: "run a image",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "it",
					Usage: "enable a tty",
				},
				&cli.StringFlag{
					Name:    "memory",
					Aliases: []string{"m"},
					Usage:   "set memory limit (MB)",
				},
				&cli.StringFlag{
					Name:  "cpu-quote",
					Usage: "set cpu quote limit (cpu period is 1000us)",
				},
			},
			Action: func(context *cli.Context) error {
				// run a command
				args := context.Args()
				if args.Len() < 1 {
					fmt.Println("run args must more than 1")
					os.Exit(1)
				}
				var resouseCfgs []*cgroup.CgroupConfig
				if len(context.String("memory")) != 0 {
					memoryLimitMB, err := strconv.Atoi(context.String("memory"))
					// TODO user may pass 100m
					if err != nil {
						fmt.Println("memory limit should be a integer 1")
						os.Exit(1)
					}
					resouseCfgs = append(resouseCfgs, &cgroup.CgroupConfig{
						Limiter: limiter.NewLimiter(limiter.MEMORY_USAGE_IN_BYTES),
						Value:   strconv.Itoa(memoryLimitMB << 20),
					})
				} else if (len(context.String("cpu-quote"))) != 0 {
					resouseCfgs = append(resouseCfgs, &cgroup.CgroupConfig{
						Limiter: limiter.NewLimiter(limiter.CPU_QUOTE),
						Value:   context.String("cpu-quote"),
					})
					resouseCfgs = append(resouseCfgs, &cgroup.CgroupConfig{
						Limiter: limiter.NewLimiter(limiter.CPU_PERIOD),
						Value:   "1000",
					})
				}
				Run(context.Args().First(), context.Bool("it"), resouseCfgs)
				return nil
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func Run(cmd string, tty bool, resourceCfgs []*cgroup.CgroupConfig) {
	// create a child process
	process := exec.Command("/proc/self/exe", "init", cmd)
	process.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWPID | syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
		Unshareflags: syscall.CLONE_NEWNS,
	}
	if tty {
		process.Stdin = os.Stdin
		process.Stdout = os.Stdout
		process.Stderr = os.Stderr
	}
	if err := process.Start(); err != nil {
		log.Fatal(err)
	}
	log.Infof("process %d is running", process.Process.Pid)

	// limit resource
	manager := cgroup.NewCgroupManager("mydocker-"+strconv.Itoa(process.Process.Pid), resourceCfgs)
	if err := manager.Apply(process.Process.Pid); err != nil {
		log.Fatal(err)
	}
	process.Wait()

	if err := manager.Destory(); err != nil {
		log.Fatal(err)
	}
	log.Infof("process %d is exit", process.Process.Pid)
}

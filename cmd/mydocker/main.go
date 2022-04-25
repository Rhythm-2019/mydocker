package main

import (
	"fmt"
	"github.com/rhythm_2019/mydocker/cfg"
	"github.com/rhythm_2019/mydocker/limiter"
	"github.com/rhythm_2019/mydocker/model"
	"github.com/rhythm_2019/mydocker/toolkit"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var Version = "<UNDEFINED>"

func main() {
	app := cli.NewApp()
	app.Name = "mydocker"
	app.Usage = "a simple container runtime implementation like docker"

	app.Before = func(context *cli.Context) error {
		log.SetLevel(log.DebugLevel)
		cfg.Init()
		model.Init()
		return nil
	}
	app.After = func(context *cli.Context) error {
		model.ReleaseDB()
		return nil
	}
	app.Commands = []*cli.Command{
		{
			Name:  "version",
			Usage: "show mydocker version",
			Action: func(context *cli.Context) error {
				fmt.Println(Version)
				return nil
			},
		},
		{
			Name:  "pull",
			Usage: "make a new image",
			Action: func(context *cli.Context) error {
				if context.Args().Len() < 1 {
					log.Fatalf("must pass a local image tar path")
				}
				if err := model.NewImage(context.Args().First()); err != nil {
					log.Fatalf("add a new image failed, detail is %v", err)
				}
				return nil
			},
		},
		{
			Name:  "images",
			Usage: "list all image",
			Action: func(context *cli.Context) error {
				records, err := model.ListImageRecord()
				if err != nil {
					log.Fatalf("list images failed, detail is %v", err)
				}
				fmt.Printf("ID\tNAME\tCREATE_TIME\n")
				for _, image := range records {
					fmt.Printf("%s\t%s\t%s\n", image.Id, image.Name, image.CreateTime)
				}
				return nil
			},
		},
		{
			Name:  "init",
			Usage: "initial container",
			Action: func(context *cli.Context) error {
				log.Debugf("init a process[%d]", syscall.Getpid())
				// 获取容器编号、
				fd := os.NewFile(uintptr(3), "pipe")
				containerId, err := ioutil.ReadAll(fd)
				if err != nil {
					log.Fatalf("init: read container id failed, detail is %v", err)
				}
				defer fd.Close()
				container, err := model.GetContainerRecord(string(containerId))
				if err != nil {
					log.Fatalf("init: get container record failed, detail is %v", err)
				}
				container.Init()
				container.Start()
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
				&cli.StringFlag{
					Name:  "name",
					Usage: "set a container name",
				},
				&cli.BoolFlag{
					Name:  "detach, d",
					Usage: "detach, run in background",
				},
				&cli.StringSliceFlag{
					Name:  "e",
					Usage: "set environment like key:value",
				},
				&cli.StringSliceFlag{
					Name:  "v",
					Usage: "make a bind mount",
				},
			},
			Action: func(context *cli.Context) error {
				args := context.Args()
				if args.Len() < 2 {
					log.Fatalf("run args must more than 2")
				}
				// image & cmd
				var imageName = args.Get(0)
				var command = args.Slice()[1:]

				image, err := model.FindImage(imageName)
				if err != nil {
					log.Fatal(err)
				}

				// tty
				var tty, detach = context.Bool("it"), context.Bool("d")
				if (tty && detach) || (!tty && !detach) {
					log.Fatalf("tty and detach just use one")
				}

				// resource
				var resouseCfgs []*model.CgroupConfig
				if len(context.String("memory")) != 0 {
					memoryLimitMB, err := strconv.Atoi(context.String("memory"))
					// TODO user may pass 100m
					if err != nil {
						log.Fatalf("memory limit should be a integer 1")
					}
					resouseCfgs = append(resouseCfgs, &model.CgroupConfig{
						Limiter: limiter.NewLimiter(limiter.MEMORY_USAGE_IN_BYTES),
						Value:   strconv.Itoa(memoryLimitMB << 20),
					})
				} else if (len(context.String("cpu-quote"))) != 0 {
					resouseCfgs = append(resouseCfgs, &model.CgroupConfig{
						Limiter: limiter.NewLimiter(limiter.CPU_QUOTE),
						Value:   context.String("cpu-quote"),
					})
					resouseCfgs = append(resouseCfgs, &model.CgroupConfig{
						Limiter: limiter.NewLimiter(limiter.CPU_PERIOD),
						Value:   "1000",
					})
				}

				// id and name
				containerId, containerName := toolkit.SnowflakeId(), context.String("name")
				if len(containerName) == 0 {
					containerName = containerId
				}

				// env
				envSlice := context.StringSlice("e")
				envKV := make(map[string]string)
				for _, kv := range envSlice {
					r, _ := regexp.Compile("(.+):(.+)")
					kvBytes := r.FindSubmatch([]byte(kv))
					if kvBytes == nil {
						continue
					}
					envKV[string(kvBytes[0])] = string(kvBytes[1])
				}

				// mount
				volumnArgs := context.StringSlice("v")
				var volumns []*model.Volumn
				for _, volumnArg := range volumnArgs {
					volumnArgsSliece := strings.Split(volumnArg, ":")
					var hostDIr, volumnDir string
					if len(volumnArgsSliece) == 1 && len(volumnArgsSliece[0]) != 0 {
						hostDIr = filepath.Join(cfg.StorePath(), containerId, "_data", volumnArgsSliece[0])
						volumnDir = volumnArgsSliece[0]
					} else if len(volumnArgsSliece) == 2 {
						hostDIr = volumnArgsSliece[0]
						volumnDir = volumnArgsSliece[1]
					}
					volumns = append(volumns, &model.Volumn{
						HostDir:   hostDIr,
						VolumnDIr: volumnDir,
					})
				}

				ctn := &model.Container{
					BaseModel: model.BaseModel{
						Id:         containerId,
						Name:       containerName,
						CreateTime: time.Now().Format("2006-01-02 15:04:05"),
					},
					Command:     command,
					Environment: envKV,
					Tty:         tty,
					Image:       image,
					LogFile:     filepath.Join(cfg.StorePath(), containerId, "container.log"),
					MergeDir:    filepath.Join(cfg.StorePath(), containerId, "merge"),
					UpperDir:    filepath.Join(cfg.StorePath(), containerId, "upper"),
					WorkDir:     filepath.Join(cfg.StorePath(), containerId, "work"),
					Volumns:     volumns,
				}
				model.NewCgroupManager(ctn, resouseCfgs)
				ctn.Bootstrap()
				return nil
			},
		},
		{
			Name: "rm",
			Action: func(context *cli.Context) error {
				if context.Args().Len() < 1 {
					log.Fatalf("must pass conatiner name or id")
				}
				input := context.Args().First()
				container, err := model.GetContainerRecord(input)
				if err != nil {
					container, err = model.GetContainerRecordByName(input)
					if err != nil {
						log.Fatalf("container record no found")
					}
				}
				if err = container.Delete(); err != nil {
					log.Fatal(err)
				}
				return nil
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

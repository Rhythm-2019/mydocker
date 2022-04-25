# mydocker 

本项目是作者参考《自动动手写 Docker》编写的类 Docker 工具。编写该项目的目的如下：
1. **了解 Docker 黑魔法**
2. 支持更多的网络模式
3. 支持远程容器管理
4. 支持容器指标、日志上报

当然啦目前是比不上 k8s 的 kubelet，作者会持续学习并迭代，打造一款可以在生产环境上使用的基础设施中的部件。


##  开发计划

1. 完成基础的 docker-cli 并支持下面命令
* images
* ps
* run
* exec
* rm
* rmi
* save
* network
* start
* stop

2. 开发 mydockerd 容器管理工具，支持下面功能
* 将 image 和 container 和 image 的信息保存到内存中
* 修改容器启动模型，先启动 pause，再启动 entrypoint，方便支持远程方式管理容器（包括网络的调整、流量控制等）
* 支持容器健康检查
* 支持指标上报、日志采集


##  开发进度
* 支持下面功能

##  Qucik start
安装 Go 环境，拉取仓库，运行下面指令
```bash
make install
./bin/mydocker
```

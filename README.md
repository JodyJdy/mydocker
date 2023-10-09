# mydocker

支持以下命令
* run        启动容器
* ps         列出所有容器
* logs       打印容器日志
* exec       在容器中执行命令
* stop       停止容器
* remove     删除容器
* buildBase  构建基础镜像
* images     展示镜像
* build      构建镜像
* network    创建容器网络
* portmap    管理端口映射
* save       保存容器为tar文件

## buildBase

容器启动需要一个镜像，该镜像要包含必要的linux的可执行文件，解压docker的busybox镜像，从中取出部分文件，打包成busybox.tar使用；
也可以使用alpine.tar,里面包含apk包管理工具，可以下载curl等工具

```shell
./mydocker  buildBase   busybox.tar
```
此时可以看到多另一个名称为base的镜像
```shell
./mydocker images

ID          NAME        VERSION     FROM        EXPOSE      CREATED
base        base                                []          2023-10-08 11:31:29
```
## network

用于创建网络/删除网络，支持的子命令有

* create  创建网络
* list    列出创建的网络
* remove  删除网络

### create 

支持的参数有

* --driver 只支持 bridge的网络驱动
* --subnet 指定子网信息

```shell
./mydocker network create --subnet 10.72.0.0/24 --driver bridge  testbridge
```
此时通过list能够查看到网络，使用ip段的第一个地址，作为网络的地址，也是linux bridge的地址
```shell
./mydocker network list
NAME         IpRange        Driver
testbridge   10.72.0.1/24   bridge
```
查看网卡信息
```shell
testbridge: flags=4099<UP,BROADCAST,MULTICAST>  mtu 1500
        inet 10.72.0.1  netmask 255.255.255.0  broadcast 10.72.0.255
        inet6 fe80::2c8b:f7ff:fe53:c420  prefixlen 64  scopeid 0x20<link>
        ether a2:92:54:e9:64:bd  txqueuelen 1000  (Ethernet)
        RX packets 0  bytes 0 (0.0 B)
        RX errors 0  dropped 0  overruns 0  frame 0
        TX packets 0  bytes 0 (0.0 B)
        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0

```

## portmap

端口转发并没有使用linux防火墙来实现，wsl测试未生效；手动实现了一个端口转发

用于实现端口转发，支持的命令有

* start    启动端口映射服务
* forward  手动配置的端口转发

## run 
run 命令支持的参数有

* -ti  交互式启动容器， 关闭tty，容器也会关闭 
* -m 设置容器的内存限制，例如:   -m 100m   限制内存为100m
* -cpushare 设置cpu时间片权重， 例如:  --cpushare 510
*  -cpuset 设置cpu核心数，例如:  --cpuset 2
* -v 挂载volume，可挂载多个
* -d    后台运行进程
* -name 容器名称  container name
* -e 设置环境变量
* -image 镜像id前缀 或者 镜像名称
* -net 指定容器所属的网络
  有四种模式:
    *  不配置  容器无网络信息
    * host 和宿主机共享网络  
    * container:容器标识   和指定容器共享网络  
    * bridgeName 指定网络
* -p 配置端口映射 
* -resolv  配置域名解析文件，默认是宿主机上面的

启动一个交互式进程
```shell
./mydocker run -ti -image base  sh
```
启动一个后台进程
```shell
./mydocker run -d -image base  top
```
启动带有网络的进程
```shell
./mydocker run -ti -image base -net testbridge sh

查看ip地址
# ifconfig
cif-64229 Link encap:Ethernet  HWaddr 06:DC:E7:76:89:5E
          inet addr:10.72.0.2  Bcast:10.72.0.255  Mask:255.255.255.0
          inet6 addr: fe80::4dc:e7ff:fe76:895e/64 Scope:Link
          UP BROADCAST RUNNING MULTICAST  MTU:1500  Metric:1
          RX packets:8 errors:0 dropped:0 overruns:0 frame:0
          TX packets:6 errors:0 dropped:0 overruns:0 carrier:0
          collisions:0 txqueuelen:1000
          RX bytes:736 (736.0 B)  TX bytes:516 (516.0 B)

lo        Link encap:Local Loopback
          inet addr:127.0.0.1  Mask:255.0.0.0
          inet6 addr: ::1/128 Scope:Host
          UP LOOPBACK RUNNING  MTU:65536  Metric:1
          RX packets:0 errors:0 dropped:0 overruns:0 frame:0
          TX packets:0 errors:0 dropped:0 overruns:0 carrier:0
          collisions:0 txqueuelen:1000
          RX bytes:0 (0.0 B)  TX bytes:0 (0.0 B)
          
          
容器中使用nc 监听一个端口
nc -lp 80

宿主机上面使用:
telnet  容器ip  80

可以连接成功

```

启动带有端口转发的进程
```shell
./mydocker run -ti -image base -net testbridge -p 3307:3307  sh
 # ifconfig
cif-21714 Link encap:Ethernet  HWaddr 46:95:42:D8:40:01
          inet addr:10.72.0.4  Bcast:10.72.0.255  Mask:255.255.255.0
          inet6 addr: fe80::4495:42ff:fed8:4001/64 Scope:Link
          UP BROADCAST RUNNING MULTICAST  MTU:1500  Metric:1
          RX packets:10 errors:0 dropped:0 overruns:0 frame:0
          TX packets:8 errors:0 dropped:0 overruns:0 carrier:0
          collisions:0 txqueuelen:1000
          RX bytes:876 (876.0 B)  TX bytes:656 (656.0 B)

lo        Link encap:Local Loopback
          inet addr:127.0.0.1  Mask:255.0.0.0
          inet6 addr: ::1/128 Scope:Host
          UP LOOPBACK RUNNING  MTU:65536  Metric:1
          RX packets:0 errors:0 dropped:0 overruns:0 frame:0
          TX packets:0 errors:0 dropped:0 overruns:0 carrier:0
          collisions:0 txqueuelen:1000
          RX bytes:0 (0.0 B)  TX bytes:0 (0.0 B)
          

容器中使用nc监听端口

nc -lp 3307

宿主机上面使用

telnet 宿主机ip  3307

发现会转发到 容器中去 （需要开启 portmap 服务）

```
启动带有卷挂载的进程
```shell
./mydocker run -ti -image base -v 宿主机目录:容器目录  sh
```
## exec

进入容器
```shell
./mydocker exec  容器id/容器名称  命令

```
## stop
停止容器
```shell
./mydocker stop 容器id/容器名称 
```

## remove

移除容器
```shell
./mydocker remove 容器id/容器名称
```

## save
容器打包成tar包
```shell
./mydocker save  -o 保存的文件名 -c  容器标识

```
## build

例如如下的dockerfile

```dockerfile
FROM base
RUN touch x.txt
RUN mkdir /home/jdy
WORKDIR /home/jdy
ENTRYPOINT sleep 99999
```

```shell
./mydocker build -f dockerfile -t xx:0.01
```
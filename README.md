# ACMBot

![Badge](https://img.shields.io/badge/OneBot-v11-black)
![Badge](https://img.shields.io/badge/go-%3E%3D1.20-30dff3?logo=go)

## 项目介绍
这是一个使用Go语言，基于onebot11协议开发的QQBot项目，主要提供比赛查询，个人信息查询，群友排行等功能

## TODO
### 个人信息展示
- [x] CodeForces | usage: `cf [username]`
- [x] CodeForces Rating曲线图 | Usage: `rating [username]`
- [ ] AtCoder
- [ ] NowCoder
### 近期比赛
- [x] CodeForces | usage: `近期cf`
- [x] AtCoder | usage: `近期比赛`
- [x] NowCoder | usage: `近期比赛`
- [x] Luogu | usage: `近期比赛`
### 其他
- [ ] 群内排行
- [ ] ...

## 如何运行

```shell
git clone https://github.com/YourSuzumiya/ACMBot
cd ACMBot
go mod tidy
go run ./main.go
```
第一次启动会自动生成配置文件，填写好相关内容之后启动即可正常运行

## 与NapCat联通

配置文件(onebot.json)内开启websocket，地址类似于`ws://localhost:8088`
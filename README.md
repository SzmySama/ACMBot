# hycBot

## 项目介绍
为了响应hyc关于解决广大JUFE学子在应对程序设计竞赛过程中需要进行的各种日常训练赛事中出现的容易漏掉比赛的问题的指示精神，hycBot项目应运而生，为广大学子提供了便利的查询服务，通过此项目，大家可以在QQ上一目了然的看到近日的训练安排，从而大幅提升训练效率

## 如何开始
首先请安装依赖

你至少需要安装 `git`, `python3.12.x`, `python-pdm` AKA `pdm`, `nb-cli`

在那之后 你可以开始本项目的本地部署

具体操作如下

```shell
git clone https://github.com/lordmystery/hello-world 
cd hello-world
pdm insatll # 安装依赖
```

然后就可以运行本项目了

```shell
# 假设您的当前目录在本项目内

# *nix
source .venv/bin/activate

# Windows
.venv/bin/activate

nb run
```

## 如何与消息发送端联通
以`NapCat`为例，
请开启`反向WebSocket`，请使用OneBotV11API，地址填写`ws://127.0.0.1:8080/onebot/v11/ws`

## TODO

- [ ] `洛谷` 
- [x] `AtCoder`
- [x] `CodeForces`
- [ ] `NowCoder`

## 配置
请在本项目下创建`.env`文件，添加`hycBot__Codeforces__secret`和`hycBot__Codeforces__key`，否则可能不能使用codeforces
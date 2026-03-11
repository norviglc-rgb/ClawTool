# clawtool

[中文](#中文) | [English](#english) | [日本語](#日本語)

## 中文

`clawtool` 是一个给 OpenClaw 用的命令行控制工具。它的目标很直接：

- 检查本机环境
- 管理本地和远程 profile
- 先规划，再执行
- 保留状态、备份和日志
- 出问题时可以验证、回滚、修复

当前已经可用的命令：

- `detect`
- `doctor`
- `init`
- `profile list/show/create/validate/use`
- `plan`
- `show`
- `apply`
- `verify`
- `inspect`
- `status`
- `logs`
- `rollback`
- `repair`
- `remote plan`
- `remote apply`
- `remote verify`
- `remote exec`

### 安装前需要什么

如果你只是想使用源码方式运行：

- Go `1.24` 或更高版本
- Git

如果你还想跑完整开发检查，另外需要：

- `make`
- PowerShell `7+`

### 如何安装

直接编译当前仓库：

```powershell
git clone https://github.com/norviglc-rgb/ClawTool.git
cd ClawTool
go build -o clawtool ./cmd/clawtool
```

也可以不编译，直接运行：

```powershell
go run ./cmd/clawtool --help
```

### 本地最常用的操作

先初始化工作区：

```powershell
go run ./cmd/clawtool init
```

查看环境和健康检查：

```powershell
go run ./cmd/clawtool detect
go run ./cmd/clawtool doctor
```

查看当前 profile，规划并执行：

```powershell
go run ./cmd/clawtool profile list
go run ./cmd/clawtool plan
go run ./cmd/clawtool apply
go run ./cmd/clawtool verify
```

查看状态、日志、回滚：

```powershell
go run ./cmd/clawtool inspect
go run ./cmd/clawtool status
go run ./cmd/clawtool logs --tail 20
go run ./cmd/clawtool rollback
```

### 如何创建远程 SSH profile

最简单的方式是直接创建：

```powershell
go run ./cmd/clawtool profile create remote-demo `
  --kind ssh `
  --address ssh.example.internal `
  --user deploy `
  --port 22 `
  --key-path C:\Users\you\.ssh\id_ed25519 `
  --host-key-strategy known_hosts
```

参数说明：

- `--address`: 远程主机地址
- `--user`: SSH 用户，不填时会尝试使用当前系统用户
- `--port`: SSH 端口，不填时默认 `22`
- `--key-path`: 私钥路径，不填时会尝试使用 `~/.ssh/` 下的默认私钥
- `--host-key-strategy`: 主机密钥校验策略，当前支持 `known_hosts` 和 `insecure`

### 远程操作怎么用

先看计划和预检查：

```powershell
go run ./cmd/clawtool remote plan remote-demo
go run ./cmd/clawtool remote apply remote-demo
go run ./cmd/clawtool remote verify remote-demo
```

执行一条远程命令：

```powershell
go run ./cmd/clawtool remote exec remote-demo uname -a
```

说明：

- `remote verify` 只做确定性的本地预检查，不会真的联网修改远程主机
- `remote apply` 会把当前 profile 渲染成远程配置，先备份旧配置，再写到远程主机
- `remote exec` 会真的发起 SSH 连接
- 如果远程命令返回非 0 退出码，`clawtool` 也会返回相同的退出码

### 开发时常用命令

```powershell
go test ./...
go vet ./...
make check
pwsh ./scripts/dev.ps1 check
```

## English

`clawtool` is a CLI control plane for OpenClaw. It helps you inspect the environment, manage profiles, plan changes, apply them safely, verify results, collect logs, roll back, and run deterministic repair steps.

Requirements:

- Go `1.24+`
- Git
- `make` and PowerShell `7+` for the full developer workflow

Build and run:

```bash
git clone https://github.com/norviglc-rgb/ClawTool.git
cd ClawTool
go build -o clawtool ./cmd/clawtool
./clawtool --help
```

Common local flow:

```bash
./clawtool init
./clawtool detect
./clawtool doctor
./clawtool plan
./clawtool apply
./clawtool verify
```

Remote SSH flow:

```bash
./clawtool profile create remote-demo \
  --kind ssh \
  --address ssh.example.internal \
  --user deploy \
  --port 22 \
  --key-path ~/.ssh/id_ed25519 \
  --host-key-strategy known_hosts

./clawtool remote plan remote-demo
./clawtool remote apply remote-demo
./clawtool remote verify remote-demo
./clawtool remote exec remote-demo uname -a
```

## 日本語

`clawtool` は OpenClaw 向けの CLI 制御ツールです。環境確認、profile 管理、plan、apply、verify、rollback、logs、repair、そして SSH ベースの基本的なリモート操作をまとめて扱えます。

必要なもの:

- Go `1.24` 以上
- Git
- 開発用には `make` と PowerShell `7+`

基本的な使い方:

```bash
go build -o clawtool ./cmd/clawtool
./clawtool init
./clawtool plan
./clawtool apply
./clawtool verify
```

SSH profile とリモート実行:

```bash
./clawtool profile create remote-demo \
  --kind ssh \
  --address ssh.example.internal \
  --user deploy \
  --port 22 \
  --key-path ~/.ssh/id_ed25519 \
  --host-key-strategy known_hosts

./clawtool remote verify remote-demo
./clawtool remote apply remote-demo
./clawtool remote exec remote-demo uname -a
```

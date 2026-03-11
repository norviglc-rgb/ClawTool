# clawtool

[中文](#中文) | [English](#english) | [日本語](#日本語)

## 中文

`clawtool` 是一个面向 OpenClaw 的 CLI-first 控制平面，用于标准化安装、配置、验证、回滚、日志与修复流程。

当前进度：

- 已完成 Phase 0 基础设施
- 已完成本地生命周期核心链路
- 已实现 `detect`、`doctor`、`init`、`profile`、`plan`、`show`
- 已实现 `apply`、`verify`、`inspect`、`status`
- 已实现 `logs`、`rollback`、`repair`

### 快速开始

```powershell
go run ./cmd/clawtool --help
go run ./cmd/clawtool init
go run ./cmd/clawtool plan
go run ./cmd/clawtool apply
go run ./cmd/clawtool verify --lang zh-CN
```

### 已实现命令

- `clawtool detect`
- `clawtool doctor`
- `clawtool init`
- `clawtool profile list`
- `clawtool profile show`
- `clawtool profile create`
- `clawtool profile validate`
- `clawtool profile use`
- `clawtool plan`
- `clawtool show`
- `clawtool apply`
- `clawtool verify`
- `clawtool inspect`
- `clawtool status`
- `clawtool logs`
- `clawtool rollback`
- `clawtool repair`

### 开发命令

```powershell
make check
pwsh ./scripts/dev.ps1 check
```

## English

`clawtool` is a CLI-first control plane for OpenClaw. It standardizes installation, configuration, verification, rollback, logging, and repair workflows.

Current status:

- Phase 0 foundation is complete
- The local lifecycle chain is in place
- `detect`, `doctor`, `init`, `profile`, `plan`, and `show` are implemented
- `apply`, `verify`, `inspect`, and `status` are implemented
- `logs`, `rollback`, and `repair` are implemented

### Quickstart

```bash
go run ./cmd/clawtool --help
go run ./cmd/clawtool init
go run ./cmd/clawtool plan
go run ./cmd/clawtool apply
go run ./cmd/clawtool verify --lang en
```

### Implemented Commands

- `clawtool detect`
- `clawtool doctor`
- `clawtool init`
- `clawtool profile list`
- `clawtool profile show`
- `clawtool profile create`
- `clawtool profile validate`
- `clawtool profile use`
- `clawtool plan`
- `clawtool show`
- `clawtool apply`
- `clawtool verify`
- `clawtool inspect`
- `clawtool status`
- `clawtool logs`
- `clawtool rollback`
- `clawtool repair`

### Developer Commands

```bash
make check
pwsh ./scripts/dev.ps1 check
```

## 日本語

`clawtool` は OpenClaw 向けの CLI-first コントロールプレーンです。インストール、設定、検証、ロールバック、ログ収集、修復フローを標準化します。

現在の進捗:

- Phase 0 の基盤は完了
- ローカルライフサイクルの主要チェーンは実装済み
- `detect`、`doctor`、`init`、`profile`、`plan`、`show` を実装済み
- `apply`、`verify`、`inspect`、`status` を実装済み
- `logs`、`rollback`、`repair` を実装済み

### クイックスタート

```bash
go run ./cmd/clawtool --help
go run ./cmd/clawtool init
go run ./cmd/clawtool plan
go run ./cmd/clawtool apply
go run ./cmd/clawtool verify --lang ja
```

### 実装済みコマンド

- `clawtool detect`
- `clawtool doctor`
- `clawtool init`
- `clawtool profile list`
- `clawtool profile show`
- `clawtool profile create`
- `clawtool profile validate`
- `clawtool profile use`
- `clawtool plan`
- `clawtool show`
- `clawtool apply`
- `clawtool verify`
- `clawtool inspect`
- `clawtool status`
- `clawtool logs`
- `clawtool rollback`
- `clawtool repair`

### 開発コマンド

```bash
make check
pwsh ./scripts/dev.ps1 check
```

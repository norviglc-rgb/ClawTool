# clawtool

[中文](#中文) | [English](#english) | [日本語](#日本語)

## 中文

`clawtool` 是一个面向 OpenClaw 的 CLI 优先控制平面，目标是标准化安装、配置、验证、诊断、回滚与修复流程。

当前仓库已经完成 Phase 0，并进入 Phase 1：

- 已接入 `cobra`、`go-i18n`、`yaml.v3`
- 已具备多语言 CLI 基础设施
- 已实现最小可用的 `detect`、`doctor`、`init`、`profile`
- 已实现基础闭环版本的 `plan`、`apply`、`verify`
- `plan` 已输出文件级差异动作，`apply` 已具备基础备份与幂等行为

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

### 开发命令

```powershell
make check
pwsh ./scripts/dev.ps1 check
```

### 代理说明

如果需要下载 Go 依赖，当前环境可使用：

- SOCKS5: `127.0.0.1:10808`
- HTTP/HTTPS: `127.0.0.1:10809`

## English

`clawtool` is a CLI-first control plane for OpenClaw. It standardizes installation, configuration, verification, diagnostics, rollback, and repair workflows.

Phase 0 is complete, Phase 1 is in place, and Phase 2 has started:

- `cobra`, `go-i18n`, and `yaml.v3` are wired in
- Multilingual CLI infrastructure is in place
- `detect`, `doctor`, `init`, and `profile` are minimally functional
- `plan`, `apply`, and `verify` now provide a basic local lifecycle loop
- `plan` now reports file-level and content-level diffs
- `plan --out` now saves a reusable plan artifact, and `show` can inspect it
- `inspect` and `status` now provide detailed and compact lifecycle summaries
- `logs` and `rollback` now provide the first operational completeness features

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
- `clawtool apply`
- `clawtool verify`
- `clawtool inspect`
- `clawtool status`
- `clawtool logs`
- `clawtool rollback`

### Developer Commands

```bash
make check
pwsh ./scripts/dev.ps1 check
```

### Proxy Notes

If Go dependencies need to be downloaded in this environment, the following local proxies are available:

- SOCKS5: `127.0.0.1:10808`
- HTTP/HTTPS: `127.0.0.1:10809`

## 日本語

`clawtool` は OpenClaw 向けの CLI ファーストなコントロールプレーンです。インストール、設定、検証、診断、ロールバック、修復のワークフローを標準化します。

現在は Phase 0 を完了し、Phase 1 に入っています。

- `cobra`、`go-i18n`、`yaml.v3` を導入済み
- 多言語 CLI 基盤を構築済み
- `detect`、`doctor`、`init`、`profile` は最小実装済み
- `plan`、`apply`、`verify` も基本的なローカルライフサイクルとして動作
- `plan` はファイル単位の変更動作を表示し、`apply` は基本的なバックアップと冪等性を備える

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
- `clawtool apply`
- `clawtool verify`

### 開発コマンド

```bash
make check
pwsh ./scripts/dev.ps1 check
```

### プロキシ情報

この環境で Go 依存関係をダウンロードする場合は、以下のローカルプロキシを利用できます。

- SOCKS5: `127.0.0.1:10808`
- HTTP/HTTPS: `127.0.0.1:10809`

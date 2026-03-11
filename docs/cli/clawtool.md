# clawtool CLI

[中文](#中文) | [English](#english) | [日本語](#日本語)

## 中文

### 当前命令

- `detect`
- `doctor`
- `init`
- `profile list`
- `profile show <name>`
- `profile create <name> [--kind local|ssh] [--address value]`
- `profile validate <name>`
- `profile use <name>`
- `plan`
- `show <plan-file>`
- `apply`
- `verify`
- `inspect`
- `status`
- `logs [--tail N] [--since RFC3339] [--bundle]`
- `rollback [backup-id]`

### 全局参数

- `--lang`
- `--json`

### 当前状态

- `detect`、`doctor`、`init`、`profile` 已有最小可用实现
- `plan`、`apply`、`verify` 已实现基础本地闭环
- `plan` 会显示 `create/update/noop`
- `apply` 会显示 `Changed` 与 `Backup Path`

## English

### Current Commands

- `detect`
- `doctor`
- `init`
- `profile list`
- `profile show <name>`
- `profile create <name> [--kind local|ssh] [--address value]`
- `profile validate <name>`
- `profile use <name>`
- `plan`
- `apply`
- `verify`
- `inspect`
- `status`

### Global Flags

- `--lang`
- `--json`

### Current Status

- `detect`, `doctor`, `init`, and `profile` have a minimal functional implementation
- `plan`, `apply`, and `verify` now provide a basic local lifecycle loop
- `plan` shows `create/update/noop` and field-level content diffs
- `plan --out` writes a saved plan artifact and `show` reads it back
- `apply` shows `Changed` and `Backup Path`
- `inspect` shows detailed lifecycle state including backup inventory
- `status` shows compact lifecycle state
- `logs` shows lifecycle entries and can create a support bundle zip
- `rollback` restores the latest or selected backup and verifies the restored state

## 日本語

### 現在のコマンド

- `detect`
- `doctor`
- `init`
- `profile list`
- `profile show <name>`
- `profile create <name> [--kind local|ssh] [--address value]`
- `profile validate <name>`
- `profile use <name>`
- `plan`
- `apply`
- `verify`

### グローバルフラグ

- `--lang`
- `--json`

### 現在の状態

- `detect`、`doctor`、`init`、`profile` は最小実装済み
- `plan`、`apply`、`verify` も基本的なローカルライフサイクルとして動作
- `plan` は `create/update/noop` を表示する
- `apply` は `Changed` と `Backup Path` を表示する

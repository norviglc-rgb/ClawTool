# Architecture

[中文](#中文) | [English](#english) | [日本語](#日本語)

## 中文

### 目标

`clawtool` 采用分层设计，确保业务逻辑、渲染、多语言、平台适配和自动化彼此解耦。

### 当前结构

- `cmd/clawtool`
  二进制入口。
- `internal/app`
  本地工作区业务逻辑，当前覆盖 `detect`、`doctor`、`init`、`profile`、`plan`、`apply`、`verify`。
- `internal/cli`
  Cobra 命令树、参数绑定和结果渲染入口。
- `internal/core`
  稳定类型、错误码和核心数据模型。
- `internal/i18n`
  语言解析和消息加载。
- `internal/render`
  人类可读输出与 JSON 输出。
- `internal/state`
  生命周期状态持久化。
- `internal/schema`
  配置与 manifest 校验。

### 当前阶段边界

- 已实现本地工作区初始化与配置文件管理
- 已实现基础环境探测和确定性健康检查
- 已实现基础版 `plan/apply/verify` 执行路径：
  基于活动 profile 生成有效配置、记录计划并更新状态
- `plan` 会产出确定性的文件动作，`apply` 会在覆盖前保存旧有效配置备份
- 远程执行与诊断增强仍保留在后续阶段

## English

### Goal

`clawtool` uses a layered architecture so business logic, rendering, localization, platform adapters, and automation remain decoupled.

### Current Layout

- `cmd/clawtool`
  Binary entry point.
- `internal/app`
  Local workspace business logic, currently covering `detect`, `doctor`, `init`, `profile`, `plan`, `apply`, `verify`, `inspect`, `status`, `logs`, and `rollback`.
- `internal/cli`
  Cobra command tree, flag binding, and result presentation entry points.
- `internal/core`
  Stable types, error codes, and core data models.
- `internal/i18n`
  Locale resolution and message loading.
- `internal/render`
  Human-readable and JSON presentation.
- `internal/state`
  Lifecycle state persistence.
- `internal/logs`
  Structured lifecycle log storage and retrieval.
- `internal/schema`
  Profile and manifest validation.

### Current Scope Boundary

- Local workspace initialization and profile management are implemented
- Basic environment detection and deterministic health checks are implemented
- A basic `plan/apply/verify` execution path now exists:
  it resolves the active profile, renders effective config, records plan output, and updates state
- `plan` produces deterministic file actions, and `apply` preserves previous effective config as a backup before overwrite
- `logs` reads structured lifecycle events and can bundle logs with current state summaries
- `rollback` restores the latest or selected backup, creates a pre-rollback backup, updates state, and runs verification
- Remote execution and richer diagnostics remain for later phases

## 日本語

### 目的

`clawtool` は、ビジネスロジック、レンダリング、多言語化、プラットフォーム適応、自動化を疎結合に保つためにレイヤード構成を採用します。

### 現在の構成

- `cmd/clawtool`
  バイナリエントリポイント。
- `internal/app`
  ローカルワークスペースの業務ロジック。現在は `detect`、`doctor`、`init`、`profile`、`plan`、`apply`、`verify` を担当。
- `internal/cli`
  Cobra のコマンドツリー、フラグ定義、結果表示の入口。
- `internal/core`
  安定した型、エラーコード、コアデータモデル。
- `internal/i18n`
  ロケール解決とメッセージ読み込み。
- `internal/render`
  人間向け出力と JSON 出力。
- `internal/state`
  ライフサイクル状態の永続化。
- `internal/schema`
  プロファイルおよび manifest の検証。

### 現在のスコープ境界

- ローカルワークスペース初期化とプロファイル管理は実装済み
- 基本的な環境検出と決定論的ヘルスチェックは実装済み
- 基本版の `plan/apply/verify` 実行経路も追加済み：
  アクティブプロファイルの解決、生成設定の出力、計画記録、状態更新を行う
- `plan` は決定論的なファイル変更動作を返し、`apply` は上書き前に旧生成設定をバックアップする
- リモート実行と診断拡張は後続フェーズで扱う

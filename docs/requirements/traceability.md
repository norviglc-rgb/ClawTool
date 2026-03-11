# Requirement Traceability

[中文](#中文) | [English](#english)

## 中文

这个文件用于把 `docs/requirements/clawtool_prd_v1.md` 的关键需求映射到当前实现状态。

状态说明：
- `done`：代码、测试、文档基本已具备
- `partial`：已有骨架或最小实现，但未达到 PRD 完整要求
- `missing`：尚未实现

| ID | Requirement | Status | Notes |
|---|---|---|---|
| F-001 | `detect` 环境探测 | partial | 已有本地探测与 JSON 输出；尚未覆盖 backend 建议、WSL2 视角、更多 OpenClaw 事实。 |
| F-002 | `doctor` 编排 | partial | 已有本地 deterministic health checks；尚未编排官方 `openclaw doctor`。 |
| F-003 | 安装编排器 | missing | 当前仍未实现独立的安装编排命令链。 |
| F-004 | `plan/apply/state/rollback` | partial | 已有基础闭环；缺 saved-plan apply、`state_id`、更多 rollback 选项。 |
| F-005 | 动态 schema + patch | missing | 目前仅有基础 schema 校验，没有动态 schema/patch 引擎。 |
| F-006 | profile / inventory / target / group | partial | 仅有 profile；缺 inventory、target、group 模型和命令。 |
| F-007 | contexts / variants / dependencies | missing | 尚未实现。 |
| F-008 | verify 验收式校验 | partial | 已有 verify；缺 OpenClaw status/health/channels/plugins/skills/security 聚合。 |
| F-009 | security | missing | 尚未实现命令组。 |
| F-010 | secrets | missing | 尚未实现。 |
| F-011 | logs / diagnose / bundle | partial | `logs` 与 bundle 已有基础能力；`diagnose` 尚未实现，日志分类也未完善。 |
| F-012 | Windows-WSL2 | missing | 尚未按 PRD 模式建模。 |
| F-013 | presets | missing | 尚未实现。 |
| F-014 | i18n | partial | 已有多语言基础；需补 PRD 要求的 golden / key 检测。 |
| F-015 | 自动化与可集成性 | partial | 已有 `--json`、部分自动化；`--non-interactive`、退出码规范仍未补齐。 |

## English

This file maps key requirements from `docs/requirements/clawtool_prd_v1.md` to the current implementation status.

Status:
- `done`: code, tests, and docs are largely in place
- `partial`: skeleton or minimum implementation exists, but PRD coverage is incomplete
- `missing`: not implemented yet

| ID | Requirement | Status | Notes |
|---|---|---|---|
| F-001 | `detect` environment discovery | partial | Local detection and JSON output exist; backend recommendation, WSL2 view, and richer OpenClaw facts are still missing. |
| F-002 | `doctor` orchestration | partial | Local deterministic checks exist; official `openclaw doctor` orchestration is not implemented yet. |
| F-003 | install orchestrator | missing | No dedicated install orchestration command chain yet. |
| F-004 | `plan/apply/state/rollback` | partial | Basic lifecycle exists; saved-plan semantics, `state_id`, and richer rollback options are still missing. |
| F-005 | dynamic schema + patch | missing | Only basic schema validation exists today. |
| F-006 | profile / inventory / target / group | partial | Only profile exists; inventory, target, and group are still missing. |
| F-007 | contexts / variants / dependencies | missing | Not implemented yet. |
| F-008 | acceptance-style verify | partial | Verify exists; OpenClaw status/health/channels/plugins/skills/security aggregation is still missing. |
| F-009 | security | missing | Command group not implemented yet. |
| F-010 | secrets | missing | Not implemented yet. |
| F-011 | logs / diagnose / bundle | partial | Basic `logs` and bundle exist; `diagnose` and richer source classification are still missing. |
| F-012 | Windows-WSL2 | missing | Not modeled per the PRD yet. |
| F-013 | presets | missing | Not implemented yet. |
| F-014 | i18n | partial | Multilingual base exists; PRD-grade golden/key validation still needs work. |
| F-015 | automation and integrability | partial | `--json` and some automation exist; `--non-interactive`, exit-code policy, and stronger golden/CI contracts are still missing. |

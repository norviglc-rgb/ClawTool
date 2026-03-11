# Clawtool 需求说明书（PRD）

版本：v1.0  
状态：可直接用于研发排期 / Codex 实施  
适用阶段：v0.1 / v0.2 / v0.3  
文档语言：zh-CN  

---

## 1. 文档目的

本文档定义 `clawtool` 的产品定位、范围边界、核心数据模型、命令设计、自动化与测试要求。

本 PRD 的目标不是描述“另一个 OpenClaw CLI”，而是定义一个**位于 OpenClaw 官方 CLI 之上的控制面（control plane）**：

- 负责多环境编排
- 负责计划、状态、回滚与验收
- 负责配置抽象与目标管理
- 负责诊断包、规则诊断与 AI 辅助判断
- 在必要时调用 OpenClaw 官方 CLI 获取事实、执行修复、收集状态

本文档将作为 Codex 的直接开发依据。实现过程中，若与本文档冲突，以本文档为准。

---

## 2. 产品定位

### 2.1 一句话定位

`clawtool` 是 OpenClaw 的**控制面与编排层**，而不是 OpenClaw 官方 CLI 的替代品。

### 2.2 核心价值

`clawtool` 的差异化能力应集中在以下层面：

1. **Profile / Inventory / Target Group 管理**
2. **Plan / Apply / State / Rollback 闭环**
3. **动态配置建模与 patch / diff 能力**
4. **验收式 verify，而不是单点存活检查**
5. **诊断包、规则诊断、AI 辅助诊断**
6. **自动化友好：JSON 输出、非交互、稳定退出码、CI 集成**

### 2.3 明确边界

`clawtool` **不应重写** OpenClaw 官方 CLI 已具备且维护中的底层能力，例如：

- `doctor`
- `config get/set/unset`
- `gateway health/status/install/start/stop`
- `logs`
- `security audit`
- `channels status/logs`
- `plugins doctor`
- `skills check`

`clawtool` **可以并且应该**在合适的场景下调用官方 CLI，将其作为：

- 事实来源（source of truth）
- 底层执行器（executor）
- 规则诊断输入源（diagnostic input）
- AI 判断上下文来源（AI context provider）

---

## 3. 设计原则

### 3.1 不与官方 CLI 重叠，而是编排官方 CLI

默认策略：

- 优先**调用** OpenClaw 官方 CLI
- 在官方 CLI 无法满足“多目标编排 / 状态管理 / 工件化 plan / 脱敏诊断 / inventory”时，由 `clawtool` 自行补齐
- 只有在官方 CLI 无稳定接口或无法满足自动化要求时，才增加 `native` 实现

### 3.2 CLI-first

先实现 CLI，再考虑 GUI/TUI。未来 GUI 只作为 `clawtool` 的上层调用者，不得另起一套业务逻辑。

### 3.3 Automation-first

所有关键命令必须支持：

- `--json`
- `--non-interactive`
- 稳定退出码
- 机器可读错误对象
- 幂等执行

### 3.4 Safe-by-default

默认要求：

- plan 先行
- apply 默认只消费已保存的 plan 工件
- 默认备份
- 默认脱敏
- 默认 rootless / agentless 优先

### 3.5 i18n from day one

国际化不是后补功能。所有用户可见文本、错误消息、提示、表格标题都必须可本地化。

---

## 4. 背景与事实依据

以下结论已纳入需求边界：

1. OpenClaw 官方 CLI 当前已经覆盖 `doctor / config / gateway / logs / security / channels / plugins / skills / status / health` 等大量底层能力，因此 `clawtool` 不应重复建设同一层。  
2. `openclaw doctor` 已经是修复 + 迁移工具，支持 `--repair`、`--force`、`--non-interactive`、`--deep` 等自动化与修复模式。  
3. OpenClaw 官方配置体系是 schema 驱动的，支持 `config.patch`、`config set/unset`，插件也要求声明 `configSchema`。  
4. OpenClaw 在 Windows 上的推荐运行方式是 WSL2，CLI 与 Gateway 运行在 Linux 内。  
5. `openclaw security audit` 已提供安全审计与可选修复基础。  
6. `openclaw status`、`openclaw health --json`、`channels status`、`plugins doctor`、`skills check` 可作为 verify 与 diagnose 的事实输入。  
7. OpenTofu 的 saved plan / apply saved plan / JSON 输出 / state 思路适合借鉴到 `clawtool` 的计划与状态模型。  
8. chezmoi 的 secrets 加密、单二进制、rootless 设计适合借鉴到本项目。  
9. Bolt 的 inventory/group/transport 模型，以及 Comtrya 的 contexts/variants/dependencies 模型，适合借鉴到 profile/inventory 体系。

---

## 5. 目标与非目标

### 5.1 产品目标

#### G1. 建立 OpenClaw 控制面
支持以声明式方式管理 OpenClaw 的安装编排、配置编排、状态、回滚和验收。

#### G2. 建立可审计的 plan/apply 工作流
实现 plan 工件化、状态持久化、变更可审计、apply 可重放。

#### G3. 建立可自动化的多目标模型
通过 inventory / target / group / profile / vars 支持本机与远程环境管理。

#### G4. 建立可执行的诊断体系
通过官方 CLI 输出、系统事实、配置 diff、日志与安全检查生成规则诊断与 AI 分析上下文。

#### G5. 建立工程化质量底座
从第一天开始支持 i18n、自动化测试、golden 测试、CI、回归验证。

### 5.2 非目标

以下内容不是 v0.1 的目标：

1. 重写 OpenClaw 的底层 `doctor`、`security audit`、`gateway` 管理逻辑
2. 自建完整 provider/resource DSL
3. 构建复杂 GUI 控制台
4. 无结构化输入前提下的 AI 自动修复闭环
5. 原生 Windows 深度支持（v0.1 仅以 WSL2 为正式支持路径）

---

## 6. 目标用户与场景

### 6.1 用户类型

1. **本机高级用户**：在个人电脑上稳定部署和维护 OpenClaw
2. **团队运维 / 工程师**：在多台服务器、边缘节点、开发机上统一管理 OpenClaw
3. **AI 自动化调用方**：将 `clawtool` 作为机器可读、可判断、可恢复的运维接口

### 6.2 核心场景

#### S1. 本机快速落地
用户希望在 macOS / Linux / Windows-WSL2 上安装并配置 OpenClaw，获得可回滚、可验收的结果。

#### S2. 多机统一管理
用户希望通过 inventory 管理多个目标，统一执行 plan / apply / verify / diagnose。

#### S3. 故障分析
用户希望在 OpenClaw 故障时，一键收集脱敏诊断包并输出可执行建议。

#### S4. AI 辅助运维
用户希望让 AI 基于结构化事实（doctor、health、status、config diff、logs）生成判断，而不是直接基于零散日志猜测。

---

## 7. 总体架构要求

### 7.1 分层架构

建议采用以下逻辑分层：

1. **CLI Layer**：命令解析、交互、输出格式化
2. **Application Layer**：命令编排、流程控制、状态管理
3. **Backend Adapter Layer**：适配 OpenClaw CLI / Linux toolkit / Native executor
4. **Config & Schema Layer**：profile、manifest、inventory、diff、patch、schema merge
5. **Execution Layer**：local / SSH 执行、文件同步、命令调用、超时与重试
6. **Diagnostics Layer**：日志聚合、规则诊断、AI 上下文构建、bundle 导出
7. **i18n Layer**：消息键、语言包、语言回退

### 7.2 后端适配器体系（必须实现）

定义统一 backend 接口，至少支持：

- `openclaw-cli`（v0.1 必做）
- `linux-toolkit`（v0.2 建议）
- `native`（预留，不要求 v0.1 完成）

#### Backend 统一能力契约

每个 backend 必须定义：

- `DetectCapabilities()`
- `Plan()`
- `Apply()`
- `VerifyInputs()`
- `CollectDiagnostics()`
- `Repair()`
- `RollbackSupport()`

#### 默认选择规则

- 本机已有 OpenClaw CLI 时，默认 `openclaw-cli`
- Linux 服务器预设场景可选 `linux-toolkit`
- 仅当后两者无法满足时使用 `native`

---

## 8. 功能需求

## 8.1 F-001 环境探测 detect

### 目标
识别运行环境、OpenClaw 可用性、backend 能力、自动化条件与风险点。

### 输入
- 本机或目标主机
- 可选 profile / target

### 输出
- 平台（macOS / Linux / windows-wsl2）
- shell / runtime / package manager 事实
- OpenClaw CLI 是否可用
- OpenClaw 版本
- Gateway 可达性
- 配置路径 / 状态路径 / 工作区路径
- systemd / launchd / WSL / SSH 能力
- backend 选择建议
- JSON 输出

### 说明
`detect` 是 `clawtool` 自有能力，不调用官方 CLI 也必须能工作；但若检测到 OpenClaw CLI，则应补充其版本、profile 与状态路径信息。

---

## 8.2 F-002 doctor 编排

### 目标
提供统一入口，编排 OpenClaw 官方 `doctor` 与本项目自有检查，形成结构化结果。

### 要求
- `clawtool doctor` 不直接重写官方诊断规则
- 优先调用 `openclaw doctor`
- 支持模式：
  - `check`
  - `migrate`
  - `repair safe`
  - `repair aggressive`
- 支持透传或映射：
  - `--yes`
  - `--non-interactive`
  - `--deep`
- 统一输出问题模型：
  - code
  - severity
  - source
  - summary
  - remediation
  - safe_to_auto_fix

### 实现要求
若官方 CLI 提供 `--json` 则优先解析 JSON；若无 JSON，则允许解析稳定文本模式，但必须封装解析器并加 golden 测试。

---

## 8.3 F-003 install 改为安装编排器

### 目标
`install` 负责“编排安装”，而不是自行重写 OpenClaw 安装器。

### 说明
默认流程：

1. detect
2. 选择 backend
3. 调用官方 `setup / onboard / gateway install` 或平台预设逻辑
4. 写入 `clawtool state`
5. 执行 verify

### 必须支持
- 本机安装
- 远程安装（通过 inventory/transport）
- 非交互模式
- 幂等重复执行

### 明确禁止
v0.1 不允许自行重新实现一套与官方 onboarding 完全重叠的安装流程。

---

## 8.4 F-004 plan / apply / state / rollback

### 目标
实现 `clawtool` 的核心控制面闭环。

### plan 能力（必须）
- `clawtool plan --out plan.json`
- `clawtool plan --json`
- `clawtool show plan.json`
- 产生 machine-readable 工件
- 生成字段级 diff 与动作图
- 标记风险级别
- 标记是否需要重启/重载
- 标记是否需要人工确认

### apply 能力（必须）
- `clawtool apply <plan-file>` 为默认路径
- `clawtool apply --auto-plan` 为显式非审计路径
- apply 前验证 plan 与环境兼容性
- apply 中自动备份
- apply 后更新 state

### state 能力（必须）
- 记录最近一次成功 apply 的目标状态摘要
- 记录 backend、profile、target、配置指纹、版本、时间戳
- 记录快照链

### rollback 能力（必须）
- 支持 `rollback --to <state-id>`
- 支持 `--dry-run`
- 支持 `--config-only`
- 支持 `--service-only`

---

## 8.5 F-005 动态 schema + patch 引擎

### 目标
建立基于 OpenClaw schema 的配置建模能力，而不是维护孤立静态 schema。

### 必须能力
- 获取或装载 OpenClaw 当前 config schema
- 合并插件 `configSchema`
- 支持 `set / unset / patch`
- 支持 JSON merge patch 语义
- 生成字段级 diff
- 支持 schema 版本识别
- 支持配置校验与错误定位

### 设计原则
- `clawtool` 内部可以维护补充 schema，但不能与 OpenClaw 官方 schema 脱节
- profile/manifest 的表达层可以是本项目自定义结构
- 最终落地到 OpenClaw 配置时，必须经过官方 schema 校验

---

## 8.6 F-006 profile / inventory / target / group 模型

### 目标
支撑多目标、多环境和配置继承。

### 必须数据结构

#### profile
至少包含：
- `name`
- `extends`
- `backend`
- `preset`
- `vars`
- `secretRefs`
- `policy`
- `targetSelector`

#### inventory
至少包含：
- `targets`
- `groups`
- `labels`
- `transport`
- `vars`
- `extends`

#### target
至少包含：
- `id`
- `address`
- `platform`
- `transport`
- `labels`
- `vars`

### 命令要求
- `clawtool target list`
- `clawtool target show <id>`
- `clawtool remote apply --target <id>`
- `clawtool remote apply --group <name>`

### 明确要求
不允许长期以 `--host` 单 flag 作为远程能力主入口。正式模型必须以 target/group 为核心。

---

## 8.7 F-007 contexts / variants / dependencies

### 目标
让计划不只是文件 diff，而是“动作图 + 条件分支 + 依赖图”。

### contexts
系统上下文至少应包括：
- OS
- architecture
- current user
- home dir
- shell
- WSL 状态
- systemd/launchd 可用性
- OpenClaw CLI 能力
- Node/Bun/pnpm 存在性

### variants
支持按条件切换：
- OS 变体
- backend 变体
- WSL 变体
- 安全级别变体

### dependencies
manifest/action 必须支持：
- `depends_on`
- `before`
- `after`

---

## 8.8 F-008 verify 验收式校验

### 目标
`verify` 不是“命令能不能执行”，而是一次完整的部署验收。

### 必须纳入的验收项
- 配置加载成功
- OpenClaw Gateway 健康
- `status` / `health` 正常
- `channels status` 正常或告警可解释
- `plugins doctor` 无阻断问题
- `skills check` 无阻断问题
- `security audit` 通过或已知告警可接受
- provider auth 状态正常
- drift 检测
- 是否需要重启/重载

### 输出要求
- 整体结果：PASS/WARN/FAIL
- 分项结果
- 建议动作
- JSON 报告

---

## 8.9 F-009 安全基线 security

### 目标
把 OpenClaw 官方安全审计纳入 `clawtool` 的 profile/preset/verify/diagnose 体系。

### 命令组
- `clawtool security audit`
- `clawtool security fix`
- `clawtool profile preset local-safe|shared-inbox-safe|public-gateway`

### 必须能力
- 调用 `openclaw security audit`
- 支持 `--deep`
- 支持 `--fix` 协调
- 统一输出安全问题模型
- 在 plan 阶段标注高风险配置

---

## 8.10 F-010 secrets 管理

### 目标
在 profile/vars 中安全引用敏感信息。

### 必须能力
- `secretRef`
- `secrets.enc.yaml`
- 支持 `age`
- 支持 `gpg`
- plan 输出自动脱敏
- diagnose bundle 自动脱敏
- 日志默认脱敏

### 约束
- 不允许将 secrets 明文写入 plan 工件
- 不允许将 secrets 明文写入 state
- 不允许将 secrets 原值出现在错误消息中

---

## 8.11 F-011 logs / diagnose / bundle

### 目标
把日志查看升级为支持运维与 AI 的问题收集系统。

### logs 必须能力
- `clawtool logs --tail`
- `clawtool logs --bundle`
- 统一来源分类：
  - clawtool
  - openclaw-cli
  - gateway
  - channels
  - system service
- 问题分类码
- 推荐下一步命令
- 默认脱敏

### diagnose 必须能力
- 汇总 detect / doctor / verify / security / logs / config diff / state
- 生成 bundle
- 输出规则诊断
- 生成 AI 可消费的上下文 JSON

### AI 辅助要求
AI 只基于结构化输入给建议，不直接获得未经脱敏的原始敏感内容。

---

## 8.12 F-012 Windows-WSL2 模式

### 目标
将 Windows 支持正式收敛为 `windows-wsl2` 模式。

### 必须能力
- 识别宿主 Windows 与 WSL 发行版
- 检查 WSL 版本
- 检查 systemd 状态
- 检查路径映射
- 检查端口与访问链路
- 输出 Windows 视角与 WSL 视角的事实

### 范围说明
v0.1 不承诺“原生 Windows 完全一致体验”。正式支持路径为 WSL2。

---

## 8.13 F-013 presets

### 目标
提供比手写 manifest 更易用的部署预设。

### 首批预设
- `local-dev`
- `laptop-safe`
- `server-nginx-tailscale`
- `remote-gateway`

### 预设要求
- 可被 profile 引用
- 可覆盖 vars
- 可输出 plan diff
- 可被 verify/diagnose 解释

---

## 8.14 F-014 i18n 国际化

### 目标
所有用户可见输出都可本地化，同时保持机器接口稳定。

### 必须语言
- `en-US`
- `zh-CN`

### 必须规则
1. 业务逻辑层不得硬编码用户可见字符串
2. 所有字符串必须使用 message key
3. CLI human output 可本地化
4. JSON 输出中的：
   - `code`
   - `field`
   - `severity`
   - `kind`
   必须稳定且不受 locale 影响
5. 错误对象必须同时包含：
   - 稳定 code
   - 默认英文 reason
   - 本地化 message
6. 缺失翻译必须回退到英文，不得输出空字符串

### 测试要求
- 缺失 key 检测
- 未使用 key 检测
- snapshot/golden 覆盖中英两套输出

---

## 8.15 F-015 自动化与可集成性

### 目标
将 `clawtool` 设计成 AI / CI / shell automation 均可安全调用的接口。

### 必须要求
- 所有核心命令支持 `--json`
- 所有核心命令支持 `--non-interactive`
- 退出码规范化
- 支持 `--plain` 或无 ANSI 模式
- 允许以保存 plan 的方式进入 CI 审核流
- 允许以 diagnose bundle 进入 AI 分析流

### 退出码建议
- `0`：成功
- `2`：检测到 drift / warn 但未失败
- `3`：verify 未通过
- `4`：plan 无法生成
- `5`：apply 失败
- `6`：backend 不可用
- `7`：安全基线失败
- `8`：配置 schema 校验失败
- `9`：目标不可达

---

## 9. 命令面设计

## 9.1 v0.1 必做命令

- `clawtool detect`
- `clawtool init`
- `clawtool profile`
- `clawtool target`
- `clawtool plan`
- `clawtool show`
- `clawtool apply`
- `clawtool verify`
- `clawtool status`
- `clawtool rollback`
- `clawtool logs`
- `clawtool diagnose`
- `clawtool security`

## 9.2 v0.2 建议补充

- `clawtool repair`
- `clawtool remote`
- `clawtool secrets`
- `clawtool backend`
- `clawtool doctor`（统一编排入口进一步增强）

---

## 10. 数据模型与目录结构要求

### 10.1 本地目录建议

```text
.clawtool/
  config.yaml
  inventory.yaml
  profiles/
  presets/
  plans/
  state/
  backups/
  bundles/
  secrets/
  cache/
```

### 10.2 plan 文件

plan 至少包含：
- metadata
- target selection
- backend
- input fingerprint
- variable resolution summary
- redacted diff
- action graph
- restart implications
- rollback hint
- created_at

### 10.3 state 文件

state 至少包含：
- state_id
- plan_id
- target(s)
- backend
- openclaw_version
- config_fingerprint
- artifact pointers
- apply result summary
- verify summary
- created_at

---

## 11. 非功能需求

### 11.1 平台支持

v0.1：
- macOS
- Linux
- Windows-WSL2

### 11.2 分发

- 首选单二进制分发
- rootless 优先
- agentless 优先

### 11.3 可靠性

- apply 幂等
- 出错保留中间工件
- 默认备份
- rollback 可审计

### 11.4 可观测性

- 结构化日志
- trace/span 预留
- bundle 可复现问题上下文

### 11.5 性能

- 常规 detect < 3s（本机，无网络）
- verify < 15s（单机，标准模式）
- plan 对单目标配置 diff 可在 5s 内完成（正常配置规模）

---

## 12. 自动化测试要求（必须严格执行）

### 12.1 测试金字塔

#### 单元测试（必须）
覆盖：
- profile/inventory 解析
- vars / secretRefs 解析
- schema merge
- patch/diff 引擎
- i18n message lookup
- error code mapping
- redaction
- backend capability selection

#### 集成测试（必须）
覆盖：
- 与 OpenClaw CLI 的进程交互
- `detect -> plan -> apply -> verify` 主链路
- `doctor` 编排与结果归一
- `status/health/security/channels/plugins/skills` 聚合
- saved plan -> apply saved plan

#### Golden / Snapshot 测试（必须）
覆盖：
- `--json` 输出
- 中英文 human 输出
- plan/show/verify/diagnose 报告
- 错误消息
- doctor/plain-output 解析器

#### E2E 测试（必须）
覆盖：
- macOS 本机场景
- Linux 本机场景
- Linux SSH 远程场景
- Windows-WSL2 场景（可先 nightly）

### 12.2 测试分层要求

- 所有文本解析器必须有 golden fixture
- 所有 CLI JSON 输出必须有 schema 校验测试
- 所有 i18n message key 必须检查双语覆盖率
- 所有 redaction 规则必须有反泄露测试
- 所有 rollback 流程必须有回放测试

### 12.3 Mock 与真实测试原则

- 对 OpenClaw CLI 交互可先用 fixture/mock 覆盖主要路径
- 但 v0.1 发布前必须至少保留一组真实 OpenClaw CLI 集成测试
- 真实测试必须覆盖至少一个 verify 成功路径和一个 diagnose 失败路径

---

## 13. CI / 自动化开发要求

### 13.1 必须提供的自动化命令

仓库必须提供统一入口，例如 `Makefile` 或 `Taskfile`：

- `make fmt`
- `make lint`
- `make test`
- `make test-unit`
- `make test-integration`
- `make test-golden`
- `make test-e2e`
- `make build`
- `make ci`

### 13.2 CI 门禁

PR 必须至少通过：
- build
- lint
- unit tests
- golden tests
- integration tests（可按变更范围分层触发）

### 13.3 变更规则

任何新增命令或输出变更，必须同步更新：
- i18n 语言包
- JSON schema / golden fixture
- 文档示例
- 测试用例

### 13.4 自动化开发要求

Codex 或其他自动化开发代理在提交实现时，必须：
- 优先复用 backend adapter，不得绕过分层直接散落调用 OpenClaw CLI
- 所有新输出先定义结构，再定义文案
- 所有命令先定义 exit code 和 JSON contract，再实现 human output
- 先补测试，再补实现，至少保证 golden 与单元测试覆盖

---

## 14. 版本规划

## 14.1 v0.1（必须交付）

范围：
1. `openclaw-cli` backend
2. detect
3. init
4. profile / inventory / target 基础模型
5. plan / show / apply / state / rollback
6. verify
7. security audit 聚合
8. logs / diagnose / bundle
9. i18n 基础设施
10. 单元 + 集成 + golden 测试

### v0.1 验收标准
- 单目标可完成 `detect -> plan -> apply(plan) -> verify`
- 能调用 OpenClaw 官方 CLI 聚合健康、安全、插件、渠道、skills 状态
- 生成 plan 工件和 state 快照
- 失败时可生成脱敏 bundle
- 中英文输出可切换
- 核心命令支持 `--json`

## 14.2 v0.2（建议）

范围：
1. remote/group apply
2. secrets 管理
3. doctor 编排增强
4. linux-toolkit backend
5. presets
6. repair 基础规则引擎

## 14.3 v0.3（建议）

范围：
1. native backend 的局部实现
2. 更深入的 drift 管理
3. AI 诊断增强
4. TUI / GUI 预研

---

## 15. 风险与约束

### R1. 官方 CLI 输出变更风险
若过度依赖文本解析，随着官方 CLI 文案调整会破坏兼容性。

**要求：** 优先使用 JSON 输出；如必须解析文本，需封装解析器并用 golden fixture 锁定。

### R2. schema 演进风险
OpenClaw schema 与插件 schema 可能持续演进。

**要求：** 引入 schema version、兼容层与快照测试。

### R3. WSL2 复杂度
Windows 支持路径必须清晰收敛，否则测试矩阵会迅速爆炸。

**要求：** 正式支持限定为 `windows-wsl2`，原生 Windows 仅标注未来方向。

### R4. AI 输出不可控
若直接让 AI 读取原始日志和敏感配置，可能造成误判或泄露。

**要求：** 先做规则归一与脱敏，再喂给 AI。

---

## 16. 与 OpenClaw 官方 CLI 的职责划分

### 16.1 官方 CLI 负责
- 实际 OpenClaw 配置读写与 schema 校验
- gateway 运行、健康、服务管理
- channels / plugins / skills / security 等域内能力
- doctor 修复与迁移

### 16.2 clawtool 负责
- 多目标抽象
- profile / inventory / presets
- plan/apply/state/rollback
- 统一 verify / diagnose 报告
- 诊断包与脱敏
- 自动化契约（JSON、exit code、saved plan）
- AI 可消费上下文构造

### 16.3 clawtool 调用官方 CLI 的原则

1. 优先调用官方 CLI 获取事实
2. 优先调用官方 CLI 执行域内修复
3. 对官方 CLI 的结果做归一、增强、聚合、脱敏
4. 非必要不复制底层逻辑

---

## 17. 研发实施建议（给 Codex）

### 17.1 推荐技术栈
- 语言：Go
- CLI：Cobra
- 配置：Viper（可选）
- 测试：Go test + golden fixtures
- JSON schema 校验：可选成熟库
- SSH：Go SSH 生态库

### 17.2 建议开发顺序
1. 定义错误码、JSON contract、i18n key 体系
2. 实现 backend adapter 接口与 `openclaw-cli` backend
3. 实现 detect / plan / show / apply / state
4. 实现 verify/security/logs/diagnose 聚合
5. 补齐 inventory/target/remote
6. 补齐 secrets / presets / repair

### 17.3 代码约束
- 不允许在业务逻辑层直接拼接用户可见字符串
- 不允许跨层直接 exec OpenClaw CLI
- 所有外部命令调用必须经 executor/backend 层
- 所有输出结构必须先定义 contract 再实现

---

## 18. 参考依据

以下资料是本 PRD 的设计参考：

- OpenClaw CLI 参考
- OpenClaw Doctor
- OpenClaw 配置系统
- OpenClaw Windows (WSL2)
- OpenClaw Security / Health / Channels / Plugins / Skills 文档
- OpenTofu plan / apply / JSON format / state 思路
- chezmoi encryption / single binary / rootless 思路
- Puppet Bolt inventory / transport 模型
- Comtrya contexts / variants 思路


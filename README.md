# Syzygy MCP - 朔望连珠范式

<div align="center">

**端到端测试固化工具 | E2E Test Crystallization Framework**

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![MCP Protocol](https://img.shields.io/badge/MCP-Model_Context_Protocol-blue)](https://modelcontextprotocol.io/)

[English](README.en.md) | **中文**

</div>

---

## <a id="chinese"></a>🌟 什么是 Syzygy？

**Syzygy（朔望连珠）** 是一个基于 MCP（Model Context Protocol）的端到端测试固化工具，专为 AI 辅助开发设计。它将 **UI 交互、网络请求、数据库状态** 三层证据固化为可复跑的规范（spec），实现真正的"三层对齐"验证。

### 核心理念

```
Define（定义） → Act（执行） → Observe（观察） → Align（对齐） → Crystallize（固化）
```

- **Define**：定义 UI 行为、API 预期、DB 状态
- **Act**：执行真实的浏览器操作
- **Observe**：捕获网络请求和数据库变更
- **Align**：验证三层证据是否对齐
- **Crystallize**：固化为可复跑的 JSON spec

### 为什么会有这个工具？

| 传统 E2E 测试 | Syzygy 范式 |
|-------------|------------|
| ❌ 只验证 UI 表象 | ✅ 验证 UI + Net + DB 三层 |
| ❌ 难以调试失败原因 | ✅ 自动生成失败取证（截图/HTML/JSON） |
| ❌ 维护成本高 | ✅ AI 辅助生成和维护 |
| ❌ 无法感知代码变更影响 | ✅ 基于元数据的智能变更感知 |
| ❌ 测试与业务脱节 | ✅ Spec 即文档，文档即测试 |

---

## 🚀 快速开始

### 前置要求

- Node.js 18+
- MySQL 5.7+ (用于 DB 断言)
- AI 助手支持 MCP 协议 (如 Claude Code、Windsurf 等)

### 安装

```bash
# 1. 克隆仓库
git clone https://github.com/cookchen233/syzygy-mcp-go.git
cd syzygy-mcp-go

# 2. 编译 MCP 服务
go build -o bin/syzygy-mcp ./cmd/syzygy-mcp

# 3. 安装回放引擎 (Replay Engine)
cd runner-node
npm install
npx playwright install

# 4. 配置 MCP Host（以 Claude Code 为例）
# 编辑 ~/Library/Application Support/Claude/claude_desktop_config.json
```

### MCP 配置示例

```json
{
  "mcpServers": {
    "syzygy-mcp": {
      "command": "/path/to/syzygy-mcp-go/bin/syzygy-mcp",
      "env": {
        "SYZYGY_HOME": "/Users/<you>/.syzygy-mcp"
      }
    }
  }
}
```

说明：
- Syzygy MCP 会把**配置与项目元信息**存放在 `SYZYGY_HOME`（默认 `~/.syzygy-mcp`）
- 多项目通过 `project_key` 分区：
  - `~/.syzygy-mcp/projects/<project_key>/config.json`
  - `~/.syzygy-mcp/projects/<project_key>/units/<unit_id>.json`
- spec/截图等**资源文件**不建议放在 `SYZYGY_HOME`，应通过 `syzygy_project_init(artifacts_dir=...)` 指定

---

## 📖 使用示例

### 0. 初始化项目运行配置（强制）

在首次使用某个项目（`project_key`）前，必须先调用 `syzygy_project_init` 写入项目级运行配置（如 BASE_URL / MYSQL_* / artifacts 目录 / 回放引擎命令）。
后续 `syzygy_unit_start` 与 `syzygy_replay` 会强制检查该 `project_key` 是否已初始化。

### 1. 使用 AI 助手创建单元

在 AI 助手对话中：

```
请使用 Syzygy 范式固化"用户登录"功能：
1. 打开登录页
2. 填写手机号和密码
3. 点击登录按钮
4. 验证跳转到工作台
5. 验证数据库中 last_login_at 已更新
```

AI 助手会自动调用 Syzygy MCP 工具：
- `syzygy_unit_start` - 创建单元
- `syzygy_step_append` - 添加 UI 步骤
- `syzygy_dbcheck_append` - 添加 DB 断言
- `syzygy_crystallize` - 生成 spec.json

### 2. 生成的 Spec 示例

```json
{
  "unit_id": "user.login.v1",
  "title": "User login flow (UI + Net + DB)",
  "steps": [
    {"name": "goto login", "ui": {"op": "ui.goto", "url": "${base_url}/login"}},
    {"name": "fill mobile", "ui": {"op": "ui.fill", "selector": "input[name=mobile]", "value": "${mobile}"}},
    {"name": "fill password", "ui": {"op": "ui.fill", "selector": "input[name=password]", "value": "${password}"}},
    {"name": "click login", "ui": {"op": "ui.click", "selector": "button[type=submit]"}, 
     "net": {"must": [{"method": "POST", "url_contains": "/api/login", "expect_json": {"code": "0"}}]}}
  ],
  "db_checks": [
    {"sql": "SELECT last_login_at FROM users WHERE mobile = :mobile", 
     "assert": {"last_login_at": "not_null"}}
  ]
}
```

### 3. 复跑验证

```bash
# 方式 1：使用 AI 助手
# 在对话中：请回放 user.login.v1
```

---

## 🛠️ MCP 工具列表

| 工具 | 功能 | 参数 |
|------|------|------|
| `syzygy_project_init` | 初始化项目运行配置 | `project_key`, `env`, `runner_command`, `runner_dir`, `artifacts_dir` |
| `syzygy_unit_start` | 创建并开始一个单元 | `project_key`, `unit_id`, `title`, `env`, `variables` |
| `syzygy_step_append` | 追加单个步骤 | `project_key`, `unit_id`, `run_id`, `step` |
| `syzygy_steps_append_batch` | 批量追加步骤 | `project_key`, `unit_id`, `run_id`, `steps` |
| `syzygy_anchor_set` | 设置数据锚点 | `project_key`, `unit_id`, `run_id`, `key`, `value` |
| `syzygy_dbcheck_append` | 追加数据库断言 | `project_key`, `unit_id`, `run_id`, `db_check` |
| `syzygy_crystallize` | 生成固化产物 | `project_key`, `unit_id`, `run_id`, `template`, `output_dir` |
| `syzygy_replay` | 回放固化用例 | `project_key`, `unit_id`, `run_id`, `env`, `command` |
| `syzygy_selfcheck` | 自查单元合规性 | `project_key`, `unit_id`, `run_id` |
| `syzygy_unit_meta_set` | 设置单元元数据 | `project_key`, `unit_id`, `meta` |
| `syzygy_plan_impacted_units` | 规划受影响的单元 | `project_key`, `changed_files`, `changed_apis`, `changed_tables` |

### 🔍 syzygy_selfcheck 工具详解

**syzygy_selfcheck** 是强制合规性检查工具，用于验证单元是否完全符合 Syzygy 范式要求。

#### 检查项目
- ✅ **固化完成** - 验证 `syzygy_crystallize` 已执行
- ✅ **回放验证** - 验证 `syzygy_replay` 已执行且成功
- ✅ **三层对齐** - 验证 UI/Net/DB 三层验证完整
- ✅ **交付格式** - 验证元数据完整

#### 使用示例
```bash
# AI 助手自动调用（推荐）
syzygy_selfcheck(unit_id="user.login.v1", run_id="run_xxx")

# 返回结果示例
{
  "all_passed": true,
  "summary": "🟢 SYZYGY SELFCHECK PASSED - All checks completed successfully",
  "checks": [
    {"name": "crystallize_completed", "passed": true, "message": "Crystallize has been executed"},
    {"name": "replay_verified", "passed": true, "message": "✅ Replay verification successful"},
    {"name": "three_layer_alignment", "passed": true, "message": "Three-layer alignment check"},
    {"name": "delivery_format", "passed": true, "message": "Delivery format check"}
  ]
}
```

#### 强制调用顺序
```
1. syzygy_unit_start
2. syzygy_step_append(s)
3. syzygy_dbcheck_append(s)
4. syzygy_crystallize
5. syzygy_replay
6. syzygy_selfcheck ← 【强制步骤】
```

**注意**：`syzygy_selfcheck` 必须在所有开发完成后调用，只有返回 `all_passed: true` 才算完成 Syzygy 流程。

---

## 📁 项目结构

```
syzygy-mcp-go/
├── cmd/
│   └── syzygy-mcp/          # MCP 服务入口
├── internal/
│   ├── application/         # 应用层（服务、工具注册）
│   ├── domain/              # 领域层（单元、步骤、断言）
│   └── infrastructure/      # 基础设施层（文件存储）
├── runner-node/             # 回放引擎（Node.js + Playwright）
│   └── package.json
├── examples/                # 示例 spec 文件
└── README.md
```

---

## 🎯 最佳实践

### 1. Spec 命名规范

```
<module>.<action>.<version>.spec.json
```

示例：
- `user.login.v1.spec.json`
- `order.create.v2.spec.json`
- `product.update.v1.spec.json`

### 2. 添加元数据

```json
{
  "unit_id": "user.login.v1",
  "metadata": {
    "module": "auth",
    "affects": {
      "apis": ["/api/login"],
      "tables": ["users"],
      "ui_routes": ["/login"],
      "controllers": ["AuthController"],
      "views": ["Login.vue"]
    },
    "created_at": "2026-01-01",
    "last_verified": "2026-01-01"
  }
}
```

### 3. 变更感知

当代码变更时，使用元数据自动识别需要重跑的 spec：

```bash
# 检测 git diff 并推荐需要重跑的 spec
./check-affected-specs.sh

# 批量执行指定模块的 spec
./run-all-specs.sh auth
```

---

## 📄 开源协议

MIT License - 详见 [LICENSE](LICENSE) 文件

---

## 🤝 贡献指南

欢迎贡献！请随时提交 Pull Request。

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

---

## 📮 联系方式

- GitHub: [@cookchen233](https://github.com/cookchen233)
- Issues: [GitHub Issues](https://github.com/cookchen233/syzygy-mcp-go/issues)

---

## 🙏 致谢

- [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) - 支持 AI 工具集成的协议
- [Playwright](https://playwright.dev/) - 浏览器自动化框架
- [Go](https://go.dev/) - Go 编程语言

---

<div align="center">

**用 ❤️ 为 AI 辅助开发而生**

⭐ 如果觉得有帮助，请给个 Star！

</div>

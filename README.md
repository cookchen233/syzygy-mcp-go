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

- Go 1.22+
- Node.js 18+
- MySQL 5.7+ (用于 DB 断言)
- AI 助手支持 MCP 协议 (如 Claude Desktop、Windsurf 等)

### 安装

```bash
# 1. 克隆仓库
git clone https://github.com/cookchen233/syzygy-mcp-go.git
cd syzygy-mcp-go

# 2. 编译 MCP 服务
go build -o bin/syzygy-mcp ./cmd/syzygy-mcp

# 3. 安装 Node.js Runner
cd runner-node
npm install
npx playwright install

# 4. 配置 MCP Host（以 Claude Desktop 为例）
# 编辑 ~/Library/Application Support/Claude/claude_desktop_config.json
```

### MCP 配置示例

```json
{
  "mcpServers": {
    "syzygy-mcp": {
      "command": "/path/to/syzygy-mcp-go/bin/syzygy-mcp",
      "env": {
        "SYZYGY_DATA_DIR": "/path/to/your-project/syzygy-data",
        "SYZYGY_RUNNER_DIR": "/path/to/syzygy-mcp-go/runner-node"
      }
    }
  }
}
```

---

## 📖 使用示例

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
- `syzygy.unit_start` - 创建单元
- `syzygy.step_append` - 添加 UI 步骤
- `syzygy.dbcheck_append` - 添加 DB 断言
- `syzygy.crystallize` - 生成 spec.json

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

# 方式 2：直接命令行
BASE_URL='https://your-app.com' \
MYSQL_HOST='127.0.0.1' MYSQL_PORT='3306' \
MYSQL_USER='root' MYSQL_PASSWORD='password' MYSQL_DATABASE='mydb' \
HEADLESS='1' \
node ./runner-node/bin/syzygy-runner.js /path/to/user.login.v1.spec.json
```

---

## 🛠️ MCP 工具列表

| 工具 | 功能 | 参数 |
|------|------|------|
| `syzygy.unit_start` | 创建并开始一个单元 | `unit_id`, `title`, `env`, `variables` |
| `syzygy.step_append` | 追加单个步骤 | `unit_id`, `run_id`, `step` |
| `syzygy.steps_append_batch` | 批量追加步骤 | `unit_id`, `run_id`, `steps` |
| `syzygy.anchor_set` | 设置数据锚点 | `unit_id`, `run_id`, `key`, `value` |
| `syzygy.dbcheck_append` | 追加数据库断言 | `unit_id`, `run_id`, `db_check` |
| `syzygy.crystallize` | 生成固化产物 | `unit_id`, `run_id`, `template`, `output_dir` |
| `syzygy.replay` | 回放固化用例 | `unit_id`, `run_id`, `env`, `command` |
| `syzygy.unit_meta_set` | 设置单元元数据 | `unit_id`, `meta` |
| `syzygy.plan_impacted_units` | 规划受影响的单元 | `changed_files`, `changed_apis`, `changed_tables` |

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
├── runner-node/             # Node.js + Playwright 执行器
│   ├── bin/
│   │   └── syzygy-runner.js # 主执行器
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

## 🌍 <a id="english"></a>English Documentation

### What is Syzygy?

**Syzygy** is an MCP-based E2E test crystallization framework designed for AI-assisted development. It solidifies **UI interactions, network requests, and database states** into replayable specs, achieving true "three-layer alignment" verification.

### Quick Start

```bash
# 1. Clone repository
git clone https://github.com/cookchen233/syzygy-mcp-go.git

# 2. Build MCP service
go build -o bin/syzygy-mcp ./cmd/syzygy-mcp

# 3. Install Node.js runner
cd runner-node && npm install && npx playwright install

# 4. Configure MCP host (e.g., Claude Desktop)
# Edit ~/Library/Application Support/Claude/claude_desktop_config.json
```

### Core Concepts

- **Define**: Define UI behavior, API expectations, DB states
- **Act**: Execute real browser operations
- **Observe**: Capture network requests and database changes
- **Align**: Verify three-layer evidence alignment
- **Crystallize**: Solidify into replayable JSON specs

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## 📮 Contact

- GitHub: [@cookchen233](https://github.com/cookchen233)
- Issues: [GitHub Issues](https://github.com/cookchen233/syzygy-mcp-go/issues)

---

## 🙏 Acknowledgments

- [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) - The protocol that powers AI-tool integration
- [Playwright](https://playwright.dev/) - Browser automation framework
- [Go](https://go.dev/) - The Go programming language

---

<div align="center">

**Made with ❤️ for AI-assisted development**

⭐ Star this repo if you find it helpful!

</div>

# Syzygy MCP - E2E Test Crystallization Framework

<div align="center">

**AI-Powered End-to-End Testing | Three-Layer Alignment Verification**

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![MCP Protocol](https://img.shields.io/badge/MCP-Model_Context_Protocol-blue)](https://modelcontextprotocol.io/)

[中文](README.md) | **English**

</div>

---

## 🌟 What is Syzygy?

**Syzygy** is an MCP-based (Model Context Protocol) end-to-end test crystallization framework designed for AI-assisted development. It solidifies **UI interactions, network requests, and database states** into replayable specs, achieving true "three-layer alignment" verification.

### Core Philosophy

```
Define → Act → Observe → Align → Crystallize
```

- **Define**: Define UI behavior, API expectations, DB states
- **Act**: Execute real browser operations
- **Observe**: Capture network requests and database changes
- **Align**: Verify three-layer evidence alignment
- **Crystallize**: Solidify into replayable JSON specs

### Why This Tool Exists?

| Traditional E2E Testing | Syzygy Paradigm |
|------------------------|-----------------|
| ❌ Only verifies UI appearance | ✅ Verifies UI + Net + DB three layers |
| ❌ Hard to debug failures | ✅ Auto-generates failure artifacts (screenshots/HTML/JSON) |
| ❌ High maintenance cost | ✅ AI-assisted generation and maintenance |
| ❌ Cannot detect code change impact | ✅ Smart change detection based on metadata |
| ❌ Tests disconnected from business | ✅ Spec is documentation, documentation is test |

---

## 🚀 Quick Start

### Prerequisites

- Go 1.22+
- Node.js 18+
- MySQL 5.7+ (for DB assertions)
- AI assistant with MCP support (e.g., Claude Code, Windsurf)

### Installation

```bash
# 1. Clone repository
git clone https://github.com/cookchen233/syzygy-mcp-go.git
cd syzygy-mcp-go

# 2. Build MCP service
go build -o bin/syzygy-mcp ./cmd/syzygy-mcp

# 3. Install Node.js Runner
cd runner-node
npm install
npx playwright install

# 4. Configure MCP Host (e.g., Claude Code)
# Edit ~/Library/Application Support/Claude/claude_desktop_config.json
```

### MCP Configuration Example

```json
{
  "mcpServers": {
    "syzygy-mcp": {
      "command": "/path/to/syzygy-mcp-go/bin/syzygy-mcp",
      "env": {
        "SYZYGY_DATA_DIR": "/path/to/your-project/syzygy-data",
        "SYZYGY_RUNNER_DIR": "/path/to/syzygy-mcp-go/runner-node",
        "SYZYGY_ARTIFACTS_DIR": "/path/to/your-project/syzygy-artifacts"
      }
    }
  }
}
```

---

## 📖 Usage Examples

### 1. Create Unit with AI Assistant

In AI assistant conversation:

```
Please use Syzygy paradigm to crystallize "user login" feature:
1. Open login page
2. Fill in mobile and password
3. Click login button
4. Verify redirect to dashboard
5. Verify last_login_at updated in database
```

AI assistant will automatically call Syzygy MCP tools:
- `syzygy.unit_start` - Create unit
- `syzygy.step_append` - Add UI steps
- `syzygy.dbcheck_append` - Add DB assertions
- `syzygy.crystallize` - Generate spec.json

### 2. Generated Spec Example

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

### 3. Replay Verification

```bash
# Method 1: Use AI assistant
# In conversation: Please replay user.login.v1

# Method 2: Direct command line
BASE_URL='https://your-app.com' \
MYSQL_HOST='127.0.0.1' MYSQL_PORT='3306' \
MYSQL_USER='root' MYSQL_PASSWORD='password' MYSQL_DATABASE='mydb' \
HEADLESS='1' \
node ./runner-node/bin/syzygy-runner.js /path/to/user.login.v1.spec.json
```

---

## 🛠️ MCP Tools

| Tool | Function                        | Parameters |
|------|---------------------------------|------------|
| `syzygy.unit_start` | Create and start a unit         | `unit_id`, `title`, `env`, `variables` |
| `syzygy.step_append` | Append single step              | `unit_id`, `run_id`, `step` |
| `syzygy.steps_append_batch` | Batch append steps              | `unit_id`, `run_id`, `steps` |
| `syzygy.anchor_set` | Set data anchor                 | `unit_id`, `run_id`, `key`, `value` |
| `syzygy.dbcheck_append` | Append database assertion       | `unit_id`, `run_id`, `db_check` |
| `syzygy.crystallize` | Generate crystallized artifacts | `unit_id`, `run_id`, `template`, `output_dir` |
| `syzygy.replay` | Replay crystallized spec        | `unit_id`, `run_id`, `env`, `command` |
| `syzygy.selfcheck` | Self-check unit compliance      | `unit_id`, `run_id` |
| `syzygy.unit_meta_set` | Set unit metadata               | `unit_id`, `meta` |
| `syzygy.plan_impacted_units` | Plan impacted units             | `changed_files`, `changed_apis`, `changed_tables` |

### 🔍 syzygy.selfcheck Tool Details

**syzygy.selfcheck** is a mandatory compliance checking tool that validates whether a unit fully complies with Syzygy paradigm requirements.

#### Check Items
- ✅ **Crystallization Complete** - Verifies `syzygy_crystallize` has been executed
- ✅ **Replay Verified** - Verifies `syzygy_replay` has been executed and succeeded
- ✅ **Three-Layer Alignment** - Verifies UI/Net/DB three-layer verification is complete
- ✅ **Delivery Format** - Verifies metadata completeness

#### Usage Example
```bash
# AI auto-call (recommended)
syzygy_selfcheck(unit_id="user.login.v1", run_id="run_xxx")

# Return result example
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

#### Mandatory Call Order
```
1. syzygy_unit_start
2. syzygy_step_append(s)
3. syzygy_dbcheck_append(s)
4. syzygy_crystallize
5. syzygy_replay
6. syzygy_selfcheck ← 【Mandatory Step】
```

**Note**: `syzygy_selfcheck` must be called after all development is complete. Only when it returns `all_passed: true` is the Syzygy process considered complete.

---

## 📁 Project Structure

```
syzygy-mcp-go/
├── cmd/
│   └── syzygy-mcp/          # MCP service entry point
├── internal/
│   ├── application/         # Application layer (services, tool registry)
│   ├── domain/              # Domain layer (units, steps, assertions)
│   └── infrastructure/      # Infrastructure layer (file storage)
├── runner-node/             # Node.js + Playwright executor
│   ├── bin/
│   │   └── syzygy-runner.js # Main executor
│   └── package.json
├── examples/                # Example spec files
└── README.md
```

---

## 🎯 Best Practices

### 1. Spec Naming Convention

```
<module>.<action>.<version>.spec.json
```

Examples:
- `user.login.v1.spec.json`
- `order.create.v2.spec.json`
- `product.update.v1.spec.json`

### 2. Add Metadata

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

### 3. Change Detection

When code changes, use metadata to automatically identify specs that need rerunning:

```bash
# Detect git diff and recommend specs to rerun
./check-affected-specs.sh

# Batch execute specs for specific module
./run-all-specs.sh auth
```

### 4. Recommended Project Structure

```
your-project/
├── syzygy-specs/              # Spec files (committed to git)
│   ├── README.md
│   ├── check-affected-specs.sh
│   ├── run-all-specs.sh
│   └── *.spec.json
├── syzygy-data/               # Runtime data (gitignored)
│   └── units/
└── syzygy-artifacts/          # Failure artifacts (gitignored)
    └── <timestamp>/
        ├── screenshot.png
        ├── page.html
        └── error.json
```

---

## 🔧 Environment Variables

### For Crystallization

- `SYZYGY_DATA_DIR`: Directory for runtime unit data (default: `./syzygy-data`)
- `SYZYGY_RUNNER_DIR`: Path to runner-node directory
- `SYZYGY_ARTIFACTS_DIR`: Directory for artifacts output (default: `./syzygy-artifacts`)

### For Replay

- `BASE_URL`: Frontend application URL (e.g., `https://your-app.com`)
- `MYSQL_HOST`: MySQL host (required for DB assertions)
- `MYSQL_PORT`: MySQL port (default: 3306)
- `MYSQL_USER`: MySQL username
- `MYSQL_PASSWORD`: MySQL password
- `MYSQL_DATABASE`: MySQL database name
- `HEADLESS`: Browser headless mode (`0` for headed, `1` for headless)

---

## 🌐 Integration with CI/CD

### GitHub Actions Example

```yaml
name: Syzygy E2E Tests

on: [push, pull_request]

jobs:
  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '18'
      
      - name: Install dependencies
        run: |
          cd runner-node
          npm install
          npx playwright install --with-deps
      
      - name: Run all specs
        env:
          BASE_URL: ${{ secrets.BASE_URL }}
          MYSQL_HOST: ${{ secrets.MYSQL_HOST }}
          MYSQL_USER: ${{ secrets.MYSQL_USER }}
          MYSQL_PASSWORD: ${{ secrets.MYSQL_PASSWORD }}
          MYSQL_DATABASE: ${{ secrets.MYSQL_DATABASE }}
          HEADLESS: '1'
        run: |
          cd syzygy-specs
          ./run-all-specs.sh
```

---

## 📚 Advanced Features

### 1. Prerequisites Chain

Build dependent test scenarios:

```json
{
  "unit_id": "order.checkout.v1",
  "prerequisites": ["user.login.v1", "product.add_to_cart.v1"],
  "steps": [...]
}
```

### 2. Data Anchors

Extract dynamic values from responses:

```json
{
  "name": "create order",
  "ui": {"op": "ui.click", "selector": "button.checkout"},
  "net": {
    "must": [{
      "method": "POST",
      "url_contains": "/api/orders",
      "expect_json": {"code": "0"},
      "anchors": {"order_id": "data.order_id"}
    }]
  }
}
```

Use anchor in subsequent steps:

```json
{
  "name": "verify order",
  "db_check": {
    "sql": "SELECT * FROM orders WHERE id = :order_id",
    "params": {"order_id": "${order_id}"}
  }
}
```

### 3. Failure Artifacts

When replay fails, automatically generates:
- `screenshot.png` - Page screenshot at failure point
- `page.html` - Full page HTML
- `error.json` - Error details and stack trace

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

### Development Guidelines

- Follow Go best practices and conventions
- Add tests for new features
- Update documentation accordingly
- Ensure all tests pass before submitting PR

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

## 📖 Related Resources

- [MCP Specification](https://spec.modelcontextprotocol.io/)
- [Playwright Documentation](https://playwright.dev/docs/intro)
- [Three-Layer Alignment Testing Pattern](https://github.com/cookchen233/syzygy-mcp-go/wiki)

---

<div align="center">

**Made with ❤️ for AI-assisted development**

⭐ Star this repo if you find it helpful!

[Report Bug](https://github.com/cookchen233/syzygy-mcp-go/issues) · [Request Feature](https://github.com/cookchen233/syzygy-mcp-go/issues)

</div>

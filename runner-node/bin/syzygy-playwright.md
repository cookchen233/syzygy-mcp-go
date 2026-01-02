# Syzygy Playwright (Playwright Proxy)

`syzygy-playwright.js` 是一个基于 `syzygy-runner` 逻辑衍生的轻量级代理工具，专门用于解决原生 Playwright 在 H5 (Uni-app) 或复杂 SPA 环境下的操作稳定性问题。

## 核心特性

1.  **稳健点击 (Robust Click)**: 自动处理组件识别偏差。如果原生 `page.click()` 失败或超时，自动回退到 `page.evaluate()` 执行 JS 点击。
2.  **SPA 等待增强**: 提供 `page.waitReady(selector)` 补丁，针对 Hash 路由和异步渲染提供更可靠的可见性检查。
3.  **移动端预设**: 内置 iPhone 12 Pro 模拟配置（视口、UserAgent、触摸事件）。
4.  **无缝集成**: 接受 `AGENT_CODE` 环境变量，直接执行 AI 生成的脚本片段。

## 使用方式

### 1. 命令行调用
```bash
AGENT_CODE='await page.goto("..."); await page.click(".login-btn");' \
MOBILE=1 \
node ./runner-node/bin/syzygy-playwright.js
```

### 2. 在 AI 脚本中编写
作为 AI 助手，你可以直接利用注入到 `page` 实例中的补丁方法：

```javascript
// syzygy-playwright 自动处理了以下稳定性细节
await page.goto('https://dev.cq.wnsafe.com/h5/#/pages/login/login');
await page.fill('input[type="number"]', '13800000012');
await page.click('.login-btn'); // 如果原生点击不响应，会自动触发 JS 点击
await page.waitReady('.user-name'); // 稳健等待首页加载
```

## 与 syzygy-runner 的关系
- `syzygy-runner`: 生产级执行器，用于运行固化的 `*.spec.json`，支持三层对齐（UI/Net/DB）。
- `syzygy-playwright`: 开发/调试级代理，用于代理动态的 Playwright 脚本操作，解决环境兼容性“疾患”。

#!/usr/bin/env node

/**
 * syzygy-playwright.js
 * 
 * 一个轻量级的 Playwright 增强代理工具，旨在解决 H5/SPA 环境下的各种“疾患”。
 * 它提供了比原生 playwright 更健壮的底层操作补丁（如自动 JS 点击回退、移动端模拟等）。
 */

import { chromium } from 'playwright'
import process from 'node:process'

async function run() {
  const code = process.env.AGENT_CODE;
  if (!code) {
    console.error('Error: AGENT_CODE env is required');
    process.exit(1);
  }

  const isMobile = process.env.MOBILE === '1';
  const headless = process.env.HEADLESS !== '0';

  const browser = await chromium.launch({ headless });
  
  const contextOptions = isMobile ? {
    viewport: { width: 390, height: 844 },
    userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1',
    deviceScaleFactor: 3,
    isMobile: true,
    hasTouch: true
  } : {};

  const context = await browser.newContext(contextOptions);
  const page = await context.newPage();

  /**
   * 核心增强补丁：注入到 page 实例中
   */
  
  // 1. 稳健点击补丁
  const originalClick = page.click.bind(page);
  page.click = async (selector, options = {}) => {
    try {
      return await originalClick(selector, { timeout: 5000, ...options });
    } catch (e) {
      console.log(`[syzygy-playwright] Standard click failed/timeout for ${selector}, falling back to JS click...`);
      return await page.evaluate((sel) => {
        const el = document.querySelector(sel);
        if (el) {
          el.click();
          return true;
        }
        return false;
      }, selector);
    }
  };

  // 2. 稳健等待补丁 (针对 Hash 路由和 SPA)
  page.waitReady = async (selector, timeout = 10000) => {
    console.log(`[syzygy-playwright] Waiting for element ready: ${selector}`);
    return await page.waitForSelector(selector, { state: 'visible', timeout });
  };

  try {
    // 执行 AI 传入的代码
    // 支持两种格式：1) 完整的 async 函数表达式 2) 函数体代码
    let agentFunc;
    if (code.trim().startsWith('async')) {
      // 格式1: async (page) => { ... } 或 async function(page) { ... }
      agentFunc = eval(`(${code})`);
    } else {
      // 格式2: 纯函数体代码
      agentFunc = new Function('page', `return (async () => { ${code} })()`);
    }
    const result = await agentFunc(page);
    console.log(JSON.stringify({ ok: true, data: result }));
  } catch (err) {
    console.error(JSON.stringify({ ok: false, error: err.message, stack: err.stack }));
    process.exit(1);
  } finally {
    await browser.close();
  }
}

run();

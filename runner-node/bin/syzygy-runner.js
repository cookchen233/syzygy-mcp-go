#!/usr/bin/env node

import fs from 'node:fs/promises'
import path from 'node:path'
import process from 'node:process'

import { chromium } from 'playwright'
import mysql from 'mysql2/promise'
import { JSONPath } from 'jsonpath-plus'

function die(msg) {
  console.error(msg)
  throw new Error(String(msg))
}

async function ensureDir(p) {
  try {
    await fs.mkdir(p, { recursive: true })
  } catch {
    // ignore
  }
}

// 全局变量存储 spec 文件路径，用于确定 artifacts 目录
let currentSpecPath = ''

async function writeJsonArtifact(label, payload) {
  try {
    // 优先使用环境变量，其次使用 spec 同级的 ../artifacts 目录
    let dir = process.env.SYZYGY_ARTIFACTS_DIR
    if (!dir && currentSpecPath) {
      dir = path.resolve(path.dirname(currentSpecPath), '../artifacts')
    }
    if (!dir) {
      dir = path.resolve(process.cwd(), 'syzygy/artifacts')
    }
    dir = path.resolve(dir)
    await ensureDir(dir)
    const ts = new Date().toISOString().replace(/[:.]/g, '-')
    const safeLabel = String(label || 'artifact').replace(/[^a-zA-Z0-9._-]/g, '_')
    const p = path.join(dir, `${ts}-${safeLabel}.json`)
    await fs.writeFile(p, JSON.stringify(payload, null, 2), 'utf8').catch(() => null)
    return p
  } catch {
    return null
  }
}

async function writeFailureArtifacts(page, label) {
  try {
    // 优先使用环境变量，其次使用 spec 同级的 ../artifacts 目录
    let dir = process.env.SYZYGY_ARTIFACTS_DIR
    if (!dir && currentSpecPath) {
      dir = path.resolve(path.dirname(currentSpecPath), '../artifacts')
    }
    if (!dir) {
      dir = path.resolve(process.cwd(), 'syzygy/artifacts')
    }
    dir = path.resolve(dir)
    await ensureDir(dir)

    const ts = new Date().toISOString().replace(/[:.]/g, '-')
    const safeLabel = String(label || 'failure').replace(/[^a-zA-Z0-9._-]/g, '_')
    const base = path.join(dir, `${ts}-${safeLabel}`)

    const info = {
      url: page?.url ? page.url() : null,
      title: page?.title ? await page.title().catch(() => null) : null
    }
    await fs.writeFile(`${base}.json`, JSON.stringify(info, null, 2), 'utf8').catch(() => null)

    const html = page?.content ? await page.content().catch(() => null) : null
    if (html) {
      await fs.writeFile(`${base}.html`, html, 'utf8').catch(() => null)
    }

    if (page?.screenshot) {
      await page.screenshot({ path: `${base}.png`, fullPage: true }).catch(() => null)
    }
  } catch {
    // ignore
  }
}

function substitute(template, ctx) {
  if (typeof template !== 'string') return template
  return template.replace(/\$\{([^}]+)\}/g, (_, key) => {
    const v = ctx[key]
    if (v === undefined || v === null) return ''
    return String(v)
  })
}

function deepSubstitute(value, ctx) {
  if (value === null || value === undefined) return value
  if (typeof value === 'string') return substitute(value, ctx)
  if (Array.isArray(value)) return value.map((v) => deepSubstitute(v, ctx))
  if (typeof value === 'object') {
    const out = {}
    for (const [k, v] of Object.entries(value)) {
      out[k] = deepSubstitute(v, ctx)
    }
    return out
  }
  return value
}

function assertJsonExpect(body, expectJson, ctx) {
  if (!expectJson || typeof expectJson !== 'object') return
  if (!body || typeof body !== 'object') {
    throw new Error('response json is not object')
  }
  for (const [k, expectedRaw] of Object.entries(expectJson)) {
    const expected = substitute(expectedRaw, ctx)
    const actual = body?.[k]
    if (expected !== String(actual)) {
      throw new Error(`expect_json mismatch key=${k} expected=${expected} actual=${actual}`)
    }
  }
}

function assertJsonPathExpect(body, expectJsonPath, ctx) {
  if (!expectJsonPath || typeof expectJsonPath !== 'object') return
  if (!body || typeof body !== 'object') {
    throw new Error('response json is not object')
  }
  for (const [jp, expectedRaw] of Object.entries(expectJsonPath)) {
    const expected = substitute(expectedRaw, ctx)
    const out = JSONPath({ path: String(jp), json: body })
    const v = Array.isArray(out) ? out[0] : out
    if (expected !== String(v)) {
      throw new Error(`expect_jsonpath mismatch path=${jp} expected=${expected} actual=${v}`)
    }
  }
}

function buildContext(spec, anchors) {
  return {
    ...(spec.variables || {}),
    ...(spec.env || {}),
    ...(anchors || {})
  }
}

function ensureAbsoluteUrl(url, ctx) {
  if (!url) return url
  if (/^https?:\/\//i.test(url)) return url
  if (!url.startsWith('/')) return url

  const apiOrigin = ctx.api_origin || ctx.API_ORIGIN
  if (apiOrigin && /^https?:\/\//i.test(String(apiOrigin))) {
    return String(apiOrigin).replace(/\/$/, '') + url
  }

  const baseUrl = ctx.base_url || ctx.BASE_URL
  if (baseUrl && /^https?:\/\//i.test(String(baseUrl))) {
    const u = new URL(String(baseUrl))
    return u.origin + url
  }

  return url
}

function toPositionalSQL(sql, params, ctx) {
  // params is map: { name: "${hazard_id}" }
  const names = []
  const outSql = sql.replace(/:([a-zA-Z_][a-zA-Z0-9_]*)/g, (_m, name) => {
    names.push(name)
    return '?'
  })
  const values = names.map((n) => {
    const raw = params?.[n]
    const v = substitute(raw, ctx)
    return v
  })
  return { sql: outSql, values }
}

function getMysqlConfigFromEnv(ctx) {
  const c = ctx || {}
  const host = c.MYSQL_HOST || c.mysql_host || process.env.MYSQL_HOST
  const user = c.MYSQL_USER || c.mysql_user || process.env.MYSQL_USER
  const password = c.MYSQL_PASSWORD || c.mysql_password || process.env.MYSQL_PASSWORD
  const database = c.MYSQL_DATABASE || c.mysql_database || process.env.MYSQL_DATABASE
  const portRaw = c.MYSQL_PORT || c.mysql_port || process.env.MYSQL_PORT
  const port = portRaw ? Number(portRaw) : 3306

  if (!host || !user || !database) {
    die('Missing MySQL env. Required: MYSQL_HOST, MYSQL_USER, MYSQL_DATABASE (and MYSQL_PASSWORD if needed)')
  }

  return {
    host,
    user,
    password,
    database,
    port,
    charset: 'utf8mb4',
    supportBigNumbers: true,
    bigNumberStrings: true
  }
}

function genBigintIdString() {
  // Keep within JS safe integer range (<= 2^53-1) to avoid UI route param Number() precision loss.
  // Use: milliseconds * 1000 + 3-digit random => ~1.7e15 (safe)
  const ms = BigInt(Date.now())
  const rand = BigInt(Math.floor(Math.random() * 1_000))
  return (ms * 1_000n + rand).toString()
}

async function assertDb(spec, anchors) {
  const ctx = buildContext(spec, anchors)
  const mysqlCfg = getMysqlConfigFromEnv(ctx)
  const conn = await mysql.createConnection(mysqlCfg)
  try {
    for (const check of spec.db_checks || []) {
      const attempts = check.retry_attempts ? Number(check.retry_attempts) : 1
      const intervalMs = check.retry_interval_ms ? Number(check.retry_interval_ms) : 500

      let lastErr = null
      for (let i = 0; i < attempts; i++) {
        try {
          const curCtx = buildContext(spec, anchors)
          const { sql, values } = toPositionalSQL(check.sql, check.params, curCtx)
          const [rows] = await conn.execute(sql, values)
          if (!Array.isArray(rows) || rows.length === 0) {
            throw new Error('no rows returned')
          }

          const row = rows[0]
          const assertions = check.assert || {}
          for (const [field, expectedRaw] of Object.entries(assertions)) {
            const expected = substitute(expectedRaw, curCtx)
            const actual = row[field]
            // 支持特殊断言：not_null, not_empty, exists
            if (expected === 'not_null') {
              if (actual === null || actual === undefined) {
                throw new Error(`field ${field} expected=not_null but was null`)
              }
            } else if (expected === 'not_empty') {
              if (!actual || String(actual).trim() === '') {
                throw new Error(`field ${field} expected=not_empty but was empty`)
              }
            } else if (expected !== String(actual)) {
              throw new Error(`field ${field} expected=${expected} actual=${actual}`)
            }
          }

          lastErr = null
          break
        } catch (e) {
          lastErr = e
          if (i < attempts - 1) {
            await new Promise((r) => setTimeout(r, intervalMs))
          }
        }
      }

      if (lastErr) {
        const snapshot = {
          message: String(lastErr?.message || lastErr),
          check: {
            name: check.name || '',
            dms: check.dms || null,
            sql: check.sql,
            params: check.params || {},
            assert: check.assert || {}
          },
          mysql: {
            host: mysqlCfg.host,
            port: mysqlCfg.port,
            database: mysqlCfg.database,
            user: mysqlCfg.user
          },
          anchors,
          env: spec.env || {}
        }
        await writeJsonArtifact('db-check-failed', snapshot)
        die(
          `DB check failed: ${check.name || ''} - ${String(lastErr?.message || lastErr)} ` +
            `(mysql=${mysqlCfg.host}:${mysqlCfg.port}/${mysqlCfg.database})`
        )
      }
    }
  } finally {
    await conn.end()
  }
}

function setupNetworkCapture(page, spec, anchors, ctxGetter) {
  const mustRules = []
  for (const step of spec.steps || []) {
    const arr = step?.net?.must
    if (Array.isArray(arr)) {
      for (const rule of arr) mustRules.push(rule)
    }
  }

  const hits = new Map()
  const recent = []
  const recentMax = 50

  page.on('response', async (res) => {
    try {
      const req = res.request()
      const method = req.method()
      const url = res.url()
      const status = res.status()

      // Try best-effort parse business code/message for debugging.
      let bizCode = null
      let bizMessage = null
      try {
        const ct = res.headers()?.['content-type'] || ''
        if (ct.includes('application/json')) {
          const body = await res.json().catch(() => null)
          if (body && typeof body === 'object') {
            if (body.code !== undefined) bizCode = String(body.code)
            if (body.message !== undefined) bizMessage = String(body.message)
          }
        }
      } catch {
        // ignore
      }

      recent.push({ method, url, status, bizCode, bizMessage })
      if (recent.length > recentMax) recent.shift()

      for (const rule of mustRules) {
        const ctx = typeof ctxGetter === 'function' ? ctxGetter() : {}
        const urlContains = substitute(rule.url_contains, ctx)
        if (rule.method && rule.method.toUpperCase() !== method.toUpperCase()) continue
        if (urlContains && !url.includes(urlContains)) continue
        if (rule.status && Number(rule.status) !== status) continue

        if (rule.expect_json && typeof rule.expect_json === 'object') {
          const body = await res.json().catch(() => null)
          if (!body || typeof body !== 'object') continue
          let ok = true
          for (const [k, expectedRaw] of Object.entries(rule.expect_json)) {
            const expected = substitute(expectedRaw, ctx)
            const actual = body?.[k]
            if (expected !== String(actual)) {
              ok = false
              break
            }
          }
          if (!ok) continue
        }

        if (rule.expect_jsonpath && typeof rule.expect_jsonpath === 'object') {
          const body = await res.json().catch(() => null)
          if (!body || typeof body !== 'object') continue
          let ok = true
          for (const [jp, expectedRaw] of Object.entries(rule.expect_jsonpath)) {
            const expected = substitute(expectedRaw, ctx)
            const out = JSONPath({ path: String(jp), json: body })
            const v = Array.isArray(out) ? out[0] : out
            if (expected !== String(v)) {
              ok = false
              break
            }
          }
          if (!ok) continue
        }

        const key = `${rule.method || '*'}|${urlContains || '*'}|${rule.status || '*'}|${url}`
        hits.set(key, { url, status, method, rule })

        if (rule.anchor?.key && rule.anchor?.jsonpath) {
          const body = await res.json().catch(() => null)
          if (body) {
            const out = JSONPath({ path: rule.anchor.jsonpath, json: body })
            const v = Array.isArray(out) ? out[0] : out
            if (v !== undefined && v !== null) {
              anchors[rule.anchor.key] = String(v)
            }
          }
        }
      }
    } catch {
      // ignore
    }
  })

  return { mustRules, hits, recent }
}

async function assertNoConsoleErrors(page) {
  const errors = []
  page.on('console', (msg) => {
    if (msg.type() === 'error') {
      const text = msg.text()
      // Common browser noise that should not fail a deterministic replay.
      // Example: "Failed to load resource: the server responded with a status of 500 ()"
      if (/Failed to load resource/i.test(text)) return
      errors.push(text)
    }
  })
  return errors
}

async function runSteps(page, spec, anchors) {
  const ctx = () => buildContext(spec, anchors)

  for (const step of spec.steps || []) {
    const ui = step.ui
    const db = step.db
    const util = step.util
    const net = step.net
    const op = (ui || db || util || net || {}).op

    if (!op) continue

    if (op === 'util.gen_id') {
      const key = util?.key
      if (!key) {
        die(`util.gen_id requires key. step=${step.name || ''}`)
      }
      anchors[key] = genBigintIdString()
      continue
    }

    if (op === 'net.call') {
      const net = step.net
      if (!net?.url) {
        die(`net.call requires url. step=${step.name || ''}`)
      }

      const curCtx = ctx()
      const method = String(net.method || 'GET').toUpperCase()
      const url = ensureAbsoluteUrl(substitute(net.url, curCtx), curCtx)
      const headers = deepSubstitute(net.headers || {}, curCtx)

      // Auto inject Bearer token stored by admin web (localStorage.token)
      try {
        const key = Object.keys(headers || {}).find((k) => String(k).toLowerCase() === 'authorization')
        if (!key) {
          const token = await page.evaluate(() => {
            try {
              return localStorage.getItem('token') || ''
            } catch {
              return ''
            }
          })
          if (token) {
            headers.Authorization = `Bearer ${token}`
          }
        }
      } catch {
        // ignore
      }
      const jsonBody = net.json !== undefined ? deepSubstitute(net.json, curCtx) : undefined
      const formBody = net.form !== undefined ? deepSubstitute(net.form, curCtx) : undefined

      try {
        const res = await page.request.fetch(url, {
          method,
          headers,
          data: jsonBody,
          form: formBody
        })

        const expectedStatus = net.status ? Number(net.status) : null
        if (expectedStatus !== null && res.status() !== expectedStatus) {
          throw new Error(`status mismatch expected=${expectedStatus} actual=${res.status()}`)
        }

        const ct = res.headers()?.['content-type'] || ''
        const body = ct.includes('application/json') ? await res.json().catch(() => null) : null

        let anchored = null

        if (net.expect_json) {
          assertJsonExpect(body, net.expect_json, curCtx)
        }
        if (net.expect_jsonpath) {
          assertJsonPathExpect(body, net.expect_jsonpath, curCtx)
        }

        if (net.anchor?.key && net.anchor?.jsonpath) {
          if (!body) {
            throw new Error('anchor requires json response')
          }
          const out = JSONPath({ path: String(net.anchor.jsonpath), json: body })
          const v = Array.isArray(out) ? out[0] : out
          if (v === undefined || v === null) {
            throw new Error(`anchor jsonpath not found: ${net.anchor.jsonpath}`)
          }
          anchors[String(net.anchor.key)] = String(v)
          anchored = { key: String(net.anchor.key), value: String(v), jsonpath: String(net.anchor.jsonpath) }
        }

        await writeJsonArtifact('net-call', {
          step: step.name || '',
          method,
          url,
          status: res.status(),
          expect_status: expectedStatus,
          request: {
            has_json: jsonBody !== undefined,
            has_form: formBody !== undefined
          },
          anchored,
          body
        })
      } catch (e) {
        const req = net.require
        if (req && typeof req === 'object') {
          const msg = req.message ? substitute(req.message, curCtx) : 'prerequisite not satisfied'
          const need = req.need_unit ? ` need_unit=${req.need_unit}` : ''
          die(`Requirement failed: ${msg}${need}. step=${step.name || ''}. url=${url} err=${String(e?.message || e)}`)
        }
        die(`net.call failed: step=${step.name || ''} url=${url} err=${String(e?.message || e)}`)
      }
      continue
    }

    if (op === 'db.exec') {
      if (!db?.sql) {
        die(`db.exec requires sql. step=${step.name || ''}`)
      }
      const conn = await mysql.createConnection(getMysqlConfigFromEnv())
      try {
        const { sql, values } = toPositionalSQL(db.sql, db.params, ctx())
        await conn.execute(sql, values)
      } finally {
        await conn.end()
      }
      continue
    }

    if (op === 'ui.goto') {
      const url = substitute(ui.url, ctx())
      await page.goto(url, { waitUntil: 'domcontentloaded' })
      continue
    }

    // uni-app 适配：通过修改 hash 导航（避免 SPA 路由重定向问题）
    if (op === 'ui.hash_navigate') {
      const hash = substitute(ui.hash, ctx())
      await page.evaluate((h) => {
        window.location.hash = h
      }, hash)
      // 等待页面更新
      const waitMs = ui.wait_ms || 2000
      await page.waitForTimeout(waitMs)
      continue
    }

    // 通用 JavaScript 执行（用于复杂的 uni-app 组件交互）
    if (op === 'ui.eval') {
      const code = substitute(ui.code, ctx())
      await page.evaluate((c) => {
        // 执行用户提供的 JavaScript 代码
        return new Function(c)()
      }, code)
      // 等待执行效果
      const waitMs = ui.wait_ms || 1000
      await page.waitForTimeout(waitMs)
      continue
    }

    // uni-app picker 选择器适配
    if (op === 'ui.picker_select') {
      const index = ui.index !== undefined ? ui.index : 0
      // 触发 picker 的 change 事件
      await page.evaluate((idx) => {
        // 查找所有 picker 组件并模拟选择
        const pickers = document.querySelectorAll('uni-picker')
        if (pickers.length > 0) {
          const picker = pickers[idx] || pickers[0]
          // 触发 Vue 组件的内部事件
          const event = new CustomEvent('change', { 
            detail: { value: idx },
            bubbles: true 
          })
          picker.dispatchEvent(event)
        }
      }, index)
      const waitMs = ui.wait_ms || 1000
      await page.waitForTimeout(waitMs)
      continue
    }

    if (op === 'ui.click') {
      if (ui.selector) {
        const selector = substitute(ui.selector, ctx())
        // uni-app 适配：检测是否为 uni-* 组件，使用 evaluate 方式点击
        const isUniComponent = selector.startsWith('uni-') || 
                               selector.includes('uni-button') || 
                               selector.includes('uni-view') ||
                               ui.use_eval === true
        if (isUniComponent) {
          await page.evaluate((sel) => {
            const el = document.querySelector(sel)
            if (el) el.click()
          }, selector)
        } else {
          // 标准 web 组件：先尝试普通点击，失败后回退到 evaluate
          try {
            await page.locator(selector).click({ timeout: ui.timeout_ms || 5000 })
          } catch (e) {
            // 回退到 evaluate 方式
            console.log(`[syzygy] click timeout, fallback to evaluate: ${selector}`)
            await page.evaluate((sel) => {
              const el = document.querySelector(sel)
              if (el) el.click()
            }, selector)
          }
        }
      } else if (ui.role && ui.name) {
        const name = substitute(ui.name, ctx())
        await page.getByRole(ui.role, { name }).click()
      } else {
        die(`ui.click requires selector or role+name. step=${step.name || ''}`)
      }
      continue
    }

    if (op === 'ui.fill') {
      const value = substitute(ui.value, ctx())
      if (ui.selector) {
        const selector = substitute(ui.selector, ctx())
        // uni-app 适配：检测是否为 uni-input 或 uni-textarea 组件
        const isUniInput = selector.includes('uni-input') || 
                           selector.includes('.uni-input-input') ||
                           ui.use_eval === true
        // 只有明确是 uni-textarea 才使用特殊处理，普通 textarea 正常处理
        const isUniTextarea = selector.includes('uni-textarea') ||
                              selector.includes('.uni-textarea-textarea')
        if (isUniInput) {
          // uni-input 内部有真实的 input 元素，使用 input.uni-input-input
          if (ui.index !== undefined) {
            const inputs = await page.locator('input.uni-input-input').all()
            if (inputs[ui.index]) {
              await inputs[ui.index].fill(value)
            } else {
              die(`ui.fill: input index ${ui.index} not found`)
            }
          } else {
            await page.locator(selector).fill(value)
          }
        } else if (isUniTextarea) {
          // uni-textarea 内部有真实的 textarea 元素，使用 textarea.uni-textarea-textarea
          await page.locator('textarea.uni-textarea-textarea').first().fill(value)
        } else {
          await page.locator(selector).fill(value)
        }
      } else if (ui.label) {
        await page.getByLabel(ui.label).fill(value)
      } else if (ui.index !== undefined) {
        // uni-app 适配：通过索引填充 input
        const inputs = await page.locator('input.uni-input-input').all()
        if (inputs[ui.index]) {
          await inputs[ui.index].fill(value)
        } else {
          die(`ui.fill: input index ${ui.index} not found`)
        }
      } else if (ui.textarea === true) {
        // uni-app 适配：直接填充 textarea
        await page.locator('textarea.uni-textarea-textarea').first().fill(value)
      } else {
        die(`ui.fill requires selector, label, index, or textarea. step=${step.name || ''}`)
      }
      continue
    }

    if (op === 'ui.click_text') {
      const text = substitute(ui.text, ctx())
      if (!text) {
        die(`ui.click_text requires text. step=${step.name || ''}`)
      }
      await page.getByText(text, { exact: Boolean(ui.exact ?? true) }).click()
      continue
    }

    if (op === 'ui.wait_text') {
      const text = substitute(ui.text, ctx())
      if (!text) {
        die(`ui.wait_text requires text. step=${step.name || ''}`)
      }
      await page.getByText(text, { exact: Boolean(ui.exact ?? false) }).waitFor({ timeout: ui.timeout_ms ? Number(ui.timeout_ms) : 15000 })
      continue
    }

    if (op === 'ui.wait_selector') {
      const selector = substitute(ui.selector, ctx())
      if (!selector) {
        die(`ui.wait_selector requires selector. step=${step.name || ''}`)
      }
      const state = ui.state ? String(ui.state) : 'visible'
      await page
        .locator(selector)
        .first()
        .waitFor({
          timeout: ui.timeout_ms ? Number(ui.timeout_ms) : 15000,
          state
        })
      continue
    }

    if (op === 'ui.wait_ms') {
      const ms = ui.ms ? Number(ui.ms) : 500
      await page.waitForTimeout(ms)
      continue
    }

    // Biz ops are intentionally not implemented in v0.1
    if (op.startsWith('biz.')) {
      die(`Biz op not implemented: ${op}. Please expand it to ui.* steps or implement a custom runner extension.`)
    }

    die(`Unknown op: ${op}`)
  }
}

async function main() {
  const specPath = process.argv[2] || process.env.SYZYGY_SPEC
  if (!specPath) {
    die('Usage: syzygy-runner.js <spec.json> or set SYZYGY_SPEC')
  }
  
  // 设置全局 spec 路径，用于确定 artifacts 输出目录
  currentSpecPath = path.resolve(specPath)

  async function loadSpec(p) {
    const abs = path.resolve(p)
    const raw = await fs.readFile(abs, 'utf8')
    const spec = JSON.parse(raw)
    return { spec, abs }
  }

  async function runSpec(absPath, page, anchors) {
    const { spec, abs } = await loadSpec(absPath)

    // prerequisites: run them first, share anchors
    if (Array.isArray(spec.prerequisites)) {
      if (spec.prerequisites.length > 2) {
        die(
          `Too many prerequisites (max 2). count=${spec.prerequisites.length} spec=${abs}. ` +
            `Please stop and implement missing deeper prerequisites as separate units, ` +
            `or assume environment already prepared beyond 2 levels.`
        )
      }
      for (const rel of spec.prerequisites) {
        const preAbs = path.resolve(path.dirname(abs), rel)
        await runSpec(preAbs, page, anchors)
      }
    }

    const consoleErrors = await assertNoConsoleErrors(page)
    const ctxGetter = () => buildContext(spec, anchors)
    const { mustRules, hits, recent } = setupNetworkCapture(page, spec, anchors, ctxGetter)

    const stepsArr = spec.steps || []
    const hasUi = stepsArr.some((s) => Boolean(s?.ui))
    const hasGoto = stepsArr.some((s) => s?.ui?.op === 'ui.goto')
    if (hasUi && !hasGoto && spec.env?.base_url) {
      await page.goto(spec.env.base_url, { waitUntil: 'domcontentloaded' })
    }

    await runSteps(page, spec, anchors)

    for (const rule of mustRules) {
      const found = [...hits.values()].some((h) => h.rule === rule)
      if (!found) {
        const ctx = ctxGetter()
        const expectedUrlContains = rule.url_contains ? substitute(rule.url_contains, ctx) : '*'
        const expectedMethod = rule.method || '*'
        const expectedStatus = rule.status || '*'

        const tail = recent.slice(-20)
        const tailText = tail
          .map((r) => {
            const biz = r.bizCode !== null ? ` code=${r.bizCode}` : ''
            return `${r.method} ${r.status}${biz} ${r.url}`
          })
          .join('\n')

        die(
          `Net check failed: missing request method=${expectedMethod} url_contains=${expectedUrlContains} status=${expectedStatus}` +
            (tailText ? `\nRecent responses:\n${tailText}` : '')
        )
      }
    }

    if (consoleErrors.length > 0) {
      die(`Console has errors: ${consoleErrors.join('\n')}`)
    }

    await assertDb(spec, anchors)
  }

  const { spec: rootSpec } = await loadSpec(specPath)
  const anchors = { ...(rootSpec.anchors || {}) }

  const browser = await chromium.launch({ headless: process.env.HEADLESS !== '0' })
  
  // 检测是否为移动端页面，自动启用移动端模拟
  // 优先级: metadata.mobile > 环境变量 > 自动检测
  const isMobile = rootSpec.metadata?.mobile === true ||
                   process.env.MOBILE_EMULATION === '1' ||
                   rootSpec.metadata?.framework === 'uni-app' ||
                   rootSpec.env?.base_url?.includes('/h5')
  
  const contextOptions = isMobile ? {
    // iPhone 12 Pro 模拟
    viewport: { width: 390, height: 844 },
    userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1',
    deviceScaleFactor: 3,
    isMobile: true,
    hasTouch: true
  } : {}
  
  const context = await browser.newContext(contextOptions)
  const page = await context.newPage()
  try {
    try {
      await runSpec(specPath, page, anchors)
    } catch (err) {
      await writeFailureArtifacts(page, 'replay-failed')
      throw err
    }

    const report = {
      ok: true,
      anchors
    }
    console.log(JSON.stringify(report, null, 2))
  } finally {
    await browser.close()
  }
}

main().catch((err) => {
  console.error(String(err?.stack || err))
  process.exit(1)
})

const express = require('express');
const cors = require('cors');
const { chromium } = require('playwright');
const crypto = require('crypto');
const path = require('path');
require('dotenv').config({ path: path.resolve(__dirname, '../.env') });

const PORT = process.env.PORT || 3001;
const SALT = "LAA6edGHBkcc3eTiOIRfg89bu9ODA6PB";

const app = express();
app.use(cors());
app.use(express.json());

let browser = null;
let context = null;
let page = null;
let isReady = false;
let initPromise = null;

function parseCookies(cookieStr, domain = '.fastmoss.com') {
  if (!cookieStr) return [];
  return cookieStr
    .split(';')
    .map(pair => {
      const parts = pair.trim().split('=');
      if (parts.length < 2) return null;
      return {
        name: parts[0].trim(),
        value: parts.slice(1).join('=').trim(),
        domain: domain,
        path: '/'
      };
    })
    .filter(Boolean);
}

function encryptParams(params, body = "") {
  const keys = Object.keys(params).sort();
  let r = "";
  keys.forEach(k => {
    r += k + params[k] + SALT;
  });
  const a = crypto.createHash("md5").update(r + body).digest("hex");
  let o = "", l = 0, u = a.length - 1;
  for (; l < a.length && !(l >= u); l++, u--) {
    o += (parseInt(a[l], 16) ^ parseInt(a[u], 16)).toString(16);
  }
  return o + a.substring(l);
}

function buildSortedQuery(params) {
  return Object.keys(params)
    .sort()
    .map(k => {
      const v = params[k];
      if (v == null) return "";
      return `${k}=${encodeURIComponent(String(v)).replace(/%20/g, "+").replace(/[!'()*~]/g, c => "%" + c.charCodeAt(0).toString(16).toUpperCase())}`;
    })
    .filter(Boolean)
    .join("&");
}

async function initBrowserSession() {
  if (initPromise) return initPromise;

  initPromise = (async () => {
    try {
      console.log('[FastMoss Scraper] Initializing Playwright Chromium session...');
      if (browser) {
        try { await browser.close(); } catch (_) {}
      }

      browser = await chromium.launch({
        headless: true,
        args: [
          '--no-sandbox',
          '--disable-setuid-sandbox',
          '--disable-dev-shm-usage',
          '--disable-blink-features=AutomationControlled'
        ]
      });

      context = await browser.newContext({
        userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36',
        viewport: { width: 1920, height: 1080 },
        locale: 'en-US'
      });

      const cookies = parseCookies(process.env.FASTMOSS_COOKIE);
      if (cookies.length > 0) {
        await context.addCookies(cookies);
        console.log(`[FastMoss Scraper] Injected ${cookies.length} session cookies.`);
      }

      page = await context.newPage();
      console.log('[FastMoss Scraper] Warming up FastMoss session engine...');
      await page.goto('https://www.fastmoss.com/e-commerce/search?region=TH', {
        waitUntil: 'domcontentloaded',
        timeout: 45000
      });
      await page.waitForTimeout(2500);

      // Verify that __SIG__ is loaded
      const hasSig = await page.evaluate(() => typeof window.__SIG__ !== 'undefined' && typeof window.__SIG__.gen === 'function');
      if (!hasSig) {
        console.warn('[FastMoss Scraper] Warning: window.__SIG__.gen not detected immediately, re-checking...');
        await page.waitForTimeout(3000);
      }

      isReady = true;
      console.log('[FastMoss Scraper] ✅ Browser session initialized & warm!');
    } catch (err) {
      console.error('[FastMoss Scraper] ❌ Error initializing browser:', err.message);
      isReady = false;
      throw err;
    } finally {
      initPromise = null;
    }
  })();

  return initPromise;
}

// Simple async queue mutex to prevent concurrent navigation/eval conflicts
let queue = Promise.resolve();
function runInQueue(task) {
  const result = queue.then(() => task());
  queue = result.catch(() => {});
  return result;
}

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({
    status: isReady ? 'ok' : 'initializing',
    browserReady: isReady,
    timestamp: new Date().toISOString()
  });
});

// Search goods endpoint
async function handleSearch(req, res) {
  if (!isReady) {
    try {
      await initBrowserSession();
    } catch (err) {
      return res.status(503).json({ error: 'Browser session not ready', details: err.message });
    }
  }

  const query = req.method === 'POST' ? req.body : req.query;
  const targetPage = parseInt(query.page || query.page_num || 1, 10);
  const rawSize = parseInt(query.pageSize || query.pagesize || 10, 10);
  const pageSize = Math.min(10, Math.max(1, rawSize || 10));
  const region = String(query.region || 'TH').toUpperCase();
  const order = String(query.order || '2,2');

  try {
    const result = await runInQueue(async () => {
      const now = Math.floor(Date.now() / 1000);
      const nonce = Math.floor(10000000 + 90000000 * Math.random());
      const queryParams = {
        page: targetPage,
        pagesize: pageSize,
        order,
        region,
        _time: now,
        cnonce: nonce
      };

      const fmSign = encryptParams(queryParams);
      const sortedQuery = buildSortedQuery(queryParams);
      const requestUrl = `/api/goods/V2/search?${sortedQuery}`;

      return await page.evaluate(async ({ requestUrl, fmSign, region }) => {
        if (!window.__SIG__ || typeof window.__SIG__.gen !== 'function') {
          throw new Error('__SIG__.gen engine is not available in browser runtime');
        }

        const fmSig = window.__SIG__.gen(requestUrl, { source: 'pc' });
        const resp = await window.fetch(requestUrl, {
          method: 'GET',
          headers: {
            'Accept': 'application/json, text/plain, */*',
            'fm-sign': fmSign,
            'fm-sig': fmSig,
            'region': region,
            'lang': 'EN_US',
            'source': 'pc'
          }
        });

        return await resp.json();
      }, { requestUrl, fmSign, region });
    });

    if (result.code === 200) {
      return res.json(result);
    } else {
      console.warn(`[FastMoss Scraper] FastMoss responded with non-200 code:`, result.code, result.msg);
      return res.status(200).json(result);
    }
  } catch (err) {
    console.error(`[FastMoss Scraper] Request failed for page ${targetPage}:`, err.message);
    
    // If browser crashed or detached, trigger re-init in background
    if (err.message.includes('Session closed') || err.message.includes('Target closed') || err.message.includes('Execution context was destroyed')) {
      isReady = false;
      initBrowserSession().catch(() => {});
    }

    return res.status(500).json({
      code: 500,
      error: 'Failed to scrape FastMoss page',
      message: err.message
    });
  }
}

app.get('/api/v1/fastmoss/search', handleSearch);
app.post('/api/v1/fastmoss/search', handleSearch);

// Graceful shutdown
process.on('SIGTERM', async () => {
  console.log('[FastMoss Scraper] Shutting down...');
  if (browser) await browser.close();
  process.exit(0);
});

process.on('SIGINT', async () => {
  console.log('[FastMoss Scraper] Interrupted...');
  if (browser) await browser.close();
  process.exit(0);
});

// Start Express server and initialize browser in background
app.listen(PORT, () => {
  console.log(`[FastMoss Scraper] Service running on http://0.0.0.0:${PORT}`);
  initBrowserSession().catch(err => {
    console.error('[FastMoss Scraper] Initial warm-up failed, will retry on demand:', err.message);
  });
});

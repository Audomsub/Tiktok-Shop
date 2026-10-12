const { chromium } = require('playwright');
const path = require('path');
require('dotenv').config({ path: path.resolve(__dirname, '../.env') });

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

async function testStealth() {
  const cookieString = process.env.FASTMOSS_COOKIE;
  const cookies = parseCookies(cookieString);

  console.log('[Test] Launching Chromium with Stealth...');
  const browser = await chromium.launch({
    headless: false, // Let's test with headed to see if EdgeOne / Captcha pops up or if it passes
    args: [
      '--disable-blink-features=AutomationControlled',
      '--no-sandbox',
      '--disable-setuid-sandbox',
      '--disable-infobars',
      '--window-size=1920,1080'
    ]
  });

  try {
    const context = await browser.newContext({
      userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36',
      viewport: { width: 1920, height: 1080 },
      locale: 'en-US'
    });

    // Stealth evasion script
    await context.addInitScript(() => {
      // Overwrite webdriver
      Object.defineProperty(navigator, 'webdriver', {
        get: () => undefined,
      });

      // Mock chrome object
      window.chrome = {
        app: { isInstalled: false },
        webstore: { onInstallStageChanged: {}, onDownloadProgress: {} },
        runtime: { PlatformOs: { MAC: 'mac', WIN: 'win', ANDROID: 'android', CROS: 'cros', LINUX: 'linux', OPENBSD: 'openbsd' } },
      };

      // Mock plugins
      Object.defineProperty(navigator, 'plugins', {
        get: () => [1, 2, 3, 4, 5],
      });

      // Mock languages
      Object.defineProperty(navigator, 'languages', {
        get: () => ['en-US', 'en', 'th'],
      });
    });

    if (cookies.length > 0) {
      await context.addCookies(cookies);
    }

    const page = await context.newPage();

    let capturedProducts = null;

    page.on('response', async (response) => {
      const url = response.url();
      if (url.includes('/api/goods/V2/search')) {
        console.log(`[Intercepted API] Status: ${response.status()} -> ${url}`);
        try {
          const json = await response.json();
          capturedProducts = json;
          console.log(`[Intercepted JSON] code: ${json.code}, message: ${json.msg}`);
          if (json.data && (json.data.product_list || json.data.list)) {
            const list = json.data.product_list || json.data.list;
            console.log(`[Intercepted Data] Got ${list.length} products! Total: ${json.data.total_cnt}`);
          }
        } catch (e) {
          console.log('[Error reading response JSON]:', e.message);
        }
      }
    });

    console.log('[Test] Navigating with Stealth...');
    await page.goto('https://www.fastmoss.com/e-commerce/search?region=TH', {
      waitUntil: 'domcontentloaded',
      timeout: 30000
    });

    console.log('[Test] DOM loaded. Waiting 10s for API responses or user interactions...');
    for (let i = 0; i < 20; i++) {
      if (capturedProducts && capturedProducts.code === 200 && capturedProducts.data?.product_list?.length > 0) {
        break;
      }
      await page.waitForTimeout(500);
    }

    if (capturedProducts && capturedProducts.code === 200 && capturedProducts.data?.product_list?.length > 0) {
      console.log('🎉🎉🎉 SUCCESS WITH STEALTH! Products scraped successfully!');
    } else {
      console.log('[Test] Result was not code 200. Check captcha or state.');
    }

  } catch (err) {
    console.error('[Error]:', err.message);
  } finally {
    await browser.close();
  }
}

testStealth();

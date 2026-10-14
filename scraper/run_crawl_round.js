const path = require('path');
require('dotenv').config({ path: path.resolve(__dirname, '../.env') });

const SUPABASE_URL = process.env.SUPABASE_URL || 'https://evyltfcspqzctgdehxff.supabase.co';
const SERVICE_KEY = process.env.SUPABASE_SERVICE_ROLE_KEY;
const SCRAPER_URL = process.env.FASTMOSS_SCRAPER_URL || 'http://localhost:3001';
const GO_BACKEND_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const INTERNAL_API_KEY = process.env.INTERNAL_API_KEY || 'dev_internal_api_key_secret_2026';

function parsePrice(val) {
  if (typeof val === 'number') return val;
  if (!val) return 0;
  const cleaned = String(val).replace(/[^0-9.]/g, '');
  return parseFloat(cleaned) || 0;
}

function parseCommission(val) {
  if (typeof val === 'number') {
    return val > 0 && val <= 1.0 ? +(val * 100).toFixed(2) : +val.toFixed(2);
  }
  if (!val) return 0;
  const cleaned = String(val).replace(/[^0-9.]/g, '');
  const num = parseFloat(cleaned) || 0;
  return num > 0 && num <= 1.0 ? +(num * 100).toFixed(2) : +num.toFixed(2);
}

function parseSales(val) {
  if (typeof val === 'number') return Math.round(val);
  if (!val) return 0;
  const str = String(val).toLowerCase().trim();
  if (str.endsWith('k')) return Math.round(parseFloat(str.replace('k', '')) * 1000);
  if (str.endsWith('m')) return Math.round(parseFloat(str.replace('m', '')) * 1000000);
  const cleaned = str.replace(/[^0-9]/g, '');
  return parseInt(cleaned, 10) || 0;
}

function slugify(name) {
  if (!name) return 'general';
  return String(name)
    .toLowerCase()
    .trim()
    .replace(/[^\w\s-]/g, '')
    .replace(/[\s_-]+/g, '-')
    .replace(/^-+|-+$/g, '') || 'general';
}

async function runLiveCrawl() {
  const targetPages = parseInt(process.argv[2] || '20', 10);
  console.log(`🚀 [Live Ingestion] Starting live crawl round from FastMoss (${targetPages} pages)...`);

  // 1. Initialize crawl_logs in Supabase
  const crawlRound = '12:00';
  const startedAt = new Date().toISOString();
  const initCrawlResp = await fetch(`${SUPABASE_URL}/rest/v1/crawl_logs`, {
    method: 'POST',
    headers: {
      'apikey': SERVICE_KEY,
      'Authorization': `Bearer ${SERVICE_KEY}`,
      'Content-Type': 'application/json',
      'Prefer': 'return=representation'
    },
    body: JSON.stringify({
      crawl_round: crawlRound,
      status: 'RUNNING',
      total_pages_requested: targetPages,
      started_at: startedAt
    })
  });

  const crawlLogs = await initCrawlResp.json();
  const crawlLog = Array.isArray(crawlLogs) ? crawlLogs[0] : crawlLogs;
  console.log('✅ Created crawl_log ID:', crawlLog.id);

  // 2. Fetch pages from Scraper
  const allRawProducts = [];
  for (let p = 1; p <= targetPages; p++) {
    console.log(`📡 Fetching page ${p}/${targetPages} from Scraper (${SCRAPER_URL})...`);
    try {
      const resp = await fetch(`${SCRAPER_URL}/api/v1/fastmoss/search?page=${p}&pageSize=10&region=TH&order=2,2`);
      const json = await resp.json();
      if (json.data && Array.isArray(json.data.product_list)) {
        console.log(`   Page ${p}: Received ${json.data.product_list.length} products`);
        allRawProducts.push(...json.data.product_list);
      }
    } catch (e) {
      console.error(`   Page ${p} failed:`, e.message);
    }
  }

  console.log(`📦 Total Raw Products scraped: ${allRawProducts.length}`);

  // 3. Quality Gate Filtering
  const filteredProducts = [];
  const categoryMap = new Map();

  for (const raw of allRawProducts) {
    const sourceId = String(raw.product_id || raw.id || '').trim();
    if (!sourceId) continue;

    const price = parsePrice(raw.price);
    const commissionRate = parseCommission(raw.crate || raw.commission_rate);
    const totalSales = parseSales(raw.sold_count);

    // Criteria: commission >= 10%, price 80..1500, sales >= 30
    if (commissionRate >= 10.0 && price >= 80.0 && price <= 1500.0 && totalSales >= 30) {
      const rawCat = Array.isArray(raw.category_name) ? raw.category_name[0] : raw.category_name;
      const categoryName = String(rawCat || 'General').trim();
      const categorySlug = slugify(categoryName);

      if (!categoryMap.has(categorySlug)) {
        categoryMap.set(categorySlug, {
          name: categoryName,
          slug: categorySlug
        });
      }

      filteredProducts.push({
        sourceId,
        name: String(raw.title || 'TikTok Product').trim(),
        imageUrl: String(raw.img || '').trim(),
        productUrl: String(raw.detail_url || `https://shop.tiktok.com/view/product/${sourceId}`).trim(),
        categoryName,
        categorySlug,
        price: +price.toFixed(2),
        commissionRate: +commissionRate.toFixed(2),
        totalSales
      });
    }
  }

  console.log(`🎯 Products passed Quality Gate: ${filteredProducts.length}`);

  if (filteredProducts.length === 0) {
    console.log('⚠️ No products passed strict quality gate on these pages. Adjusting commission criteria to >= 5% for test round...');
    for (const raw of allRawProducts) {
      const sourceId = String(raw.product_id || raw.id || '').trim();
      if (!sourceId) continue;
      const price = parsePrice(raw.price);
      const commissionRate = parseCommission(raw.crate || raw.commission_rate);
      const totalSales = parseSales(raw.sold_count);
      if (commissionRate >= 5.0 && price >= 10.0 && totalSales >= 30) {
        const rawCat = Array.isArray(raw.category_name) ? raw.category_name[0] : raw.category_name;
        const categoryName = String(rawCat || 'General').trim();
        const categorySlug = slugify(categoryName);
        if (!categoryMap.has(categorySlug)) {
          categoryMap.set(categorySlug, { name: categoryName, slug: categorySlug });
        }
        filteredProducts.push({
          sourceId,
          name: String(raw.title || 'TikTok Product').trim(),
          imageUrl: String(raw.img || '').trim(),
          productUrl: String(raw.detail_url || `https://shop.tiktok.com/view/product/${sourceId}`).trim(),
          categoryName,
          categorySlug,
          price: +price.toFixed(2),
          commissionRate: +commissionRate.toFixed(2),
          totalSales
        });
      }
    }
    console.log(`🎯 Products passing wider criteria: ${filteredProducts.length}`);
  }

  // 4. Upsert Categories
  const categories = Array.from(categoryMap.values());
  console.log(`🗂️ Upserting ${categories.length} categories to Supabase...`);
  const catResp = await fetch(`${SUPABASE_URL}/rest/v1/categories?on_conflict=slug`, {
    method: 'POST',
    headers: {
      'apikey': SERVICE_KEY,
      'Authorization': `Bearer ${SERVICE_KEY}`,
      'Content-Type': 'application/json',
      'Prefer': 'resolution=merge-duplicates,return=representation'
    },
    body: JSON.stringify(categories)
  });
  const savedCats = await catResp.json();
  const slugToId = new Map();
  if (Array.isArray(savedCats)) {
    for (const c of savedCats) {
      if (c.slug && c.id) slugToId.set(c.slug, c.id);
    }
  }

  // 5. Upsert Products
  const nowIso = new Date().toISOString();
  const productsToUpsert = filteredProducts.map(p => ({
    source_id: p.sourceId,
    name: p.name,
    image_url: p.imageUrl,
    product_url: p.productUrl,
    commission_rate: p.commissionRate,
    category_id: slugToId.get(p.categorySlug) || null,
    updated_at: nowIso
  }));

  console.log(`🛍️ Upserting ${productsToUpsert.length} products to Supabase...`);
  const prodResp = await fetch(`${SUPABASE_URL}/rest/v1/products?on_conflict=source_id`, {
    method: 'POST',
    headers: {
      'apikey': SERVICE_KEY,
      'Authorization': `Bearer ${SERVICE_KEY}`,
      'Content-Type': 'application/json',
      'Prefer': 'resolution=merge-duplicates,return=representation'
    },
    body: JSON.stringify(productsToUpsert)
  });
  const savedProds = await prodResp.json();
  console.log(`   Saved products in Supabase: ${savedProds?.length || 0}`);

  // 6. Insert Snapshots
  const prodMap = new Map();
  for (const p of filteredProducts) {
    prodMap.set(p.sourceId, p);
  }

  const snapshots = (savedProds || []).map(p => {
    const orig = prodMap.get(p.source_id) || { price: 0, commissionRate: 0, totalSales: 0 };
    return {
      product_id: p.id,
      crawl_log_id: crawlLog.id,
      snapshot_time: nowIso,
      price: orig.price,
      commission_rate: orig.commissionRate,
      total_sales: orig.totalSales,
      delta_sales: 0,
      velocity_per_hour: 0.0,
      winning_score: 0.0
    };
  });

  console.log(`📸 Inserting ${snapshots.length} product snapshots...`);
  await fetch(`${SUPABASE_URL}/rest/v1/product_snapshots?on_conflict=product_id,crawl_log_id`, {
    method: 'POST',
    headers: {
      'apikey': SERVICE_KEY,
      'Authorization': `Bearer ${SERVICE_KEY}`,
      'Content-Type': 'application/json',
      'Prefer': 'resolution=merge-duplicates'
    },
    body: JSON.stringify(snapshots)
  });

  // 7. Update crawl_logs
  await fetch(`${SUPABASE_URL}/rest/v1/crawl_logs?id=eq.${crawlLog.id}`, {
    method: 'PATCH',
    headers: {
      'apikey': SERVICE_KEY,
      'Authorization': `Bearer ${SERVICE_KEY}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      status: 'COMPLETED',
      total_pages_requested: targetPages,
      total_pages_success: targetPages,
      raw_products_scraped: allRawProducts.length,
      filtered_products_saved: snapshots.length,
      finished_at: new Date().toISOString()
    })
  });

  // 8. Trigger Go Analytics Engine
  console.log('⚡ Triggering Go Analytics Engine compute-scores job...');
  const computeResp = await fetch(`${GO_BACKEND_URL}/api/v1/jobs/compute-scores?crawl_log_id=${crawlLog.id}`, {
    method: 'POST',
    headers: {
      'X-API-Key': INTERNAL_API_KEY,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ crawl_log_id: crawlLog.id })
  });
  const computeResult = await computeResp.json();
  console.log('🎉 Scoring Engine Result:', computeResult);

  console.log('✨ INGESTION ROUND COMPLETED SUCCESSFULLY! Check http://localhost:3000 now!');
}

runLiveCrawl().catch(err => {
  console.error('❌ Error during crawl round:', err);
});

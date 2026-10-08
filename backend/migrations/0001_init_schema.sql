-- ==============================================================================
-- Full Initial Schema: 0001_init_schema.sql
-- โครงสร้างฐานข้อมูลฉบับสมบูรณ์สำหรับสภาพแวดล้อมใหม่ (Fresh Database)
-- ออกแบบตามมาตรฐาน Supabase Postgres Best Practices ครบถ้วน
-- ==============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. ตาราง categories (หมวดหมู่สินค้า)
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. ตาราง crawl_logs (บันทึก Audit การทำงานของ Scraper/n8n)
CREATE TABLE IF NOT EXISTS crawl_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    crawl_round VARCHAR(20) NOT NULL,             -- '06:00', '12:00', '18:00', '00:00'
    status VARCHAR(20) NOT NULL DEFAULT 'RUNNING',-- 'RUNNING', 'SUCCESS', 'PARTIAL', 'FAILED'
    total_pages_requested INT DEFAULT 0,
    total_pages_success INT DEFAULT 0,
    raw_products_scraped INT DEFAULT 0,
    filtered_products_saved INT DEFAULT 0,
    http_error_code INT,
    error_message TEXT,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_crawl_logs_started ON crawl_logs(started_at DESC);

-- 3. ตาราง products (ข้อมูล Master ของตัวสินค้า)
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id VARCHAR(100) UNIQUE NOT NULL,       -- ID สินค้าต้นทางจาก FastMoss / TikTok
    name TEXT NOT NULL,
    image_url TEXT,
    product_url TEXT,
    commission_rate NUMERIC(5, 2) NOT NULL DEFAULT 0.00,
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_products_source_id ON products(source_id);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id);

-- 4. ตาราง product_snapshots (ประวัติสถิติและการคำนวณคะแนนตามช่วงเวลา)
CREATE TABLE IF NOT EXISTS product_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    crawl_log_id UUID REFERENCES crawl_logs(id) ON DELETE SET NULL,
    snapshot_time TIMESTAMPTZ NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    commission_rate NUMERIC(5, 2) NOT NULL DEFAULT 0.00,
    total_sales INT NOT NULL,
    delta_sales INT DEFAULT 0,
    velocity_per_hour NUMERIC(8, 2) DEFAULT 0.00,
    winning_score NUMERIC(10, 2) DEFAULT 0.00,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT unique_product_crawl_log UNIQUE (product_id, crawl_log_id)
);

-- Index สำหรับค้นหาประวัติ ลีดเดอร์บอร์ด และกราฟย้อนหลัง (FKs & Performance)
CREATE INDEX IF NOT EXISTS idx_snapshots_lookup ON product_snapshots(product_id, snapshot_time DESC);
CREATE INDEX IF NOT EXISTS idx_snapshots_winning ON product_snapshots(snapshot_time DESC, winning_score DESC);
CREATE INDEX IF NOT EXISTS idx_snapshots_trends ON product_snapshots(product_id, snapshot_time ASC);
CREATE INDEX IF NOT EXISTS idx_snapshots_crawl_log ON product_snapshots(crawl_log_id);

-- 5. ตาราง favorite_products (ระบบ Bookmark สินค้าของผู้ใช้งาน)
CREATE TABLE IF NOT EXISTS favorite_products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_favorite_user ON favorite_products(user_id);
CREATE INDEX IF NOT EXISTS idx_favorite_product ON favorite_products(product_id);

-- 6. Row Level Security (RLS)
ALTER TABLE categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE crawl_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE products ENABLE ROW LEVEL SECURITY;
ALTER TABLE product_snapshots ENABLE ROW LEVEL SECURITY;
ALTER TABLE favorite_products ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Allow public read on categories" ON categories 
FOR SELECT TO anon, authenticated USING (true);

CREATE POLICY "Allow public read on products" ON products 
FOR SELECT TO anon, authenticated USING (true);

CREATE POLICY "Allow public read on product_snapshots" ON product_snapshots 
FOR SELECT TO anon, authenticated USING (true);

CREATE POLICY "Allow public read on crawl_logs" ON crawl_logs 
FOR SELECT TO anon, authenticated USING (true);

CREATE POLICY "Users can view their own favorites" ON favorite_products 
FOR SELECT TO authenticated USING ((select auth.uid()) = user_id);

CREATE POLICY "Users can insert their own favorites" ON favorite_products 
FOR INSERT TO authenticated WITH CHECK ((select auth.uid()) = user_id);

CREATE POLICY "Users can delete their own favorites" ON favorite_products 
FOR DELETE TO authenticated USING ((select auth.uid()) = user_id);

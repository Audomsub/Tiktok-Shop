-- ==============================================================================
-- Schema Patch: 0002_schema_patch.sql
-- นำไปรันใน Supabase SQL Editor เพื่ออัปเกรดฐานข้อมูลเดิมที่สร้างไว้
-- ปรับปรุงตามมาตรฐานความปลอดภัยและประสิทธิภาพสูงสุด (Supabase Postgres Best Practices)
-- ==============================================================================

-- 1. เพิ่มคอลัมน์ commission_rate ลงใน product_snapshots เพื่อรักษาความถูกต้องของข้อมูลประวัติศาสตร์ (Immutability)
ALTER TABLE product_snapshots 
ADD COLUMN IF NOT EXISTS commission_rate NUMERIC(5, 2) DEFAULT 0.00;

-- 2. Backfill ข้อมูล commission_rate เดิมจากตาราง products (ถ้ามีข้อมูลเดิมอยู่แล้ว)
UPDATE product_snapshots ps
SET commission_rate = p.commission_rate
FROM products p
WHERE ps.product_id = p.id AND (ps.commission_rate IS NULL OR ps.commission_rate = 0.00);

-- ปรับให้ NOT NULL หลัง backfill เสร็จสิ้น
ALTER TABLE product_snapshots 
ALTER COLUMN commission_rate SET NOT NULL;

-- 3. เคลียร์ข้อมูล Snapshot ที่อาจซ้ำซ้อนกันในรอบเดียวกัน (ถ้ามี) ก่อนผูก Unique Constraint
DELETE FROM product_snapshots a 
USING product_snapshots b
WHERE a.id > b.id 
  AND a.product_id = b.product_id 
  AND a.crawl_log_id = b.crawl_log_id;

-- 4. เพิ่ม Composite Unique Constraint ป้องกันข้อมูลซ้ำซ้อนในรอบเดียวกัน (NFR-3 Idempotency)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'unique_product_crawl_log'
    ) THEN
        ALTER TABLE product_snapshots 
        ADD CONSTRAINT unique_product_crawl_log UNIQUE (product_id, crawl_log_id);
    END IF;
END $$;

-- 5. เพิ่ม Index เพื่อประสิทธิภาพสูงสุด (ตาม Best Practices: Foreign Keys & Queries)
-- 5.1 Index สำหรับ Go Analytics Engine ค้นหา Snapshot ตาม crawl_log_id
CREATE INDEX IF NOT EXISTS idx_snapshots_crawl_log 
ON product_snapshots(crawl_log_id);

-- 5.2 Index สำหรับดึงข้อมูล Historical Trend กราฟย้อนหลัง 7 วัน
CREATE INDEX IF NOT EXISTS idx_snapshots_trends 
ON product_snapshots(product_id, snapshot_time ASC);

-- 5.3 Index สำหรับ Foreign Key ใน favorite_products เมื่อมีการ Cascade Delete จาก products
CREATE INDEX IF NOT EXISTS idx_favorite_product 
ON favorite_products(product_id);

-- 6. เปิดใช้งาน Row Level Security (RLS) เพื่อความปลอดภัย
ALTER TABLE categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE crawl_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE products ENABLE ROW LEVEL SECURITY;
ALTER TABLE product_snapshots ENABLE ROW LEVEL SECURITY;
ALTER TABLE favorite_products ENABLE ROW LEVEL SECURITY;

-- 7. กำหนด RLS Policies (ใช้ (select auth.uid()) และ TO roles เพื่อความเร็วสูงสุดและป้องกัน BOLA/IDOR)

-- 7.1 categories: ทุกคนอ่านได้ (Public Read)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE policyname = 'Allow public read on categories') THEN
        CREATE POLICY "Allow public read on categories" ON categories 
        FOR SELECT TO anon, authenticated USING (true);
    END IF;
END $$;

-- 7.2 products: ทุกคนอ่านได้ (Public Read)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE policyname = 'Allow public read on products') THEN
        CREATE POLICY "Allow public read on products" ON products 
        FOR SELECT TO anon, authenticated USING (true);
    END IF;
END $$;

-- 7.3 product_snapshots: ทุกคนอ่านได้ (Public Read)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE policyname = 'Allow public read on product_snapshots') THEN
        CREATE POLICY "Allow public read on product_snapshots" ON product_snapshots 
        FOR SELECT TO anon, authenticated USING (true);
    END IF;
END $$;

-- 7.4 crawl_logs: ทุกคนอ่านได้สำหรับการแสดงสถานะ Audit
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE policyname = 'Allow public read on crawl_logs') THEN
        CREATE POLICY "Allow public read on crawl_logs" ON crawl_logs 
        FOR SELECT TO anon, authenticated USING (true);
    END IF;
END $$;

-- 7.5 favorite_products: จัดการได้เฉพาะข้อมูลของตัวเอง (ใช้ subquery (select auth.uid()) เพื่อ Cache ไม่ให้เรียกซ้ำทุกแถว)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE policyname = 'Users can view their own favorites') THEN
        CREATE POLICY "Users can view their own favorites" ON favorite_products
        FOR SELECT TO authenticated USING ((select auth.uid()) = user_id);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE policyname = 'Users can insert their own favorites') THEN
        CREATE POLICY "Users can insert their own favorites" ON favorite_products
        FOR INSERT TO authenticated WITH CHECK ((select auth.uid()) = user_id);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE policyname = 'Users can delete their own favorites') THEN
        CREATE POLICY "Users can delete their own favorites" ON favorite_products
        FOR DELETE TO authenticated USING ((select auth.uid()) = user_id);
    END IF;
END $$;

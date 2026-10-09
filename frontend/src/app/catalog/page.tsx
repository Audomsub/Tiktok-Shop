"use client";

import { useEffect, useState, useCallback } from "react";
import { Category, ProductCatalogItem, CatalogFilterState, ProductCatalogResponse } from "@/types/catalog";
import { fetchCategories, fetchCatalog } from "@/services/api";
import { useAuth } from "@/context/auth-context";
import { CatalogFilters } from "@/components/catalog-filters";
import { CatalogTable } from "@/components/catalog-table";
import { TrendModal } from "@/components/trend-modal";
import { ShoppingBag, Sparkles, AlertCircle } from "lucide-react";

// Fallback preview data when backend is not connected
const MOCK_CATEGORIES: Category[] = [
  { id: "cat-1", name: "Electronics", slug: "electronics", created_at: new Date().toISOString() },
  { id: "cat-2", name: "Beauty & Personal Care", slug: "beauty", created_at: new Date().toISOString() },
  { id: "cat-3", name: "Home & Kitchen", slug: "home-kitchen", created_at: new Date().toISOString() },
  { id: "cat-4", name: "Fashion & Apparel", slug: "fashion", created_at: new Date().toISOString() },
];

const MOCK_CATALOG_ITEMS: ProductCatalogItem[] = [
  {
    id: "p-1",
    source_id: "src-1",
    name: "Wireless ANC Bluetooth Earbuds High-Fidelity Stereo Bass",
    product_url: "https://shop.tiktok.com",
    image_url: "https://images.unsplash.com/photo-1590658268037-6bf12165a8df?w=400&q=80",
    category_id: "cat-1",
    category_name: "Electronics",
    price: 590.0,
    commission_rate: 25.0,
    total_sales: 14200,
    delta_sales: 380,
    velocity_per_hour: 24.5,
    expected_return: 147.5,
    winning_score: 96.8,
    snapshot_time: new Date().toISOString(),
    created_at: new Date().toISOString(),
  },
  {
    id: "p-2",
    source_id: "src-2",
    name: "Korean Collagen Facial Firming Serum Anti-Aging Glow",
    product_url: "https://shop.tiktok.com",
    image_url: "https://images.unsplash.com/photo-1620916566398-39f1143ab7be?w=400&q=80",
    category_id: "cat-2",
    category_name: "Beauty & Personal Care",
    price: 390.0,
    commission_rate: 30.0,
    total_sales: 9800,
    delta_sales: 240,
    velocity_per_hour: 16.0,
    expected_return: 117.0,
    winning_score: 91.2,
    snapshot_time: new Date().toISOString(),
    created_at: new Date().toISOString(),
  },
  {
    id: "p-3",
    source_id: "src-3",
    name: "Smart Water Bottle with LED Temperature Display 500ml",
    product_url: "https://shop.tiktok.com",
    image_url: "https://images.unsplash.com/photo-1602143407151-7111542de6e8?w=400&q=80",
    category_id: "cat-3",
    category_name: "Home & Kitchen",
    price: 299.0,
    commission_rate: 22.0,
    total_sales: 6400,
    delta_sales: 180,
    velocity_per_hour: 11.2,
    expected_return: 65.78,
    winning_score: 84.5,
    snapshot_time: new Date().toISOString(),
    created_at: new Date().toISOString(),
  },
  {
    id: "p-4",
    source_id: "src-4",
    name: "Oversized Cotton Graphic Vintage Drop-Shoulder Tee",
    product_url: "https://shop.tiktok.com",
    image_url: "https://images.unsplash.com/photo-1521572267360-ee0c2909d518?w=400&q=80",
    category_id: "cat-4",
    category_name: "Fashion & Apparel",
    price: 350.0,
    commission_rate: 20.0,
    total_sales: 8500,
    delta_sales: 120,
    velocity_per_hour: 9.4,
    expected_return: 70.0,
    winning_score: 78.6,
    snapshot_time: new Date().toISOString(),
    created_at: new Date().toISOString(),
  },
];

export default function CatalogPage() {
  const { accessToken } = useAuth();
  const [categories, setCategories] = useState<Category[]>([]);
  const [catalogData, setCatalogData] = useState<ProductCatalogResponse>({
    page: 1,
    limit: 20,
    total_count: 0,
    total_pages: 0,
    items: [],
  });
  const [filters, setFilters] = useState<CatalogFilterState>({
    q: "",
    category_id: "",
    min_price: undefined,
    max_price: undefined,
    min_commission: undefined,
    sort_by: "winning_score",
    sort_order: "desc",
    page: 1,
    limit: 20,
  });
  const [isLoading, setIsLoading] = useState(true);
  const [isUsingFallback, setIsUsingFallback] = useState(false);
  const [selectedProduct, setSelectedProduct] = useState<{ id: string; name: string } | null>(null);

  // Load Categories once on mount
  useEffect(() => {
    async function loadCategories() {
      try {
        const cats = await fetchCategories();
        if (cats && cats.length > 0) {
          setCategories(cats);
        } else {
          setCategories(MOCK_CATEGORIES);
        }
      } catch {
        setCategories(MOCK_CATEGORIES);
      }
    }
    loadCategories();
  }, []);

  // Fetch catalog data when filters change
  const loadCatalog = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await fetchCatalog(filters, accessToken || undefined);
      if (data && data.items && data.items.length > 0) {
        setCatalogData(data);
        setIsUsingFallback(false);
      } else {
        // Fallback for preview mode if live data is empty
        setCatalogData({
          page: 1,
          limit: filters.limit || 20,
          total_count: MOCK_CATALOG_ITEMS.length,
          total_pages: 1,
          items: MOCK_CATALOG_ITEMS,
        });
        setIsUsingFallback(true);
      }
    } catch {
      setCatalogData({
        page: 1,
        limit: filters.limit || 20,
        total_count: MOCK_CATALOG_ITEMS.length,
        total_pages: 1,
        items: MOCK_CATALOG_ITEMS,
      });
      setIsUsingFallback(true);
    } finally {
      setIsLoading(false);
    }
  }, [filters, accessToken]);

  useEffect(() => {
    loadCatalog();
  }, [loadCatalog]);

  const handleFilterChange = (newFilters: Partial<CatalogFilterState>) => {
    setFilters((prev) => ({ ...prev, ...newFilters }));
  };

  const handleReset = () => {
    setFilters({
      q: "",
      category_id: "",
      min_price: undefined,
      max_price: undefined,
      min_commission: undefined,
      sort_by: "winning_score",
      sort_order: "desc",
      page: 1,
      limit: 20,
    });
  };

  const handleSort = (column: string) => {
    setFilters((prev) => {
      if (prev.sort_by === column) {
        return {
          ...prev,
          sort_order: prev.sort_order === "asc" ? "desc" : "asc",
          page: 1,
        };
      }
      return {
        ...prev,
        sort_by: column,
        sort_order: "desc",
        page: 1,
      };
    });
  };

  const handlePageChange = (newPage: number) => {
    setFilters((prev) => ({ ...prev, page: newPage }));
  };

  const handlePageSizeChange = (newSize: number) => {
    setFilters((prev) => ({ ...prev, limit: newSize, page: 1 }));
  };

  return (
    <div className="container mx-auto px-4 sm:px-6 py-8 max-w-7xl space-y-6">
      {/* Header Banner */}
      <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-4 pb-6 border-b border-border/50">
        <div>
          <div className="inline-flex items-center gap-2 rounded-full bg-secondary/80 border border-border/60 px-3 py-1 text-xs font-semibold text-[#25f4ee] mb-3">
            <Sparkles className="h-3.5 w-3.5" />
            <span>Search & Filter Catalog</span>
          </div>
          <h1 className="text-3xl font-extrabold tracking-tight text-foreground flex items-center gap-2.5">
            <ShoppingBag className="h-7 w-7 text-primary" />
            <span>Product Catalog</span>
          </h1>
          <p className="mt-1 text-sm text-muted-foreground max-w-2xl">
            Filter TikTok Shop affiliate opportunities by category, price range, commission yields, and algorithmic score.
          </p>
        </div>
      </div>

      {/* Preview Mode Notice */}
      {isUsingFallback && (
        <div className="flex items-center gap-3 rounded-xl bg-amber-500/10 border border-amber-500/20 p-4 text-xs text-amber-200">
          <AlertCircle className="h-4 w-4 flex-shrink-0 text-amber-400" />
          <p>
            <strong>Preview Mode:</strong> Displaying simulated product records. As soon as the scraper runs, live items will populate this catalog.
          </p>
        </div>
      )}

      {/* Filter Toolbar */}
      <CatalogFilters
        categories={categories}
        filters={filters}
        onFilterChange={handleFilterChange}
        onReset={handleReset}
      />

      {/* Catalog Table */}
      <CatalogTable
        items={catalogData.items}
        totalCount={catalogData.total_count}
        totalPages={catalogData.total_pages}
        currentPage={filters.page || 1}
        pageSize={filters.limit || 20}
        isLoading={isLoading}
        sortBy={filters.sort_by}
        sortOrder={filters.sort_order}
        onSort={handleSort}
        onPageChange={handlePageChange}
        onPageSizeChange={handlePageSizeChange}
        onSelectProduct={(id, name) => setSelectedProduct({ id, name })}
      />

      {/* Historical Trend Analysis Modal */}
      {selectedProduct && (
        <TrendModal
          productId={selectedProduct.id}
          productName={selectedProduct.name}
          onClose={() => setSelectedProduct(null)}
        />
      )}
    </div>
  );
}

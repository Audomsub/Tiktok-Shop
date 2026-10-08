export interface Category {
  id: string;
  name: string;
  slug: string;
  created_at: string;
}

export interface ProductCatalogItem {
  id: string;
  source_id: string;
  name: string;
  product_url: string;
  image_url: string;
  category_id?: string;
  category_name: string;
  price: number;
  commission_rate: number;
  total_sales: number;
  delta_sales: number;
  velocity_per_hour: number;
  expected_return: number;
  winning_score: number;
  snapshot_time: string;
  created_at: string;
}

export interface ProductCatalogResponse {
  page: number;
  limit: number;
  total_count: number;
  total_pages: number;
  items: ProductCatalogItem[];
}

export interface CatalogFilterState {
  q?: string;
  category_id?: string;
  min_price?: number;
  max_price?: number;
  min_commission?: number;
  sort_by?: string;
  sort_order?: "asc" | "desc";
  page?: number;
  limit?: number;
}

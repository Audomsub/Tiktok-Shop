export interface FavoriteItem {
  product_id: string;
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
  favorited_at: string;
}

export interface FavoriteListResponse {
  total: number;
  items: FavoriteItem[];
}

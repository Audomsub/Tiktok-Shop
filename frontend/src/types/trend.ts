export interface TrendPoint {
  snapshot_time: string;
  price: number;
  commission_rate: number;
  total_sales: number;
  delta_sales: number;
  velocity_per_hour: number;
  expected_return: number;
  winning_score: number;
}

export interface ProductTrendsResponse {
  product_id: string;
  product_name: string;
  days: number;
  total_points: number;
  points: TrendPoint[];
}

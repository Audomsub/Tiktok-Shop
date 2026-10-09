export interface LeaderboardBadge {
  is_viral_surge: boolean;     // 🔥 Velocity >= 10 pcs/hr or top 15%
  is_high_commission: boolean; // 💎 Commission Rate >= 20%
  is_high_yield: boolean;      // 💰 Expected Return >= 100 THB
  is_winning_pick: boolean;    // 🏆 Winning Score >= 80
}

export interface LeaderboardItem {
  rank: number;
  product_id: string;
  name: string;
  product_url: string;
  image_url: string;
  category_id?: string;
  category_name: string;
  price: number;
  total_sales: number;
  delta_sales: number;
  velocity_per_hour: number;
  commission_rate: number;
  expected_return: number;
  winning_score: number;
  snapshot_time: string;
  badges: LeaderboardBadge;
  is_favorited?: boolean;
}

export interface LeaderboardResponse {
  crawl_log_id: string;
  crawl_round: string;
  updated_at: string;
  total: number;
  items: LeaderboardItem[];
}

"use client";

import { useEffect, useState } from "react";
import { LeaderboardResponse, LeaderboardItem } from "@/types/leaderboard";
import { fetchLeaderboard } from "@/services/api";
import { LeaderboardCard } from "@/components/leaderboard-card";
import {
  Trophy,
  RefreshCw,
  Clock,
  CheckCircle2,
  AlertCircle,
  Sparkles,
} from "lucide-react";

// Mock data used when backend is offline in preview/dev mode
const MOCK_FALLBACK_LEADERBOARD: LeaderboardResponse = {
  crawl_log_id: "demo-crawl-log",
  crawl_round: "12:00",
  updated_at: new Date().toISOString(),
  total: 6,
  items: [
    {
      rank: 1,
      product_id: "prod-mock-1",
      name: "Wireless ANC Bluetooth Earbuds High-Fidelity Stereo Bass",
      product_url: "https://shop.tiktok.com",
      image_url: "https://images.unsplash.com/photo-1590658268037-6bf12165a8df?w=400&q=80",
      category_name: "Electronics",
      price: 590.0,
      total_sales: 14200,
      delta_sales: 380,
      velocity_per_hour: 24.5,
      commission_rate: 25.0,
      expected_return: 147.5,
      winning_score: 96.8,
      snapshot_time: new Date().toISOString(),
      badges: {
        is_viral_surge: true,
        is_high_commission: true,
        is_high_yield: true,
        is_winning_pick: true,
      },
    },
    {
      rank: 2,
      product_id: "prod-mock-2",
      name: "Korean Collagen Facial Firming Serum Anti-Aging Glow",
      product_url: "https://shop.tiktok.com",
      image_url: "https://images.unsplash.com/photo-1620916566398-39f1143ab7be?w=400&q=80",
      category_name: "Beauty & Personal Care",
      price: 390.0,
      total_sales: 9800,
      delta_sales: 240,
      velocity_per_hour: 16.0,
      commission_rate: 30.0,
      expected_return: 117.0,
      winning_score: 91.2,
      snapshot_time: new Date().toISOString(),
      badges: {
        is_viral_surge: true,
        is_high_commission: true,
        is_high_yield: true,
        is_winning_pick: true,
      },
    },
    {
      rank: 3,
      product_id: "prod-mock-3",
      name: "Smart Water Bottle with LED Temperature Display 500ml",
      product_url: "https://shop.tiktok.com",
      image_url: "https://images.unsplash.com/photo-1602143407151-7111542de6e8?w=400&q=80",
      category_name: "Home & Kitchen",
      price: 299.0,
      total_sales: 6400,
      delta_sales: 180,
      velocity_per_hour: 11.2,
      commission_rate: 22.0,
      expected_return: 65.78,
      winning_score: 84.5,
      snapshot_time: new Date().toISOString(),
      badges: {
        is_viral_surge: true,
        is_high_commission: true,
        is_high_yield: false,
        is_winning_pick: true,
      },
    },
    {
      rank: 4,
      product_id: "prod-mock-4",
      name: "Ergonomic Memory Foam Lumbar Support Cushion for Office",
      product_url: "https://shop.tiktok.com",
      image_url: "https://images.unsplash.com/photo-1584100936595-c0654b55a2e2?w=400&q=80",
      category_name: "Office Supplies",
      price: 450.0,
      total_sales: 3200,
      delta_sales: 90,
      velocity_per_hour: 7.5,
      commission_rate: 20.0,
      expected_return: 90.0,
      winning_score: 79.1,
      snapshot_time: new Date().toISOString(),
      badges: {
        is_viral_surge: false,
        is_high_commission: true,
        is_high_yield: false,
        is_winning_pick: false,
      },
    },
    {
      rank: 5,
      product_id: "prod-mock-5",
      name: "Mini Portable Lint Remover Electric Clothes Fabric Shaver",
      product_url: "https://shop.tiktok.com",
      image_url: "https://images.unsplash.com/photo-1581578731548-c64695cc6952?w=400&q=80",
      category_name: "Home Appliances",
      price: 189.0,
      total_sales: 18200,
      delta_sales: 110,
      velocity_per_hour: 8.8,
      commission_rate: 15.0,
      expected_return: 28.35,
      winning_score: 72.4,
      snapshot_time: new Date().toISOString(),
      badges: {
        is_viral_surge: false,
        is_high_commission: false,
        is_high_yield: false,
        is_winning_pick: false,
      },
    },
    {
      rank: 6,
      product_id: "prod-mock-6",
      name: "UV LED Nail Lamp Quick Drying Gel Polish Curing Light",
      product_url: "https://shop.tiktok.com",
      image_url: "https://images.unsplash.com/photo-1632345031435-8727f6897d53?w=400&q=80",
      category_name: "Beauty & Personal Care",
      price: 320.0,
      total_sales: 5100,
      delta_sales: 60,
      velocity_per_hour: 5.0,
      commission_rate: 25.0,
      expected_return: 80.0,
      winning_score: 70.8,
      snapshot_time: new Date().toISOString(),
      badges: {
        is_viral_surge: false,
        is_high_commission: true,
        is_high_yield: false,
        is_winning_pick: false,
      },
    },
  ],
};

export default function LeaderboardPage() {
  const [data, setData] = useState<LeaderboardResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isUsingFallback, setIsUsingFallback] = useState(false);

  const loadData = async (isManualRefresh = false) => {
    if (isManualRefresh) setIsRefreshing(true);
    try {
      const result = await fetchLeaderboard(10);
      // If backend returned items, use them; if empty, use fallback in demo environment
      if (result && result.items && result.items.length > 0) {
        setData(result);
        setIsUsingFallback(false);
      } else {
        setData(MOCK_FALLBACK_LEADERBOARD);
        setIsUsingFallback(true);
      }
    } catch {
      // Graceful fallback to demo data if backend connection is unavailable
      setData(MOCK_FALLBACK_LEADERBOARD);
      setIsUsingFallback(true);
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  return (
    <div className="container mx-auto px-4 sm:px-6 py-8 max-w-7xl">
      {/* Header Banner */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-6 pb-8 border-b border-border/50">
        <div>
          <div className="inline-flex items-center gap-2 rounded-full bg-secondary/80 border border-border/60 px-3 py-1 text-xs font-semibold text-[#25f4ee] mb-3">
            <Sparkles className="h-3.5 w-3.5" />
            <span>AI Winning Radar</span>
          </div>
          <h1 className="text-3xl sm:text-4xl font-extrabold tracking-tight text-foreground">
            Top 10 Winning Leaderboard
          </h1>
          <p className="mt-2 text-sm sm:text-base text-muted-foreground max-w-2xl">
            Real-time algorithmic ranking weighted by sales velocity (50%), commission rate (30%), and expected return (20%).
          </p>
        </div>

        {/* Metadata Badges & Refresh Action */}
        <div className="flex flex-wrap items-center gap-3">
          {data && (
            <div className="flex items-center gap-2 rounded-xl bg-secondary/60 border border-border/60 px-3.5 py-2 text-xs text-muted-foreground">
              <Clock className="h-4 w-4 text-primary" />
              <span>
                Round: <strong className="text-foreground">{data.crawl_round || "Latest"}</strong>
              </span>
              <span className="text-border">|</span>
              <span>
                Ranked: <strong className="text-foreground">{data.total} Items</strong>
              </span>
            </div>
          )}

          <button
            onClick={() => loadData(true)}
            disabled={isLoading || isRefreshing}
            className="inline-flex items-center gap-2 rounded-xl bg-primary hover:bg-primary/90 text-primary-foreground px-4 py-2 text-xs font-bold shadow-md shadow-primary/20 transition-all disabled:opacity-50"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${isRefreshing ? "animate-spin" : ""}`} />
            <span>{isRefreshing ? "Updating..." : "Refresh"}</span>
          </button>
        </div>
      </div>

      {/* Demo Warning Notice if Go server is not yet returning crawled live data */}
      {isUsingFallback && (
        <div className="my-6 flex items-center gap-3 rounded-xl bg-amber-500/10 border border-amber-500/20 p-4 text-xs text-amber-200">
          <AlertCircle className="h-4 w-4 flex-shrink-0 text-amber-400" />
          <p>
            <strong>Preview Mode:</strong> Displaying simulated analytics items. Once the n8n crawler performs a live scraping cycle and stores snapshots, real-time products will display automatically.
          </p>
        </div>
      )}

      {/* Loading Skeleton */}
      {isLoading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 pt-8">
          {[...Array(6)].map((_, i) => (
            <div
              key={i}
              className="h-72 rounded-2xl bg-secondary/30 animate-pulse border border-border/40 p-5 flex flex-col justify-between"
            >
              <div className="space-y-3">
                <div className="h-6 w-24 bg-secondary/70 rounded-full" />
                <div className="h-16 w-full bg-secondary/50 rounded-xl" />
                <div className="h-4 w-3/4 bg-secondary/50 rounded" />
              </div>
              <div className="h-10 w-full bg-secondary/70 rounded-xl" />
            </div>
          ))}
        </div>
      ) : data?.items && data.items.length > 0 ? (
        /* Leaderboard Cards Grid */
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 pt-8">
          {data.items.map((item: LeaderboardItem) => (
            <LeaderboardCard key={item.product_id} item={item} />
          ))}
        </div>
      ) : (
        /* Empty State */
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <div className="h-14 w-14 rounded-2xl bg-secondary flex items-center justify-center text-muted-foreground mb-4">
            <Trophy className="h-7 w-7" />
          </div>
          <h3 className="text-lg font-bold text-foreground">No Snapshot Records Found</h3>
          <p className="mt-1 text-sm text-muted-foreground max-w-md">
            The database does not have completed crawl rounds yet. Trigger the crawler to populate product data.
          </p>
        </div>
      )}
    </div>
  );
}

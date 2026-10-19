"use client";

import { useEffect, useState, useMemo } from "react";
import Link from "next/link";
import { LeaderboardResponse, LeaderboardItem } from "@/types/leaderboard";
import { fetchLeaderboard } from "@/services/api";
import { useAuth } from "@/context/auth-context";
import { LeaderboardCard } from "@/components/leaderboard-card";
import { TrendModal } from "@/components/trend-modal";
import {
  Trophy,
  RefreshCw,
  Clock,
  AlertCircle,
  Sparkles,
  Search,
  X,
  ShoppingBag,
  ArrowRight,
} from "lucide-react";

// Mock data used ONLY when backend connection is unavailable
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
  ],
};

export default function LeaderboardPage() {
  const { accessToken } = useAuth();
  const [data, setData] = useState<LeaderboardResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isUsingFallback, setIsUsingFallback] = useState(false);
  const [selectedProduct, setSelectedProduct] = useState<{ id: string; name: string } | null>(null);

  // Search & Filter state on homepage
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedCategory, setSelectedCategory] = useState("");

  const loadData = async (isManualRefresh = false) => {
    if (isManualRefresh) setIsRefreshing(true);
    try {
      const result = await fetchLeaderboard(20, accessToken || undefined);
      if (result && result.items && result.items.length > 0) {
        setData(result);
        setIsUsingFallback(false);
      } else {
        setData(MOCK_FALLBACK_LEADERBOARD);
        setIsUsingFallback(true);
      }
    } catch {
      setData(MOCK_FALLBACK_LEADERBOARD);
      setIsUsingFallback(true);
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [accessToken]);

  // Extract unique categories from items
  const uniqueCategories = useMemo(() => {
    if (!data?.items) return [];
    const set = new Set<string>();
    data.items.forEach((item) => {
      if (item.category_name) set.add(item.category_name);
    });
    return Array.from(set).sort();
  }, [data?.items]);

  // Live filter leaderboard items
  const filteredItems = useMemo(() => {
    if (!data?.items) return [];
    return data.items.filter((item) => {
      const q = searchQuery.toLowerCase().trim();
      const matchesQuery =
        !q ||
        item.name.toLowerCase().includes(q) ||
        (item.category_name && item.category_name.toLowerCase().includes(q));
      const matchesCat = !selectedCategory || item.category_name === selectedCategory;
      return matchesQuery && matchesCat;
    });
  }, [data?.items, searchQuery, selectedCategory]);

  return (
    <div className="container mx-auto px-4 sm:px-6 py-8 max-w-7xl space-y-6">
      {/* Header Banner */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-6 pb-6 border-b border-border/50">
        <div>
          <div className="inline-flex items-center gap-2 rounded-full bg-secondary/80 border border-border/60 px-3 py-1 text-xs font-semibold text-[#25f4ee] mb-3">
            <Sparkles className="h-3.5 w-3.5" />
            <span>AI Winning Radar</span>
          </div>
          <h1 className="text-3xl sm:text-4xl font-extrabold tracking-tight text-foreground">
            Top Winning Leaderboard
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

      {/* Offline Mode Warning */}
      {isUsingFallback && (
        <div className="flex items-center gap-3 rounded-xl bg-amber-500/10 border border-amber-500/20 p-4 text-xs text-amber-200">
          <AlertCircle className="h-4 w-4 flex-shrink-0 text-amber-400" />
          <p>
            <strong>Preview Mode:</strong> Displaying simulated analytics items. Once the n8n crawler performs a live scraping cycle and stores snapshots, real-time products will display automatically.
          </p>
        </div>
      )}

      {/* Real-time Search & Filter Toolbar on Homepage */}
      <div className="rounded-2xl bg-card border border-border/60 p-4 shadow-sm">
        <div className="grid grid-cols-1 md:grid-cols-12 gap-3 items-center">
          {/* Search Box */}
          <div className="relative md:col-span-6">
            <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <input
              type="text"
              placeholder="Search winning products by title or keyword..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full rounded-xl bg-secondary/50 border border-border/60 pl-10 pr-10 py-2.5 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/40 focus:border-primary transition-all"
            />
            {searchQuery && (
              <button
                type="button"
                onClick={() => setSearchQuery("")}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
              >
                <X className="h-4 w-4" />
              </button>
            )}
          </div>

          {/* Category Filter */}
          <div className="md:col-span-3">
            <select
              value={selectedCategory}
              onChange={(e) => setSelectedCategory(e.target.value)}
              className="w-full rounded-xl bg-secondary/50 border border-border/60 px-3.5 py-2.5 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary/40 focus:border-primary transition-all cursor-pointer"
            >
              <option value="">All Categories</option>
              {uniqueCategories.map((cat) => (
                <option key={cat} value={cat}>
                  {cat}
                </option>
              ))}
            </select>
          </div>

          {/* Link to Full Catalog */}
          <div className="md:col-span-3 flex justify-end">
            <Link
              href={searchQuery ? `/catalog?q=${encodeURIComponent(searchQuery)}` : "/catalog"}
              className="w-full inline-flex items-center justify-center gap-2 rounded-xl bg-secondary hover:bg-secondary/80 text-foreground px-4 py-2.5 text-xs font-semibold border border-border/60 transition-colors"
            >
              <ShoppingBag className="h-4 w-4 text-[#25f4ee]" />
              <span>Explore Full Catalog</span>
              <ArrowRight className="h-3.5 w-3.5" />
            </Link>
          </div>
        </div>
      </div>

      {/* Loading Skeleton */}
      {isLoading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 pt-4">
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
      ) : filteredItems.length > 0 ? (
        /* Leaderboard Cards Grid */
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 pt-4">
          {filteredItems.map((item: LeaderboardItem) => (
            <LeaderboardCard
              key={item.product_id}
              item={item}
              onSelectProduct={(id, name) => setSelectedProduct({ id, name })}
            />
          ))}
        </div>
      ) : (
        /* Search Empty State */
        <div className="rounded-3xl border border-dashed border-border/80 bg-card/40 p-12 text-center max-w-xl mx-auto my-8">
          <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-secondary/80 border border-border/60 text-muted-foreground mb-4">
            <Trophy className="h-7 w-7 stroke-1" />
          </div>
          <h3 className="text-lg font-bold text-foreground">No Leaderboard Items Match</h3>
          <p className="mt-2 text-xs sm:text-sm text-muted-foreground leading-relaxed">
            No products in the Top Leaderboard match &quot;{searchQuery}&quot;. You can clear your search or explore the full catalog containing all products.
          </p>
          <div className="mt-6 flex flex-wrap items-center justify-center gap-3">
            <button
              type="button"
              onClick={() => {
                setSearchQuery("");
                setSelectedCategory("");
              }}
              className="inline-flex items-center gap-1.5 rounded-xl bg-secondary hover:bg-secondary/80 text-foreground px-4 py-2 text-xs font-semibold border border-border/60 transition-colors"
            >
              <RefreshCw className="h-3.5 w-3.5" />
              <span>Clear Filter</span>
            </button>
            <Link
              href={searchQuery ? `/catalog?q=${encodeURIComponent(searchQuery)}` : "/catalog"}
              className="inline-flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-[#fe2c55] to-[#fe2c55]/90 hover:opacity-95 text-white px-4 py-2 text-xs font-semibold shadow-md shadow-[#fe2c55]/20 transition-all"
            >
              <ShoppingBag className="h-3.5 w-3.5" />
              <span>Search in Full Catalog</span>
            </Link>
          </div>
        </div>
      )}

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

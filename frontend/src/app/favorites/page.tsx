"use client";

import React, { useEffect, useState, useMemo } from "react";
import Link from "next/link";
import { useAuth } from "@/context/auth-context";
import { useFavorites } from "@/context/favorites-context";
import { fetchFavorites } from "@/services/api";
import { FavoriteItem } from "@/types/favorite";
import { TrendModal } from "@/components/trend-modal";
import { formatTHB, formatNumber } from "@/lib/utils";
import {
  Bookmark,
  Heart,
  TrendingUp,
  ExternalLink,
  Activity,
  ShoppingBag,
  Zap,
  Sparkles,
  LogIn,
  Trash2,
  Percent,
  Coins,
  RefreshCw,
  AlertCircle,
  ArrowRight,
} from "lucide-react";

export default function FavoritesPage() {
  const { user, accessToken, isLoading: isAuthLoading, openAuthModal } = useAuth();
  const { toggleFavorite } = useFavorites();

  const [favorites, setFavorites] = useState<FavoriteItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedProduct, setSelectedProduct] = useState<{ id: string; name: string } | null>(null);

  useEffect(() => {
    let active = true;

    async function loadFavorites() {
      if (!accessToken) {
        setIsLoading(false);
        setFavorites([]);
        return;
      }

      setIsLoading(true);
      setError(null);

      try {
        const res = await fetchFavorites(accessToken);
        if (active) {
          setFavorites(res.items || []);
        }
      } catch (err: unknown) {
        const errObj = err as Error;
        console.error("Failed to load favorites list:", errObj);
        if (active) {
          setError(errObj.message || "Failed to load bookmarks.");
        }
      } finally {
        if (active) {
          setIsLoading(false);
        }
      }
    }

    if (!isAuthLoading) {
      loadFavorites();
    }

    return () => {
      active = false;
    };
  }, [accessToken, isAuthLoading]);

  // Handle optimistic removal from local view when unfavoriting
  const handleRemoveFavorite = async (productId: string) => {
    // Optimistically filter out from local view
    setFavorites((prev) => prev.filter((item) => item.product_id !== productId));
    await toggleFavorite(productId);
  };

  // Aggregated KPIs
  const kpiStats = useMemo(() => {
    if (favorites.length === 0) {
      return { total: 0, avgCommission: 0, avgPrice: 0, totalYield: 0, maxScore: 0 };
    }
    const total = favorites.length;
    const avgCommission = favorites.reduce((sum, item) => sum + item.commission_rate, 0) / total;
    const avgPrice = favorites.reduce((sum, item) => sum + item.price, 0) / total;
    const totalYield = favorites.reduce((sum, item) => sum + item.expected_return, 0);
    const maxScore = Math.max(...favorites.map((item) => item.winning_score));

    return { total, avgCommission, avgPrice, totalYield, maxScore };
  }, [favorites]);

  // 1. Loading State (Auth Check)
  if (isAuthLoading) {
    return (
      <div className="container mx-auto px-4 sm:px-6 py-12 flex flex-col items-center justify-center min-h-[50vh]">
        <div className="h-10 w-10 animate-spin rounded-full border-4 border-[#fe2c55] border-t-transparent mb-4" />
        <p className="text-sm text-muted-foreground animate-pulse">Checking creator session...</p>
      </div>
    );
  }

  // 2. Unauthenticated State (Prompt User to Sign In)
  if (!user) {
    return (
      <div className="container mx-auto px-4 sm:px-6 py-16 max-w-3xl">
        <div className="rounded-3xl border border-border/60 bg-gradient-to-b from-card to-card/60 p-8 sm:p-12 text-center shadow-xl relative overflow-hidden">
          {/* Subtle background glow */}
          <div className="absolute top-0 left-1/2 -translate-x-1/2 w-96 h-48 bg-gradient-to-r from-[#fe2c55]/15 to-[#25f4ee]/15 blur-3xl pointer-events-none" />

          <div className="relative z-10 flex flex-col items-center">
            <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-tr from-[#fe2c55] to-[#25f4ee] p-0.5 shadow-xl shadow-[#fe2c55]/20 mb-6">
              <div className="flex h-full w-full items-center justify-center rounded-[14px] bg-background">
                <Bookmark className="h-8 w-8 text-[#fe2c55]" />
              </div>
            </div>

            <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight text-foreground">
              Sign In to View Bookmarks
            </h1>
            <p className="mt-3 text-sm text-muted-foreground max-w-lg leading-relaxed">
              Save high-potential TikTok affiliate items to your personal watchlist, monitor hourly momentum curves, and plan your production pipeline in one place.
            </p>

            <div className="mt-8 flex flex-col sm:flex-row items-center gap-3">
              <button
                type="button"
                onClick={() => openAuthModal("Sign in to access and manage your bookmarked winning products.")}
                className="w-full sm:w-auto inline-flex items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-[#fe2c55] to-[#fe2c55]/90 hover:opacity-95 text-white px-6 py-3 text-sm font-semibold shadow-lg shadow-[#fe2c55]/25 transition-all"
              >
                <LogIn className="h-4 w-4" />
                <span>Sign In / Create Account</span>
              </button>

              <Link
                href="/"
                className="w-full sm:w-auto inline-flex items-center justify-center gap-2 rounded-xl bg-secondary hover:bg-secondary/80 text-foreground px-6 py-3 text-sm font-semibold border border-border/60 transition-colors"
              >
                <span>Explore Leaderboard</span>
                <ArrowRight className="h-4 w-4" />
              </Link>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // 3. Authenticated State
  return (
    <div className="container mx-auto px-4 sm:px-6 py-8">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-8">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight text-foreground">
              My Saved Bookmarks
            </h1>
            <span className="inline-flex items-center rounded-full bg-[#fe2c55]/10 px-2.5 py-0.5 text-xs font-semibold text-[#fe2c55] border border-[#fe2c55]/20">
              {favorites.length} saved
            </span>
          </div>
          <p className="text-sm text-muted-foreground mt-1">
            Tracked products for your upcoming TikTok video showcases and affiliate campaigns.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Link
            href="/catalog"
            className="inline-flex items-center gap-2 rounded-xl bg-secondary hover:bg-secondary/80 text-foreground px-4 py-2 text-xs font-semibold border border-border/60 transition-colors"
          >
            <ShoppingBag className="h-3.5 w-3.5" />
            <span>Discover More Products</span>
          </Link>
        </div>
      </div>

      {/* KPI Stats Ribbon */}
      {favorites.length > 0 && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3 sm:gap-4 mb-8">
          <div className="rounded-2xl bg-card border border-border/60 p-4 shadow-sm">
            <span className="block text-xs font-medium text-muted-foreground">Total Bookmarked</span>
            <div className="flex items-baseline gap-2 mt-1">
              <span className="text-2xl font-black text-foreground">{kpiStats.total}</span>
              <span className="text-xs text-muted-foreground">items</span>
            </div>
          </div>

          <div className="rounded-2xl bg-card border border-border/60 p-4 shadow-sm">
            <span className="block text-xs font-medium text-muted-foreground">Avg Commission Rate</span>
            <div className="flex items-baseline gap-2 mt-1">
              <span className="text-2xl font-black text-cyan-300">
                {kpiStats.avgCommission.toFixed(1)}%
              </span>
              <span className="text-xs text-muted-foreground">per sale</span>
            </div>
          </div>

          <div className="rounded-2xl bg-card border border-border/60 p-4 shadow-sm">
            <span className="block text-xs font-medium text-muted-foreground">Total Potential Yield</span>
            <div className="flex items-baseline gap-2 mt-1">
              <span className="text-2xl font-black text-emerald-400">
                {formatTHB(kpiStats.totalYield)}
              </span>
            </div>
          </div>

          <div className="rounded-2xl bg-card border border-border/60 p-4 shadow-sm">
            <span className="block text-xs font-medium text-muted-foreground">Top Winning Score</span>
            <div className="flex items-baseline gap-2 mt-1">
              <span className="text-2xl font-black text-[#25f4ee]">
                {kpiStats.maxScore.toFixed(1)}
              </span>
              <span className="text-xs text-muted-foreground">/ 100</span>
            </div>
          </div>
        </div>
      )}

      {/* Error Notice */}
      {error && (
        <div className="mb-6 rounded-2xl bg-destructive/15 border border-destructive/30 p-4 flex items-center gap-3 text-sm text-destructive">
          <AlertCircle className="h-5 w-5 flex-shrink-0" />
          <p>{error}</p>
        </div>
      )}

      {/* Loading Skeleton */}
      {isLoading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          {[...Array(6)].map((_, i) => (
            <div
              key={i}
              className="rounded-2xl bg-card border border-border/40 p-5 space-y-4 animate-pulse"
            >
              <div className="flex gap-4">
                <div className="h-20 w-20 rounded-xl bg-secondary/70 flex-shrink-0" />
                <div className="flex-1 space-y-2">
                  <div className="h-4 w-3/4 bg-secondary/70 rounded" />
                  <div className="h-3 w-1/2 bg-secondary/50 rounded" />
                </div>
              </div>
              <div className="h-16 bg-secondary/40 rounded-xl" />
            </div>
          ))}
        </div>
      ) : favorites.length === 0 ? (
        /* Empty State */
        <div className="rounded-3xl border border-dashed border-border/80 bg-card/40 p-12 text-center max-w-xl mx-auto my-8">
          <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-secondary/80 border border-border/60 text-muted-foreground mb-4">
            <Bookmark className="h-7 w-7 stroke-1" />
          </div>
          <h3 className="text-lg font-bold text-foreground">No Bookmarked Products Yet</h3>
          <p className="mt-2 text-xs sm:text-sm text-muted-foreground leading-relaxed">
            You haven&apos;t added any products to your bookmarks. Browse the real-time Winning Leaderboard or Product Catalog to discover and bookmark trending items.
          </p>
          <div className="mt-6 flex flex-wrap items-center justify-center gap-3">
            <Link
              href="/"
              className="inline-flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-[#fe2c55] to-[#fe2c55]/90 hover:opacity-95 text-white px-4 py-2 text-xs font-semibold shadow-md shadow-[#fe2c55]/20 transition-all"
            >
              <Sparkles className="h-3.5 w-3.5" />
              <span>Winning Leaderboard</span>
            </Link>
            <Link
              href="/catalog"
              className="inline-flex items-center gap-1.5 rounded-xl bg-secondary hover:bg-secondary/80 text-foreground px-4 py-2 text-xs font-semibold border border-border/60 transition-colors"
            >
              <ShoppingBag className="h-3.5 w-3.5" />
              <span>Product Catalog</span>
            </Link>
          </div>
        </div>
      ) : (
        /* Favorites Grid */
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          {favorites.map((item) => (
            <div
              key={item.product_id}
              className="group relative flex flex-col justify-between overflow-hidden rounded-2xl bg-card border border-border/60 hover:border-border transition-all duration-300 hover:-translate-y-1 hover:shadow-xl"
            >
              {/* Card Body */}
              <div className="p-5 pb-3">
                {/* Header: Category & Score */}
                <div className="flex items-center justify-between gap-2 mb-3">
                  <span className="inline-flex items-center rounded-lg bg-secondary/80 px-2.5 py-1 text-[11px] font-semibold text-muted-foreground border border-border/40">
                    {item.category_name || "General"}
                  </span>

                  <div className="flex items-center gap-2">
                    <div className="flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-[#fe2c55]/20 to-[#25f4ee]/20 border border-[#fe2c55]/30 px-2.5 py-1 shadow-sm">
                      <Zap className="h-3.5 w-3.5 text-[#25f4ee] fill-[#25f4ee]" />
                      <span className="text-xs font-extrabold text-foreground">
                        {item.winning_score.toFixed(1)}
                      </span>
                    </div>

                    {/* Unfavorite / Remove Button */}
                    <button
                      type="button"
                      onClick={() => handleRemoveFavorite(item.product_id)}
                      className="rounded-xl p-1.5 bg-[#fe2c55]/15 border border-[#fe2c55]/30 text-[#fe2c55] hover:bg-destructive/20 hover:border-destructive/40 transition-colors"
                      title="Remove from bookmarks"
                      aria-label="Remove from bookmarks"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>
                </div>

                {/* Product Info */}
                <div className="flex gap-4">
                  <div className="relative h-20 w-20 flex-shrink-0 overflow-hidden rounded-xl bg-secondary/50 border border-border/40">
                    {item.image_url ? (
                      // eslint-disable-next-line @next/next/no-img-element
                      <img
                        src={item.image_url}
                        alt={item.name}
                        className="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
                      />
                    ) : (
                      <div className="flex h-full w-full items-center justify-center text-muted-foreground">
                        <ShoppingBag className="h-8 w-8 stroke-1" />
                      </div>
                    )}
                  </div>

                  <div className="flex flex-col justify-center min-w-0 flex-1">
                    <h3 className="line-clamp-2 text-sm font-semibold text-foreground group-hover:text-primary transition-colors leading-snug">
                      {item.name}
                    </h3>
                    <p className="mt-1 text-xs text-muted-foreground">
                      Total Sales:{" "}
                      <span className="font-semibold text-foreground">
                        {formatNumber(item.total_sales)}
                      </span>{" "}
                      units
                    </p>
                  </div>
                </div>
              </div>

              {/* Metrics Matrix */}
              <div className="p-5 pt-0">
                <div className="grid grid-cols-3 gap-2 rounded-xl bg-secondary/40 p-3 border border-border/40 text-center mb-4">
                  <div>
                    <span className="block text-[10px] text-muted-foreground uppercase font-medium">
                      Price
                    </span>
                    <span className="text-xs font-bold text-foreground">
                      {formatTHB(item.price)}
                    </span>
                  </div>
                  <div>
                    <span className="block text-[10px] text-muted-foreground uppercase font-medium">
                      Speed
                    </span>
                    <span className="inline-flex items-center justify-center gap-0.5 text-xs font-bold text-emerald-400">
                      <TrendingUp className="h-3 w-3" />
                      {item.velocity_per_hour > 0 ? `+${item.velocity_per_hour.toFixed(1)}` : "0.0"}
                      /hr
                    </span>
                  </div>
                  <div>
                    <span className="block text-[10px] text-muted-foreground uppercase font-medium">
                      Est. Yield
                    </span>
                    <span className="text-xs font-bold text-amber-300">
                      {formatTHB(item.expected_return)}
                    </span>
                  </div>
                </div>

                {/* Actions */}
                <div className="flex gap-2">
                  <button
                    type="button"
                    onClick={() =>
                      setSelectedProduct({ id: item.product_id, name: item.name })
                    }
                    className="flex-1 inline-flex items-center justify-center gap-1.5 rounded-xl bg-secondary hover:bg-secondary/80 py-2 text-xs font-semibold text-foreground border border-border/60 transition-colors"
                  >
                    <Activity className="h-3.5 w-3.5 text-[#25f4ee]" />
                    <span>Trends</span>
                  </button>

                  <a
                    href={item.product_url || `https://shop.tiktok.com`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex-1 inline-flex items-center justify-center gap-1.5 rounded-xl bg-gradient-to-r from-secondary to-secondary/80 hover:from-[#fe2c55] hover:to-[#fe2c55]/90 hover:text-white py-2 text-xs font-semibold text-foreground transition-all duration-200 border border-border/60 hover:border-transparent"
                  >
                    <span>Shop</span>
                    <ExternalLink className="h-3.5 w-3.5" />
                  </a>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Historical Trend Chart Modal */}
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

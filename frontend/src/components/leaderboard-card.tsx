"use client";

import React, { useState } from "react";
import { LeaderboardItem } from "@/types/leaderboard";
import { formatTHB, formatNumber } from "@/lib/utils";
import { useFavorites } from "@/context/favorites-context";
import {
  Flame,
  Percent,
  Coins,
  Trophy,
  ExternalLink,
  Zap,
  ShoppingBag,
  TrendingUp,
  Activity,
  Heart,
} from "lucide-react";

interface LeaderboardCardProps {
  item: LeaderboardItem;
  onSelectProduct?: (productId: string, productName: string) => void;
}

export function LeaderboardCard({ item, onSelectProduct }: LeaderboardCardProps) {
  const [imageError, setImageError] = useState(false);
  const { isFavorited, toggleFavorite } = useFavorites();
  const favorited = isFavorited(item.product_id) || Boolean(item.is_favorited && !isFavorited(item.product_id) ? false : isFavorited(item.product_id));

  // Rank styling configuration
  const getRankBadge = (rank: number) => {
    switch (rank) {
      case 1:
        return {
          label: "1st Place",
          badgeClass: "bg-amber-500/20 text-amber-300 border-amber-500/40 shadow-amber-500/20",
          medal: "🥇",
          glow: "border-amber-500/30 hover:border-amber-500/60 shadow-lg shadow-amber-500/5",
        };
      case 2:
        return {
          label: "2nd Place",
          badgeClass: "bg-slate-300/20 text-slate-200 border-slate-300/40 shadow-slate-300/20",
          medal: "🥈",
          glow: "border-slate-400/30 hover:border-slate-400/60 shadow-lg shadow-slate-400/5",
        };
      case 3:
        return {
          label: "3rd Place",
          badgeClass: "bg-amber-700/20 text-amber-500 border-amber-700/40 shadow-amber-700/20",
          medal: "🥉",
          glow: "border-amber-700/30 hover:border-amber-700/60 shadow-lg shadow-amber-700/5",
        };
      default:
        return {
          label: `#${rank}`,
          badgeClass: "bg-secondary text-muted-foreground border-border/50",
          medal: `#${rank}`,
          glow: "border-border/60 hover:border-border",
        };
    }
  };

  const rankInfo = getRankBadge(item.rank);

  return (
    <div
      className={`group relative flex flex-col justify-between overflow-hidden rounded-2xl bg-card border ${rankInfo.glow} transition-all duration-300 hover:-translate-y-1 hover:shadow-xl`}
    >
      {/* Top Banner & Rank Medal */}
      <div className="relative p-5 pb-3">
        <div className="flex items-center justify-between gap-2 mb-3">
          <div
            className={`inline-flex items-center gap-1.5 rounded-full border px-3 py-1 text-xs font-bold shadow-sm ${rankInfo.badgeClass}`}
          >
            <span>{rankInfo.medal}</span>
            <span>{rankInfo.label}</span>
          </div>

          <div className="flex items-center gap-2">
            {/* Winning Score Badge */}
            <div className="flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-[#fe2c55]/20 to-[#25f4ee]/20 border border-[#fe2c55]/30 px-3 py-1 shadow-sm">
              <Zap className="h-3.5 w-3.5 text-[#25f4ee] fill-[#25f4ee]" />
              <span className="text-xs font-medium text-muted-foreground">Score</span>
              <span className="text-sm font-extrabold text-foreground tracking-tight">
                {item.winning_score.toFixed(1)}
              </span>
            </div>

            {/* Bookmark Heart Button */}
            <button
              type="button"
              onClick={(e) => {
                e.stopPropagation();
                toggleFavorite(item.product_id);
              }}
              className={`rounded-xl p-1.5 border transition-all duration-200 ${
                favorited
                  ? "bg-[#fe2c55]/20 text-[#fe2c55] border-[#fe2c55]/40 hover:bg-[#fe2c55]/30 shadow-sm shadow-[#fe2c55]/20"
                  : "bg-secondary/60 text-muted-foreground border-border/60 hover:text-[#fe2c55] hover:bg-secondary"
              }`}
              title={favorited ? "Remove from bookmarks" : "Save to bookmarks"}
              aria-label={favorited ? "Remove from bookmarks" : "Save to bookmarks"}
            >
              <Heart
                className={`h-4 w-4 transition-transform active:scale-125 ${
                  favorited ? "fill-[#fe2c55] text-[#fe2c55]" : ""
                }`}
              />
            </button>
          </div>
        </div>

        {/* Product Media & Title */}
        <div className="flex gap-4">
          <div className="relative h-20 w-20 flex-shrink-0 overflow-hidden rounded-xl bg-secondary/50 border border-border/40">
            {item.image_url && !imageError ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img
                src={item.image_url}
                alt={item.name}
                className="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
                onError={() => setImageError(true)}
              />
            ) : (
              <div className="flex h-full w-full items-center justify-center text-muted-foreground">
                <ShoppingBag className="h-8 w-8 stroke-1" />
              </div>
            )}
          </div>

          <div className="flex flex-col justify-center min-w-0 flex-1">
            <span className="text-[11px] font-medium text-muted-foreground uppercase tracking-wider">
              {item.category_name || "General"}
            </span>
            <h3 className="line-clamp-2 text-sm font-semibold text-foreground group-hover:text-primary transition-colors leading-snug">
              {item.name}
            </h3>
            <p className="mt-1 text-xs text-muted-foreground">
              Total Sales: <span className="font-semibold text-foreground">{formatNumber(item.total_sales)}</span> units
            </p>
          </div>
        </div>

        {/* Badges Ribbon (🔥, 💎, 💰, 🏆) */}
        <div className="mt-3 flex flex-wrap gap-1.5">
          {item.badges.is_winning_pick && (
            <span className="inline-flex items-center gap-1 rounded-md bg-yellow-500/10 px-2 py-0.5 text-[11px] font-medium text-yellow-300 border border-yellow-500/25">
              <Trophy className="h-3 w-3 text-yellow-400" />
              Winning Pick
            </span>
          )}
          {item.badges.is_viral_surge && (
            <span className="inline-flex items-center gap-1 rounded-md bg-orange-500/10 px-2 py-0.5 text-[11px] font-medium text-orange-400 border border-orange-500/25">
              <Flame className="h-3 w-3 text-orange-400" />
              Viral Surge
            </span>
          )}
          {item.badges.is_high_commission && (
            <span className="inline-flex items-center gap-1 rounded-md bg-cyan-500/10 px-2 py-0.5 text-[11px] font-medium text-cyan-300 border border-cyan-500/25">
              <Percent className="h-3 w-3 text-cyan-400" />
              High Comm ({item.commission_rate.toFixed(0)}%)
            </span>
          )}
          {item.badges.is_high_yield && (
            <span className="inline-flex items-center gap-1 rounded-md bg-emerald-500/10 px-2 py-0.5 text-[11px] font-medium text-emerald-400 border border-emerald-500/25">
              <Coins className="h-3 w-3 text-emerald-400" />
              High Yield
            </span>
          )}
        </div>
      </div>

      {/* Metrics Matrix */}
      <div className="p-5 pt-0">
        <div className="grid grid-cols-3 gap-2 rounded-xl bg-secondary/40 p-3 border border-border/40 text-center mb-4">
          <div>
            <span className="block text-[10px] text-muted-foreground uppercase font-medium">Price</span>
            <span className="text-xs font-bold text-foreground">{formatTHB(item.price)}</span>
          </div>
          <div>
            <span className="block text-[10px] text-muted-foreground uppercase font-medium">Hourly Speed</span>
            <span className="inline-flex items-center justify-center gap-0.5 text-xs font-bold text-emerald-400">
              <TrendingUp className="h-3 w-3" />
              {item.velocity_per_hour > 0 ? `+${item.velocity_per_hour.toFixed(1)}` : "0.0"}/hr
            </span>
          </div>
          <div>
            <span className="block text-[10px] text-muted-foreground uppercase font-medium">Est. Yield</span>
            <span className="text-xs font-bold text-amber-300">{formatTHB(item.expected_return)}</span>
          </div>
        </div>

        {/* Action Buttons: View Trends & TikTok Shop Link */}
        <div className="flex gap-2">
          {onSelectProduct && (
            <button
              type="button"
              onClick={() => onSelectProduct(item.product_id, item.name)}
              className="flex-1 inline-flex items-center justify-center gap-1.5 rounded-xl bg-secondary hover:bg-secondary/80 py-2.5 text-xs font-semibold text-foreground border border-border/60 transition-colors"
            >
              <Activity className="h-3.5 w-3.5 text-[#25f4ee]" />
              <span>Trends</span>
            </button>
          )}
          <a
            href={item.product_url || `https://shop.tiktok.com`}
            target="_blank"
            rel="noopener noreferrer"
            className="flex-1 inline-flex items-center justify-center gap-1.5 rounded-xl bg-gradient-to-r from-secondary to-secondary/80 hover:from-[#fe2c55] hover:to-[#fe2c55]/90 hover:text-white py-2.5 text-xs font-semibold text-foreground transition-all duration-200 border border-border/60 hover:border-transparent"
          >
            <span>Shop</span>
            <ExternalLink className="h-3.5 w-3.5" />
          </a>
        </div>
      </div>
    </div>
  );
}

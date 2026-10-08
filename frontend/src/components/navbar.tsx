"use client";

import Link from "next/link";
import { TrendingUp, ShoppingBag, Bookmark, Sparkles, Activity } from "lucide-react";

export function Navbar() {
  return (
    <header className="sticky top-0 z-50 w-full border-b border-border/50 bg-background/80 backdrop-blur-md">
      <div className="container mx-auto flex h-16 items-center justify-between px-4 sm:px-6">
        {/* Brand Logo */}
        <Link href="/" className="flex items-center gap-2.5 group">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-tr from-[#fe2c55] to-[#25f4ee] p-0.5 shadow-lg shadow-[#fe2c55]/20 transition-transform group-hover:scale-105">
            <div className="flex h-full w-full items-center justify-center rounded-[10px] bg-background">
              <TrendingUp className="h-5 w-5 text-foreground" />
            </div>
          </div>
          <div>
            <div className="flex items-center gap-1.5">
              <span className="font-bold text-lg tracking-tight bg-gradient-to-r from-white via-slate-200 to-slate-400 bg-clip-text text-transparent">
                AffiliatePulse
              </span>
              <span className="inline-flex items-center rounded-full bg-[#fe2c55]/10 px-2 py-0.5 text-[10px] font-semibold text-[#fe2c55] border border-[#fe2c55]/20">
                PRO
              </span>
            </div>
            <p className="text-[11px] text-muted-foreground hidden sm:block">
              TikTok Shop Winning Analytics
            </p>
          </div>
        </Link>

        {/* Navigation Links */}
        <nav className="flex items-center gap-1 sm:gap-2">
          <Link
            href="/"
            className="flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm font-medium bg-secondary text-foreground shadow-sm transition-colors hover:bg-secondary/80"
          >
            <Sparkles className="h-4 w-4 text-[#25f4ee]" />
            <span>Leaderboard</span>
          </Link>

          <Link
            href="/catalog"
            className="flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground hover:bg-secondary/50"
          >
            <ShoppingBag className="h-4 w-4" />
            <span className="hidden sm:inline">Product Catalog</span>
          </Link>

          <Link
            href="/favorites"
            className="flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground hover:bg-secondary/50"
          >
            <Bookmark className="h-4 w-4" />
            <span className="hidden sm:inline">Bookmarks</span>
          </Link>
        </nav>

        {/* System Status Indicator */}
        <div className="hidden md:flex items-center gap-2 text-xs text-muted-foreground bg-secondary/40 border border-border/40 rounded-full px-3 py-1">
          <span className="relative flex h-2 w-2">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
            <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
          </span>
          <Activity className="h-3.5 w-3.5 text-emerald-400" />
          <span>Real-time Scoring Live</span>
        </div>
      </div>
    </header>
  );
}

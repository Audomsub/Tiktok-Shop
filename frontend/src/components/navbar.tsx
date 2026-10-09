"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useAuth } from "@/context/auth-context";
import { useFavorites } from "@/context/favorites-context";
import {
  TrendingUp,
  ShoppingBag,
  Bookmark,
  Sparkles,
  Activity,
  LogIn,
  LogOut,
  User as UserIcon,
} from "lucide-react";

export function Navbar() {
  const pathname = usePathname();
  const { user, signOut, openAuthModal, isLoading } = useAuth();
  const { favoritesCount } = useFavorites();

  const isHome = pathname === "/";
  const isCatalog = pathname.startsWith("/catalog");
  const isFavorites = pathname.startsWith("/favorites");

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
            className={`flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm font-medium transition-colors ${
              isHome
                ? "bg-secondary text-foreground shadow-sm"
                : "text-muted-foreground hover:text-foreground hover:bg-secondary/50"
            }`}
          >
            <Sparkles className="h-4 w-4 text-[#25f4ee]" />
            <span>Leaderboard</span>
          </Link>

          <Link
            href="/catalog"
            className={`flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm font-medium transition-colors ${
              isCatalog
                ? "bg-secondary text-foreground shadow-sm"
                : "text-muted-foreground hover:text-foreground hover:bg-secondary/50"
            }`}
          >
            <ShoppingBag className="h-4 w-4" />
            <span className="hidden sm:inline">Product Catalog</span>
          </Link>

          <Link
            href="/favorites"
            className={`relative flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm font-medium transition-colors ${
              isFavorites
                ? "bg-secondary text-foreground shadow-sm"
                : "text-muted-foreground hover:text-foreground hover:bg-secondary/50"
            }`}
          >
            <Bookmark className="h-4 w-4" />
            <span className="hidden sm:inline">Bookmarks</span>
            {user && favoritesCount > 0 && (
              <span className="inline-flex items-center justify-center rounded-full bg-[#fe2c55] px-1.5 py-0.2 text-[10px] font-bold text-white min-w-4 text-center">
                {favoritesCount}
              </span>
            )}
          </Link>
        </nav>

        {/* Right Section: System Indicator & Auth Actions */}
        <div className="flex items-center gap-3">
          {/* Real-time Indicator */}
          <div className="hidden lg:flex items-center gap-2 text-xs text-muted-foreground bg-secondary/40 border border-border/40 rounded-full px-3 py-1">
            <span className="relative flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
              <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
            </span>
            <Activity className="h-3.5 w-3.5 text-emerald-400" />
            <span>Scoring Live</span>
          </div>

          {/* Auth State Button / Profile */}
          {!isLoading && (
            <div>
              {user ? (
                <div className="flex items-center gap-2">
                  <div className="hidden sm:flex items-center gap-1.5 rounded-xl bg-secondary/60 border border-border/60 px-2.5 py-1 text-xs">
                    <UserIcon className="h-3.5 w-3.5 text-[#25f4ee]" />
                    <span className="max-w-[130px] truncate text-foreground font-medium" title={user.email}>
                      {user.email}
                    </span>
                  </div>
                  <button
                    type="button"
                    onClick={() => signOut()}
                    className="inline-flex items-center gap-1.5 rounded-xl border border-border/60 bg-secondary/40 hover:bg-destructive/20 hover:text-destructive hover:border-destructive/40 px-3 py-1.5 text-xs font-semibold text-muted-foreground transition-colors"
                    title="Sign Out"
                  >
                    <LogOut className="h-3.5 w-3.5" />
                    <span className="hidden sm:inline">Sign Out</span>
                  </button>
                </div>
              ) : (
                <button
                  type="button"
                  onClick={() => openAuthModal()}
                  className="inline-flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-[#fe2c55] to-[#fe2c55]/90 hover:opacity-95 text-white px-3.5 py-1.5 text-xs font-semibold shadow-md shadow-[#fe2c55]/20 transition-all"
                >
                  <LogIn className="h-3.5 w-3.5" />
                  <span>Sign In</span>
                </button>
              )}
            </div>
          )}
        </div>
      </div>
    </header>
  );
}

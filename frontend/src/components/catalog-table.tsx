"use client";

import React, { useState } from "react";
import { ProductCatalogItem, CatalogFilterState } from "@/types/catalog";
import { formatTHB, formatNumber } from "@/lib/utils";
import {
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  ExternalLink,
  ShoppingBag,
  TrendingUp,
  Zap,
  ChevronLeft,
  ChevronRight,
  Activity,
} from "lucide-react";

interface CatalogTableProps {
  items: ProductCatalogItem[];
  totalCount: number;
  totalPages: number;
  currentPage: number;
  pageSize: number;
  isLoading: boolean;
  sortBy?: string;
  sortOrder?: "asc" | "desc";
  onSort: (column: string) => void;
  onPageChange: (newPage: number) => void;
  onPageSizeChange: (newSize: number) => void;
  onSelectProduct?: (productId: string, productName: string) => void;
}

export function CatalogTable({
  items,
  totalCount,
  totalPages,
  currentPage,
  pageSize,
  isLoading,
  sortBy,
  sortOrder,
  onSort,
  onPageChange,
  onPageSizeChange,
  onSelectProduct,
}: CatalogTableProps) {
  const [failedImages, setFailedImages] = useState<Record<string, boolean>>({});

  const handleImageError = (id: string) => {
    setFailedImages((prev) => ({ ...prev, [id]: true }));
  };

  const renderSortIcon = (column: string) => {
    if (sortBy !== column) {
      return <ArrowUpDown className="h-3.5 w-3.5 text-muted-foreground/60" />;
    }
    return sortOrder === "asc" ? (
      <ArrowUp className="h-3.5 w-3.5 text-primary" />
    ) : (
      <ArrowDown className="h-3.5 w-3.5 text-primary" />
    );
  };

  const startItem = totalCount === 0 ? 0 : (currentPage - 1) * pageSize + 1;
  const endItem = Math.min(currentPage * pageSize, totalCount);

  return (
    <div className="rounded-2xl bg-card border border-border/60 overflow-hidden shadow-sm flex flex-col">
      {/* Table Container */}
      <div className="overflow-x-auto">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="border-b border-border/60 bg-secondary/30 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              <th className="py-3.5 px-4">Product</th>
              <th className="py-3.5 px-4 hidden md:table-cell">Category</th>
              <th
                className="py-3.5 px-4 cursor-pointer hover:text-foreground transition-colors"
                onClick={() => onSort("price")}
              >
                <div className="flex items-center gap-1.5">
                  <span>Price</span>
                  {renderSortIcon("price")}
                </div>
              </th>
              <th
                className="py-3.5 px-4 cursor-pointer hover:text-foreground transition-colors"
                onClick={() => onSort("commission_rate")}
              >
                <div className="flex items-center gap-1.5">
                  <span>Commission</span>
                  {renderSortIcon("commission_rate")}
                </div>
              </th>
              <th
                className="py-3.5 px-4 cursor-pointer hover:text-foreground transition-colors hidden sm:table-cell"
                onClick={() => onSort("total_sales")}
              >
                <div className="flex items-center gap-1.5">
                  <span>Total Sales</span>
                  {renderSortIcon("total_sales")}
                </div>
              </th>
              <th
                className="py-3.5 px-4 cursor-pointer hover:text-foreground transition-colors"
                onClick={() => onSort("velocity")}
              >
                <div className="flex items-center gap-1.5">
                  <span>Hourly Speed</span>
                  {renderSortIcon("velocity")}
                </div>
              </th>
              <th
                className="py-3.5 px-4 cursor-pointer hover:text-foreground transition-colors"
                onClick={() => onSort("winning_score")}
              >
                <div className="flex items-center gap-1.5">
                  <span>Score</span>
                  {renderSortIcon("winning_score")}
                </div>
              </th>
              <th className="py-3.5 px-4 text-right">Action</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/40 text-sm">
            {isLoading ? (
              // Skeleton rows
              [...Array(pageSize)].map((_, i) => (
                <tr key={i} className="animate-pulse">
                  <td className="py-4 px-4">
                    <div className="flex items-center gap-3">
                      <div className="h-12 w-12 rounded-xl bg-secondary/70 flex-shrink-0" />
                      <div className="space-y-2 flex-1">
                        <div className="h-3.5 w-3/4 bg-secondary/70 rounded" />
                        <div className="h-2.5 w-1/3 bg-secondary/50 rounded" />
                      </div>
                    </div>
                  </td>
                  <td className="py-4 px-4 hidden md:table-cell">
                    <div className="h-4 w-20 bg-secondary/60 rounded" />
                  </td>
                  <td className="py-4 px-4">
                    <div className="h-4 w-16 bg-secondary/60 rounded" />
                  </td>
                  <td className="py-4 px-4">
                    <div className="h-4 w-12 bg-secondary/60 rounded" />
                  </td>
                  <td className="py-4 px-4 hidden sm:table-cell">
                    <div className="h-4 w-16 bg-secondary/60 rounded" />
                  </td>
                  <td className="py-4 px-4">
                    <div className="h-4 w-14 bg-secondary/60 rounded" />
                  </td>
                  <td className="py-4 px-4">
                    <div className="h-5 w-12 bg-secondary/70 rounded-full" />
                  </td>
                  <td className="py-4 px-4 text-right">
                    <div className="h-8 w-20 bg-secondary/60 rounded-xl ml-auto" />
                  </td>
                </tr>
              ))
            ) : items.length === 0 ? (
              <tr>
                <td colSpan={8} className="py-16 text-center text-muted-foreground">
                  <div className="flex flex-col items-center justify-center">
                    <ShoppingBag className="h-10 w-10 text-muted-foreground/50 mb-3" />
                    <p className="font-semibold text-foreground text-base">No products match your criteria</p>
                    <p className="text-xs text-muted-foreground mt-1">Try resetting filters or adjusting search parameters</p>
                  </div>
                </td>
              </tr>
            ) : (
              items.map((item) => (
                <tr
                  key={item.id}
                  className="hover:bg-secondary/20 transition-colors group"
                >
                  {/* Product Title & Thumbnail */}
                  <td className="py-3.5 px-4 min-w-[240px]">
                    <div className="flex items-center gap-3">
                      <div className="relative h-12 w-12 flex-shrink-0 overflow-hidden rounded-xl bg-secondary/50 border border-border/40">
                        {item.image_url && !failedImages[item.id] ? (
                          // eslint-disable-next-line @next/next/no-img-element
                          <img
                            src={item.image_url}
                            alt={item.name}
                            className="h-full w-full object-cover"
                            onError={() => handleImageError(item.id)}
                          />
                        ) : (
                          <div className="flex h-full w-full items-center justify-center text-muted-foreground">
                            <ShoppingBag className="h-5 w-5 stroke-1" />
                          </div>
                        )}
                      </div>
                      <div className="min-w-0 flex-1">
                        <a
                          href={item.product_url || "#"}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="font-semibold text-foreground line-clamp-1 group-hover:text-primary transition-colors text-xs sm:text-sm"
                        >
                          {item.name}
                        </a>
                        <span className="text-[11px] text-muted-foreground block truncate md:hidden">
                          {item.category_name}
                        </span>
                      </div>
                    </div>
                  </td>

                  {/* Category */}
                  <td className="py-3.5 px-4 hidden md:table-cell">
                    <span className="inline-flex items-center rounded-lg bg-secondary/80 px-2.5 py-1 text-xs font-medium text-muted-foreground border border-border/40">
                      {item.category_name}
                    </span>
                  </td>

                  {/* Price */}
                  <td className="py-3.5 px-4 font-bold text-foreground">
                    {formatTHB(item.price)}
                  </td>

                  {/* Commission */}
                  <td className="py-3.5 px-4">
                    <span className="inline-flex items-center rounded-md bg-cyan-500/10 px-2 py-0.5 text-xs font-bold text-cyan-300 border border-cyan-500/20">
                      {item.commission_rate.toFixed(0)}%
                    </span>
                    <span className="block text-[10px] text-muted-foreground mt-0.5">
                      Yield {formatTHB(item.expected_return)}
                    </span>
                  </td>

                  {/* Total Sales */}
                  <td className="py-3.5 px-4 hidden sm:table-cell font-medium text-foreground">
                    {formatNumber(item.total_sales)} units
                  </td>

                  {/* Hourly Velocity */}
                  <td className="py-3.5 px-4">
                    <span className="inline-flex items-center gap-1 text-xs font-bold text-emerald-400">
                      <TrendingUp className="h-3 w-3" />
                      +{item.velocity_per_hour.toFixed(1)}/hr
                    </span>
                  </td>

                  {/* Winning Score */}
                  <td className="py-3.5 px-4">
                    <div className="inline-flex items-center gap-1 rounded-xl bg-gradient-to-r from-[#fe2c55]/15 to-[#25f4ee]/15 border border-[#fe2c55]/30 px-2.5 py-1 text-xs font-extrabold text-foreground">
                      <Zap className="h-3 w-3 text-[#25f4ee] fill-[#25f4ee]" />
                      <span>{item.winning_score.toFixed(1)}</span>
                    </div>
                  </td>

                  {/* Action Links */}
                  <td className="py-3.5 px-4 text-right">
                    <div className="inline-flex items-center gap-1.5 justify-end">
                      {onSelectProduct && (
                        <button
                          type="button"
                          onClick={() => onSelectProduct(item.id, item.name)}
                          className="inline-flex items-center gap-1 rounded-xl bg-secondary hover:bg-secondary/80 px-2.5 py-1.5 text-xs font-semibold text-foreground border border-border/60 transition-colors"
                          title="View Historical Growth Curve"
                        >
                          <Activity className="h-3 w-3 text-[#25f4ee]" />
                          <span className="hidden sm:inline">Trends</span>
                        </button>
                      )}
                      <a
                        href={item.product_url || "https://shop.tiktok.com"}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="inline-flex items-center gap-1.5 rounded-xl bg-secondary hover:bg-primary hover:text-white px-2.5 py-1.5 text-xs font-semibold text-foreground border border-border/60 transition-all hover:border-transparent"
                      >
                        <span>Shop</span>
                        <ExternalLink className="h-3 w-3" />
                      </a>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination Controls Footer */}
      <div className="flex flex-col sm:flex-row items-center justify-between gap-4 p-4 border-t border-border/40 bg-secondary/10">
        <div className="flex items-center gap-3 text-xs text-muted-foreground">
          <span>
            Showing <strong className="text-foreground">{startItem}</strong> -{" "}
            <strong className="text-foreground">{endItem}</strong> of{" "}
            <strong className="text-foreground">{totalCount}</strong> products
          </span>

          <div className="flex items-center gap-1.5 pl-3 border-l border-border/40">
            <span>Per page:</span>
            <select
              value={pageSize}
              onChange={(e) => onPageSizeChange(Number(e.target.value))}
              className="rounded-lg bg-secondary border border-border/60 px-2 py-1 text-xs text-foreground cursor-pointer focus:outline-none"
            >
              <option value={10}>10</option>
              <option value={20}>20</option>
              <option value={50}>50</option>
            </select>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => onPageChange(currentPage - 1)}
            disabled={currentPage <= 1 || isLoading}
            className="inline-flex items-center gap-1 rounded-xl bg-secondary hover:bg-secondary/80 disabled:opacity-40 disabled:pointer-events-none px-3 py-1.5 text-xs font-medium text-foreground border border-border/40 transition-colors"
          >
            <ChevronLeft className="h-3.5 w-3.5" />
            <span>Prev</span>
          </button>

          <span className="text-xs font-semibold text-foreground px-2">
            Page {currentPage} of {totalPages || 1}
          </span>

          <button
            onClick={() => onPageChange(currentPage + 1)}
            disabled={currentPage >= totalPages || isLoading}
            className="inline-flex items-center gap-1 rounded-xl bg-secondary hover:bg-secondary/80 disabled:opacity-40 disabled:pointer-events-none px-3 py-1.5 text-xs font-medium text-foreground border border-border/40 transition-colors"
          >
            <span>Next</span>
            <ChevronRight className="h-3.5 w-3.5" />
          </button>
        </div>
      </div>
    </div>
  );
}

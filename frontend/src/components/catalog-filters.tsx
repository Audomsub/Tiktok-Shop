"use client";

import React, { useState, useEffect } from "react";
import { Category, CatalogFilterState } from "@/types/catalog";
import { downloadCatalogCSV } from "@/services/api";
import { Search, RotateCcw, X, Download, Loader2 } from "lucide-react";

interface CatalogFiltersProps {
  categories: Category[];
  filters: CatalogFilterState;
  onFilterChange: (newFilters: Partial<CatalogFilterState>) => void;
  onReset: () => void;
}

export function CatalogFilters({
  categories,
  filters,
  onFilterChange,
  onReset,
}: CatalogFiltersProps) {
  const [searchTerm, setSearchTerm] = useState(filters.q || "");
  const [isExporting, setIsExporting] = useState(false);
  const [exportError, setExportError] = useState<string | null>(null);

  // Debounce search input by 300ms
  useEffect(() => {
    const timer = setTimeout(() => {
      if (searchTerm !== (filters.q || "")) {
        onFilterChange({ q: searchTerm, page: 1 });
      }
    }, 300);

    return () => clearTimeout(timer);
  }, [searchTerm, filters.q, onFilterChange]);

  const handleExportCSV = async () => {
    try {
      setIsExporting(true);
      setExportError(null);
      await downloadCatalogCSV(filters);
    } catch (err: unknown) {
      const errObj = err as Error;
      console.error("Export CSV error:", errObj);
      setExportError(errObj.message || "Failed to export CSV.");
    } finally {
      setIsExporting(false);
    }
  };

  return (
    <div className="rounded-2xl bg-card border border-border/60 p-4 sm:p-5 shadow-sm space-y-4">
      {/* Search Bar & Category Dropdown */}
      <div className="grid grid-cols-1 md:grid-cols-12 gap-3">
        {/* Debounced Search Input */}
        <div className="relative md:col-span-7">
          <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search products by title or keyword..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full rounded-xl bg-secondary/50 border border-border/60 pl-10 pr-10 py-2.5 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/40 focus:border-primary transition-all"
          />
          {searchTerm && (
            <button
              onClick={() => {
                setSearchTerm("");
                onFilterChange({ q: "", page: 1 });
              }}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
            >
              <X className="h-4 w-4" />
            </button>
          )}
        </div>

        {/* Dynamic Category Selector */}
        <div className="md:col-span-5">
          <select
            value={filters.category_id || ""}
            onChange={(e) =>
              onFilterChange({
                category_id: e.target.value || undefined,
                page: 1,
              })
            }
            className="w-full rounded-xl bg-secondary/50 border border-border/60 px-3.5 py-2.5 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary/40 focus:border-primary transition-all cursor-pointer"
          >
            <option value="">All Categories ({categories.length})</option>
            {categories.map((cat) => (
              <option key={cat.id} value={cat.id}>
                {cat.name}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Numeric Filter Sliders / Inputs */}
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-3 pt-2 border-t border-border/40 items-center">
        {/* Min Price */}
        <div>
          <label className="block text-xs font-medium text-muted-foreground mb-1">
            Min Price (฿): {filters.min_price || 0}
          </label>
          <input
            type="range"
            min={0}
            max={2000}
            step={50}
            value={filters.min_price || 0}
            onChange={(e) =>
              onFilterChange({
                min_price: Number(e.target.value) || undefined,
                page: 1,
              })
            }
            className="w-full accent-primary cursor-pointer h-1.5 bg-secondary rounded-lg"
          />
        </div>

        {/* Max Price */}
        <div>
          <label className="block text-xs font-medium text-muted-foreground mb-1">
            Max Price (฿): {filters.max_price ? `${filters.max_price}` : "No Limit"}
          </label>
          <input
            type="range"
            min={100}
            max={3000}
            step={100}
            value={filters.max_price || 3000}
            onChange={(e) =>
              onFilterChange({
                max_price: Number(e.target.value) < 3000 ? Number(e.target.value) : undefined,
                page: 1,
              })
            }
            className="w-full accent-primary cursor-pointer h-1.5 bg-secondary rounded-lg"
          />
        </div>

        {/* Min Commission Rate */}
        <div>
          <label className="block text-xs font-medium text-muted-foreground mb-1">
            Min Commission: {filters.min_commission || 0}%
          </label>
          <input
            type="range"
            min={0}
            max={40}
            step={5}
            value={filters.min_commission || 0}
            onChange={(e) =>
              onFilterChange({
                min_commission: Number(e.target.value) || undefined,
                page: 1,
              })
            }
            className="w-full accent-[#25f4ee] cursor-pointer h-1.5 bg-secondary rounded-lg"
          />
        </div>

        {/* Actions: Reset & Export to CSV */}
        <div className="flex items-center justify-end gap-2 pt-2 sm:pt-0">
          <button
            type="button"
            onClick={() => {
              setSearchTerm("");
              onReset();
            }}
            className="inline-flex items-center gap-1.5 text-xs font-medium text-muted-foreground hover:text-foreground bg-secondary/60 hover:bg-secondary px-3 py-2 rounded-xl border border-border/40 transition-colors"
          >
            <RotateCcw className="h-3.5 w-3.5" />
            <span>Reset</span>
          </button>

          <button
            type="button"
            onClick={handleExportCSV}
            disabled={isExporting}
            className="inline-flex items-center gap-1.5 text-xs font-semibold text-foreground bg-secondary hover:bg-secondary/80 hover:text-primary px-3 py-2 rounded-xl border border-border/60 shadow-sm transition-all disabled:opacity-50"
            title="Download matching products as CSV"
          >
            {isExporting ? (
              <Loader2 className="h-3.5 w-3.5 animate-spin text-[#25f4ee]" />
            ) : (
              <Download className="h-3.5 w-3.5 text-[#25f4ee]" />
            )}
            <span>{isExporting ? "Exporting..." : "Export CSV"}</span>
          </button>
        </div>
      </div>

      {exportError && (
        <div className="text-xs text-destructive bg-destructive/15 border border-destructive/30 rounded-xl p-2.5">
          {exportError}
        </div>
      )}
    </div>
  );
}

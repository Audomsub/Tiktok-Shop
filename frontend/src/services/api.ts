import { LeaderboardResponse } from "@/types/leaderboard";
import { Category, ProductCatalogResponse, CatalogFilterState } from "@/types/catalog";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export async function fetchLeaderboard(limit: number = 10): Promise<LeaderboardResponse> {
  try {
    const res = await fetch(`${API_BASE_URL}/api/v1/leaderboard?limit=${limit}`, {
      cache: "no-store", // Always fetch fresh realtime analytics
      headers: {
        Accept: "application/json",
      },
    });

    if (!res.ok) {
      throw new Error(`Failed to fetch leaderboard: HTTP ${res.status}`);
    }

    return await res.json();
  } catch (error) {
    console.error("API Error in fetchLeaderboard:", error);
    throw error;
  }
}

export async function fetchCategories(): Promise<Category[]> {
  try {
    const res = await fetch(`${API_BASE_URL}/api/v1/categories`, {
      cache: "no-store",
      headers: {
        Accept: "application/json",
      },
    });

    if (!res.ok) {
      throw new Error(`Failed to fetch categories: HTTP ${res.status}`);
    }

    return await res.json();
  } catch (error) {
    console.error("API Error in fetchCategories:", error);
    throw error;
  }
}

export async function fetchCatalog(filter: CatalogFilterState = {}): Promise<ProductCatalogResponse> {
  try {
    const params = new URLSearchParams();
    if (filter.q) params.set("q", filter.q);
    if (filter.category_id) params.set("category_id", filter.category_id);
    if (filter.min_price !== undefined) params.set("min_price", filter.min_price.toString());
    if (filter.max_price !== undefined) params.set("max_price", filter.max_price.toString());
    if (filter.min_commission !== undefined) params.set("min_commission", filter.min_commission.toString());
    if (filter.sort_by) params.set("sort_by", filter.sort_by);
    if (filter.sort_order) params.set("sort_order", filter.sort_order);
    if (filter.page) params.set("page", filter.page.toString());
    if (filter.limit) params.set("limit", filter.limit.toString());

    const res = await fetch(`${API_BASE_URL}/api/v1/products?${params.toString()}`, {
      cache: "no-store",
      headers: {
        Accept: "application/json",
      },
    });

    if (!res.ok) {
      throw new Error(`Failed to fetch product catalog: HTTP ${res.status}`);
    }

    return await res.json();
  } catch (error) {
    console.error("API Error in fetchCatalog:", error);
    throw error;
  }
}

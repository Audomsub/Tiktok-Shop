"use client";

import React, { createContext, useContext, useEffect, useState, useCallback } from "react";
import { useAuth } from "@/context/auth-context";
import { fetchFavorites, addFavorite, removeFavorite } from "@/services/api";
import { FavoriteItem } from "@/types/favorite";

const LOCAL_STORAGE_IDS_KEY = "tiktok_shop_favorites_ids";
const LOCAL_STORAGE_ITEMS_KEY = "tiktok_shop_favorites_items";

interface FavoritesContextType {
  favoriteIds: Set<string>;
  favoriteItems: FavoriteItem[];
  favoritesCount: number;
  isLoadingFavorites: boolean;
  isFavorited: (productId: string) => boolean;
  toggleFavorite: (productId: string, itemData?: Partial<FavoriteItem>) => Promise<boolean>;
  refreshFavorites: () => Promise<void>;
}

const FavoritesContext = createContext<FavoritesContextType | undefined>(undefined);

export function FavoritesProvider({ children }: { children: React.ReactNode }) {
  const { user, accessToken } = useAuth();
  const [favoriteIds, setFavoriteIds] = useState<Set<string>>(new Set());
  const [favoriteItems, setFavoriteItems] = useState<FavoriteItem[]>([]);
  const [isLoadingFavorites, setIsLoadingFavorites] = useState(false);

  // Load favorites from localStorage on initial client mount
  useEffect(() => {
    try {
      const storedIdsJson = localStorage.getItem(LOCAL_STORAGE_IDS_KEY);
      const storedItemsJson = localStorage.getItem(LOCAL_STORAGE_ITEMS_KEY);

      if (storedIdsJson) {
        const parsedIds = JSON.parse(storedIdsJson);
        if (Array.isArray(parsedIds)) {
          setFavoriteIds(new Set(parsedIds));
        }
      }

      if (storedItemsJson) {
        const parsedItems = JSON.parse(storedItemsJson);
        if (Array.isArray(parsedItems)) {
          setFavoriteItems(parsedItems);
        }
      }
    } catch (e) {
      console.warn("Failed to read favorites from localStorage:", e);
    }
  }, []);

  // Fetch authenticated favorites from Supabase when logged in
  const loadFavorites = useCallback(async () => {
    if (!accessToken) {
      return;
    }

    try {
      setIsLoadingFavorites(true);
      const res = await fetchFavorites(accessToken);
      const apiItems = res.items || [];
      const ids = new Set(apiItems.map((item) => item.product_id));

      setFavoriteIds(ids);
      setFavoriteItems(apiItems);

      // Sync to local storage as cache
      try {
        localStorage.setItem(LOCAL_STORAGE_IDS_KEY, JSON.stringify(Array.from(ids)));
        localStorage.setItem(LOCAL_STORAGE_ITEMS_KEY, JSON.stringify(apiItems));
      } catch (err) {
        console.warn("Could not cache favorites to localStorage:", err);
      }
    } catch (err) {
      console.error("Failed to load user favorites from API:", err);
    } finally {
      setIsLoadingFavorites(false);
    }
  }, [accessToken]);

  useEffect(() => {
    if (user && accessToken) {
      loadFavorites();
    }
  }, [user, accessToken, loadFavorites]);

  const isFavorited = useCallback(
    (productId: string) => {
      return favoriteIds.has(productId);
    },
    [favoriteIds]
  );

  const toggleFavorite = async (productId: string, itemData?: Partial<FavoriteItem>): Promise<boolean> => {
    const wasFavorited = favoriteIds.has(productId);

    // 1. Optimistic UI update for IDs
    const nextIds = new Set(favoriteIds);
    let nextItems: FavoriteItem[] = [...favoriteItems];

    if (wasFavorited) {
      nextIds.delete(productId);
      nextItems = nextItems.filter((i) => i.product_id !== productId);
    } else {
      nextIds.add(productId);
      const newItem: FavoriteItem = {
        product_id: productId,
        source_id: itemData?.source_id || productId,
        name: itemData?.name || "Product Opportunity",
        product_url: itemData?.product_url || "https://shop.tiktok.com",
        image_url: itemData?.image_url || "",
        category_id: itemData?.category_id || "",
        category_name: itemData?.category_name || "General",
        price: itemData?.price || 0,
        commission_rate: itemData?.commission_rate || 0,
        total_sales: itemData?.total_sales || 0,
        delta_sales: itemData?.delta_sales || 0,
        velocity_per_hour: itemData?.velocity_per_hour || 0,
        expected_return: itemData?.expected_return || 0,
        winning_score: itemData?.winning_score || 0,
        snapshot_time: itemData?.snapshot_time || new Date().toISOString(),
        favorited_at: new Date().toISOString(),
      };
      nextItems.unshift(newItem);
    }

    setFavoriteIds(nextIds);
    setFavoriteItems(nextItems);

    // Persist to localStorage immediately (Works 100% for guests!)
    try {
      localStorage.setItem(LOCAL_STORAGE_IDS_KEY, JSON.stringify(Array.from(nextIds)));
      localStorage.setItem(LOCAL_STORAGE_ITEMS_KEY, JSON.stringify(nextItems));
    } catch (e) {
      console.warn("Failed to write favorites to localStorage:", e);
    }

    // 2. If logged in, sync to Supabase backend API
    if (user && accessToken) {
      try {
        if (wasFavorited) {
          await removeFavorite(productId, accessToken);
        } else {
          await addFavorite(productId, accessToken);
        }
      } catch (err) {
        console.error("Failed to sync favorite with backend API:", err);
      }
    }

    return !wasFavorited;
  };

  return (
    <FavoritesContext.Provider
      value={{
        favoriteIds,
        favoriteItems,
        favoritesCount: favoriteIds.size,
        isLoadingFavorites,
        isFavorited,
        toggleFavorite,
        refreshFavorites: loadFavorites,
      }}
    >
      {children}
    </FavoritesContext.Provider>
  );
}

export function useFavorites() {
  const context = useContext(FavoritesContext);
  if (!context) {
    throw new Error("useFavorites must be used within a FavoritesProvider");
  }
  return context;
}

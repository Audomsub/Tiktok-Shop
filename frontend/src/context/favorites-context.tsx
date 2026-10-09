"use client";

import React, { createContext, useContext, useEffect, useState, useCallback } from "react";
import { useAuth } from "@/context/auth-context";
import { fetchFavorites, addFavorite, removeFavorite } from "@/services/api";

interface FavoritesContextType {
  favoriteIds: Set<string>;
  favoritesCount: number;
  isLoadingFavorites: boolean;
  isFavorited: (productId: string) => boolean;
  toggleFavorite: (productId: string) => Promise<boolean>;
  refreshFavorites: () => Promise<void>;
}

const FavoritesContext = createContext<FavoritesContextType | undefined>(undefined);

export function FavoritesProvider({ children }: { children: React.ReactNode }) {
  const { user, accessToken, openAuthModal } = useAuth();
  const [favoriteIds, setFavoriteIds] = useState<Set<string>>(new Set());
  const [isLoadingFavorites, setIsLoadingFavorites] = useState(false);

  const loadFavorites = useCallback(async () => {
    if (!accessToken) {
      setFavoriteIds(new Set());
      return;
    }

    try {
      setIsLoadingFavorites(true);
      const res = await fetchFavorites(accessToken);
      const ids = new Set((res.items || []).map((item) => item.product_id));
      setFavoriteIds(ids);
    } catch (err) {
      console.error("Failed to load user favorites:", err);
    } finally {
      setIsLoadingFavorites(false);
    }
  }, [accessToken]);

  useEffect(() => {
    if (user && accessToken) {
      loadFavorites();
    } else {
      setFavoriteIds(new Set());
    }
  }, [user, accessToken, loadFavorites]);

  const isFavorited = useCallback(
    (productId: string) => {
      return favoriteIds.has(productId);
    },
    [favoriteIds]
  );

  const toggleFavorite = async (productId: string): Promise<boolean> => {
    // Acceptance criterion 6: Prompt appears if unauthenticated user attempts to bookmark
    if (!user || !accessToken) {
      openAuthModal("Please sign in or create an account to save winning products to your favorites shortlist.");
      return false;
    }

    const wasFavorited = favoriteIds.has(productId);

    // Acceptance criterion 3: Optimistic UI update
    setFavoriteIds((prev) => {
      const next = new Set(prev);
      if (wasFavorited) {
        next.delete(productId);
      } else {
        next.add(productId);
      }
      return next;
    });

    try {
      if (wasFavorited) {
        await removeFavorite(productId, accessToken);
      } else {
        await addFavorite(productId, accessToken);
      }
      return !wasFavorited;
    } catch (err) {
      console.error("Failed to toggle favorite API:", err);
      // Rollback optimistic update on error
      setFavoriteIds((prev) => {
        const rollback = new Set(prev);
        if (wasFavorited) {
          rollback.add(productId);
        } else {
          rollback.delete(productId);
        }
        return rollback;
      });
      return wasFavorited;
    }
  };

  return (
    <FavoritesContext.Provider
      value={{
        favoriteIds,
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

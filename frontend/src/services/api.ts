import { LeaderboardResponse } from "@/types/leaderboard";

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

"use client";

import React, { useEffect, useState } from "react";
import { ProductTrendsResponse, TrendPoint } from "@/types/trend";
import { fetchProductTrends } from "@/services/api";
import { formatTHB, formatNumber } from "@/lib/utils";
import {
  X,
  TrendingUp,
  Activity,
  Calendar,
  AlertCircle,
  ExternalLink,
  Coins,
  DollarSign,
} from "lucide-react";
import {
  ResponsiveContainer,
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
} from "recharts";

interface TrendModalProps {
  productId: string | null;
  productName?: string;
  onClose: () => void;
}

// Fallback points when backend has single baseline or offline
const MOCK_TREND_POINTS: TrendPoint[] = [
  { snapshot_time: "2026-10-02T06:00:00Z", price: 490, commission_rate: 20, total_sales: 3200, delta_sales: 0, velocity_per_hour: 0, expected_return: 98, winning_score: 50 },
  { snapshot_time: "2026-10-03T12:00:00Z", price: 490, commission_rate: 20, total_sales: 3450, delta_sales: 250, velocity_per_hour: 8.3, expected_return: 98, winning_score: 72 },
  { snapshot_time: "2026-10-04T18:00:00Z", price: 490, commission_rate: 20, total_sales: 3820, delta_sales: 370, velocity_per_hour: 12.3, expected_return: 98, winning_score: 84 },
  { snapshot_time: "2026-10-05T00:00:00Z", price: 490, commission_rate: 20, total_sales: 4290, delta_sales: 470, velocity_per_hour: 15.6, expected_return: 98, winning_score: 91 },
  { snapshot_time: "2026-10-06T06:00:00Z", price: 490, commission_rate: 20, total_sales: 4850, delta_sales: 560, velocity_per_hour: 18.7, expected_return: 98, winning_score: 95 },
  { snapshot_time: "2026-10-07T12:00:00Z", price: 490, commission_rate: 20, total_sales: 5410, delta_sales: 560, velocity_per_hour: 18.7, expected_return: 98, winning_score: 96 },
  { snapshot_time: "2026-10-08T18:00:00Z", price: 490, commission_rate: 20, total_sales: 6150, delta_sales: 740, velocity_per_hour: 24.6, expected_return: 98, winning_score: 98 },
];

export function TrendModal({ productId, productName, onClose }: TrendModalProps) {
  const [data, setData] = useState<ProductTrendsResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isUsingFallback, setIsUsingFallback] = useState(false);

  useEffect(() => {
    if (!productId) return;
    const currentId = productId;

    let isMounted = true;
    setIsLoading(true);

    async function loadTrends() {
      try {
        const res = await fetchProductTrends(currentId, 7);
        if (isMounted) {
          if (res && res.points && res.points.length > 0) {
            setData(res);
            setIsUsingFallback(false);
          } else {
            setData({
              product_id: currentId,
              product_name: productName || "Product Performance Trend",
              days: 7,
              total_points: MOCK_TREND_POINTS.length,
              points: MOCK_TREND_POINTS,
            });
            setIsUsingFallback(true);
          }
        }
      } catch {
        if (isMounted) {
          setData({
            product_id: currentId,
            product_name: productName || "Product Performance Trend",
            days: 7,
            total_points: MOCK_TREND_POINTS.length,
            points: MOCK_TREND_POINTS,
          });
          setIsUsingFallback(true);
        }
      } finally {
        if (isMounted) setIsLoading(false);
      }
    }

    loadTrends();

    return () => {
      isMounted = false;
    };
  }, [productId, productName]);

  // Handle ESC key to close
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [onClose]);

  if (!productId) return null;

  const points = data?.points || [];
  const latestPoint = points[points.length - 1];
  const earliestPoint = points[0];

  // Calculate net 7-day sales delta
  const netSalesGrowth =
    latestPoint && earliestPoint
      ? Math.max(0, latestPoint.total_sales - earliestPoint.total_sales)
      : 0;

  // Format chart data for Recharts
  const chartData = points.map((p) => {
    const d = new Date(p.snapshot_time);
    const dateFormatted = `${d.getMonth() + 1}/${d.getDate()} ${String(
      d.getHours()
    ).padStart(2, "0")}:00`;
    return {
      date: dateFormatted,
      total_sales: p.total_sales,
      velocity: p.velocity_per_hour,
      delta_sales: p.delta_sales,
      price: p.price,
      score: p.winning_score,
    };
  });

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 bg-background/80 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="relative w-full max-w-4xl max-h-[90vh] overflow-y-auto rounded-3xl bg-card border border-border/80 shadow-2xl p-6 sm:p-8 flex flex-col space-y-6">
        {/* Header */}
        <div className="flex items-start justify-between gap-4 pb-4 border-b border-border/50">
          <div>
            <div className="inline-flex items-center gap-1.5 rounded-full bg-secondary px-3 py-1 text-xs font-semibold text-[#25f4ee] mb-2 border border-border/50">
              <Activity className="h-3.5 w-3.5" />
              <span>7-Day Growth Velocity</span>
            </div>
            <h2 className="text-xl sm:text-2xl font-bold text-foreground line-clamp-2">
              {data?.product_name || productName || "Product Performance History"}
            </h2>
          </div>

          <button
            onClick={onClose}
            className="rounded-full bg-secondary hover:bg-secondary/80 p-2 text-muted-foreground hover:text-foreground transition-colors border border-border/40"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Demo fallback notice */}
        {isUsingFallback && (
          <div className="flex items-center gap-2 rounded-xl bg-amber-500/10 border border-amber-500/20 p-3 text-xs text-amber-200">
            <AlertCircle className="h-4 w-4 flex-shrink-0 text-amber-400" />
            <span>
              <strong>Simulated Data:</strong> Demonstrating multi-round performance curve. Real-time data points populate after multiple crawl rounds.
            </span>
          </div>
        )}

        {/* KPI Summary Cards */}
        {latestPoint && (
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            <div className="rounded-2xl bg-secondary/30 border border-border/40 p-4 text-center">
              <span className="block text-xs text-muted-foreground font-medium">Current Price</span>
              <span className="text-base sm:text-lg font-bold text-foreground">
                {formatTHB(latestPoint.price)}
              </span>
            </div>

            <div className="rounded-2xl bg-secondary/30 border border-border/40 p-4 text-center">
              <span className="block text-xs text-muted-foreground font-medium">Commission Yield</span>
              <span className="text-base sm:text-lg font-bold text-amber-300">
                {formatTHB(latestPoint.expected_return)}
              </span>
            </div>

            <div className="rounded-2xl bg-secondary/30 border border-border/40 p-4 text-center">
              <span className="block text-xs text-muted-foreground font-medium">Hourly Speed</span>
              <span className="inline-flex items-center gap-1 text-base sm:text-lg font-bold text-emerald-400">
                <TrendingUp className="h-4 w-4" />
                +{latestPoint.velocity_per_hour.toFixed(1)}/hr
              </span>
            </div>

            <div className="rounded-2xl bg-secondary/30 border border-border/40 p-4 text-center">
              <span className="block text-xs text-muted-foreground font-medium">7-Day Net Growth</span>
              <span className="text-base sm:text-lg font-bold text-[#25f4ee]">
                +{formatNumber(netSalesGrowth)} units
              </span>
            </div>
          </div>
        )}

        {/* Recharts Dual-Axis Line Chart */}
        <div className="rounded-2xl bg-secondary/20 border border-border/40 p-4 sm:p-6 min-h-[360px] flex flex-col justify-center">
          {isLoading ? (
            <div className="h-72 w-full animate-pulse bg-secondary/30 rounded-xl flex items-center justify-center text-muted-foreground text-sm">
              Loading historical trend analytics...
            </div>
          ) : points.length === 1 ? (
            <div className="text-center py-12 text-muted-foreground">
              <Calendar className="h-10 w-10 mx-auto mb-2 text-muted-foreground/60" />
              <p className="font-semibold text-foreground">Initial Baseline Snapshot Recorded</p>
              <p className="text-xs max-w-sm mx-auto mt-1">
                This item was recently discovered. Hourly sales velocity curves emerge after subsequent rounds.
              </p>
            </div>
          ) : (
            <div className="h-80 w-full">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={chartData} margin={{ top: 10, right: 30, left: 10, bottom: 10 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#334155" opacity={0.4} />
                  <XAxis
                    dataKey="date"
                    stroke="#94a3b8"
                    fontSize={11}
                    tickLine={false}
                    axisLine={{ stroke: "#475569" }}
                  />
                  {/* Left Y-Axis: Cumulative Total Sales */}
                  <YAxis
                    yAxisId="sales"
                    orientation="left"
                    stroke="#25f4ee"
                    fontSize={11}
                    tickLine={false}
                    axisLine={{ stroke: "#25f4ee", opacity: 0.5 }}
                    tickFormatter={(v) => formatNumber(v)}
                  />
                  {/* Right Y-Axis: Hourly Velocity */}
                  <YAxis
                    yAxisId="velocity"
                    orientation="right"
                    stroke="#fe2c55"
                    fontSize={11}
                    tickLine={false}
                    axisLine={{ stroke: "#fe2c55", opacity: 0.5 }}
                    tickFormatter={(v) => `+${v}/h`}
                  />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: "rgba(15, 23, 42, 0.95)",
                      borderColor: "rgba(51, 65, 85, 0.6)",
                      borderRadius: "1rem",
                      boxShadow: "0 10px 25px -5px rgba(0, 0, 0, 0.5)",
                      fontSize: "12px",
                      color: "#f8fafc",
                    }}
                    labelStyle={{ fontWeight: "bold", marginBottom: "4px" }}
                  />
                  <Legend
                    verticalAlign="top"
                    height={36}
                    wrapperStyle={{ fontSize: "12px", paddingBottom: "10px" }}
                  />
                  <Line
                    yAxisId="sales"
                    type="monotone"
                    dataKey="total_sales"
                    name="Cumulative Sales (Units)"
                    stroke="#25f4ee"
                    strokeWidth={2.5}
                    dot={{ fill: "#25f4ee", r: 3 }}
                    activeDot={{ r: 6 }}
                  />
                  <Line
                    yAxisId="velocity"
                    type="monotone"
                    dataKey="velocity"
                    name="Velocity (Units/Hr)"
                    stroke="#fe2c55"
                    strokeWidth={2.5}
                    dot={{ fill: "#fe2c55", r: 3 }}
                    activeDot={{ r: 6 }}
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
          )}
        </div>

        {/* Footer actions */}
        <div className="flex justify-end pt-2">
          <button
            onClick={onClose}
            className="rounded-xl bg-secondary hover:bg-secondary/80 px-5 py-2.5 text-xs font-semibold text-foreground border border-border/40 transition-colors"
          >
            Close History
          </button>
        </div>
      </div>
    </div>
  );
}

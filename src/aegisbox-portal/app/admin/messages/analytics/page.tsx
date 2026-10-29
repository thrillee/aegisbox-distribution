"use client";

import { useState, useEffect, useCallback } from "react";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from "recharts";
import {
  MessageSquare,
  TrendingUp,
  TrendingDown,
  Clock,
  CheckCircle,
  XCircle,
  DollarSign,
  Filter,
} from "lucide-react";
import { PageHeader } from "@/components/dashboard/page-header";
import { DataCard } from "@/components/dashboard/data-card";
import { Button } from "@/components/ui/button";
import { PageLoadingSkeleton } from "@/components/routing/loading-components";
import { formatMoney } from "@/lib/money";
import { reportDate, reportDayStart } from "@/lib/time";

interface Analytics {
  overview: {
    totalMessages: number;
    delivered: number;
    deliveredRate: string;
    failed: number;
    failedRate: string;
    pending: number;
  };
  charged: { currency: string; amount: number }[];
  byDay: {
    date: string;
    received: number;
    delivered: number;
    failed: number;
  }[];
  byServiceProvider: {
    id: number;
    name: string;
    currency: string;
    totalMessages: number;
    delivered: number;
    failed: number;
    charged: number;
  }[];
  byMno: {
    id: number;
    name: string;
    totalMessages: number;
    delivered: number;
    failed: number;
    deliveryRate: string;
  }[];
}

export default function AnalyticsPage() {
  const [analytics, setAnalytics] = useState<Analytics | null>(null);
  const [loading, setLoading] = useState(true);
  const [initialLoad, setInitialLoad] = useState(true);
  // Last 30 days in Lagos time by default.
  const [filters, setFilters] = useState(() => ({
    date_from: reportDate(reportDayStart(29)),
    date_to: reportDate(),
  }));
  const [activeTab, setActiveTab] = useState<"bySP" | "byMno">("byMno");

  const fetchAnalytics = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      if (filters.date_from) params.set("date_from", filters.date_from);
      if (filters.date_to) params.set("date_to", filters.date_to);

      const res = await fetch(`/api/admin/messages/analytics?${params}`);
      if (res.ok) {
        const data = await res.json();
        setAnalytics(data);
      }
    } catch (error) {
      console.error("Error fetching analytics:", error);
    } finally {
      setLoading(false);
      setInitialLoad(false);
    }
  }, [filters]);

  useEffect(() => {
    fetchAnalytics();
  }, [fetchAnalytics]);

  const statusData = analytics
    ? [
        {
          name: "Delivered",
          value: analytics.overview.delivered,
          color: "#10B981",
        },
        { name: "Failed", value: analytics.overview.failed, color: "#EF4444" },
        {
          name: "Pending",
          value: analytics.overview.pending,
          color: "#F59E0B",
        },
      ]
    : [];

  const dailyChartData =
    analytics?.byDay.map((day) => ({
      date: new Date(`${day.date}T00:00:00Z`).toLocaleDateString("en-US", {
        month: "short",
        day: "numeric",
        timeZone: "UTC",
      }),
      Received: day.received,
      Delivered: day.delivered,
      Failed: day.failed,
    })) || [];

  if (initialLoad) {
    return <PageLoadingSkeleton />;
  }

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Message Analytics"
        description="Track message delivery performance and trends"
        actions={
          <div className="flex items-center gap-2">
            <input
              type="date"
              value={filters.date_from}
              onChange={(e) =>
                setFilters((prev) => ({ ...prev, date_from: e.target.value }))
              }
              className="px-3 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white"
            />
            <span className="text-zinc-400">to</span>
            <input
              type="date"
              value={filters.date_to}
              onChange={(e) =>
                setFilters((prev) => ({ ...prev, date_to: e.target.value }))
              }
              className="px-3 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white"
            />
          </div>
        }
      />

      {loading ? (
        <div className="flex items-center justify-center py-12">
          <MessageSquare className="h-8 w-8 animate-spin text-zinc-500" />
        </div>
      ) : analytics ? (
        <>
          <div className="grid gap-4 md:grid-cols-3 xl:grid-cols-5">
            <DataCard title="">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-blue-500/20">
                  <MessageSquare className="h-5 w-5 text-blue-400" />
                </div>
                <div>
                  <p className="text-sm text-zinc-400">Total Messages</p>
                  <p className="text-2xl font-bold text-white">
                    {analytics.overview.totalMessages.toLocaleString()}
                  </p>
                </div>
              </div>
            </DataCard>
            <DataCard title="">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-emerald-500/20">
                  <CheckCircle className="h-5 w-5 text-emerald-400" />
                </div>
                <div>
                  <p className="text-sm text-zinc-400">Delivered</p>
                  <p className="text-2xl font-bold text-emerald-400">
                    {analytics.overview.deliveredRate}%
                  </p>
                  <p className="text-xs text-zinc-500">{analytics.overview.delivered.toLocaleString()} messages</p>
                </div>
              </div>
            </DataCard>
            <DataCard title="">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-red-500/20">
                  <XCircle className="h-5 w-5 text-red-400" />
                </div>
                <div>
                  <p className="text-sm text-zinc-400">Failed</p>
                  <p className="text-2xl font-bold text-red-400">
                    {analytics.overview.failedRate}%
                  </p>
                  <p className="text-xs text-zinc-500">{analytics.overview.failed.toLocaleString()} messages, rejected included</p>
                </div>
              </div>
            </DataCard>
            <DataCard title="">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-amber-500/20">
                  <Clock className="h-5 w-5 text-amber-400" />
                </div>
                <div>
                  <p className="text-sm text-zinc-400">Pending</p>
                  <p className="text-2xl font-bold text-amber-400">
                    {analytics.overview.pending.toLocaleString()}
                  </p>
                </div>
              </div>
            </DataCard>
            <DataCard title="">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-violet-500/20">
                  <DollarSign className="h-5 w-5 text-violet-400" />
                </div>
                <div>
                  <p className="text-sm text-zinc-400">Charged, less refunds</p>
                  {analytics.charged.length === 0 ? (
                    <p className="text-2xl font-bold text-white">-</p>
                  ) : (
                    analytics.charged.map((c) => (
                      <p key={c.currency} className="text-xl font-bold text-white">
                        {formatMoney(c.amount, c.currency)}
                      </p>
                    ))
                  )}
                </div>
              </div>
            </DataCard>
          </div>

          <div className="grid gap-6 lg:grid-cols-3">
            <DataCard
              title="Daily Traffic"
              subtitle="Message volume over time"
              className="lg:col-span-2"
            >
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={dailyChartData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#3f3f46" />
                  <XAxis dataKey="date" stroke="#a1a1aa" fontSize={12} />
                  <YAxis stroke="#a1a1aa" fontSize={12} />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: "#18181b",
                      border: "1px solid #3f3f46",
                      borderRadius: "8px",
                      color: "#fff",
                    }}
                  />
                  <Bar
                    dataKey="Received"
                    fill="#6366f1"
                    radius={[4, 4, 0, 0]}
                  />
                  <Bar
                    dataKey="Delivered"
                    fill="#10b981"
                    radius={[4, 4, 0, 0]}
                  />
                  <Bar dataKey="Failed" fill="#ef4444" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </DataCard>

            <DataCard
              title="Delivery Status"
              subtitle="Message delivery breakdown"
            >
              <ResponsiveContainer width="100%" height={250}>
                <PieChart>
                  <Pie
                    data={statusData}
                    cx="50%"
                    cy="50%"
                    innerRadius={60}
                    outerRadius={80}
                    paddingAngle={5}
                    dataKey="value"
                  >
                    {statusData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip
                    contentStyle={{
                      backgroundColor: "#18181b",
                      border: "1px solid #3f3f46",
                      borderRadius: "8px",
                      color: "#fff",
                    }}
                  />
                </PieChart>
              </ResponsiveContainer>
              <div className="flex justify-center gap-4 mt-4">
                {statusData.map((item) => (
                  <div key={item.name} className="flex items-center gap-2">
                    <div
                      className="w-3 h-3 rounded-full"
                      style={{ backgroundColor: item.color }}
                    />
                    <span className="text-sm text-zinc-400">
                      {item.name}: {item.value.toLocaleString()}
                    </span>
                  </div>
                ))}
              </div>
            </DataCard>
          </div>

          <DataCard
            title="Breakdown by Entity"
            subtitle="Message statistics by service provider and MNO"
          >
            <div className="border-b border-white/10">
              <div className="flex">
                <button
                  onClick={() => setActiveTab("bySP")}
                  className={`px-6 py-3 text-sm font-medium transition-colors ${
                    activeTab === "bySP"
                      ? "border-b-2 border-white text-white"
                      : "text-zinc-400 hover:text-white"
                  }`}
                >
                  By Service Provider
                </button>
                <button
                  onClick={() => setActiveTab("byMno")}
                  className={`px-6 py-3 text-sm font-medium transition-colors ${
                    activeTab === "byMno"
                      ? "border-b-2 border-white text-white"
                      : "text-zinc-400 hover:text-white"
                  }`}
                >
                  By MNO
                </button>
              </div>
            </div>

            <div className="p-6">
              {activeTab === "bySP" && (
                <div className="overflow-x-auto">
                  {analytics.byServiceProvider.length === 0 ? (
                    <div className="text-center py-8 text-zinc-500">
                      No service provider data available yet
                    </div>
                  ) : (
                    <table className="w-full">
                      <thead>
                        <tr className="text-left text-xs font-semibold text-zinc-400 uppercase">
                          <th className="pb-3">Service Provider</th>
                          <th className="pb-3 text-right">Messages</th>
                          <th className="pb-3 text-right">Delivered</th>
                          <th className="pb-3 text-right">Failed</th>
                          <th className="pb-3 text-right">Charged, less refunds</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-white/10">
                        {analytics.byServiceProvider.map((sp) => (
                          <tr key={`${sp.id}-${sp.currency}`}>
                            <td className="py-3 text-white">{sp.name}</td>
                            <td className="py-3 text-right text-zinc-400">
                              {sp.totalMessages.toLocaleString()}
                            </td>
                            <td className="py-3 text-right text-emerald-400">
                              {sp.delivered.toLocaleString()}
                            </td>
                            <td className="py-3 text-right text-red-400">
                              {sp.failed.toLocaleString()}
                            </td>
                            <td className="py-3 text-right font-medium text-white">
                              {formatMoney(sp.charged, sp.currency)}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  )}
                </div>
              )}

              {activeTab === "byMno" && (
                <div className="overflow-x-auto">
                  {analytics.byMno.length === 0 ? (
                    <div className="text-center py-8 text-zinc-500">
                      No routing data available yet
                    </div>
                  ) : (
                    <table className="w-full">
                      <thead>
                        <tr className="text-left text-xs font-semibold text-zinc-400 uppercase">
                          <th className="pb-3">Route (MNO)</th>
                          <th className="pb-3 text-right">Messages</th>
                          <th className="pb-3 text-right">Delivered</th>
                          <th className="pb-3 text-right">Failed</th>
                          <th className="pb-3 text-right">Rate</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-white/10">
                        {analytics.byMno.map((mno) => (
                          <tr key={mno.id || "none"}>
                            <td className="py-3 text-white">{mno.name}</td>
                            <td className="py-3 text-right text-zinc-400">
                              {mno.totalMessages.toLocaleString()}
                            </td>
                            <td className="py-3 text-right text-emerald-400">
                              {mno.delivered.toLocaleString()}
                            </td>
                            <td className="py-3 text-right text-red-400">
                              {mno.failed.toLocaleString()}
                            </td>
                            <td className="py-3 text-right font-medium">
                              <span
                                className={
                                  Number(mno.deliveryRate) > 80
                                    ? "text-emerald-400"
                                    : Number(mno.deliveryRate) > 50
                                      ? "text-amber-400"
                                      : "text-red-400"
                                }
                              >
                                {mno.deliveryRate}%
                              </span>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  )}
                </div>
              )}
            </div>
          </DataCard>
        </>
      ) : (
        <div className="text-center py-12">
          <MessageSquare className="h-12 w-12 text-zinc-500 mx-auto mb-4" />
          <p className="text-zinc-500">Failed to load analytics</p>
        </div>
      )}
    </div>
  );
}

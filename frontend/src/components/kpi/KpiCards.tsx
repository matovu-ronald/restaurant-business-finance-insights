'use client';

import { formatCurrency, formatPercent } from '@/lib/dates';
import { KPITotals, useKPIFreshness } from '@/hooks/use-kpi';

interface KPICardsProps {
  totals: KPITotals | undefined;
  isLoading: boolean;
  error: Error | null;
  freshnessTimestamp: string | undefined;
}

interface KPICardProps {
  title: string;
  value: string;
  subtitle?: string;
  trend?: 'up' | 'down' | 'neutral';
  isLoading: boolean;
}

function KPICard({ title, value, subtitle, trend, isLoading }: KPICardProps) {
  if (isLoading) {
    return (
      <div className="rounded-lg bg-white p-6 shadow-sm border border-gray-200 animate-pulse">
        <div className="h-4 bg-gray-200 rounded w-24 mb-2"></div>
        <div className="h-8 bg-gray-200 rounded w-32 mb-1"></div>
        <div className="h-3 bg-gray-200 rounded w-20"></div>
      </div>
    );
  }

  const trendColors = {
    up: 'text-green-600',
    down: 'text-red-600',
    neutral: 'text-gray-500',
  };

  return (
    <div className="rounded-lg bg-white p-6 shadow-sm border border-gray-200">
      <h3 className="text-sm font-medium text-gray-500">{title}</h3>
      <p className="mt-2 text-3xl font-semibold text-gray-900">{value}</p>
      {subtitle && (
        <p className={`mt-1 text-sm ${trend ? trendColors[trend] : 'text-gray-500'}`}>
          {subtitle}
        </p>
      )}
    </div>
  );
}

export function KPICards({ totals, isLoading, error, freshnessTimestamp }: KPICardsProps) {
  const freshness = useKPIFreshness(freshnessTimestamp);

  if (error) {
    return (
      <div className="rounded-lg bg-red-50 p-6 border border-red-200">
        <p className="text-red-700">Failed to load KPIs: {error.message}</p>
      </div>
    );
  }

  const grossMarginPct = totals && totals.revenue > 0
    ? (totals.grossMargin / totals.revenue) * 100
    : 0;

  const netProfitPct = totals && totals.revenue > 0
    ? (totals.netProfit / totals.revenue) * 100
    : 0;

  return (
    <div className="space-y-4">
      {/* Freshness indicator */}
      <div className="flex items-center justify-end">
        <span
          className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
            freshness.isStale
              ? 'bg-yellow-100 text-yellow-800'
              : 'bg-green-100 text-green-800'
          }`}
        >
          <span
            className={`w-2 h-2 rounded-full mr-1.5 ${
              freshness.isStale ? 'bg-yellow-400' : 'bg-green-400'
            }`}
          ></span>
          {freshness.message}
        </span>
      </div>

      {/* Primary KPIs */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <KPICard
          title="Revenue"
          value={isLoading ? '-' : formatCurrency(totals?.revenue ?? 0)}
          subtitle={`${totals?.covers ?? 0} covers`}
          isLoading={isLoading}
        />
        <KPICard
          title="Gross Margin"
          value={isLoading ? '-' : formatCurrency(totals?.grossMargin ?? 0)}
          subtitle={`${formatPercent(grossMarginPct)} of revenue`}
          trend={grossMarginPct >= 65 ? 'up' : grossMarginPct >= 60 ? 'neutral' : 'down'}
          isLoading={isLoading}
        />
        <KPICard
          title="Labor Cost"
          value={isLoading ? '-' : formatCurrency(totals?.laborCost ?? 0)}
          subtitle={`${formatPercent(totals?.laborPct ?? 0)} of revenue`}
          trend={(totals?.laborPct ?? 0) <= 30 ? 'up' : (totals?.laborPct ?? 0) <= 35 ? 'neutral' : 'down'}
          isLoading={isLoading}
        />
        <KPICard
          title="Net Profit"
          value={isLoading ? '-' : formatCurrency(totals?.netProfit ?? 0)}
          subtitle={`${formatPercent(netProfitPct)} margin`}
          trend={netProfitPct >= 10 ? 'up' : netProfitPct >= 5 ? 'neutral' : 'down'}
          isLoading={isLoading}
        />
      </div>

      {/* Secondary KPIs */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <KPICard
          title="Avg Check"
          value={isLoading ? '-' : formatCurrency(totals?.avgCheck ?? 0)}
          isLoading={isLoading}
        />
        <KPICard
          title="COGS"
          value={isLoading ? '-' : formatCurrency(totals?.cogs ?? 0)}
          isLoading={isLoading}
        />
        <KPICard
          title="Discounts"
          value={isLoading ? '-' : formatCurrency(totals?.discounts ?? 0)}
          isLoading={isLoading}
        />
        <KPICard
          title="Comps"
          value={isLoading ? '-' : formatCurrency(totals?.comps ?? 0)}
          isLoading={isLoading}
        />
      </div>
    </div>
  );
}

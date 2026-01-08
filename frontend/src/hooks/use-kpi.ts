'use client';

import { useQuery } from '@tanstack/react-query';
import { apiClient } from '@/lib/api-client';
import { queryKeys } from '@/lib/query-client';

export interface KPITotals {
  revenue: number;
  cogs: number;
  grossMargin: number;
  laborCost: number;
  laborPct: number;
  opex: number;
  netProfit: number;
  covers: number;
  avgCheck: number;
  discounts: number;
  comps: number;
  freshnessTimestamp: string;
}

export interface KPISummary {
  label: string;
  display_name: string;
  revenue: number;
  cogs: number;
  grossMargin: number;
  laborCost: number;
  laborPct: number;
  opex: number;
  netProfit: number;
  covers: number;
  avgCheck: number;
  discounts: number;
  comps: number;
}

export interface DailyKPIResponse {
  freshnessTimestamp: string;
  range: string;
  totals: KPITotals;
  byChannel: KPISummary[];
  byDaypart: KPISummary[];
}

export type DateRange = '30d' | 'ytd' | 'trailing12m';

interface UseKPIParams {
  range?: DateRange;
  date?: string;
}

// Transform snake_case API response to camelCase
function transformKPIData(data: Record<string, unknown>): KPITotals {
  return {
    revenue: (data.revenue as number) || 0,
    cogs: (data.cogs as number) || 0,
    grossMargin: (data.gross_margin as number) || 0,
    laborCost: (data.labor_cost as number) || 0,
    laborPct: (data.labor_pct as number) || 0,
    opex: (data.opex as number) || 0,
    netProfit: (data.net_profit as number) || 0,
    covers: (data.covers as number) || 0,
    avgCheck: (data.avg_check as number) || 0,
    discounts: (data.discounts as number) || 0,
    comps: (data.comps as number) || 0,
    freshnessTimestamp: (data.freshness_timestamp as string) || '',
  };
}

function transformKPISummary(data: Record<string, unknown>): KPISummary {
  return {
    label: (data.label as string) || '',
    display_name: (data.display_name as string) || '',
    revenue: (data.revenue as number) || 0,
    cogs: (data.cogs as number) || 0,
    grossMargin: (data.gross_margin as number) || 0,
    laborCost: (data.labor_cost as number) || 0,
    laborPct: (data.labor_pct as number) || 0,
    opex: (data.opex as number) || 0,
    netProfit: (data.net_profit as number) || 0,
    covers: (data.covers as number) || 0,
    avgCheck: (data.avg_check as number) || 0,
    discounts: (data.discounts as number) || 0,
    comps: (data.comps as number) || 0,
  };
}

export function useKPI(params: UseKPIParams = {}) {
  const { range = '30d', date } = params;
  
  return useQuery({
    queryKey: queryKeys.kpi.daily({ range, date }),
    queryFn: async (): Promise<DailyKPIResponse> => {
      const searchParams = new URLSearchParams();
      searchParams.set('range', range);
      if (date) {
        searchParams.set('date', date);
      }
      const rawData = await apiClient.get<Record<string, unknown>>(`/kpi/daily?${searchParams.toString()}`);
      
      // Transform snake_case response to camelCase
      return {
        freshnessTimestamp: (rawData.freshnessTimestamp as string) || '',
        range: (rawData.range as string) || '',
        totals: transformKPIData(rawData.totals as Record<string, unknown>),
        byChannel: ((rawData.byChannel as Record<string, unknown>[]) || []).map(transformKPISummary),
        byDaypart: ((rawData.byDaypart as Record<string, unknown>[]) || []).map(transformKPISummary),
      };
    },
    staleTime: 30 * 1000, // 30 seconds
    refetchInterval: 60 * 1000, // Refresh every minute
  });
}

export function useKPIFreshness(freshnessTimestamp: string | undefined) {
  if (!freshnessTimestamp) {
    return { isStale: false, message: 'Loading...' };
  }
  
  const timestamp = new Date(freshnessTimestamp);
  const now = new Date();
  const diffMs = now.getTime() - timestamp.getTime();
  const diffMinutes = Math.floor(diffMs / (1000 * 60));
  
  if (diffMinutes < 5) {
    return { isStale: false, message: 'Just updated' };
  } else if (diffMinutes < 60) {
    return { isStale: false, message: `Updated ${diffMinutes} min ago` };
  } else if (diffMinutes < 1440) {
    const hours = Math.floor(diffMinutes / 60);
    return { isStale: true, message: `Updated ${hours}h ago` };
  } else {
    const days = Math.floor(diffMinutes / 1440);
    return { isStale: true, message: `Updated ${days}d ago` };
  }
}

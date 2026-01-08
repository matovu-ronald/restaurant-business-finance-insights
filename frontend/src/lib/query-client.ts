import { QueryClient } from "@tanstack/react-query";

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 60 * 1000, // 1 minute
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
});

// Query keys for consistent caching
export const queryKeys = {
  kpi: {
    all: ["kpi"] as const,
    daily: (params: { date?: string; range?: string; channel?: string; daypart?: string }) =>
      ["kpi", "daily", params] as const,
    drilldown: (params: { startDate: string; endDate: string; channel?: string; daypart?: string }) =>
      ["kpi", "drilldown", params] as const,
  },
  imports: {
    all: ["imports"] as const,
    list: () => ["imports", "list"] as const,
    detail: (id: string) => ["imports", id] as const,
  },
  exports: {
    all: ["exports"] as const,
    list: () => ["exports", "list"] as const,
    detail: (id: string) => ["exports", id] as const,
  },
  mappings: {
    all: ["mappings"] as const,
    list: () => ["mappings", "list"] as const,
    byType: (sourceType: string) => ["mappings", sourceType] as const,
  },
};

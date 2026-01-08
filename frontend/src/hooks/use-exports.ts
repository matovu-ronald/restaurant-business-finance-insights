import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api-client';
import { queryKeys } from '@/lib/query-client';

export interface SaleRow {
  id: string;
  occurred_at: string;
  channel: string;
  daypart: string;
  total: number;
  subtotal: number;
  tax: number;
  discounts: number;
  comps: number;
  payment_method: string;
}

export interface DrilldownResponse {
  data: SaleRow[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface DrilldownParams {
  startDate?: string;
  endDate?: string;
  channel?: string;
  daypart?: string;
  page?: number;
  pageSize?: number;
}

export function useDrilldown(params: DrilldownParams = {}) {
  const { startDate, endDate, channel, daypart, page = 1, pageSize = 50 } = params;

  return useQuery({
    queryKey: ['drilldown', 'sales', { startDate, endDate, channel, daypart, page, pageSize }],
    queryFn: async (): Promise<DrilldownResponse> => {
      const searchParams = new URLSearchParams();
      if (startDate) searchParams.set('start_date', startDate);
      if (endDate) searchParams.set('end_date', endDate);
      if (channel) searchParams.set('channel', channel);
      if (daypart) searchParams.set('daypart', daypart);
      searchParams.set('page', page.toString());
      searchParams.set('page_size', pageSize.toString());

      return apiClient.get<DrilldownResponse>(`/kpi/drilldown/sales?${searchParams.toString()}`);
    },
  });
}

export interface ExportJob {
  id: string;
  export_type: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  file_name: string;
  file_size: number;
  location_id: string;
  created_by_id: string;
  created_at: string;
  completed_at?: string;
}

export function useExports() {
  return useQuery({
    queryKey: queryKeys.exports.list(),
    queryFn: async (): Promise<ExportJob[]> => {
      return apiClient.get<ExportJob[]>('/exports');
    },
  });
}

export function useCreateExport() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (params: {
      exportType: string;
      startDate: string;
      endDate: string;
    }): Promise<Blob> => {
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080/api/v1'}/exports/pnl`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${localStorage.getItem('auth_token') ?? ''}`,
        },
        body: JSON.stringify({
          export_type: params.exportType,
          start_date: params.startDate,
          end_date: params.endDate,
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to generate export');
      }

      return response.blob();
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.exports.all });
    },
  });
}

// Helper to trigger file download
export function downloadBlob(blob: Blob, filename: string) {
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  window.URL.revokeObjectURL(url);
}

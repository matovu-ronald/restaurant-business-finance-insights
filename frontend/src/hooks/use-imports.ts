import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api-client';
import { queryKeys } from '@/lib/query-client';

export interface ImportJob {
  id: string;
  source_type: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  file_name: string;
  file_hash: string;
  total_rows: number;
  processed_rows: number;
  error_rows: number;
  location_id: string;
  mapping_id?: string;
  created_by_id: string;
  created_at: string;
  completed_at?: string;
  error_message?: string;
}

export interface ImportAnomaly {
  id: string;
  import_job_id: string;
  line_number: number;
  severity: 'error' | 'warning';
  message: string;
  raw_data?: string;
  created_at: string;
}

export interface MappingProfile {
  id: string;
  name: string;
  source_type: string;
  column_maps: Record<string, string>;
  defaults: Record<string, unknown>;
  location_id: string;
  created_by_id: string;
  created_at: string;
  updated_at: string;
}

export interface MappingsResponse {
  profiles: MappingProfile[];
  defaults: Record<string, Record<string, string>>;
}

export interface ImportJobDetailResponse {
  job: ImportJob;
  anomalies: ImportAnomaly[];
}

// Fetch import jobs list
export function useImports() {
  return useQuery({
    queryKey: queryKeys.imports.list(),
    queryFn: async (): Promise<ImportJob[]> => {
      return apiClient.get<ImportJob[]>('/imports');
    },
  });
}

// Fetch single import job with anomalies
export function useImport(id: string | undefined) {
  return useQuery({
    queryKey: queryKeys.imports.detail(id ?? ''),
    queryFn: async (): Promise<ImportJobDetailResponse> => {
      return apiClient.get<ImportJobDetailResponse>(`/imports/${id}`);
    },
    enabled: !!id,
    refetchInterval: (data) => {
      // Poll while job is in progress
      if (data?.state.data?.job.status === 'pending' || data?.state.data?.job.status === 'processing') {
        return 2000;
      }
      return false;
    },
  });
}

// Create new import
export function useCreateImport() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (params: {
      file: File;
      sourceType: string;
      mappingId?: string;
    }): Promise<ImportJob> => {
      const formData = new FormData();
      formData.append('file', params.file);
      formData.append('source_type', params.sourceType);
      if (params.mappingId) {
        formData.append('mapping_id', params.mappingId);
      }

      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080/api/v1'}/imports`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${localStorage.getItem('auth_token') ?? ''}`,
        },
        body: formData,
      });

      if (!response.ok) {
        const error = await response.text();
        throw new Error(error || 'Failed to create import');
      }

      return response.json();
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.imports.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.kpi.all });
    },
  });
}

// Fetch mapping profiles
export function useMappings() {
  return useQuery({
    queryKey: queryKeys.mappings.list(),
    queryFn: async (): Promise<MappingsResponse> => {
      return apiClient.get<MappingsResponse>('/mappings');
    },
  });
}

// Create mapping profile
export function useCreateMapping() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (params: {
      name: string;
      sourceType: string;
      columnMaps: Record<string, string>;
      defaults?: Record<string, unknown>;
    }): Promise<MappingProfile> => {
      return apiClient.post<MappingProfile>('/mappings', {
        name: params.name,
        source_type: params.sourceType,
        column_maps: params.columnMaps,
        defaults: params.defaults ?? {},
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.mappings.all });
    },
  });
}

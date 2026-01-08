'use client';

import { ImportJob, ImportAnomaly } from '@/hooks/use-imports';
import { formatDateTime } from '@/lib/dates';

interface ImportLogTableProps {
  jobs: ImportJob[];
  onSelect?: (job: ImportJob) => void;
  selectedJobId?: string;
}

export function ImportLogTable({ jobs, onSelect, selectedJobId }: ImportLogTableProps) {
  if (!jobs.length) {
    return (
      <div className="rounded-lg bg-gray-50 p-8 text-center border border-gray-200">
        <p className="text-gray-500">No imports yet</p>
        <p className="text-sm text-gray-400 mt-1">
          Upload a CSV file to get started
        </p>
      </div>
    );
  }

  const getStatusBadge = (status: ImportJob['status']) => {
    const styles = {
      pending: 'bg-yellow-100 text-yellow-800',
      processing: 'bg-blue-100 text-blue-800',
      completed: 'bg-green-100 text-green-800',
      failed: 'bg-red-100 text-red-800',
    };

    return (
      <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${styles[status]}`}>
        {status}
      </span>
    );
  };

  return (
    <div className="overflow-x-auto">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">File</th>
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Type</th>
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Progress</th>
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Date</th>
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-200">
          {jobs.map((job) => (
            <tr
              key={job.id}
              onClick={() => onSelect?.(job)}
              className={`cursor-pointer hover:bg-gray-50 ${
                selectedJobId === job.id ? 'bg-blue-50' : ''
              }`}
            >
              <td className="px-4 py-3 text-sm text-gray-900">
                <div className="truncate max-w-xs" title={job.file_name}>
                  {job.file_name}
                </div>
              </td>
              <td className="px-4 py-3 text-sm text-gray-500 capitalize">{job.source_type}</td>
              <td className="px-4 py-3 text-sm">{getStatusBadge(job.status)}</td>
              <td className="px-4 py-3 text-sm text-gray-500">
                {job.total_rows > 0 ? (
                  <span>
                    {job.processed_rows}/{job.total_rows}
                    {job.error_rows > 0 && (
                      <span className="text-red-600 ml-1">({job.error_rows} errors)</span>
                    )}
                  </span>
                ) : (
                  '-'
                )}
              </td>
              <td className="px-4 py-3 text-sm text-gray-500">
                {formatDateTime(job.created_at)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

interface AnomalyTableProps {
  anomalies: ImportAnomaly[];
}

export function AnomalyTable({ anomalies }: AnomalyTableProps) {
  if (!anomalies.length) {
    return (
      <div className="rounded-lg bg-green-50 p-4 border border-green-200">
        <p className="text-green-700 text-sm">No anomalies detected</p>
      </div>
    );
  }

  return (
    <div className="rounded-lg border border-gray-200 overflow-hidden">
      <div className="bg-red-50 px-4 py-2 border-b border-red-200">
        <p className="text-red-800 text-sm font-medium">
          {anomalies.length} anomal{anomalies.length === 1 ? 'y' : 'ies'} detected
        </p>
      </div>
      <div className="max-h-64 overflow-y-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50 sticky top-0">
            <tr>
              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">
                Line
              </th>
              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">
                Severity
              </th>
              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">
                Message
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {anomalies.map((anomaly) => (
              <tr key={anomaly.id}>
                <td className="px-4 py-2 text-sm text-gray-900">{anomaly.line_number}</td>
                <td className="px-4 py-2 text-sm">
                  <span
                    className={`px-2 py-0.5 rounded-full text-xs font-medium ${
                      anomaly.severity === 'error'
                        ? 'bg-red-100 text-red-800'
                        : 'bg-yellow-100 text-yellow-800'
                    }`}
                  >
                    {anomaly.severity}
                  </span>
                </td>
                <td className="px-4 py-2 text-sm text-gray-600">{anomaly.message}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

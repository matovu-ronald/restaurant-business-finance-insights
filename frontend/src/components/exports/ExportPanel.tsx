'use client';

import { useState } from 'react';
import { useCreateExport, downloadBlob } from '@/hooks/use-exports';
import { formatDate } from '@/lib/dates';

interface ExportPanelProps {
  defaultStartDate?: string;
  defaultEndDate?: string;
}

const EXPORT_TYPES = [
  { value: 'pnl', label: 'P&L Report', description: 'Full profit & loss breakdown by date, channel, and daypart' },
  { value: 'channel_summary', label: 'Channel Summary', description: 'Revenue and margin by sales channel' },
];

export function ExportPanel({ defaultStartDate, defaultEndDate }: ExportPanelProps) {
  const today = new Date().toISOString().split('T')[0];
  const thirtyDaysAgo = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];

  const [exportType, setExportType] = useState('pnl');
  const [startDate, setStartDate] = useState(defaultStartDate ?? thirtyDaysAgo);
  const [endDate, setEndDate] = useState(defaultEndDate ?? today);

  const createExport = useCreateExport();

  const handleExport = async () => {
    try {
      const blob = await createExport.mutateAsync({
        exportType,
        startDate,
        endDate,
      });

      const filename = `${exportType}_${startDate}_${endDate}.csv`;
      downloadBlob(blob, filename);
    } catch (error) {
      console.error('Export failed:', error);
      alert('Failed to generate export');
    }
  };

  return (
    <div className="rounded-lg bg-white p-6 shadow-sm border border-gray-200">
      <h3 className="text-lg font-medium text-gray-900 mb-4">Export Data</h3>

      <div className="space-y-4">
        {/* Export Type */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">Export Type</label>
          <div className="space-y-2">
            {EXPORT_TYPES.map((type) => (
              <label
                key={type.value}
                className={`flex items-start p-3 rounded-lg border cursor-pointer ${
                  exportType === type.value
                    ? 'border-blue-500 bg-blue-50'
                    : 'border-gray-200 hover:bg-gray-50'
                }`}
              >
                <input
                  type="radio"
                  name="exportType"
                  value={type.value}
                  checked={exportType === type.value}
                  onChange={(e) => setExportType(e.target.value)}
                  className="mt-0.5"
                />
                <div className="ml-3">
                  <div className="text-sm font-medium text-gray-900">{type.label}</div>
                  <div className="text-xs text-gray-500">{type.description}</div>
                </div>
              </label>
            ))}
          </div>
        </div>

        {/* Date Range */}
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Start Date</label>
            <input
              type="date"
              value={startDate}
              onChange={(e) => setStartDate(e.target.value)}
              max={endDate}
              className="w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 text-sm"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">End Date</label>
            <input
              type="date"
              value={endDate}
              onChange={(e) => setEndDate(e.target.value)}
              min={startDate}
              max={today}
              className="w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 text-sm"
            />
          </div>
        </div>

        {/* Export Button */}
        <button
          onClick={handleExport}
          disabled={createExport.isPending}
          className="w-full flex items-center justify-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {createExport.isPending ? (
            <>
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
              Generating...
            </>
          ) : (
            <>
              <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
              Download CSV
            </>
          )}
        </button>
      </div>
    </div>
  );
}

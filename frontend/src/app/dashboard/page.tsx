'use client';

import { useState } from 'react';
import { useKPI, DateRange } from '@/hooks/use-kpi';
import { KPICards, ChannelDaypartCharts } from '@/components/kpi';

const DATE_RANGE_OPTIONS: { value: DateRange; label: string }[] = [
  { value: '30d', label: 'Last 30 Days' },
  { value: 'ytd', label: 'Year to Date' },
  { value: 'trailing12m', label: 'Trailing 12 Months' },
];

export default function DashboardPage() {
  const [dateRange, setDateRange] = useState<DateRange>('30d');
  const { data, isLoading, error } = useKPI({ range: dateRange });

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div>
              <h1 className="text-2xl font-bold text-gray-900">
                Finance Dashboard
              </h1>
              <p className="mt-1 text-sm text-gray-500">
                The Lakehouse Restaurant • Daily Finance Snapshot
              </p>
            </div>
            
            {/* Date Range Selector */}
            <div className="flex items-center gap-2">
              <label htmlFor="date-range" className="text-sm text-gray-600">
                Period:
              </label>
              <select
                id="date-range"
                value={dateRange}
                onChange={(e) => setDateRange(e.target.value as DateRange)}
                className="rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 text-sm"
              >
                {DATE_RANGE_OPTIONS.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="space-y-8">
          {/* KPI Cards Section */}
          <section>
            <h2 className="text-lg font-semibold text-gray-900 mb-4">
              Key Performance Indicators
            </h2>
            <KPICards
              totals={data?.totals}
              isLoading={isLoading}
              error={error}
              freshnessTimestamp={data?.freshnessTimestamp}
            />
          </section>

          {/* Charts Section */}
          <section>
            <h2 className="text-lg font-semibold text-gray-900 mb-4">
              Channel & Daypart Analysis
            </h2>
            <ChannelDaypartCharts
              byChannel={data?.byChannel}
              byDaypart={data?.byDaypart}
              isLoading={isLoading}
            />
          </section>

          {/* Empty State Hint */}
          {!isLoading && !error && data && !data.totals?.revenue && (
            <div className="rounded-lg bg-blue-50 p-6 border border-blue-200">
              <div className="flex">
                <div className="flex-shrink-0">
                  <svg
                    className="h-5 w-5 text-blue-400"
                    viewBox="0 0 20 20"
                    fill="currentColor"
                  >
                    <path
                      fillRule="evenodd"
                      d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                      clipRule="evenodd"
                    />
                  </svg>
                </div>
                <div className="ml-3">
                  <h3 className="text-sm font-medium text-blue-800">
                    No data available
                  </h3>
                  <p className="mt-2 text-sm text-blue-700">
                    Import your POS sales data to see financial metrics and trends.
                    Navigate to Imports to upload a CSV file.
                  </p>
                </div>
              </div>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}

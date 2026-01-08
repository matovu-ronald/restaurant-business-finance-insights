'use client';

import { useState } from 'react';
import { useDrilldown, SaleRow } from '@/hooks/use-exports';
import { ExportPanel } from '@/components/exports';
import { formatCurrency } from '@/lib/dates';

const CHANNELS = [
  { value: '', label: 'All Channels' },
  { value: 'dine-in', label: 'Dine In' },
  { value: 'takeaway', label: 'Takeaway' },
  { value: 'pickup', label: 'Pickup' },
  { value: 'catering', label: 'Catering' },
];

const DAYPARTS = [
  { value: '', label: 'All Dayparts' },
  { value: 'breakfast', label: 'Breakfast' },
  { value: 'lunch', label: 'Lunch' },
  { value: 'dinner', label: 'Dinner' },
];

export default function DrilldownPage() {
  const today = new Date().toISOString().split('T')[0];
  const thirtyDaysAgo = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];

  const [startDate, setStartDate] = useState(thirtyDaysAgo);
  const [endDate, setEndDate] = useState(today);
  const [channel, setChannel] = useState('');
  const [daypart, setDaypart] = useState('');
  const [page, setPage] = useState(1);

  const { data, isLoading, error } = useDrilldown({
    startDate,
    endDate,
    channel: channel || undefined,
    daypart: daypart || undefined,
    page,
    pageSize: 50,
  });

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <h1 className="text-2xl font-bold text-gray-900">Sales Drill-Down</h1>
          <p className="mt-1 text-sm text-gray-500">
            View individual transactions and export data
          </p>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
          {/* Left Sidebar - Filters & Export */}
          <div className="lg:col-span-1 space-y-6">
            {/* Filters */}
            <div className="rounded-lg bg-white p-4 shadow-sm border border-gray-200 space-y-4">
              <h3 className="font-medium text-gray-900">Filters</h3>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Start Date</label>
                <input
                  type="date"
                  value={startDate}
                  onChange={(e) => { setStartDate(e.target.value); setPage(1); }}
                  max={endDate}
                  className="w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 text-sm"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">End Date</label>
                <input
                  type="date"
                  value={endDate}
                  onChange={(e) => { setEndDate(e.target.value); setPage(1); }}
                  min={startDate}
                  max={today}
                  className="w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 text-sm"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Channel</label>
                <select
                  value={channel}
                  onChange={(e) => { setChannel(e.target.value); setPage(1); }}
                  className="w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 text-sm"
                >
                  {CHANNELS.map((c) => (
                    <option key={c.value} value={c.value}>{c.label}</option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Daypart</label>
                <select
                  value={daypart}
                  onChange={(e) => { setDaypart(e.target.value); setPage(1); }}
                  className="w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 text-sm"
                >
                  {DAYPARTS.map((d) => (
                    <option key={d.value} value={d.value}>{d.label}</option>
                  ))}
                </select>
              </div>
            </div>

            {/* Export Panel */}
            <ExportPanel defaultStartDate={startDate} defaultEndDate={endDate} />
          </div>

          {/* Main Content - Data Table */}
          <div className="lg:col-span-3">
            <div className="rounded-lg bg-white shadow-sm border border-gray-200">
              <div className="px-4 py-3 border-b border-gray-200 flex items-center justify-between">
                <h2 className="font-medium text-gray-900">
                  Transactions
                  {data && (
                    <span className="ml-2 text-sm text-gray-500">
                      ({data.total.toLocaleString()} total)
                    </span>
                  )}
                </h2>
              </div>

              {error && (
                <div className="p-4 bg-red-50 text-red-700">
                  Failed to load transactions: {error.message}
                </div>
              )}

              {isLoading && (
                <div className="p-8 text-center">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
                  <p className="mt-2 text-gray-500">Loading...</p>
                </div>
              )}

              {!isLoading && !error && data && (
                <>
                  <div className="overflow-x-auto">
                    <table className="min-w-full divide-y divide-gray-200">
                      <thead className="bg-gray-50">
                        <tr>
                          <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Date/Time</th>
                          <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Channel</th>
                          <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Daypart</th>
                          <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Total</th>
                          <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Tax</th>
                          <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Discounts</th>
                          <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Payment</th>
                        </tr>
                      </thead>
                      <tbody className="bg-white divide-y divide-gray-200">
                        {data.data?.map((row) => (
                          <tr key={row.id} className="hover:bg-gray-50">
                            <td className="px-4 py-3 text-sm text-gray-900 whitespace-nowrap">{row.occurred_at}</td>
                            <td className="px-4 py-3 text-sm text-gray-500">{row.channel}</td>
                            <td className="px-4 py-3 text-sm text-gray-500">{row.daypart}</td>
                            <td className="px-4 py-3 text-sm text-gray-900 text-right font-medium">{formatCurrency(row.total)}</td>
                            <td className="px-4 py-3 text-sm text-gray-500 text-right">{formatCurrency(row.tax)}</td>
                            <td className="px-4 py-3 text-sm text-gray-500 text-right">
                              {row.discounts > 0 ? formatCurrency(row.discounts) : '-'}
                            </td>
                            <td className="px-4 py-3 text-sm text-gray-500">{row.payment_method}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>

                  {/* Pagination */}
                  {data.total_pages > 1 && (
                    <div className="px-4 py-3 border-t border-gray-200 flex items-center justify-between">
                      <p className="text-sm text-gray-500">
                        Page {data.page} of {data.total_pages}
                      </p>
                      <div className="flex gap-2">
                        <button
                          onClick={() => setPage((p) => Math.max(1, p - 1))}
                          disabled={page === 1}
                          className="px-3 py-1 text-sm border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                          Previous
                        </button>
                        <button
                          onClick={() => setPage((p) => Math.min(data.total_pages, p + 1))}
                          disabled={page === data.total_pages}
                          className="px-3 py-1 text-sm border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                          Next
                        </button>
                      </div>
                    </div>
                  )}

                  {data.data?.length === 0 && (
                    <div className="p-8 text-center text-gray-500">
                      No transactions found for the selected filters
                    </div>
                  )}
                </>
              )}
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}

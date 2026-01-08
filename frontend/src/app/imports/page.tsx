'use client';

import { useState, useRef } from 'react';
import {
  useImports,
  useImport,
  useCreateImport,
  useMappings,
  ImportJob,
} from '@/hooks/use-imports';
import { MappingProfileForm, ImportLogTable, AnomalyTable } from '@/components/imports';

const SOURCE_TYPES = [
  { value: 'pos', label: 'POS / Sales', description: 'Sales transactions from your point of sale system' },
  { value: 'payroll', label: 'Payroll', description: 'Employee wages and labor costs' },
  { value: 'inventory', label: 'Inventory', description: 'Stock snapshots and valuations' },
];

export default function ImportsPage() {
  const [sourceType, setSourceType] = useState('pos');
  const [mappingId, setMappingId] = useState<string | undefined>();
  const [selectedJob, setSelectedJob] = useState<ImportJob | undefined>();
  const [isDragging, setIsDragging] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const { data: jobs, isLoading: jobsLoading } = useImports();
  const { data: selectedJobDetail } = useImport(selectedJob?.id);
  const { data: mappings } = useMappings();
  const createImport = useCreateImport();

  const handleFileSelect = async (file: File) => {
    if (!file.name.endsWith('.csv')) {
      alert('Please select a CSV file');
      return;
    }

    try {
      const job = await createImport.mutateAsync({
        file,
        sourceType,
        mappingId,
      });
      setSelectedJob(job);
    } catch (error) {
      console.error('Import failed:', error);
      alert(error instanceof Error ? error.message : 'Import failed');
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);

    const file = e.dataTransfer.files[0];
    if (file) {
      handleFileSelect(file);
    }
  };

  const handleFileInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      handleFileSelect(file);
    }
    // Reset input
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <h1 className="text-2xl font-bold text-gray-900">Import Data</h1>
          <p className="mt-1 text-sm text-gray-500">
            Upload CSV files from your POS, payroll, or inventory systems
          </p>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Left Column - Upload Form */}
          <div className="lg:col-span-1 space-y-6">
            {/* Source Type Selection */}
            <div className="rounded-lg bg-white p-4 shadow-sm border border-gray-200">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Data Source
              </label>
              <div className="space-y-2">
                {SOURCE_TYPES.map((type) => (
                  <label
                    key={type.value}
                    className={`flex items-start p-3 rounded-lg border cursor-pointer ${
                      sourceType === type.value
                        ? 'border-blue-500 bg-blue-50'
                        : 'border-gray-200 hover:bg-gray-50'
                    }`}
                  >
                    <input
                      type="radio"
                      name="sourceType"
                      value={type.value}
                      checked={sourceType === type.value}
                      onChange={(e) => setSourceType(e.target.value)}
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

            {/* Mapping Profile */}
            <div className="rounded-lg bg-white p-4 shadow-sm border border-gray-200">
              <MappingProfileForm
                mappings={mappings}
                selectedMappingId={mappingId}
                onSelect={setMappingId}
                sourceType={sourceType}
              />
            </div>

            {/* File Upload */}
            <div
              className={`rounded-lg bg-white p-6 shadow-sm border-2 border-dashed transition-colors ${
                isDragging
                  ? 'border-blue-500 bg-blue-50'
                  : 'border-gray-300 hover:border-gray-400'
              }`}
              onDragOver={(e) => {
                e.preventDefault();
                setIsDragging(true);
              }}
              onDragLeave={() => setIsDragging(false)}
              onDrop={handleDrop}
            >
              <input
                ref={fileInputRef}
                type="file"
                accept=".csv"
                onChange={handleFileInput}
                className="hidden"
                id="file-upload"
              />
              <label
                htmlFor="file-upload"
                className="flex flex-col items-center cursor-pointer"
              >
                <svg
                  className="w-12 h-12 text-gray-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
                  />
                </svg>
                <p className="mt-2 text-sm text-gray-600">
                  <span className="font-medium text-blue-600">Click to upload</span> or drag and drop
                </p>
                <p className="mt-1 text-xs text-gray-500">CSV files only</p>
              </label>

              {createImport.isPending && (
                <div className="mt-4 flex items-center justify-center">
                  <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
                  <span className="ml-2 text-sm text-gray-600">Uploading...</span>
                </div>
              )}
            </div>
          </div>

          {/* Right Column - Import History & Details */}
          <div className="lg:col-span-2 space-y-6">
            {/* Import History */}
            <div className="rounded-lg bg-white shadow-sm border border-gray-200">
              <div className="px-4 py-3 border-b border-gray-200">
                <h2 className="font-medium text-gray-900">Import History</h2>
              </div>
              <div className="p-4">
                {jobsLoading ? (
                  <div className="animate-pulse space-y-3">
                    <div className="h-10 bg-gray-100 rounded"></div>
                    <div className="h-10 bg-gray-100 rounded"></div>
                    <div className="h-10 bg-gray-100 rounded"></div>
                  </div>
                ) : (
                  <ImportLogTable
                    jobs={jobs ?? []}
                    onSelect={setSelectedJob}
                    selectedJobId={selectedJob?.id}
                  />
                )}
              </div>
            </div>

            {/* Selected Job Details */}
            {selectedJobDetail && (
              <div className="rounded-lg bg-white shadow-sm border border-gray-200">
                <div className="px-4 py-3 border-b border-gray-200 flex items-center justify-between">
                  <h2 className="font-medium text-gray-900">Import Details</h2>
                  <button
                    onClick={() => setSelectedJob(undefined)}
                    className="text-gray-400 hover:text-gray-500"
                  >
                    ✕
                  </button>
                </div>
                <div className="p-4 space-y-4">
                  <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 text-sm">
                    <div>
                      <p className="text-gray-500">File</p>
                      <p className="font-medium truncate" title={selectedJobDetail.job.file_name}>
                        {selectedJobDetail.job.file_name}
                      </p>
                    </div>
                    <div>
                      <p className="text-gray-500">Status</p>
                      <p className="font-medium capitalize">{selectedJobDetail.job.status}</p>
                    </div>
                    <div>
                      <p className="text-gray-500">Processed</p>
                      <p className="font-medium">
                        {selectedJobDetail.job.processed_rows} / {selectedJobDetail.job.total_rows}
                      </p>
                    </div>
                    <div>
                      <p className="text-gray-500">Errors</p>
                      <p className={`font-medium ${selectedJobDetail.job.error_rows > 0 ? 'text-red-600' : 'text-green-600'}`}>
                        {selectedJobDetail.job.error_rows}
                      </p>
                    </div>
                  </div>

                  {selectedJobDetail.job.error_message && (
                    <div className="rounded-lg bg-red-50 p-3 border border-red-200">
                      <p className="text-red-800 text-sm">{selectedJobDetail.job.error_message}</p>
                    </div>
                  )}

                  <AnomalyTable anomalies={selectedJobDetail.anomalies ?? []} />
                </div>
              </div>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}

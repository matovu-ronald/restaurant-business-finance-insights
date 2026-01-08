'use client';

import { useState } from 'react';
import { MappingProfile, MappingsResponse, useCreateMapping } from '@/hooks/use-imports';

interface MappingProfileFormProps {
  mappings: MappingsResponse | undefined;
  selectedMappingId: string | undefined;
  onSelect: (id: string | undefined) => void;
  sourceType: string;
}

const SOURCE_TYPE_FIELDS: Record<string, string[]> = {
  pos: ['date', 'time', 'total', 'subtotal', 'tax', 'discounts', 'comps', 'payment_method', 'channel', 'server'],
  payroll: ['period_start', 'period_end', 'employee_name', 'hours_worked', 'hourly_rate', 'total_wages', 'superannuation', 'tax_withheld'],
  inventory: ['snapshot_date', 'item_name', 'category', 'quantity', 'unit', 'unit_cost', 'total_value'],
};

export function MappingProfileForm({
  mappings,
  selectedMappingId,
  onSelect,
  sourceType,
}: MappingProfileFormProps) {
  const [isCreating, setIsCreating] = useState(false);
  const [newProfileName, setNewProfileName] = useState('');
  const [columnMaps, setColumnMaps] = useState<Record<string, string>>({});
  const createMapping = useCreateMapping();

  const profiles = mappings?.profiles?.filter((p) => p.source_type === sourceType) ?? [];
  const defaultMapping = mappings?.defaults?.[sourceType] ?? {};
  const targetFields = SOURCE_TYPE_FIELDS[sourceType] ?? [];

  const handleCreate = async () => {
    if (!newProfileName.trim()) return;

    try {
      const profile = await createMapping.mutateAsync({
        name: newProfileName,
        sourceType,
        columnMaps,
      });
      onSelect(profile.id);
      setIsCreating(false);
      setNewProfileName('');
      setColumnMaps({});
    } catch (error) {
      console.error('Failed to create mapping:', error);
    }
  };

  if (isCreating) {
    return (
      <div className="rounded-lg border border-gray-200 bg-white p-4 space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="font-medium text-gray-900">Create Mapping Profile</h3>
          <button
            type="button"
            onClick={() => setIsCreating(false)}
            className="text-gray-400 hover:text-gray-500"
          >
            ✕
          </button>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700">Profile Name</label>
          <input
            type="text"
            value={newProfileName}
            onChange={(e) => setNewProfileName(e.target.value)}
            placeholder="e.g., Square POS Export"
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 text-sm"
          />
        </div>

        <div className="space-y-2">
          <p className="text-sm font-medium text-gray-700">Column Mappings</p>
          <p className="text-xs text-gray-500">Map your CSV column headers to system fields</p>

          <div className="grid grid-cols-2 gap-2">
            {targetFields.map((field) => (
              <div key={field} className="flex items-center gap-2">
                <span className="text-sm text-gray-600 w-32">{field}</span>
                <input
                  type="text"
                  placeholder="CSV column name"
                  value={columnMaps[field] ?? defaultMapping[field] ?? ''}
                  onChange={(e) =>
                    setColumnMaps((prev) => ({ ...prev, [e.target.value]: field }))
                  }
                  className="flex-1 rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 text-xs"
                />
              </div>
            ))}
          </div>
        </div>

        <div className="flex justify-end gap-2">
          <button
            type="button"
            onClick={() => setIsCreating(false)}
            className="px-3 py-1.5 text-sm text-gray-600 hover:text-gray-800"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={handleCreate}
            disabled={!newProfileName.trim() || createMapping.isPending}
            className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
          >
            {createMapping.isPending ? 'Creating...' : 'Create'}
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-2">
      <label className="block text-sm font-medium text-gray-700">Mapping Profile</label>
      <div className="flex gap-2">
        <select
          value={selectedMappingId ?? ''}
          onChange={(e) => onSelect(e.target.value || undefined)}
          className="flex-1 rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 text-sm"
        >
          <option value="">Use default mapping</option>
          {profiles.map((profile) => (
            <option key={profile.id} value={profile.id}>
              {profile.name}
            </option>
          ))}
        </select>
        <button
          type="button"
          onClick={() => setIsCreating(true)}
          className="px-3 py-2 text-sm text-blue-600 hover:text-blue-800"
        >
          + New
        </button>
      </div>
      <p className="text-xs text-gray-500">
        Mappings define how your CSV columns map to system fields
      </p>
    </div>
  );
}

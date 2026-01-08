'use client';

import { useEffect, useRef } from 'react';
import * as d3 from 'd3';
import { KPISummary } from '@/hooks/use-kpi';
import { formatCurrency } from '@/lib/dates';

interface ChannelDaypartChartsProps {
  byChannel: KPISummary[] | undefined;
  byDaypart: KPISummary[] | undefined;
  isLoading: boolean;
}

interface BarChartProps {
  data: KPISummary[];
  title: string;
  metric: 'revenue' | 'covers' | 'grossMargin';
}

function BarChart({ data, title, metric }: BarChartProps) {
  const svgRef = useRef<SVGSVGElement>(null);

  useEffect(() => {
    if (!svgRef.current || !data.length) return;

    const svg = d3.select(svgRef.current);
    svg.selectAll('*').remove();

    const margin = { top: 20, right: 30, bottom: 40, left: 60 };
    const width = 400 - margin.left - margin.right;
    const height = 250 - margin.top - margin.bottom;

    const g = svg
      .attr('width', width + margin.left + margin.right)
      .attr('height', height + margin.top + margin.bottom)
      .append('g')
      .attr('transform', `translate(${margin.left},${margin.top})`);

    // Scales
    const x = d3
      .scaleBand()
      .domain(data.map((d) => d.display_name))
      .range([0, width])
      .padding(0.3);

    const y = d3
      .scaleLinear()
      .domain([0, d3.max(data, (d) => d[metric]) ?? 0])
      .nice()
      .range([height, 0]);

    // Color scale
    const color = d3
      .scaleOrdinal<string>()
      .domain(data.map((d) => d.label))
      .range(['#3B82F6', '#10B981', '#F59E0B', '#EF4444']);

    // Axes
    g.append('g')
      .attr('transform', `translate(0,${height})`)
      .call(d3.axisBottom(x))
      .selectAll('text')
      .attr('transform', 'rotate(-15)')
      .style('text-anchor', 'end')
      .style('font-size', '12px');

    g.append('g')
      .call(
        d3.axisLeft(y).tickFormat((d) =>
          metric === 'covers' ? d.toString() : formatCurrency(d as number)
        )
      )
      .selectAll('text')
      .style('font-size', '11px');

    // Bars
    g.selectAll('.bar')
      .data(data)
      .join('rect')
      .attr('class', 'bar')
      .attr('x', (d) => x(d.display_name) ?? 0)
      .attr('y', (d) => y(d[metric]))
      .attr('width', x.bandwidth())
      .attr('height', (d) => height - y(d[metric]))
      .attr('fill', (d) => color(d.label))
      .attr('rx', 4)
      .attr('ry', 4);

    // Value labels
    g.selectAll('.label')
      .data(data)
      .join('text')
      .attr('class', 'label')
      .attr('x', (d) => (x(d.display_name) ?? 0) + x.bandwidth() / 2)
      .attr('y', (d) => y(d[metric]) - 5)
      .attr('text-anchor', 'middle')
      .style('font-size', '11px')
      .style('fill', '#374151')
      .text((d) =>
        metric === 'covers' ? d[metric].toLocaleString() : formatCurrency(d[metric])
      );
  }, [data, metric]);

  return (
    <div className="rounded-lg bg-white p-4 shadow-sm border border-gray-200">
      <h3 className="text-sm font-medium text-gray-700 mb-4">{title}</h3>
      <svg ref={svgRef} className="w-full"></svg>
    </div>
  );
}

function LoadingChart() {
  return (
    <div className="rounded-lg bg-white p-4 shadow-sm border border-gray-200 animate-pulse">
      <div className="h-4 bg-gray-200 rounded w-32 mb-4"></div>
      <div className="h-48 bg-gray-100 rounded"></div>
    </div>
  );
}

export function ChannelDaypartCharts({
  byChannel,
  byDaypart,
  isLoading,
}: ChannelDaypartChartsProps) {
  if (isLoading) {
    return (
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <LoadingChart />
        <LoadingChart />
        <LoadingChart />
        <LoadingChart />
      </div>
    );
  }

  const hasChannelData = byChannel && byChannel.length > 0;
  const hasDaypartData = byDaypart && byDaypart.length > 0;

  if (!hasChannelData && !hasDaypartData) {
    return (
      <div className="rounded-lg bg-gray-50 p-8 text-center border border-gray-200">
        <p className="text-gray-500">No breakdown data available</p>
        <p className="text-sm text-gray-400 mt-1">
          Import sales data to see channel and daypart analysis
        </p>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
      {hasChannelData && (
        <>
          <BarChart
            data={byChannel}
            title="Revenue by Channel"
            metric="revenue"
          />
          <BarChart
            data={byChannel}
            title="Covers by Channel"
            metric="covers"
          />
        </>
      )}
      {hasDaypartData && (
        <>
          <BarChart
            data={byDaypart}
            title="Revenue by Daypart"
            metric="revenue"
          />
          <BarChart
            data={byDaypart}
            title="Gross Margin by Daypart"
            metric="grossMargin"
          />
        </>
      )}
    </div>
  );
}

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { createFileRoute } from '@tanstack/react-router';
import { apiClient } from '@/lib/apiClient';
import type { HealthStatus } from '@/types/auth.types';
import { Activity, Database, Server, RefreshCw } from 'lucide-react';

export const Route = createFileRoute('/health')({
  component: HealthDashboard,
});

function HealthDashboard() {
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [lastChecked, setLastChecked] = useState<Date>(new Date());

  const { data: healthStatus, isLoading, refetch } = useQuery({
    queryKey: ['health'],
    queryFn: async () => {
      const response = await apiClient.get<HealthStatus>('/health');
      setLastChecked(new Date());
      return response.data;
    },
    refetchInterval: autoRefresh ? 30000 : false, // Refresh every 30 seconds if enabled
  });

  const handleManualRefresh = () => {
    refetch();
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'ok':
        return 'bg-green-100 text-green-800 border-green-300';
      case 'down':
        return 'bg-red-100 text-red-800 border-red-300';
      case 'disabled':
        return 'bg-gray-100 text-gray-600 border-gray-300';
      default:
        return 'bg-gray-100 text-gray-600 border-gray-300';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'ok':
        return '🟢';
      case 'down':
        return '🔴';
      case 'disabled':
        return '⚪';
      default:
        return '⚪';
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 flex items-center gap-2">
                <Activity className="h-8 w-8 text-indigo-600" />
                System Health
              </h1>
              <p className="mt-2 text-gray-600">
                Monitor the status of backend services
              </p>
            </div>
            <button
              onClick={handleManualRefresh}
              disabled={isLoading}
              className="flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:opacity-50 transition-colors"
            >
              <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
              Refresh
            </button>
          </div>
        </div>

        {/* Auto-refresh toggle */}
        <div className="mb-6 bg-white rounded-lg border border-gray-200 p-4">
          <label className="flex items-center justify-between cursor-pointer">
            <div>
              <span className="text-sm font-medium text-gray-900">Auto-refresh</span>
              <p className="text-sm text-gray-500">Automatically check status every 30 seconds</p>
            </div>
            <div className="relative">
              <input
                type="checkbox"
                checked={autoRefresh}
                onChange={(e) => setAutoRefresh(e.target.checked)}
                className="sr-only peer"
              />
              <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-indigo-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-indigo-600"></div>
            </div>
          </label>
        </div>

        {/* Last checked */}
        <div className="mb-6 text-sm text-gray-500">
          Last checked: {lastChecked.toLocaleTimeString()}
        </div>

        {/* Status Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Database Status */}
          <div className="bg-white rounded-lg border border-gray-200 p-6 shadow-sm">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-3">
                <Database className="h-8 w-8 text-gray-700" />
                <h2 className="text-xl font-semibold text-gray-900">MongoDB</h2>
              </div>
            </div>
            {isLoading ? (
              <div className="flex items-center gap-2 text-gray-500">
                <RefreshCw className="h-4 w-4 animate-spin" />
                <span>Checking...</span>
              </div>
            ) : (
              <div className="flex items-center gap-2">
                <span className="text-2xl">{getStatusIcon(healthStatus?.database || 'down')}</span>
                <span
                  className={`px-3 py-1 rounded-full text-sm font-medium border ${getStatusColor(
                    healthStatus?.database || 'down'
                  )}`}
                >
                  {healthStatus?.database?.toUpperCase() || 'UNKNOWN'}
                </span>
              </div>
            )}
            <p className="mt-4 text-sm text-gray-600">
              Primary database for storing users and todos
            </p>
          </div>

          {/* Cache Status */}
          <div className="bg-white rounded-lg border border-gray-200 p-6 shadow-sm">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-3">
                <Server className="h-8 w-8 text-gray-700" />
                <h2 className="text-xl font-semibold text-gray-900">Redis Cache</h2>
              </div>
            </div>
            {isLoading ? (
              <div className="flex items-center gap-2 text-gray-500">
                <RefreshCw className="h-4 w-4 animate-spin" />
                <span>Checking...</span>
              </div>
            ) : (
              <div className="flex items-center gap-2">
                <span className="text-2xl">{getStatusIcon(healthStatus?.cache || 'disabled')}</span>
                <span
                  className={`px-3 py-1 rounded-full text-sm font-medium border ${getStatusColor(
                    healthStatus?.cache || 'disabled'
                  )}`}
                >
                  {healthStatus?.cache?.toUpperCase() || 'UNKNOWN'}
                </span>
              </div>
            )}
            <p className="mt-4 text-sm text-gray-600">
              {healthStatus?.cache === 'disabled'
                ? 'Caching is currently disabled'
                : 'Cache for username availability checks'}
            </p>
          </div>
        </div>

        {/* Overall Status Summary */}
        <div className="mt-8 bg-white rounded-lg border border-gray-200 p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">System Status</h3>
          {isLoading ? (
            <p className="text-gray-500">Loading...</p>
          ) : healthStatus?.database === 'ok' ? (
            <div className="flex items-center gap-2 text-green-700">
              <span className="text-xl">✓</span>
              <span className="font-medium">All critical services are operational</span>
            </div>
          ) : (
            <div className="flex items-center gap-2 text-red-700">
              <span className="text-xl">✗</span>
              <span className="font-medium">One or more critical services are down</span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

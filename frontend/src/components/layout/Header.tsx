'use client';

import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/lib/auth-context';

const roleColors: Record<string, string> = {
  owner_admin: 'bg-purple-100 text-purple-800',
  manager: 'bg-blue-100 text-blue-800',
  accountant: 'bg-green-100 text-green-800',
  viewer: 'bg-gray-100 text-gray-800',
};

const roleLabels: Record<string, string> = {
  owner_admin: 'Admin',
  manager: 'Manager',
  accountant: 'Accountant',
  viewer: 'Viewer',
};

export function Header() {
  const { user, isAuthenticated, logout } = useAuth();
  const router = useRouter();

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  return (
    <header className="bg-white shadow-sm border-b border-gray-200">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          <div className="flex items-center space-x-8">
            <Link href="/dashboard" className="flex items-center space-x-2">
              <span className="text-2xl">🏠</span>
              <span className="text-xl font-bold text-gray-900">The Lakehouse</span>
            </Link>
            
            {isAuthenticated && (
              <nav className="hidden md:flex space-x-4">
                <Link
                  href="/dashboard"
                  className="text-gray-600 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium"
                >
                  Dashboard
                </Link>
                <Link
                  href="/imports"
                  className="text-gray-600 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium"
                >
                  Imports
                </Link>
                <Link
                  href="/drilldown"
                  className="text-gray-600 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium"
                >
                  Drill-down
                </Link>
              </nav>
            )}
          </div>

          <div className="flex items-center space-x-4">
            {isAuthenticated && user ? (
              <>
                <div className="flex items-center space-x-3">
                  <span className="text-sm text-gray-600">{user.email}</span>
                  <span
                    className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                      roleColors[user.role] || 'bg-gray-100 text-gray-800'
                    }`}
                  >
                    {roleLabels[user.role] || user.role}
                  </span>
                </div>
                <button
                  onClick={handleLogout}
                  className="text-gray-500 hover:text-gray-700 px-3 py-2 rounded-md text-sm font-medium"
                >
                  Sign out
                </button>
              </>
            ) : (
              <Link
                href="/login"
                className="text-blue-600 hover:text-blue-800 px-3 py-2 rounded-md text-sm font-medium"
              >
                Sign in
              </Link>
            )}
          </div>
        </div>
      </div>
    </header>
  );
}

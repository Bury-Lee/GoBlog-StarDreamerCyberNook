import { Link, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../../store/authStore';
import { PenSquare, Home, User, LogOut, Menu, X } from 'lucide-react';
import { useState } from 'react';

export default function Navbar() {
  const { user, isAuthenticated, logout } = useAuthStore();
  const navigate = useNavigate();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const handleLogout = () => {
    logout();
    navigate('/');
    setMobileMenuOpen(false);
  };

  return (
    <nav className="fixed top-0 left-0 right-0 z-50 bg-white/95 backdrop-blur-xl border-b border-slate-200 shadow-sm">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16 lg:h-18">
          <Link to="/" className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center shadow-md">
              <span className="text-white font-bold text-lg">G</span>
            </div>
            <span className="text-xl font-bold bg-gradient-to-r from-blue-600 to-cyan-600 bg-clip-text text-transparent">
              StarDreamer
            </span>
          </Link>

          <div className="hidden lg:flex items-center gap-8">
            <Link to="/" className="text-slate-700 hover:text-blue-600 transition-colors flex items-center gap-2 font-medium">
              <Home size={18} />
              首页
            </Link>
            <Link to="/articles" className="text-slate-700 hover:text-blue-600 transition-colors font-medium">
              文章
            </Link>
            {isAuthenticated && (
              <Link
                to="/editor"
                className="btn-primary flex items-center gap-2"
              >
                <PenSquare size={18} />
                写文章
              </Link>
            )}
          </div>

          <div className="hidden lg:flex items-center gap-4">
            {isAuthenticated ? (
              <div className="flex items-center gap-4">
                <Link to="/profile" className="flex items-center gap-2 text-slate-700 hover:text-blue-600 transition-colors">
                  <div className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center">
                    <User size={16} className="text-white" />
                  </div>
                  <span className="font-medium">{user?.nick_name || user?.user_name}</span>
                </Link>
                <button
                  onClick={handleLogout}
                  className="p-2 text-slate-500 hover:text-red-500 transition-colors"
                  title="退出登录"
                >
                  <LogOut size={18} />
                </button>
              </div>
            ) : (
              <div className="flex items-center gap-3">
                <Link to="/login" className="btn-secondary">
                  登录
                </Link>
                <Link to="/register" className="btn-primary">
                  注册
                </Link>
              </div>
            )}
          </div>

          <button
            className="lg:hidden p-2 text-slate-600 hover:text-blue-600 transition-colors"
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
          >
            {mobileMenuOpen ? <X size={24} /> : <Menu size={24} />}
          </button>
        </div>
      </div>

      {mobileMenuOpen && (
        <div className="lg:hidden bg-white border-t border-slate-200 shadow-lg">
          <div className="px-4 py-5 space-y-1">
            <Link
              to="/"
              className="flex items-center gap-3 text-slate-700 hover:text-blue-600 hover:bg-blue-50 py-3 px-4 rounded-lg transition-colors"
              onClick={() => setMobileMenuOpen(false)}
            >
              <Home size={18} />
              首页
            </Link>
            <Link
              to="/articles"
              className="flex items-center gap-3 text-slate-700 hover:text-blue-600 hover:bg-blue-50 py-3 px-4 rounded-lg transition-colors"
              onClick={() => setMobileMenuOpen(false)}
            >
              <span>📚</span>
              文章
            </Link>
            {isAuthenticated ? (
              <>
                <Link
                  to="/editor"
                  className="flex items-center gap-3 text-slate-700 hover:text-blue-600 hover:bg-blue-50 py-3 px-4 rounded-lg transition-colors"
                  onClick={() => setMobileMenuOpen(false)}
                >
                  <PenSquare size={18} />
                  写文章
                </Link>
                <Link
                  to="/profile"
                  className="flex items-center gap-3 text-slate-700 hover:text-blue-600 hover:bg-blue-50 py-3 px-4 rounded-lg transition-colors"
                  onClick={() => setMobileMenuOpen(false)}
                >
                  <User size={18} />
                  个人中心
                </Link>
                <button
                  onClick={handleLogout}
                  className="flex items-center gap-3 w-full text-left text-red-500 hover:text-red-600 hover:bg-red-50 py-3 px-4 rounded-lg transition-colors"
                >
                  <LogOut size={18} />
                  退出登录
                </button>
              </>
            ) : (
              <div className="flex gap-3 pt-4 mt-2 border-t border-slate-200">
                <Link to="/login" className="btn-secondary flex-1 text-center">
                  登录
                </Link>
                <Link to="/register" className="btn-primary flex-1 text-center">
                  注册
                </Link>
              </div>
            )}
          </div>
        </div>
      )}
    </nav>
  );
}

import { useAuthStore } from '../../store/authStore';
import Layout from '../../components/layout/Layout';
import { User, Mail } from 'lucide-react';
import { Link } from 'react-router-dom';

export default function ProfilePage() {
  const { user } = useAuthStore();

  if (!user) {
    return (
      <Layout>
        <div className="min-h-[calc(100vh-64px)] flex items-center justify-center">
          <p className="text-slate-600">请先登录</p>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8 lg:py-10">
        <div className="card p-6 lg:p-8 mb-6 lg:mb-8">
          <div className="flex flex-col sm:flex-row items-center sm:items-start gap-4 sm:gap-6">
            <div className="w-20 h-20 sm:w-24 sm:h-24 rounded-2xl bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center text-white text-2xl sm:text-3xl font-bold flex-shrink-0">
              {user.nick_name?.[0] || user.user_name?.[0]}
            </div>
            <div className="flex-1 text-center sm:text-left">
              <h1 className="text-xl sm:text-2xl font-bold text-slate-800 mb-1">
                {user.nick_name || user.user_name}
              </h1>
              <p className="text-slate-600 mb-2 text-sm sm:text-base">@{user.user_name}</p>
              {user.email && (
                <p className="text-slate-500 text-xs sm:text-sm flex items-center gap-2 justify-center sm:justify-start">
                  <Mail size={14} />
                  {user.email}
                </p>
              )}
            </div>
            <Link to="/settings" className="btn-secondary text-sm w-full sm:w-auto">
              编辑资料
            </Link>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 lg:gap-6">
          <Link
            to="/my-articles"
            className="card p-5 lg:p-6 flex items-center gap-4 hover:border-blue-500/50 transition-all"
          >
            <div className="w-10 h-10 lg:w-12 lg:h-12 rounded-xl bg-blue-100 flex items-center justify-center flex-shrink-0">
              <User className="text-blue-600" size={20} />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-slate-800 font-semibold text-sm lg:text-base">我的文章</p>
              <p className="text-slate-600 text-xs lg:text-sm">查看和管理</p>
            </div>
          </Link>

          <Link
            to="/my-collections"
            className="card p-5 lg:p-6 flex items-center gap-4 hover:border-cyan-500/50 transition-all"
          >
            <div className="w-10 h-10 lg:w-12 lg:h-12 rounded-xl bg-cyan-100 flex items-center justify-center flex-shrink-0">
              <User className="text-cyan-600" size={20} />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-slate-800 font-semibold text-sm lg:text-base">我的收藏</p>
              <p className="text-slate-600 text-xs lg:text-sm">收藏的文章</p>
            </div>
          </Link>

          <Link
            to="/my-history"
            className="card p-5 lg:p-6 flex items-center gap-4 hover:border-emerald-500/50 transition-all"
          >
            <div className="w-10 h-10 lg:w-12 lg:h-12 rounded-xl bg-emerald-100 flex items-center justify-center flex-shrink-0">
              <User className="text-emerald-600" size={20} />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-slate-800 font-semibold text-sm lg:text-base">浏览历史</p>
              <p className="text-slate-600 text-xs lg:text-sm">最近浏览</p>
            </div>
          </Link>
        </div>

        {user.role === 1 && (
          <div className="mt-6 lg:mt-8">
            <h2 className="text-base lg:text-lg font-semibold text-slate-800 mb-4">管理员</h2>
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 lg:gap-4">
              <Link
                to="/admin/articles"
                className="card p-3 lg:p-4 text-center hover:border-blue-500/50 transition-all"
              >
                <p className="text-slate-800 text-xs lg:text-sm font-medium">文章管理</p>
              </Link>
              <Link
                to="/admin/comments"
                className="card p-3 lg:p-4 text-center hover:border-cyan-500/50 transition-all"
              >
                <p className="text-slate-800 text-xs lg:text-sm font-medium">评论管理</p>
              </Link>
              <Link
                to="/admin/users"
                className="card p-3 lg:p-4 text-center hover:border-amber-500/50 transition-all"
              >
                <p className="text-slate-800 text-xs lg:text-sm font-medium">用户管理</p>
              </Link>
              <Link
                to="/admin/settings"
                className="card p-3 lg:p-4 text-center hover:border-emerald-500/50 transition-all"
              >
                <p className="text-slate-800 text-xs lg:text-sm font-medium">站点设置</p>
              </Link>
            </div>
          </div>
        )}
      </div>
    </Layout>
  );
}

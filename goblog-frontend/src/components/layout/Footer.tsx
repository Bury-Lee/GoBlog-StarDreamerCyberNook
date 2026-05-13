import { Link } from 'react-router-dom';
import { Mail, Heart } from 'lucide-react';

export default function Footer() {
  return (
    <footer className="bg-gradient-to-b from-white to-slate-50 border-t border-slate-200 mt-auto">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12 lg:py-16">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8 lg:gap-12">
          <div className="lg:col-span-2">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center shadow-md">
                <span className="text-white font-bold text-lg">G</span>
              </div>
              <span className="text-xl font-bold text-slate-800">StarDreamer</span>
            </div>
            <p className="text-slate-600 max-w-md leading-relaxed mb-4">
              一个基于 Gin + GORM + Redis + Elasticsearch 的现代化博客社区平台。
              追求独立与自主的技术爱好者社区。
            </p>
            <div className="flex items-center gap-3">
              <a
                href="mailto:contact@stardreamer.com"
                className="w-10 h-10 rounded-lg bg-slate-100 hover:bg-blue-100 flex items-center justify-center text-slate-600 hover:text-blue-600 transition-colors"
                title="发送邮件"
              >
                <Mail size={20} />
              </a>
            </div>
          </div>

          <div>
            <h3 className="text-slate-800 font-semibold mb-4">快速链接</h3>
            <ul className="space-y-2">
              <li>
                <Link to="/" className="text-slate-600 hover:text-blue-600 transition-colors inline-block">
                  首页
                </Link>
              </li>
              <li>
                <Link to="/articles" className="text-slate-600 hover:text-blue-600 transition-colors inline-block">
                  文章列表
                </Link>
              </li>
              <li>
                <Link to="/about" className="text-slate-600 hover:text-blue-600 transition-colors inline-block">
                  关于我们
                </Link>
              </li>
            </ul>
          </div>

          <div>
            <h3 className="text-slate-800 font-semibold mb-4">技术栈</h3>
            <ul className="space-y-2 text-slate-600">
              <li className="flex items-center gap-2">
                <span className="w-1.5 h-1.5 rounded-full bg-blue-500"></span>
                Go + Gin 框架
              </li>
              <li className="flex items-center gap-2">
                <span className="w-1.5 h-1.5 rounded-full bg-cyan-500"></span>
                React + TypeScript
              </li>
              <li className="flex items-center gap-2">
                <span className="w-1.5 h-1.5 rounded-full bg-amber-500"></span>
                Redis 缓存
              </li>
              <li className="flex items-center gap-2">
                <span className="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
                Elasticsearch 搜索
              </li>
            </ul>
          </div>
        </div>

        <div className="border-t border-slate-200 mt-10 lg:mt-12 pt-8">
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4 text-center">
            <p className="text-slate-500 text-sm">
              © 2024 StarDreamer. 基于 GoBlog 构建。
            </p>
            <p className="text-slate-500 text-sm flex items-center gap-1">
              用 <Heart size={14} className="text-red-400" /> 打造
            </p>
          </div>
        </div>
      </div>
    </footer>
  );
}

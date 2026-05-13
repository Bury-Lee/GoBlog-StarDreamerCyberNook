import { Link } from 'react-router-dom';
import type { Article } from '../../types';
import { Eye, MessageCircle, Heart, Bookmark } from 'lucide-react';

interface ArticleCardProps {
  article: Article;
}

export default function ArticleCard({ article }: ArticleCardProps) {
  return (
    <Link
      to={`/article/${article.id}`}
      className="card p-5 lg:p-6 block group hover:transform hover:scale-[1.02] transition-all duration-300"
    >
      {article.cover && (
        <div className="mb-4 overflow-hidden rounded-lg">
          <img
            src={article.cover}
            alt={article.title}
            className="w-full h-44 lg:h-48 object-cover group-hover:scale-110 transition-transform duration-500"
          />
        </div>
      )}

      <div className="flex flex-wrap items-center gap-2 mb-3">
        {article.category && (
          <span className="px-3 py-1 text-xs font-medium bg-blue-50 text-blue-600 rounded-full">
            {article.category.name}
          </span>
        )}
        {article.tag_list?.slice(0, 2).map((tag) => (
          <span
            key={tag}
            className="px-2.5 py-0.5 text-xs bg-slate-100 text-slate-600 rounded-full"
          >
            #{tag}
          </span>
        ))}
      </div>

      <h3 className="text-lg lg:text-xl font-semibold text-slate-800 mb-2 group-hover:text-blue-600 transition-colors line-clamp-2 leading-snug">
        {article.title}
      </h3>

      {article.abstract && (
        <p className="text-slate-600 text-sm mb-4 line-clamp-2 leading-relaxed">
          {article.abstract}
        </p>
      )}

      <div className="flex items-center justify-between text-slate-500 text-sm">
        <div className="flex items-center gap-3 lg:gap-4">
          <span className="flex items-center gap-1">
            <Eye size={14} />
            <span className="hidden sm:inline">{article.look_count}</span>
          </span>
          <span className="flex items-center gap-1">
            <MessageCircle size={14} />
            <span className="hidden sm:inline">{article.comment_count}</span>
          </span>
          <span className="flex items-center gap-1">
            <Heart size={14} />
            <span className="hidden sm:inline">{article.digg_count}</span>
          </span>
        </div>
        <span className="flex items-center gap-1">
          <Bookmark size={14} />
          <span className="hidden sm:inline">{article.collect_count}</span>
        </span>
      </div>

      <div className="flex items-center gap-3 mt-4 pt-4 border-t border-slate-200">
        {article.user && (
          <>
            <div className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center text-white text-sm font-medium flex-shrink-0">
              {article.user.nick_name?.[0] || article.user.user_name?.[0]}
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm text-slate-700 truncate">{article.user.nick_name || article.user.user_name}</p>
              <p className="text-xs text-slate-500">
                {new Date(article.created_at).toLocaleDateString('zh-CN')}
              </p>
            </div>
          </>
        )}
      </div>
    </Link>
  );
}

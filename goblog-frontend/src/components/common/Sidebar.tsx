import { TrendingUp, MessageCircle, Tag, Link2 } from 'lucide-react';

interface SidebarProps {
  hotArticles?: any[];
  recentComments?: any[];
  tags?: string[];
}

export default function Sidebar({ hotArticles = [], recentComments = [], tags = [] }: SidebarProps) {
  return (
    <div className="space-y-5 lg:space-y-6">
      {hotArticles.length > 0 && (
        <div className="card p-5 lg:p-6">
          <div className="flex items-center gap-2 mb-4">
            <TrendingUp className="text-blue-500" size={20} />
            <h3 className="text-lg font-semibold text-slate-800">热门文章</h3>
          </div>
          <ul className="space-y-3">
            {hotArticles.map((article, index) => (
              <li key={article.id}>
                <a
                  href={`/article/${article.id}`}
                  className="flex items-start gap-3 group"
                >
                  <span className={`flex-shrink-0 w-6 h-6 rounded flex items-center justify-center text-sm font-bold ${
                    index < 3 ? 'bg-blue-100 text-blue-600' : 'bg-slate-200 text-slate-600'
                  }`}>
                    {index + 1}
                  </span>
                  <span className="text-sm text-slate-700 group-hover:text-blue-600 transition-colors line-clamp-2 leading-relaxed">
                    {article.title}
                  </span>
                </a>
              </li>
            ))}
          </ul>
        </div>
      )}

      {recentComments.length > 0 && (
        <div className="card p-5 lg:p-6">
          <div className="flex items-center gap-2 mb-4">
            <MessageCircle className="text-cyan-500" size={20} />
            <h3 className="text-lg font-semibold text-slate-800">最新评论</h3>
          </div>
          <ul className="space-y-3">
            {recentComments.map((comment) => (
              <li key={comment.id} className="text-sm">
                <p className="text-slate-700 line-clamp-2 leading-relaxed">{comment.content}</p>
                <p className="text-slate-500 text-xs mt-1">
                  {comment.user?.nick_name || comment.user?.user_name} · {comment.article?.title}
                </p>
              </li>
            ))}
          </ul>
        </div>
      )}

      {tags.length > 0 && (
        <div className="card p-5 lg:p-6">
          <div className="flex items-center gap-2 mb-4">
            <Tag className="text-amber-500" size={20} />
            <h3 className="text-lg font-semibold text-slate-800">标签云</h3>
          </div>
          <div className="flex flex-wrap gap-2">
            {tags.map((tag) => (
              <a
                key={tag}
                href={`/articles?tag=${tag}`}
                className="px-3 py-1.5 text-sm bg-slate-100 text-slate-700 rounded-full hover:bg-blue-100 hover:text-blue-600 transition-colors font-medium"
              >
                #{tag}
              </a>
            ))}
          </div>
        </div>
      )}

      <div className="card p-5 lg:p-6">
        <div className="flex items-center gap-2 mb-3">
          <Link2 className="text-emerald-500" size={20} />
          <h3 className="text-lg font-semibold text-slate-800">友链申请</h3>
        </div>
        <p className="text-sm text-slate-600 mb-4 leading-relaxed">
          欢迎交换友链，联系方式见关于页面。
        </p>
        <button className="btn-secondary w-full text-sm">
          申请友链
        </button>
      </div>
    </div>
  );
}

import { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import Layout from '../../components/layout/Layout';
import BannerCarousel from '../../components/common/BannerCarousel';
import ArticleCard from '../../components/common/ArticleCard';
import Sidebar from '../../components/common/Sidebar';
import { articleApi, siteApi } from '../../api';
import type { Article, Banner } from '../../types';
import { Sparkles, TrendingUp, Clock, Search } from 'lucide-react';

const sortOptions = [
  { label: '最新发布', icon: Clock, value: 0 },
  { label: '猜你喜欢', icon: Sparkles, value: 1 },
  { label: '最多回复', icon: TrendingUp, value: 2 },
];

export default function HomePage() {
  const [articles, setArticles] = useState<Article[]>([]);
  const [banners, setBanners] = useState<Banner[]>([]);
  const [loading, setLoading] = useState(true);
  const [sortType, setSortType] = useState(0);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const [articlesRes, bannersRes] = await Promise.all([
        articleApi.getArticleList({ page: 1, pageSize: 12, type: 'other', status: 2 }),
        siteApi.getBannerList(),
      ]);

      if (articlesRes.code === 200) {
        setArticles(articlesRes.data.list || []);
        setHasMore(articlesRes.data.list?.length === 12);
      }
      if (bannersRes.code === 200) {
        setBanners(bannersRes.data || []);
      }
    } catch (error) {
      console.error('数据加载失败'); // 生产环境可考虑上报错误监控系统
    } finally {
      setLoading(false);
    }
  };

  const loadMore = async () => {
    if (!hasMore || loading) return;
    try {
      const nextPage = page + 1;
      const res = await articleApi.getArticleList({ page: nextPage, pageSize: 12, type: 'other', status: 2 });
      if (res.code === 200) {
        setArticles((prev) => [...prev, ...(res.data.list || [])]);
        setPage(nextPage);
        setHasMore(res.data.list?.length === 12);
      }
    } catch (error) {
      console.error('加载更多失败');
    }
  };

  return (
    <Layout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 lg:py-12">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
        >
          <BannerCarousel banners={banners} />
        </motion.div>

        <div className="mt-10 lg:mt-12 grid grid-cols-1 lg:grid-cols-4 gap-6 lg:gap-10">
          <div className="lg:col-span-3 order-2 lg:order-1">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between mb-6 lg:mb-8 gap-4">
              <div className="flex flex-wrap gap-2">
                {sortOptions.map((option) => (
                  <button
                    key={option.value}
                    onClick={() => setSortType(option.value)}
                    className={`flex items-center gap-2 px-4 py-2 rounded-lg transition-all ${
                      sortType === option.value
                        ? 'bg-blue-100 text-blue-600'
                        : 'bg-slate-100 text-slate-600 hover:text-slate-800'
                    }`}
                  >
                    <option.icon size={16} />
                    <span className="text-sm font-medium">{option.label}</span>
                  </button>
                ))}
              </div>

              <a
                href="/articles"
                className="flex items-center gap-2 text-blue-600 hover:text-blue-700 transition-colors text-sm"
              >
                <Search size={16} />
                <span>搜索文章</span>
              </a>
            </div>

            {loading ? (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-5 lg:gap-6">
                {[...Array(6)].map((_, i) => (
                  <div key={i} className="card p-5 lg:p-6 animate-pulse">
                    <div className="w-full h-44 lg:h-48 bg-slate-200 rounded-lg mb-4" />
                    <div className="h-6 bg-slate-200 rounded w-3/4 mb-3" />
                    <div className="h-4 bg-slate-200 rounded w-full mb-2" />
                    <div className="h-4 bg-slate-200 rounded w-2/3" />
                  </div>
                ))}
              </div>
            ) : articles.length > 0 ? (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-5 lg:gap-6">
                {articles.map((article, index) => (
                  <motion.div
                    key={article.id}
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.3, delay: index * 0.05 }}
                  >
                    <ArticleCard article={article} />
                  </motion.div>
                ))}
              </div>
            ) : (
              <div className="card p-12 text-center">
                <p className="text-slate-600 mb-4">暂无文章</p>
                <a href="/editor" className="btn-primary inline-flex items-center gap-2">
                  撰写第一篇文章
                </a>
              </div>
            )}

            {hasMore && articles.length > 0 && (
              <div className="mt-10 lg:mt-12 text-center">
                <button
                  onClick={loadMore}
                  disabled={loading}
                  className="btn-secondary"
                >
                  {loading ? '加载中...' : '加载更多'}
                </button>
              </div>
            )}
          </div>

          <div className="lg:col-span-1 order-1 lg:order-2 mb-6 lg:mb-0">
            <div className="lg:sticky lg:top-24">
              <Sidebar
                hotArticles={articles.slice(0, 5)}
                tags={['React', 'TypeScript', 'Go', 'Gin', 'Redis', 'Docker']}
              />
            </div>
          </div>
        </div>
      </div>
    </Layout>
  );
}

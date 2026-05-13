import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import Layout from '../../components/layout/Layout';
import ArticleCard from '../../components/common/ArticleCard';
import Sidebar from '../../components/common/Sidebar';
import { articleApi } from '../../api';
import type { Article, Category } from '../../types';
import { Search } from 'lucide-react';

export default function ArticleListPage() {
  const [articles, setArticles] = useState<Article[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [keyword, setKeyword] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<number | null>(null);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);

  useEffect(() => {
    fetchCategories();
  }, []);

  useEffect(() => {
    fetchArticles();
  }, [selectedCategory]);

  const fetchCategories = async () => {
    try {
      const res = await articleApi.getCategories();
      if (res.code === 200) {
        setCategories(res.data || []);
      }
    } catch (error) {
      console.error('加载分类失败');
    }
  };

  const fetchArticles = async (searchPage = 1) => {
    try {
      setLoading(true);
      let res;
      if (keyword) {
        res = await articleApi.searchArticles({ keyword, page: searchPage, pageSize: 12 });
      } else {
        res = await articleApi.getArticleList({
          page: searchPage,
          pageSize: 12,
          type: 'other',
          categoryID: selectedCategory || undefined,
          status: 2,
        });
      }
      if (res.code === 200) {
        setArticles(searchPage === 1 ? res.data.list || [] : [...articles, ...(res.data.list || [])]);
        setHasMore(res.data.list?.length === 12);
        setPage(searchPage);
      }
    } catch (error) {
      console.error('加载文章失败');
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    fetchArticles(1);
  };

  const loadMore = () => {
    fetchArticles(page + 1);
  };

  return (
    <Layout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 lg:py-12">
        <form onSubmit={handleSearch} className="mb-8">
          <div className="flex flex-col sm:flex-row gap-3">
            <div className="flex-1 relative">
              <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-500" size={20} />
              <input
                type="text"
                value={keyword}
                onChange={(e) => setKeyword(e.target.value)}
                placeholder="搜索文章..."
                className="input-field pl-12 w-full"
              />
            </div>
            <button type="submit" className="btn-primary px-8 whitespace-nowrap">
              搜索
            </button>
          </div>
        </form>

        <div className="flex gap-2 mb-8 overflow-x-auto pb-2">
          <button
            onClick={() => setSelectedCategory(null)}
            className={`px-4 py-2 rounded-lg whitespace-nowrap transition-all ${
              selectedCategory === null
                ? 'bg-blue-100 text-blue-600'
                : 'bg-slate-100 text-slate-600 hover:text-slate-800'
            }`}
          >
            全部
          </button>
          {categories.map((cat) => (
            <button
              key={cat.id}
              onClick={() => setSelectedCategory(cat.id)}
              className={`px-4 py-2 rounded-lg whitespace-nowrap transition-all ${
                selectedCategory === cat.id
                  ? 'bg-blue-100 text-blue-600'
                  : 'bg-slate-100 text-slate-600 hover:text-slate-800'
              }`}
            >
              {cat.name}
            </button>
          ))}
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6 lg:gap-10">
          <div className="lg:col-span-3">
            {loading && articles.length === 0 ? (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-5 lg:gap-6">
                {[...Array(6)].map((_, i) => (
                  <div key={i} className="card p-5 lg:p-6 animate-pulse">
                    <div className="w-full h-44 lg:h-48 bg-slate-200 rounded-lg mb-4" />
                    <div className="h-6 bg-slate-200 rounded w-3/4 mb-3" />
                    <div className="h-4 bg-slate-200 rounded w-full" />
                  </div>
                ))}
              </div>
            ) : articles.length > 0 ? (
              <>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-5 lg:gap-6">
                  {articles.map((article) => (
                    <ArticleCard key={article.id} article={article} />
                  ))}
                </div>
                {hasMore && (
                  <div className="mt-10 lg:mt-12 text-center">
                    <button onClick={loadMore} disabled={loading} className="btn-secondary">
                      {loading ? '加载中...' : '加载更多'}
                    </button>
                  </div>
                )}
              </>
            ) : (
              <div className="card p-12 text-center">
                <p className="text-slate-600 mb-4">暂无文章</p>
                <Link to="/editor" className="btn-primary inline-flex items-center gap-2">
                  撰写第一篇文章
                </Link>
              </div>
            )}
          </div>

          <div className="lg:col-span-1">
            <div className="lg:sticky lg:top-24">
              <Sidebar tags={['React', 'TypeScript', 'Go', 'Gin', 'Redis', 'Docker', 'Kubernetes']} />
            </div>
          </div>
        </div>
      </div>
    </Layout>
  );
}

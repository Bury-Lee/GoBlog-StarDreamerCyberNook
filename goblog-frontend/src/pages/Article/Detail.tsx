import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import Layout from '../../components/layout/Layout';
import { articleApi, commentApi } from '../../api';
import type { Article, Comment } from '../../types';
import { Eye, Heart, MessageCircle, Bookmark, Share2, ArrowLeft } from 'lucide-react';
import { motion } from 'framer-motion';

export default function ArticleDetailPage() {
  const { id } = useParams();
  const [article, setArticle] = useState<Article | null>(null);
  const [comments, setComments] = useState<Comment[]>([]);
  const [loading, setLoading] = useState(true);
  const [newComment, setNewComment] = useState('');

  useEffect(() => {
    if (id) {
      fetchArticle(parseInt(id));
      fetchComments(parseInt(id));
    }
  }, [id]);

  const fetchArticle = async (articleId: number) => {
    try {
      setLoading(true);
      const res = await articleApi.getArticleDetail(articleId);
      if (res.code === 200) {
        setArticle(res.data);
        articleApi.lookArticle(articleId);
      }
    } catch (error) {
      console.error('加载文章失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchComments = async (articleId: number) => {
    try {
      const res = await commentApi.getCommentList(articleId);
      if (res.code === 200) {
        setComments(res.data.list || []);
      }
    } catch (error) {
      console.error('加载评论失败');
    }
  };

  const handleDigg = async () => {
    if (!id) return;
    try {
      await articleApi.diggArticle(parseInt(id));
      setArticle((prev) =>
        prev ? { ...prev, digg_count: prev.digg_count + 1 } : null
      );
    } catch (error) {
      console.error('点赞失败');
    }
  };

  const handleCollect = async () => {
    if (!id) return;
    try {
      await articleApi.collectArticle(parseInt(id));
      setArticle((prev) =>
        prev ? { ...prev, collect_count: prev.collect_count + 1 } : null
      );
    } catch (error) {
      console.error('收藏失败');
    }
  };

  const handleSubmitComment = async () => {
    if (!id || !newComment.trim()) return;
    try {
      await commentApi.createComment({
        content: newComment,
        article_id: parseInt(id),
      });
      setNewComment('');
      fetchComments(parseInt(id));
    } catch (error) {
      console.error('提交评论失败');
    }
  };

  if (loading) {
    return (
      <Layout>
        <div className="max-w-4xl mx-auto px-4 py-8 lg:py-10">
          <div className="animate-pulse space-y-4">
            <div className="h-10 bg-slate-200 rounded w-3/4" />
            <div className="h-6 bg-slate-200 rounded w-1/4" />
            <div className="h-64 lg:h-80 bg-slate-200 rounded-lg" />
          </div>
        </div>
      </Layout>
    );
  }

  if (!article) {
    return (
      <Layout>
        <div className="max-w-4xl mx-auto px-4 py-12 lg:py-16 text-center">
          <p className="text-slate-600 mb-4">文章不存在</p>
          <Link to="/" className="btn-primary inline-flex items-center gap-2">
            <ArrowLeft size={18} />
            返回首页
          </Link>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <article className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8 lg:py-10">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
        >
          <Link
            to="/"
            className="inline-flex items-center gap-2 text-slate-600 hover:text-blue-600 mb-6 lg:mb-8 transition-colors"
          >
            <ArrowLeft size={18} />
            <span className="hidden sm:inline">返回首页</span>
          </Link>

          {article.cover && (
            <img
              src={article.cover}
              alt={article.title}
              className="w-full h-64 sm:h-80 lg:h-96 object-cover rounded-2xl mb-8 lg:mb-10"
            />
          )}

          <div className="flex flex-wrap items-center gap-2 mb-4">
            {article.category && (
              <span className="px-3 py-1 text-sm bg-blue-100 text-blue-600 rounded-full font-medium">
                {article.category.name}
              </span>
            )}
            {article.tag_list?.map((tag) => (
              <span
                key={tag}
                className="px-2.5 py-0.5 text-sm bg-slate-100 text-slate-600 rounded-full"
              >
                #{tag}
              </span>
            ))}
          </div>

          <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold text-slate-800 mb-6 leading-tight">
            {article.title}
          </h1>

          <div className="flex flex-col sm:flex-row sm:items-center gap-4 mb-8 pb-8 border-b border-slate-200">
            {article.user && (
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-full bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center text-white font-medium">
                  {article.user.nick_name?.[0] || article.user.user_name?.[0]}
                </div>
                <div>
                  <p className="text-slate-800 font-medium">
                    {article.user.nick_name || article.user.user_name}
                  </p>
                  <p className="text-slate-500 text-sm">
                    {new Date(article.created_at).toLocaleDateString('zh-CN')}
                  </p>
                </div>
              </div>
            )}

            <div className="flex items-center gap-4 sm:gap-6 text-slate-600 sm:ml-auto text-sm">
              <span className="flex items-center gap-1.5">
                <Eye size={16} />
                <span className="hidden xs:inline">{article.look_count}</span>
              </span>
              <button
                onClick={handleDigg}
                className="flex items-center gap-1.5 hover:text-red-500 transition-colors"
              >
                <Heart size={16} />
                <span className="hidden xs:inline">{article.digg_count}</span>
              </button>
              <span className="flex items-center gap-1.5">
                <MessageCircle size={16} />
                <span className="hidden xs:inline">{article.comment_count}</span>
              </span>
              <button
                onClick={handleCollect}
                className="flex items-center gap-1.5 hover:text-amber-500 transition-colors"
              >
                <Bookmark size={16} />
                <span className="hidden xs:inline">{article.collect_count}</span>
              </button>
            </div>
          </div>

          <div
            className="prose max-w-none mb-10 lg:mb-12 leading-relaxed"
            dangerouslySetInnerHTML={{ __html: article.content }}
          />

          <div className="flex flex-wrap gap-3 mb-10 lg:mb-12">
            <button onClick={handleDigg} className="btn-secondary flex items-center gap-2 text-sm">
              <Heart size={16} />
              点赞
            </button>
            <button onClick={handleCollect} className="btn-secondary flex items-center gap-2 text-sm">
              <Bookmark size={16} />
              收藏
            </button>
            <button className="btn-secondary flex items-center gap-2 text-sm">
              <Share2 size={16} />
              分享
            </button>
          </div>

          <div className="card p-6 lg:p-8">
            <h3 className="text-lg lg:text-xl font-semibold text-slate-800 mb-6">
              评论 ({comments.length})
            </h3>

            {article.open_comment && (
              <div className="mb-6">
                <textarea
                  value={newComment}
                  onChange={(e) => setNewComment(e.target.value)}
                  placeholder="写下你的评论..."
                  className="input-field w-full min-h-[100px] resize-y mb-3"
                />
                <button onClick={handleSubmitComment} className="btn-primary">
                  发布评论
                </button>
              </div>
            )}

            <div className="space-y-4">
              {comments.map((comment) => (
                <div key={comment.id} className="flex gap-4 pb-4 border-b border-slate-100 last:border-0">
                  <div className="w-10 h-10 rounded-full bg-gradient-to-br from-cyan-500 to-blue-600 flex items-center justify-center text-white font-medium flex-shrink-0">
                    {comment.user?.nick_name?.[0] || comment.user?.user_name?.[0] || '?'}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex flex-wrap items-center gap-2 mb-1">
                      <span className="text-slate-800 font-medium text-sm">
                        {comment.user?.nick_name || comment.user?.user_name}
                      </span>
                      <span className="text-slate-500 text-xs">
                        {new Date(comment.created_at).toLocaleDateString('zh-CN')}
                      </span>
                    </div>
                    <p className="text-slate-700 text-sm leading-relaxed">{comment.content}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </motion.div>
      </article>
    </Layout>
  );
}

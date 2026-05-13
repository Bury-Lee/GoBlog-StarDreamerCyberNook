import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import Layout from '../../components/layout/Layout';
import { articleApi } from '../../api';
import { useAuthStore } from '../../store/authStore';
import { Save, Eye } from 'lucide-react';

export default function EditorPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { isAuthenticated } = useAuthStore();
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [abstract, setAbstract] = useState('');
  const [cover, setCover] = useState('');
  const [openComment, setOpenComment] = useState(true);
  const [isLoading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login');
      return;
    }

    if (id) {
      fetchArticle(parseInt(id));
    }
  }, [id]);

  const fetchArticle = async (articleId: number) => {
    try {
      setLoading(true);
      const res = await articleApi.getArticleDetail(articleId);
      if (res.code === 200) {
        const article = res.data;
        setTitle(article.title);
        setContent(article.content);
        setAbstract(article.abstract || '');
        setCover(article.cover || '');
        setOpenComment(article.open_comment);
      }
    } catch (error) {
      console.error('加载文章失败');
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async (asDraft = false) => {
    if (!title.trim() || !content.trim()) {
      alert('请填写标题和内容');
      return;
    }

    try {
      setSaving(true);
      const data = {
        title,
        content,
        abstract: abstract || undefined,
        cover: cover || undefined,
        openComment,
        ...(asDraft ? { status: 1 } : {}),
      };

      let res;
      if (id) {
        res = await articleApi.updateArticle({ id: parseInt(id), ...data });
      } else {
        res = await articleApi.createArticle(data);
      }

      if (res.code === 200) {
        alert(asDraft ? '保存草稿成功' : '发布成功');
        navigate('/');
      } else {
        alert(res.message || '保存失败');
      }
    } catch (error) {
      console.error('保存文章失败');
      alert('保存失败，请稍后重试');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Layout>
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8 lg:py-10">
        <div className="card p-6 lg:p-8">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between mb-6 gap-4">
            <h1 className="text-xl lg:text-2xl font-bold text-slate-800">
              {id ? '编辑文章' : '撰写文章'}
            </h1>
            <div className="flex gap-3">
              <button
                onClick={() => handleSave(true)}
                disabled={saving}
                className="btn-secondary flex items-center gap-2 text-sm"
              >
                <Save size={18} />
                保存草稿
              </button>
              <button
                onClick={() => handleSave(false)}
                disabled={saving}
                className="btn-primary flex items-center gap-2 text-sm"
              >
                <Eye size={18} />
                {saving ? '发布中...' : '发布'}
              </button>
            </div>
          </div>

          <div className="space-y-5">
            <div>
              <label className="block text-sm text-slate-700 mb-2 font-medium">封面图片URL</label>
              <input
                type="text"
                value={cover}
                onChange={(e) => setCover(e.target.value)}
                placeholder="https://example.com/image.jpg"
                className="input-field"
              />
            </div>

            <div>
              <label className="block text-sm text-slate-700 mb-2 font-medium">标题</label>
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="请输入文章标题"
                className="input-field text-base lg:text-lg"
                maxLength={32}
              />
              <p className="text-slate-500 text-xs lg:text-sm mt-1">{title.length}/32</p>
            </div>

            <div>
              <label className="block text-sm text-slate-700 mb-2 font-medium">摘要</label>
              <textarea
                value={abstract}
                onChange={(e) => setAbstract(e.target.value)}
                placeholder="请输入文章摘要（选填）"
                className="input-field min-h-[80px] resize-y"
                maxLength={256}
              />
              <p className="text-slate-500 text-xs lg:text-sm mt-1">{abstract.length}/256</p>
            </div>

            <div>
              <label className="block text-sm text-slate-700 mb-2 font-medium">内容</label>
              <textarea
                value={content}
                onChange={(e) => setContent(e.target.value)}
                placeholder="请输入文章内容（支持Markdown）"
                className="input-field min-h-[350px] sm:min-h-[400px] lg:min-h-[450px] resize-y font-mono text-sm"
              />
            </div>

            <div className="flex items-center gap-4">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={openComment}
                  onChange={(e) => setOpenComment(e.target.checked)}
                  className="w-4 h-4 rounded bg-slate-100 border-slate-300 text-blue-500 focus:ring-blue-500"
                />
                <span className="text-sm text-slate-700">开启评论</span>
              </label>
            </div>
          </div>
        </div>
      </div>
    </Layout>
  );
}

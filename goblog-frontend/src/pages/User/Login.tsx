import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import Layout from '../../components/layout/Layout';
import { userApi } from '../../api';
import { useAuthStore } from '../../store/authStore';
import { Mail, Lock, User } from 'lucide-react';

export default function LoginPage() {
  const navigate = useNavigate();
  const { setUser, setTokens } = useAuthStore();
  const [loginType, setLoginType] = useState<'用户名' | '邮箱'>('用户名');
  const [identity, setIdentity] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!identity || !password) {
      setError('请填写所有字段');
      return;
    }

    try {
      setLoading(true);
      const res = await userApi.login({
        type: loginType,
        val: identity,
        pwd: password,
      });

      if (res.code === 200) {
        setTokens(res.data.AccessToken, res.data.RefreshToken);

        const userRes = await userApi.getUserDetail();
        if (userRes.code === 200) {
          setUser(userRes.data);
        }

        navigate('/');
      } else {
        setError(res.message || '登录失败');
      }
    } catch (err: any) {
      setError(err.response?.data?.message || '登录失败，请检查用户名和密码');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Layout>
      <div className="min-h-[calc(100vh-64px)] flex items-center justify-center px-4 py-10 lg:py-16">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          className="w-full max-w-md"
        >
          <div className="card p-8 lg:p-10">
            <div className="text-center mb-8">
              <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center shadow-lg">
                <span className="text-white text-2xl font-bold">G</span>
              </div>
              <h1 className="text-2xl lg:text-3xl font-bold bg-gradient-to-r from-blue-600 to-cyan-600 bg-clip-text text-transparent mb-2">
                欢迎回来
              </h1>
              <p className="text-slate-600">登录 StarDreamer 账户</p>
            </div>

            <form onSubmit={handleSubmit} className="space-y-5">
              <div className="flex gap-2">
                <button
                  type="button"
                  onClick={() => setLoginType('用户名')}
                  className={`flex-1 py-2.5 rounded-lg transition-all font-medium text-sm ${
                    loginType === '用户名'
                      ? 'bg-blue-100 text-blue-600'
                      : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
                  }`}
                >
                  用户名登录
                </button>
                <button
                  type="button"
                  onClick={() => setLoginType('邮箱')}
                  className={`flex-1 py-2.5 rounded-lg transition-all font-medium text-sm ${
                    loginType === '邮箱'
                      ? 'bg-blue-100 text-blue-600'
                      : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
                  }`}
                >
                  邮箱登录
                </button>
              </div>

              <div>
                <label className="block text-sm text-slate-700 mb-2 font-medium">
                  {loginType === '用户名' ? '用户名' : '邮箱'}
                </label>
                <div className="relative">
                  <div className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500">
                    {loginType === '用户名' ? <User size={18} /> : <Mail size={18} />}
                  </div>
                  <input
                    type={loginType === '用户名' ? 'text' : 'email'}
                    value={identity}
                    onChange={(e) => setIdentity(e.target.value)}
                    placeholder={`请输入${loginType === '用户名' ? '用户名' : '邮箱'}`}
                    className="input-field pl-10"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm text-slate-700 mb-2 font-medium">密码</label>
                <div className="relative">
                  <div className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500">
                    <Lock size={18} />
                  </div>
                  <input
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="请输入密码"
                    className="input-field pl-10"
                  />
                </div>
              </div>

              {error && (
                <div className="p-3 rounded-lg bg-red-50 border border-red-200 text-red-600 text-sm">
                  {error}
                </div>
              )}

              <button
                type="submit"
                disabled={loading}
                className="btn-primary w-full py-3 disabled:opacity-50"
              >
                {loading ? '登录中...' : '登录'}
              </button>
            </form>

            <div className="mt-6 text-center text-sm text-slate-600">
              还没有账户？{' '}
              <Link to="/register" className="text-blue-600 hover:text-blue-700 font-medium">
                立即注册
              </Link>
            </div>
          </div>
        </motion.div>
      </div>
    </Layout>
  );
}

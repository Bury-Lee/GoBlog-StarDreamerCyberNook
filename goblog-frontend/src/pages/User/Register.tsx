import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import Layout from '../../components/layout/Layout';
import { userApi } from '../../api';
import { Mail, Lock, User } from 'lucide-react';

export default function RegisterPage() {
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [code, setCode] = useState('');
  const [password, setPassword] = useState('');
  const [nickname, setNickname] = useState('');
  const [loading, setLoading] = useState(false);
  const [sending, setSending] = useState(false);
  const [error, setError] = useState('');
  const [countdown, setCountdown] = useState(0);

  const sendCode = async () => {
    if (!email) {
      setError('请输入邮箱');
      return;
    }

    try {
      setSending(true);
      const res = await userApi.sendEmail(email);
      if (res.code === 200) {
        setCountdown(60);
        const timer = setInterval(() => {
          setCountdown((prev) => {
            if (prev <= 1) {
              clearInterval(timer);
              return 0;
            }
            return prev - 1;
          });
        }, 1000);
      } else {
        setError(res.message || '发送失败');
      }
    } catch (err: any) {
      setError(err.response?.data?.message || '发送验证码失败');
    } finally {
      setSending(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!email || !code || !password) {
      setError('请填写所有必填字段');
      return;
    }

    if (password.length < 6) {
      setError('密码至少6位');
      return;
    }

    try {
      setLoading(true);
      const res = await userApi.register({
        email,
        code,
        password,
        nickname: nickname || undefined,
      });

      if (res.code === 200) {
        navigate('/login');
      } else {
        setError(res.message || '注册失败');
      }
    } catch (err: any) {
      setError(err.response?.data?.message || '注册失败');
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
                加入我们
              </h1>
              <p className="text-slate-600">创建 StarDreamer 账户</p>
            </div>

            <form onSubmit={handleSubmit} className="space-y-5">
              <div>
                <label className="block text-sm text-slate-700 mb-2 font-medium">邮箱</label>
                <div className="relative">
                  <div className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500">
                    <Mail size={18} />
                  </div>
                  <input
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="请输入邮箱"
                    className="input-field pl-10"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm text-slate-700 mb-2 font-medium">验证码</label>
                <div className="flex gap-2">
                  <div className="relative flex-1">
                    <div className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500">
                      <Lock size={18} />
                    </div>
                    <input
                      type="text"
                      value={code}
                      onChange={(e) => setCode(e.target.value)}
                      placeholder="验证码"
                      className="input-field pl-10"
                    />
                  </div>
                  <button
                    type="button"
                    onClick={sendCode}
                    disabled={sending || countdown > 0}
                    className="px-4 py-3 bg-slate-100 hover:bg-slate-200 rounded-lg text-sm transition-colors disabled:opacity-50 whitespace-nowrap font-medium"
                  >
                    {countdown > 0 ? `${countdown}s` : sending ? '发送中...' : '获取验证码'}
                  </button>
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
                    placeholder="至少6位密码"
                    className="input-field pl-10"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm text-slate-700 mb-2 font-medium">昵称（选填）</label>
                <div className="relative">
                  <div className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500">
                    <User size={18} />
                  </div>
                  <input
                    type="text"
                    value={nickname}
                    onChange={(e) => setNickname(e.target.value)}
                    placeholder="选填，默认随机昵称"
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
                {loading ? '注册中...' : '注册'}
              </button>
            </form>

            <div className="mt-6 text-center text-sm text-slate-600">
              已有账户？{' '}
              <Link to="/login" className="text-blue-600 hover:text-blue-700 font-medium">
                立即登录
              </Link>
            </div>
          </div>
        </motion.div>
      </div>
    </Layout>
  );
}

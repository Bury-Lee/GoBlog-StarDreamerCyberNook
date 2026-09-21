import { loadEnv } from 'vite'

const FALLBACK_BASE = 'http://127.0.0.1:8080/api'

function resolveBase(): string {
  const env = loadEnv('test', process.cwd(), '')
  const base = env.VITE_API_BASE || process.env.VITE_API_BASE || FALLBACK_BASE
  return /^https?:\/\//.test(base) ? base.replace(/\/$/, '') : FALLBACK_BASE
}

export default async function setup(): Promise<void> {
  const heartbeat = `${resolveBase()}/heartbeat`
  try {
    const response = await fetch(heartbeat, { signal: AbortSignal.timeout(5000) })
    const body = (await response.json()) as { code?: number; message?: string }
    if (body?.code !== 200) {
      throw new Error(`heartbeat 返回 code=${body?.code} message=${body?.message}`)
    }
    console.log(`\n[api-test] 后端已连通: ${heartbeat}\n`)
  } catch (error) {
    throw new Error(
      [
        `无法连接后端 ${heartbeat}`,
        '请在 dist 目录启动后端可执行文件后再运行测试:',
        '  cd dist && .\\main_windows_amd64.exe',
        `原始错误: ${String(error)}`,
      ].join('\n'),
    )
  }
}

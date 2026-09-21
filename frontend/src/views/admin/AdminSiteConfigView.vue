<template>
  <div class="admin-site">
    <el-alert
      type="info"
      :closable="false"
      show-icon
      title="站点配置当前为只读,如需修改请联系开发者"
    />

    <el-tabs v-model="activeTab">
      <el-tab-pane label="站点信息" name="site">
        <EmptyState v-if="siteError" text="站点配置加载失败,请稍后重试">
          <el-button size="small" @click="loadSite">重试</el-button>
        </EmptyState>
        <div v-else class="site-grid">
          <section class="sd-panel">
            <header class="sd-panel__header">
              <span class="sd-panel__title">基础信息</span>
            </header>
            <div class="sd-panel__body">
              <el-descriptions :column="1" border>
                <el-descriptions-item label="站点标题">{{ site?.siteInfo?.title || '-' }}</el-descriptions-item>
                <el-descriptions-item label="站点 Logo">{{ site?.siteInfo?.Logo || '-' }}</el-descriptions-item>
                <el-descriptions-item label="备案号">{{ site?.siteInfo?.Beian || '-' }}</el-descriptions-item>
                <el-descriptions-item label="运行模式">{{ site?.siteInfo?.Mode === 2 ? '社区模式' : '博客模式' }}</el-descriptions-item>
                <el-descriptions-item label="项目名称">{{ site?.project?.title || '-' }}</el-descriptions-item>
                <el-descriptions-item label="项目图标">{{ site?.project?.icon || '-' }}</el-descriptions-item>
                <el-descriptions-item label="访问路径">{{ site?.project?.webPath || '-' }}</el-descriptions-item>
              </el-descriptions>
            </div>
          </section>

          <section class="sd-panel">
            <header class="sd-panel__header">
              <span class="sd-panel__title">SEO / 关于</span>
            </header>
            <div class="sd-panel__body">
              <el-descriptions :column="1" border>
                <el-descriptions-item label="关键词">{{ site?.seo?.keywords || '-' }}</el-descriptions-item>
                <el-descriptions-item label="站点描述">{{ site?.seo?.description || '-' }}</el-descriptions-item>
                <el-descriptions-item label="后端版本">{{ site?.about?.Version || '-' }}</el-descriptions-item>
                <el-descriptions-item label="建站时间">{{ site?.about?.siteDate || '-' }}</el-descriptions-item>
                <el-descriptions-item label="QQ 群">{{ site?.about?.qq || '-' }}</el-descriptions-item>
                <el-descriptions-item label="微信">{{ site?.about?.wechat || '-' }}</el-descriptions-item>
                <el-descriptions-item label="哔哩哔哩">{{ site?.about?.biliBili || '-' }}</el-descriptions-item>
                <el-descriptions-item label="GitHub">{{ site?.about?.gitHub || '-' }}</el-descriptions-item>
              </el-descriptions>
            </div>
          </section>

          <section class="sd-panel">
            <header class="sd-panel__header">
              <span class="sd-panel__title">功能开关</span>
            </header>
            <div class="sd-panel__body">
              <el-descriptions :column="1" border>
                <el-descriptions-item label="文章审核">
                  <el-tag size="small" :type="site?.article?.enableExamination ? 'warning' : 'success'">
                    {{ site?.article?.enableExamination ? '开启' : '关闭' }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="用户名密码登录">
                  <el-tag size="small" :type="site?.login?.usernamePassword ? 'success' : 'info'">
                    {{ site?.login?.usernamePassword ? '开启' : '关闭' }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="邮箱登录">
                  <el-tag size="small" :type="site?.login?.emailLogin ? 'success' : 'info'">
                    {{ site?.login?.emailLogin ? '开启' : '关闭' }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="QQ 登录">
                  <el-tag size="small" :type="site?.login?.QQLogin ? 'success' : 'info'">
                    {{ site?.login?.QQLogin ? '开启' : '未启用' }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="图形验证码">
                  <el-tag size="small" :type="site?.login?.captcha ? 'warning' : 'info'">
                    {{ site?.login?.captcha ? '开启' : '关闭' }}
                  </el-tag>
                </el-descriptions-item>
              </el-descriptions>
            </div>
          </section>

          <section class="sd-panel">
            <header class="sd-panel__header">
              <span class="sd-panel__title">首页右侧组件</span>
            </header>
            <div class="sd-panel__body">
              <ul class="site-index-list">
                <li v-for="item in site?.indexRight?.list || []" :key="item.title">
                  <span>{{ item.title }}</span>
                  <el-tag size="small" :type="item.enable ? 'success' : 'info'">
                    {{ item.enable ? '启用' : '停用' }}
                  </el-tag>
                </li>
                <li v-if="!(site?.indexRight?.list || []).length" class="sd-dim">未配置</li>
              </ul>
            </div>
          </section>
        </div>
      </el-tab-pane>

      <el-tab-pane label="邮件配置" name="email">
        <section class="sd-panel">
          <header class="sd-panel__header">
            <span class="sd-panel__title">邮件服务(敏感字段已打码)</span>
          </header>
          <div class="sd-panel__body">
            <el-descriptions v-if="email" :column="2" border>
              <el-descriptions-item label="SMTP 域名">{{ email.domain }}</el-descriptions-item>
              <el-descriptions-item label="端口">{{ email.port }}</el-descriptions-item>
              <el-descriptions-item label="发件邮箱">{{ email.sendEmail }}</el-descriptions-item>
              <el-descriptions-item label="发件昵称">{{ email.sendNickname }}</el-descriptions-item>
              <el-descriptions-item label="授权码">{{ email.authCode }}</el-descriptions-item>
              <el-descriptions-item label="SSL / TLS">
                {{ boolText(email.SSL) }} / {{ boolText(email.TLS) }}
              </el-descriptions-item>
            </el-descriptions>
            <EmptyState v-else-if="emailError" text="邮件配置加载失败,请稍后重试">
              <el-button size="small" @click="loadEmail">重试</el-button>
            </EmptyState>
            <EmptyState v-else text="尚未配置邮件服务" />
          </div>
        </section>
      </el-tab-pane>

      <el-tab-pane label="QQ 登录" name="qq">
        <section class="sd-panel">
          <header class="sd-panel__header">
            <span class="sd-panel__title">QQ 互联配置</span>
          </header>
          <div class="sd-panel__body">
            <el-descriptions v-if="qq" :column="2" border>
              <el-descriptions-item label="AppID">{{ qq.appID }}</el-descriptions-item>
              <el-descriptions-item label="AppKey">{{ qq.appKey }}</el-descriptions-item>
              <el-descriptions-item label="回调地址">{{ qq.redirect }}</el-descriptions-item>
            </el-descriptions>
            <EmptyState v-else-if="qqError" text="QQ 互联配置加载失败,请稍后重试">
              <el-button size="small" @click="loadQQ">重试</el-button>
            </EmptyState>
            <EmptyState v-else text="尚未配置 QQ 互联" />
          </div>
        </section>
      </el-tab-pane>

      <el-tab-pane label="AI 配置" name="ai">
        <section class="sd-panel">
          <header class="sd-panel__header">
            <span class="sd-panel__title">AI 服务配置</span>
          </header>
          <div class="sd-panel__body">
            <el-descriptions v-if="ai" :column="2" border>
              <el-descriptions-item label="启用 AI">{{ ai.enable ? '是' : '否' }}</el-descriptions-item>
              <el-descriptions-item label="开放对话">{{ ai.chat_enable ? '是' : '否' }}</el-descriptions-item>
              <el-descriptions-item label="定时自动审核">{{ ai.auto_review ? '是' : '否' }}</el-descriptions-item>
              <el-descriptions-item label="模型">{{ ai.model }}</el-descriptions-item>
              <el-descriptions-item label="接口类型">{{ ai.api_type }}</el-descriptions-item>
              <el-descriptions-item label="服务地址">{{ ai.host }}</el-descriptions-item>
              <el-descriptions-item label="温度 / 最大 token">
                {{ ai.temperature }} / {{ ai.max_tokens }}
              </el-descriptions-item>
              <el-descriptions-item label="助手昵称">{{ ai.nickName }}</el-descriptions-item>
              <el-descriptions-item label="平台">{{ ai.platform || '-' }}</el-descriptions-item>
            </el-descriptions>
            <EmptyState v-else-if="aiError" text="AI 配置加载失败,请稍后重试">
              <el-button size="small" @click="loadAI">重试</el-button>
            </EmptyState>
            <EmptyState v-else text="尚未配置 AI 服务" />
          </div>
        </section>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { fetchAIConfig, fetchEmailConfig, fetchQQConfig, fetchSiteConfig } from '@/api/site'
import type { AIConfig, EmailConfig, QQConfig, SiteConfig } from '@/api/types'
import { useSiteStore } from '@/stores'

const siteStore = useSiteStore()
const activeTab = ref('site')
const site = ref<SiteConfig | null>(null)
const email = ref<EmailConfig | null>(null)
const qq = ref<QQConfig | null>(null)
const ai = ref<AIConfig | null>(null)
const siteError = ref(false)
const emailError = ref(false)
const qqError = ref(false)
const aiError = ref(false)

function boolText(value: boolean | undefined): string {
  return value ? '是' : '否'
}

async function loadSite(): Promise<void> {
  try {
    site.value = await fetchSiteConfig()
    siteError.value = false
  } catch {
    site.value = null
    siteError.value = true
  }
}

async function loadEmail(): Promise<void> {
  try {
    email.value = await fetchEmailConfig()
    emailError.value = false
  } catch {
    email.value = null
    emailError.value = true
  }
}

async function loadQQ(): Promise<void> {
  try {
    qq.value = await fetchQQConfig()
    qqError.value = false
  } catch {
    qq.value = null
    qqError.value = true
  }
}

async function loadAI(): Promise<void> {
  try {
    ai.value = await fetchAIConfig()
    aiError.value = false
  } catch {
    ai.value = null
    aiError.value = true
  }
}

async function load(): Promise<void> {
  await Promise.all([loadSite(), loadEmail(), loadQQ(), loadAI()])
}

onMounted(() => {
  void siteStore.loadSite()
  void load()
})
</script>

<style scoped lang="scss">
.admin-site {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.site-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(340px, 1fr));
  gap: 18px;
}

.site-index-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  font-size: 13px;

  li {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-bottom: 10px;
    border-bottom: 1px dashed rgba(30, 43, 69, 0.7);
  }

  li:last-child {
    border-bottom: none;
    padding-bottom: 0;
  }
}
</style>

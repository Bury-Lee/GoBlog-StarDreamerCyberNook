<template>
  <div class="auth-card sd-panel">
    <div class="auth-card__head">
      <span class="auth-card__logo">
        <img v-if="siteStore.logo" :src="siteStore.logo" alt="logo" />
        <el-icon v-else :size="22"><Platform /></el-icon>
      </span>
      <h2 class="auth-card__title sd-neon-text">{{ siteStore.title }}</h2>
      <p class="sd-dim">登录后继续你的赛博之旅</p>
    </div>

    <el-tabs v-model="loginType" class="auth-card__tabs" stretch>
      <el-tab-pane v-if="siteStore.loginOptions.usernamePassword" label="用户名登录" name="用户名" />
      <el-tab-pane v-if="siteStore.loginOptions.emailLogin" label="邮箱登录" name="邮箱" />
    </el-tabs>

    <el-form ref="formRef" :model="form" :rules="rules" label-position="top" @keyup.enter="submit">
      <el-form-item prop="val">
        <el-input
          v-model="form.val"
          :placeholder="loginType === '用户名' ? '请输入用户名' : '请输入邮箱地址'"
          size="large"
          clearable
        >
          <template #prefix>
            <el-icon><component :is="loginType === '用户名' ? User : Message" /></el-icon>
          </template>
        </el-input>
      </el-form-item>

      <el-form-item prop="pwd">
        <el-input
          v-model="form.pwd"
          type="password"
          placeholder="请输入密码"
          size="large"
          show-password
        >
          <template #prefix>
            <el-icon><Lock /></el-icon>
          </template>
        </el-input>
      </el-form-item>

      <el-form-item v-if="captchaEnabled" prop="captchaCode">
        <CaptchaField
          v-model="form.captchaCode"
          :image="captcha.state.image"
          :loading="captcha.loading"
          @refresh="refreshCaptcha"
        />
      </el-form-item>

      <el-button type="primary" size="large" class="auth-card__submit" :loading="submitting" @click="submit">
        登录
      </el-button>
    </el-form>

    <div class="auth-card__footer">
      <span class="sd-dim">还没有账号?</span>
      <router-link class="sd-link" :to="{ name: 'register' }">立即注册</router-link>
      <span class="auth-card__spacer" />
      <router-link class="sd-link" :to="{ name: 'home' }">返回首页</router-link>
    </div>

    <el-alert
      v-if="!siteStore.loginOptions.usernamePassword && !siteStore.loginOptions.emailLogin"
      class="auth-card__alert"
      type="warning"
      :closable="false"
      title="站点当前未开放密码登录,请联系管理员"
      show-icon
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Lock, Message, Platform, User } from '@element-plus/icons-vue'
import CaptchaField from '@/components/common/CaptchaField.vue'
import { useCaptcha } from '@/composables/useCaptcha'
import { useSiteStore, useUserStore } from '@/stores'

const route = useRoute()
const router = useRouter()
const siteStore = useSiteStore()
const userStore = useUserStore()

const formRef = ref<FormInstance>()
const submitting = ref(false)
const loginType = ref<'用户名' | '邮箱'>('用户名')
const captcha = useCaptcha('用户名')

const captchaEnabled = computed(() => siteStore.captchaEnabled)

const form = reactive({
  val: '',
  pwd: '',
  captchaCode: '',
})

//刷新验证码:清空已输入的验证码,保证与表单校验同一数据源
async function refreshCaptcha(): Promise<void> {
  form.captchaCode = ''
  await captcha.load()
}

const rules = computed<FormRules>(() => {
  const base: FormRules = {
    val: [{ required: true, message: loginType.value === '用户名' ? '请输入用户名' : '请输入邮箱', trigger: 'blur' }],
    pwd: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  }
  if (captchaEnabled.value) {
    base.captchaCode = [{ required: true, message: '请输入图形验证码', trigger: 'blur' }]
  }
  return base
})

watch(loginType, (value) => {
  captcha.reset()
  form.captchaCode = ''
  if (captchaEnabled.value) void refreshCaptcha()
  void value
})

onMounted(() => {
  if (captchaEnabled.value) void refreshCaptcha()
  if (!siteStore.loginOptions.usernamePassword && siteStore.loginOptions.emailLogin) {
    loginType.value = '邮箱'
  }
})

function redirectAfterLogin(): void {
  const redirect = route.query.redirect
  if (typeof redirect === 'string' && redirect.startsWith('/')) {
    router.replace(redirect)
  } else {
    router.replace({ name: 'home' })
  }
}

async function submit(): Promise<void> {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  submitting.value = true
  try {
    await userStore.login({
      type: loginType.value,
      val: form.val.trim(),
      pwd: form.pwd,
      captchaID: captchaEnabled.value ? captcha.state.captchaID : undefined,
      captchaCode: captchaEnabled.value ? form.captchaCode : undefined,
    })
    ElMessage.success('登录成功')
    redirectAfterLogin()
  } catch {
    if (captchaEnabled.value) {
      await refreshCaptcha()
    }
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped lang="scss">
.auth-card {
  padding: 28px 26px 22px;
}

.auth-card__head {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  margin-bottom: 18px;
  text-align: center;
}

.auth-card__logo {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 46px;
  height: 46px;
  overflow: hidden;
  border-radius: 14px;
  border: 1px solid rgba(34, 211, 238, 0.45);
  background: rgba(34, 211, 238, 0.1);
  color: var(--sd-cyan);
  box-shadow: 0 0 22px -8px rgba(34, 211, 238, 0.95);

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}

.auth-card__title {
  font-size: 20px;
  font-weight: 700;
}

.auth-card__tabs {
  margin-bottom: 8px;
}

.auth-card__submit {
  width: 100%;
  margin-top: 4px;
}

.auth-card__footer {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 16px;
  font-size: 13px;
}

.auth-card__spacer {
  flex: 1;
}

.auth-card__alert {
  margin-top: 14px;
}
</style>

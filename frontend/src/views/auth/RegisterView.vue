<template>
  <div class="auth-card sd-panel">
    <div class="auth-card__head">
      <span class="auth-card__logo">
        <img v-if="siteStore.logo" :src="siteStore.logo" alt="logo" />
        <el-icon v-else :size="22"><Platform /></el-icon>
      </span>
      <h2 class="auth-card__title sd-neon-text">注册新账号</h2>
      <p class="sd-dim">使用邮箱验证码快速创建账号</p>
    </div>

    <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
      <el-form-item prop="email">
        <el-input v-model="form.email" placeholder="邮箱地址" size="large" clearable>
          <template #prefix>
            <el-icon><Message /></el-icon>
          </template>
        </el-input>
      </el-form-item>

      <el-form-item v-if="captchaEnabled" prop="captchaCode">
        <CaptchaField
          v-model="captcha.state.captchaCode"
          :image="captcha.state.image"
          :loading="captcha.loading"
          @refresh="captcha.load"
        />
      </el-form-item>

      <el-form-item prop="emailCode">
        <div class="auth-card__code-row">
          <el-input v-model="form.emailCode" placeholder="邮箱验证码" size="large" maxlength="8" clearable>
            <template #prefix>
              <el-icon><Key /></el-icon>
            </template>
          </el-input>
          <el-button
            size="large"
            :disabled="countdown > 0 || sending"
            :loading="sending"
            @click="sendCode"
          >
            {{ countdown > 0 ? `${countdown}s` : '发送验证码' }}
          </el-button>
        </div>
      </el-form-item>

      <el-form-item prop="password">
        <el-input v-model="form.password" type="password" placeholder="设置登录密码" size="large" show-password>
          <template #prefix>
            <el-icon><Lock /></el-icon>
          </template>
        </el-input>
      </el-form-item>

      <el-form-item prop="confirmPassword">
        <el-input
          v-model="form.confirmPassword"
          type="password"
          placeholder="再次输入密码"
          size="large"
          show-password
        >
          <template #prefix>
            <el-icon><Lock /></el-icon>
          </template>
        </el-input>
      </el-form-item>

      <el-form-item prop="nickName">
        <el-input v-model="form.nickName" placeholder="昵称(可选,会经过内容审核)" size="large" clearable>
          <template #prefix>
            <el-icon><Avatar /></el-icon>
          </template>
        </el-input>
      </el-form-item>

      <el-button
        type="primary"
        size="large"
        class="auth-card__submit"
        :loading="submitting"
        @click="submit"
      >
        注册并登录
      </el-button>
    </el-form>

    <div class="auth-card__footer">
      <span class="sd-dim">已有账号?</span>
      <router-link class="sd-link" :to="{ name: 'login' }">返回登录</router-link>
      <span class="auth-card__spacer" />
      <router-link class="sd-link" :to="{ name: 'home' }">返回首页</router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Avatar, Key, Lock, Message, Platform } from '@element-plus/icons-vue'
import CaptchaField from '@/components/common/CaptchaField.vue'
import { useCaptcha } from '@/composables/useCaptcha'
import { sendEmailCode } from '@/api/user'
import { useSiteStore, useUserStore } from '@/stores'

const router = useRouter()
const siteStore = useSiteStore()
const userStore = useUserStore()

const formRef = ref<FormInstance>()
const captcha = useCaptcha('注册')
const submitting = ref(false)
const sending = ref(false)
const countdown = ref(0)
const emailID = ref('')
let timer: ReturnType<typeof setInterval> | null = null

const captchaEnabled = computed(() => siteStore.captchaEnabled)

const form = reactive({
  email: '',
  emailCode: '',
  password: '',
  confirmPassword: '',
  nickName: '',
})

const rules = computed<FormRules>(() => {
  const base: FormRules = {
    email: [
      { required: true, message: '请输入邮箱地址', trigger: 'blur' },
      { type: 'email', message: '邮箱格式不正确', trigger: 'blur' },
    ],
    emailCode: [{ required: true, message: '请输入邮箱验证码', trigger: 'blur' }],
    password: [
      { required: true, message: '请设置密码', trigger: 'blur' },
      { min: 6, message: '密码至少 6 位', trigger: 'blur' },
    ],
    confirmPassword: [
      { required: true, message: '请再次输入密码', trigger: 'blur' },
      {
        validator: (_rule, value: string, callback) => {
          if (value !== form.password) callback(new Error('两次输入的密码不一致'))
          else callback()
        },
        trigger: 'blur',
      },
    ],
  }
  if (captchaEnabled.value) {
    base.captchaCode = [{ required: true, message: '请输入图形验证码', trigger: 'blur' }]
  }
  return base
})

function startCountdown(seconds = 60): void {
  countdown.value = seconds
  if (timer) clearInterval(timer)
  timer = setInterval(() => {
    countdown.value -= 1
    if (countdown.value <= 0 && timer) {
      clearInterval(timer)
      timer = null
    }
  }, 1000)
}

async function sendCode(): Promise<void> {
  if (!form.email || !/^\S+@\S+\.\S+$/.test(form.email)) {
    ElMessage.warning('请先输入正确的邮箱地址')
    return
  }
  if (captchaEnabled.value && !captcha.state.captchaCode) {
    ElMessage.warning('请输入图形验证码')
    return
  }
  sending.value = true
  try {
    const result = await sendEmailCode({
      type: '注册',
      email: form.email.trim(),
      captchaID: captchaEnabled.value ? captcha.state.captchaID : undefined,
      captchaCode: captchaEnabled.value ? captcha.state.captchaCode : undefined,
    })
    emailID.value = result?.emailID || ''
    ElMessage.success('验证码已发送,请查收邮箱(有效期 2 分钟)')
    startCountdown()
  } catch {
    if (captchaEnabled.value) await captcha.load()
  } finally {
    sending.value = false
  }
}

async function submit(): Promise<void> {
  if (!formRef.value) return
  if (!emailID.value) {
    ElMessage.warning('请先获取邮箱验证码')
    return
  }
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  submitting.value = true
  try {
    await userStore.register({
      emailID: emailID.value,
      emailCode: form.emailCode.trim(),
      password: form.password,
      nickName: form.nickName.trim() || undefined,
    })
    ElMessage.success('注册成功,欢迎加入')
    router.replace({ name: 'home' })
  } catch {
    if (captchaEnabled.value) await captcha.load()
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  if (captchaEnabled.value) void captcha.load()
})

onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
})
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

.auth-card__code-row {
  display: flex;
  gap: 10px;
  width: 100%;
}

.auth-card__submit {
  width: 100%;
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
</style>

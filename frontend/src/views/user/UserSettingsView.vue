<template>
  <div class="sd-page">
    <div class="sd-container">
      <section class="sd-panel settings-panel">
        <el-tabs v-model="activeTab">
          <el-tab-pane label="个人资料" name="profile">
            <el-form :model="profileForm" label-width="100px" class="settings-form">
              <el-form-item label="头像">
                <ImageUploader v-model="profileForm.avatar" :width="120" :height="120" tip="建议使用正方形图片" />
              </el-form-item>
              <el-form-item label="昵称">
                <el-input v-model="profileForm.nickName" maxlength="24" show-word-limit placeholder="展示昵称" />
              </el-form-item>
              <el-form-item label="年龄">
                <el-input-number v-model="profileForm.Age" :min="0" :max="150" />
              </el-form-item>
              <el-form-item label="个人简介">
                <el-input
                  v-model="profileForm.abstract"
                  type="textarea"
                  :rows="3"
                  resize="none"
                  maxlength="200"
                  show-word-limit
                  placeholder="介绍一下自己吧"
                />
              </el-form-item>
              <el-form-item label="兴趣标签">
                <div class="settings-tags">
                  <el-tag
                    v-for="tag in profileForm.likeTags"
                    :key="tag"
                    closable
                    @close="removeLikeTag(tag)"
                  >
                    {{ tag }}
                  </el-tag>
                  <el-input
                    v-model="likeTagInput"
                    class="settings-tags__input"
                    size="small"
                    placeholder="回车添加"
                    maxlength="20"
                    @keyup.enter="addLikeTag"
                  />
                </div>
              </el-form-item>
              <el-form-item label="联系方式">
                <div class="settings-contacts">
                  <div
                    v-for="(contact, index) in contactRows"
                    :key="index"
                    class="settings-contacts__row"
                  >
                    <el-input v-model="contact.key" placeholder="平台,如 github" class="settings-contacts__key" />
                    <el-input v-model="contact.value" placeholder="账号或链接" />
                    <el-button text type="danger" @click="removeContact(index)">
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </div>
                  <el-button text type="primary" @click="addContact">
                    <el-icon><Plus /></el-icon>
                    添加联系方式
                  </el-button>
                </div>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :loading="saving" @click="saveProfile">保存资料</el-button>
              </el-form-item>
            </el-form>
          </el-tab-pane>

          <el-tab-pane label="隐私设置" name="privacy">
            <el-form label-width="140px" class="settings-form">
              <el-form-item label="公开收藏夹">
                <el-switch v-model="privacyForm.openCollect" />
                <span class="settings-hint sd-dim">关闭后他人无法查看你的收藏夹</span>
              </el-form-item>
              <el-form-item label="公开关注列表">
                <el-switch v-model="privacyForm.openFollow" />
                <span class="settings-hint sd-dim">关闭后他人无法查看你的关注</span>
              </el-form-item>
              <el-form-item label="公开粉丝列表">
                <el-switch v-model="privacyForm.openFans" />
                <span class="settings-hint sd-dim">关闭后他人无法查看你的粉丝</span>
              </el-form-item>
              <el-form-item label="主页样式 ID">
                <el-input-number v-model="privacyForm.homeStyleID" :min="0" :max="99" />
                <span class="settings-hint sd-dim">对应前端主题编号,预留字段</span>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :loading="savingPrivacy" @click="savePrivacy">保存设置</el-button>
              </el-form-item>
            </el-form>
          </el-tab-pane>

          <el-tab-pane label="消息通知" name="message">
            <el-alert
              v-if="messageError"
              class="settings-alert"
              type="warning"
              :closable="false"
              show-icon
              :title="messageError"
            />
            <el-form label-width="140px" class="settings-form">
              <el-form-item label="评论通知">
                <el-switch v-model="messageForm.openCommentMessage" />
              </el-form-item>
              <el-form-item label="回复通知">
                <el-switch v-model="messageForm.openReplyMessage" />
              </el-form-item>
              <el-form-item label="点赞通知">
                <el-switch v-model="messageForm.openDiggMessage" />
              </el-form-item>
              <el-form-item label="收藏通知">
                <el-switch v-model="messageForm.openCollectMessage" />
              </el-form-item>
              <el-form-item label="私信通知">
                <el-switch v-model="messageForm.openPrivateMessage" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :loading="savingMessage" @click="saveMessageConf">
                  保存通知设置
                </el-button>
              </el-form-item>
            </el-form>
          </el-tab-pane>

          <el-tab-pane label="邮箱管理" name="email">
            <el-alert
              class="settings-alert"
              type="info"
              :closable="false"
              show-icon
              title="重置邮箱需要同时验证原邮箱与新邮箱,验证码有效期 2 分钟"
            />
            <el-form label-width="120px" class="settings-form">
              <el-form-item label="新邮箱">
                <el-input v-model="emailForm.email" placeholder="新邮箱地址" />
              </el-form-item>
              <el-form-item v-if="captchaEnabled" label="图形验证码">
                <CaptchaField
                  v-model="captcha.state.captchaCode"
                  :image="captcha.state.image"
                  :loading="captcha.loading"
                  @refresh="captcha.load"
                />
              </el-form-item>
              <el-form-item>
                <el-button :loading="sendingEmail" :disabled="emailCountdown > 0" @click="sendResetCodes">
                  {{ emailCountdown > 0 ? `${emailCountdown}s 后可重发` : '发送双份验证码' }}
                </el-button>
                <span class="settings-hint sd-dim">会向原邮箱与新邮箱各发送一份验证码</span>
              </el-form-item>
              <el-form-item label="新邮箱验证码">
                <el-input v-model="emailForm.emailCode" maxlength="8" placeholder="新邮箱收到的验证码" />
              </el-form-item>
              <el-form-item label="原邮箱验证码">
                <el-input v-model="emailForm.ResetEmailCode" maxlength="8" placeholder="原邮箱收到的验证码" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :loading="resettingEmail" @click="submitResetEmail">
                  确认重置邮箱
                </el-button>
              </el-form-item>
            </el-form>
          </el-tab-pane>
        </el-tabs>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Delete, Plus } from '@element-plus/icons-vue'
import ImageUploader from '@/components/common/ImageUploader.vue'
import CaptchaField from '@/components/common/CaptchaField.vue'
import { useCaptcha } from '@/composables/useCaptcha'
import { fetchMessageConf, updateMessageConf } from '@/api/message'
import { resetEmail, sendEmailCode } from '@/api/user'
import { useSiteStore, useUserStore } from '@/stores'

const siteStore = useSiteStore()
const userStore = useUserStore()
const captcha = useCaptcha('重置邮箱')

const activeTab = ref('profile')
const saving = ref(false)
const savingPrivacy = ref(false)
const savingMessage = ref(false)
const sendingEmail = ref(false)
const resettingEmail = ref(false)
const likeTagInput = ref('')
const messageError = ref('')
const emailCountdown = ref(0)
let timer: ReturnType<typeof setInterval> | null = null

const profileForm = reactive({
  avatar: '',
  nickName: '',
  Age: 0,
  abstract: '',
  likeTags: [] as string[],
})

const contactRows = ref<Array<{ key: string; value: string }>>([])

const privacyForm = reactive({
  openCollect: true,
  openFollow: true,
  openFans: true,
  homeStyleID: 0,
})

const messageForm = reactive({
  openCommentMessage: true,
  openReplyMessage: true,
  openDiggMessage: true,
  openCollectMessage: true,
  openPrivateMessage: true,
})

const emailForm = reactive({
  email: '',
  emailCode: '',
  emailID: '',
  ResetEmailID: '',
  ResetEmailCode: '',
})

const captchaEnabled = computed(() => siteStore.captchaEnabled)

function syncFromProfile(): void {
  const profile = userStore.profile
  if (!profile) return
  profileForm.avatar = profile.avatar || ''
  profileForm.nickName = profile.nickname || ''
  profileForm.Age = profile.Age || 0
  profileForm.abstract = profile.abstract || ''
  profileForm.likeTags = [...(profile.likeTags || [])]
  contactRows.value = Object.entries(profile.contactInfo || {}).map(([key, value]) => ({ key, value }))
  privacyForm.openCollect = profile.openCollect
  privacyForm.openFollow = profile.openFollow
  privacyForm.openFans = profile.openFans
  privacyForm.homeStyleID = profile.homeStyleID || 0
}

function buildContactInfo(): Record<string, string> {
  const result: Record<string, string> = {}
  contactRows.value.forEach((row) => {
    const key = row.key.trim()
    const value = row.value.trim()
    if (key && value) result[key] = value
  })
  return result
}

function addLikeTag(): void {
  const value = likeTagInput.value.trim()
  if (!value) return
  if (profileForm.likeTags.includes(value)) {
    ElMessage.warning('标签已存在')
    return
  }
  if (profileForm.likeTags.length >= 36) {
    ElMessage.warning('最多 36 个标签')
    return
  }
  profileForm.likeTags.push(value)
  likeTagInput.value = ''
}

function removeLikeTag(tag: string): void {
  profileForm.likeTags = profileForm.likeTags.filter((item) => item !== tag)
}

function addContact(): void {
  contactRows.value.push({ key: '', value: '' })
}

function removeContact(index: number): void {
  contactRows.value.splice(index, 1)
}

async function saveProfile(): Promise<void> {
  saving.value = true
  try {
    await userStore.updateProfile({
      avatar: profileForm.avatar || null,
      nickName: profileForm.nickName || null,
      Age: profileForm.Age,
      abstract: profileForm.abstract || null,
      likeTags: profileForm.likeTags,
      contactInfo: buildContactInfo(),
    })
    ElMessage.success('资料已保存')
  } catch {
    // ignore
  } finally {
    saving.value = false
  }
}

async function savePrivacy(): Promise<void> {
  savingPrivacy.value = true
  try {
    await userStore.updateProfile({
      openCollect: privacyForm.openCollect,
      openFollow: privacyForm.openFollow,
      openFans: privacyForm.openFans,
      homeStyleID: privacyForm.homeStyleID,
    })
    ElMessage.success('隐私设置已保存')
  } catch {
    // ignore
  } finally {
    savingPrivacy.value = false
  }
}

async function loadMessageConf(): Promise<void> {
  messageError.value = ''
  try {
    const data = await fetchMessageConf()
    messageForm.openCommentMessage = data.openCommentMessage
    messageForm.openReplyMessage = data.openReplyMessage
    messageForm.openDiggMessage = data.openDiggMessage
    messageForm.openCollectMessage = data.openCollectMessage
    messageForm.openPrivateMessage = data.openPrivateMessage
  } catch {
    messageError.value = '消息配置读取失败,可能尚未初始化消息设置,保存后将自动创建'
  }
}

async function saveMessageConf(): Promise<void> {
  savingMessage.value = true
  try {
    await updateMessageConf({ ...messageForm })
    ElMessage.success('通知设置已保存')
    messageError.value = ''
  } catch {
    // ignore
  } finally {
    savingMessage.value = false
  }
}

function startEmailCountdown(seconds = 60): void {
  emailCountdown.value = seconds
  if (timer) clearInterval(timer)
  timer = setInterval(() => {
    emailCountdown.value -= 1
    if (emailCountdown.value <= 0 && timer) {
      clearInterval(timer)
      timer = null
    }
  }, 1000)
}

async function sendResetCodes(): Promise<void> {
  if (!emailForm.email || !/^\S+@\S+\.\S+$/.test(emailForm.email)) {
    ElMessage.warning('请输入正确的新邮箱地址')
    return
  }
  sendingEmail.value = true
  try {
    const result = await sendEmailCode({
      type: '重置邮箱',
      email: emailForm.email.trim(),
      captchaID: captchaEnabled.value ? captcha.state.captchaID : undefined,
      captchaCode: captchaEnabled.value ? captcha.state.captchaCode : undefined,
    })
    emailForm.emailID = result?.emailID || ''
    emailForm.ResetEmailID = result?.resetEmailID || ''
    ElMessage.success('验证码已发送到原邮箱与新邮箱')
    startEmailCountdown()
  } catch {
    if (captchaEnabled.value) await captcha.load()
  } finally {
    sendingEmail.value = false
  }
}

async function submitResetEmail(): Promise<void> {
  if (!emailForm.emailID || !emailForm.ResetEmailID) {
    ElMessage.warning('请先发送验证码')
    return
  }
  if (!emailForm.emailCode || !emailForm.ResetEmailCode) {
    ElMessage.warning('请填写两份验证码')
    return
  }
  resettingEmail.value = true
  try {
    await resetEmail({
      emailID: emailForm.emailID,
      emailCode: emailForm.emailCode,
      ResetEmailID: emailForm.ResetEmailID,
      ResetEmailCode: emailForm.ResetEmailCode,
    })
    ElMessage.success('邮箱重置成功')
    emailForm.email = ''
    emailForm.emailCode = ''
    emailForm.ResetEmailCode = ''
    emailForm.emailID = ''
    emailForm.ResetEmailID = ''
    await userStore.loadProfile(true)
  } catch {
    // ignore
  } finally {
    resettingEmail.value = false
  }
}

onMounted(async () => {
  if (!userStore.profile) await userStore.loadProfile()
  syncFromProfile()
  await loadMessageConf()
  if (captchaEnabled.value) void captcha.load()
})

onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped lang="scss">
.settings-panel {
  padding: 10px 22px 24px;
}

.settings-form {
  max-width: 620px;
  padding-top: 10px;
}

.settings-hint {
  margin-left: 12px;
  font-size: 12px;
}

.settings-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.settings-tags__input {
  width: 130px;
}

.settings-contacts {
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: 100%;
}

.settings-contacts__row {
  display: flex;
  gap: 10px;
}

.settings-contacts__key {
  width: 160px;
}

.settings-alert {
  margin: 10px 0 4px;
}
</style>

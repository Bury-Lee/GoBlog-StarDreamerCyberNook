<template>
  <div class="admin-users">
    <el-tabs v-model="activeTab">
      <el-tab-pane label="用户列表" name="users">
        <section class="sd-panel admin-users__panel">
          <header class="admin-users__head">
            <el-input
              v-model="key"
              class="admin-users__search"
              placeholder="搜索昵称 / 简介"
              clearable
              @keyup.enter="search"
              @clear="search"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
            <span class="sd-dim">共 {{ count }} 位用户</span>
          </header>

          <el-table v-loading="loading" :data="list" border stripe>
            <el-table-column prop="userID" label="ID" width="80" />
            <el-table-column label="头像" width="80">
              <template #default="{ row }">
                <UserAvatar :src="row.avatar" :name="row.nickname" :size="36" />
              </template>
            </el-table-column>
            <el-table-column prop="nickname" label="昵称" width="160" />
            <el-table-column prop="abstract" label="简介" min-width="220" show-overflow-tooltip />
            <el-table-column label="操作" width="230" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="openDetail(row)">详情</el-button>
                <el-button link type="warning" @click="openEdit(row)">编辑</el-button>
                <el-button link type="info" @click="viewUserArticles(row)">文章</el-button>
              </template>
            </el-table-column>
          </el-table>

          <PaginationBar
            :page="page"
            :limit="limit"
            :count="count"
            @update:page="changePage"
            @update:limit="changeLimit"
          />
        </section>
      </el-tab-pane>

      <el-tab-pane label="登录日志" name="logs">
        <section class="sd-panel admin-users__panel">
          <header class="admin-users__head">
            <div class="admin-users__filters">
              <el-input v-model="logQuery.ip" class="admin-users__search" placeholder="IP" clearable />
              <el-input v-model="logQuery.addr" class="admin-users__search" placeholder="地址" clearable />
              <el-input v-model="logQuery.userId" class="admin-users__search" placeholder="用户 ID" clearable />
            </div>
            <el-button type="primary" @click="loadLogs">查询</el-button>
          </header>

          <el-table v-loading="logLoading" :data="logs" border stripe>
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column label="用户" width="180">
              <template #default="{ row }">
                <div class="admin-users__user">
                  <UserAvatar :src="row.userAvatar" :name="row.userNickname" :size="26" />
                  <span>{{ row.userNickname || `#${row.userID}` }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="ip" label="IP" width="150" />
            <el-table-column prop="addr" label="地址" width="160" />
            <el-table-column prop="userAgent" label="User Agent" min-width="240" show-overflow-tooltip />
            <el-table-column label="登录时间" width="170">
              <template #default="{ row }">{{ formatDate(row.createdAt) }}</template>
            </el-table-column>
          </el-table>

          <PaginationBar
            :page="logPage"
            :limit="logLimit"
            :count="logCount"
            @update:page="changeLogPage"
            @update:limit="changeLogLimit"
          />
        </section>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="detailVisible" title="用户详情" width="520px">
      <el-descriptions v-if="detail" :column="1" border>
        <el-descriptions-item label="用户 ID">{{ detail.userID }}</el-descriptions-item>
        <el-descriptions-item label="昵称">{{ detail.nickName }}</el-descriptions-item>
        <el-descriptions-item label="年龄">{{ detail.age }}</el-descriptions-item>
        <el-descriptions-item label="地区">{{ detail.region || '-' }}</el-descriptions-item>
        <el-descriptions-item label="入驻天数">{{ detail.existDay }}</el-descriptions-item>
        <el-descriptions-item label="文章数">{{ detail.articleCount }}</el-descriptions-item>
        <el-descriptions-item label="粉丝 / 关注">{{ detail.fansCount }} / {{ detail.followCount }}</el-descriptions-item>
        <el-descriptions-item label="最近登录">{{ formatDate(detail.lastLoginTime) }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <el-dialog v-model="editVisible" title="编辑用户" width="560px">
      <el-form label-width="90px">
        <el-form-item label="用户 ID">
          <el-input :model-value="String(editForm.userID)" disabled />
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="editForm.username" placeholder="系统登录名" />
        </el-form-item>
        <el-form-item label="昵称">
          <el-input v-model="editForm.nickname" />
        </el-form-item>
        <el-form-item label="头像">
          <ImageUploader v-model="editForm.avatar" :width="110" :height="110" />
        </el-form-item>
        <el-form-item label="简介">
          <el-input v-model="editForm.abstract" type="textarea" :rows="3" resize="none" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="editForm.role" class="admin-users__control">
            <el-option
              v-for="item in ROLE_OPTIONS"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitEdit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import UserAvatar from '@/components/common/UserAvatar.vue'
import ImageUploader from '@/components/common/ImageUploader.vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import { adminUpdateUserInfo, fetchLoginLogs, fetchUserBaseInfo, fetchUserList } from '@/api/user'
import type { LoginLogItem, UserBaseInfo, UserListItem } from '@/api/types'
import { ROLE_OPTIONS, formatDate } from '@/utils/format'
import { usePagination } from '@/composables/usePagination'

const router = useRouter()

const activeTab = ref('users')
const detailVisible = ref(false)
const editVisible = ref(false)
const saving = ref(false)
const detail = ref<UserBaseInfo | null>(null)

const editForm = reactive({
  userID: 0,
  username: '',
  nickname: '',
  avatar: '',
  abstract: '',
  role: 4,
})

const {
  list,
  count,
  page,
  limit,
  key,
  loading,
  load,
  search,
  changePage,
  changeLimit,
} = usePagination<UserListItem>(
  (params) => fetchUserList(params),
  () => ({}),
  { limit: 10 },
)

const logs = ref<LoginLogItem[]>([])
const logLoading = ref(false)
const logPage = ref(1)
const logLimit = ref(10)
const logCount = ref(0)
const logQuery = reactive({ ip: '', addr: '', userId: '' })

async function loadLogs(): Promise<void> {
  logLoading.value = true
  try {
    const data = await fetchLoginLogs({
      page: logPage.value,
      limit: logLimit.value,
      ip: logQuery.ip || undefined,
      addr: logQuery.addr || undefined,
      userId: logQuery.userId ? Number(logQuery.userId) : undefined,
    })
    logs.value = data?.list ?? []
    logCount.value = data?.count ?? 0
  } catch {
    logs.value = []
    logCount.value = 0
  } finally {
    logLoading.value = false
  }
}

function changeLogPage(next: number): void {
  logPage.value = next
  void loadLogs()
}

function changeLogLimit(next: number): void {
  logLimit.value = next
  logPage.value = 1
  void loadLogs()
}

async function openDetail(row: UserListItem): Promise<void> {
  detailVisible.value = true
  detail.value = null
  try {
    detail.value = await fetchUserBaseInfo(row.userID)
  } catch {
    detail.value = null
  }
}

function openEdit(row: UserListItem): void {
  editForm.userID = row.userID
  editForm.username = ''
  editForm.nickname = row.nickname
  editForm.avatar = row.avatar
  editForm.abstract = row.abstract
  editForm.role = 4
  editVisible.value = true
}

async function submitEdit(): Promise<void> {
  saving.value = true
  try {
    await adminUpdateUserInfo({
      userID: editForm.userID,
      username: editForm.username || null,
      nickname: editForm.nickname || null,
      avatar: editForm.avatar || null,
      abstract: editForm.abstract || null,
      role: editForm.role,
    })
    ElMessage.success('用户信息已更新')
    editVisible.value = false
    await load()
  } catch {
    // ignore
  } finally {
    saving.value = false
  }
}

function viewUserArticles(row: UserListItem): void {
  router.push({ name: 'admin-articles', query: { userID: String(row.userID) } })
}

watch(activeTab, (tab) => {
  if (tab === 'logs' && !logs.value.length) void loadLogs()
})

onMounted(() => {
  void load()
})
</script>

<style scoped lang="scss">
.admin-users__panel {
  padding: 18px;
}

.admin-users__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}

.admin-users__filters {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.admin-users__search {
  width: 200px;
}

.admin-users__user {
  display: flex;
  align-items: center;
  gap: 8px;
}

.admin-users__control {
  width: 100%;
}
</style>

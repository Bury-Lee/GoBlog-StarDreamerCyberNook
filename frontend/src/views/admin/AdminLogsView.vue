<template>
  <div class="admin-logs">
    <section class="sd-panel admin-logs__filters">
      <el-input
        v-model="query.key"
        class="admin-logs__input"
        placeholder="搜索标题"
        clearable
        @keyup.enter="search"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <el-select v-model="query.logType" class="admin-logs__select" placeholder="日志类型" clearable>
        <el-option v-for="item in LOG_TYPE_OPTIONS" :key="item.value" :label="item.label" :value="item.value" />
      </el-select>
      <el-select v-model="query.level" class="admin-logs__select" placeholder="级别" clearable>
        <el-option v-for="item in LOG_LEVEL_OPTIONS" :key="item.value" :label="item.label" :value="item.value" />
      </el-select>
      <el-select v-model="query.loginStatus" class="admin-logs__select" placeholder="登录状态" clearable>
        <el-option label="登录成功" :value="true" />
        <el-option label="登录失败" :value="false" />
      </el-select>
      <el-input v-model="query.ip" class="admin-logs__input" placeholder="IP" clearable />
      <el-input v-model="query.serviceName" class="admin-logs__input" placeholder="服务名" clearable />
      <el-input v-model="query.userID" class="admin-logs__input" placeholder="用户 ID" clearable />
      <el-button type="primary" @click="search">查询</el-button>
      <el-button @click="resetQuery">重置</el-button>
      <el-button type="danger" plain :disabled="!selection.length" @click="removeSelected">
        删除选中({{ selection.length }})
      </el-button>
    </section>

    <section class="sd-panel admin-logs__table">
      <el-table
        v-loading="loading"
        :data="list"
        border
        stripe
        @selection-change="onSelectionChange"
        @row-click="openDetail"
      >
        <el-table-column type="selection" width="46" />
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column label="类型" width="110">
          <template #default="{ row }">
            <el-tag size="small" effect="plain">{{ logTypeLabel(row.logType) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="级别" width="90">
          <template #default="{ row }">
            <el-tag size="small" :type="logLevelType(row.level)">{{ logLevelLabel(row.level) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip />
        <el-table-column prop="serviceName" label="服务" width="130" show-overflow-tooltip />
        <el-table-column prop="ip" label="IP" width="140" />
        <el-table-column label="登录状态" width="110">
          <template #default="{ row }">
            <el-tag size="small" :type="row.loginStatus ? 'success' : 'danger'">
              {{ row.loginStatus ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="已读" width="80">
          <template #default="{ row }">
            <el-tag size="small" :type="row.isRead ? 'info' : 'warning'">{{ row.isRead ? '已读' : '未读' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="时间" width="170">
          <template #default="{ row }">{{ formatDate(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button link type="danger" @click.stop="removeOne(row)">删除</el-button>
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

    <el-dialog v-model="detailVisible" title="日志详情" width="640px">
      <div v-if="current" class="admin-logs__detail">
        <div class="admin-logs__detail-meta">
          <el-tag size="small">{{ logTypeLabel(current.logType) }}</el-tag>
          <el-tag size="small" :type="logLevelType(current.level)">{{ logLevelLabel(current.level) }}</el-tag>
          <span class="sd-dim">{{ formatDate(current.createdAt) }}</span>
          <span class="sd-dim">{{ current.ip }} · {{ current.addr }}</span>
        </div>
        <h4 class="admin-logs__detail-title">{{ current.title }}</h4>
        <pre class="admin-logs__content">{{ current.content }}</pre>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import { fetchLogDetail, fetchLogs, removeLogs } from '@/api/ops'
import type { LogModel } from '@/api/types'
import {
  LOG_LEVEL_OPTIONS,
  LOG_TYPE_OPTIONS,
  formatDate,
  logLevelLabel,
  logLevelType,
  logTypeLabel,
} from '@/utils/format'
import { usePagination } from '@/composables/usePagination'

const query = reactive({
  key: '',
  logType: undefined as number | undefined,
  level: undefined as number | undefined,
  loginStatus: undefined as boolean | undefined,
  ip: '',
  serviceName: '',
  userID: '',
})

const selection = ref<LogModel[]>([])
const detailVisible = ref(false)
const current = ref<LogModel | null>(null)

const { list, count, page, limit, loading, load, search, changePage, changeLimit } = usePagination(
  (params) => fetchLogs(params),
  () => ({
    logType: query.logType,
    level: query.level,
    loginStatus: query.loginStatus,
    ip: query.ip || undefined,
    serviceName: query.serviceName || undefined,
    userID: query.userID ? Number(query.userID) : undefined,
  }),
  { limit: 10 },
)

function onSelectionChange(rows: LogModel[]): void {
  selection.value = rows
}

function resetQuery(): void {
  query.logType = undefined
  query.level = undefined
  query.loginStatus = undefined
  query.ip = ''
  query.serviceName = ''
  query.userID = ''
  search()
}

async function openDetail(row: LogModel): Promise<void> {
  current.value = row
  detailVisible.value = true
  try {
    await fetchLogDetail(row.id)
  } catch {
    // ignore
  }
}

async function removeOne(row: LogModel): Promise<void> {
  try {
    await ElMessageBox.confirm('确认删除该条日志?', '删除日志', { type: 'warning' })
  } catch {
    return
  }
  try {
    await removeLogs([row.id])
    ElMessage.success('删除成功')
    await load()
  } catch {
    // ignore
  }
}

async function removeSelected(): Promise<void> {
  if (!selection.value.length) return
  try {
    await ElMessageBox.confirm(`确认删除选中的 ${selection.value.length} 条日志?`, '批量删除', {
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    await removeLogs(selection.value.map((item) => item.id))
    ElMessage.success('删除成功')
    selection.value = []
    await load()
  } catch {
    // ignore
  }
}
</script>

<style scoped lang="scss">
.admin-logs {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.admin-logs__filters {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  padding: 16px;
}

.admin-logs__input {
  width: 160px;
}

.admin-logs__select {
  width: 140px;
}

.admin-logs__table {
  padding: 18px;
}

.admin-logs__detail-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  font-size: 12px;
}

.admin-logs__detail-title {
  margin: 12px 0 8px;
  font-size: 15px;
}

.admin-logs__content {
  max-height: 380px;
  overflow: auto;
  margin: 0;
  padding: 14px;
  border-radius: 10px;
  border: 1px solid var(--sd-border);
  background: rgba(7, 11, 20, 0.85);
  font-family: var(--sd-font-mono);
  font-size: 12px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>

<template>
  <div class="ticket-list">
    <el-card>
      <template #header>
        <div class="header">
          <div class="tabs">
            <el-radio-group v-model="activeTab" @change="loadTickets">
              <el-radio-button value="pending">待审批</el-radio-button>
              <el-radio-button value="approved">已批准</el-radio-button>
              <el-radio-button value="rejected">已拒绝</el-radio-button>
              <el-radio-button value="">全部</el-radio-button>
            </el-radio-group>
          </div>
          <el-button type="primary" @click="$router.push('/tickets/create')">
            创建工单
          </el-button>
        </div>
      </template>

      <el-table :data="tickets" style="width: 100%" v-loading="loading">
        <el-table-column prop="ticket_id" label="工单编号" width="150" />
        <el-table-column prop="title" label="标题" min-width="200" />
        <el-table-column prop="scene" label="场景" width="120" />
        <el-table-column prop="applicant" label="申请人" width="100" />
        <el-table-column prop="approver" label="审批人" width="100" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="$router.push(`/tickets/${row.ticket_id}`)">
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="page"
          :page-size="20"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="loadTickets"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { ticketApi } from '../api/ticket'

const tickets = ref([])
const activeTab = ref('pending')
const page = ref(1)
const total = ref(0)
const loading = ref(false)

const statusType = (status: string) => {
  const map: Record<string, string> = {
    pending: 'warning',
    approved: 'success',
    rejected: 'danger'
  }
  return map[status] || 'info'
}

const statusText = (status: string) => {
  const map: Record<string, string> = {
    pending: '待审批',
    approved: '已批准',
    rejected: '已拒绝'
  }
  return map[status] || status
}

const loadTickets = async () => {
  loading.value = true
  try {
    const { data } = await ticketApi.list({
      status: activeTab.value,
      page: page.value
    })
    tickets.value = data.tickets
    total.value = data.total
  } catch (error) {
    ElMessage.error('加载工单失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadTickets)
</script>

<style scoped>
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>

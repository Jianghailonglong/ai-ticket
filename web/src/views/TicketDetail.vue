<template>
  <div class="ticket-detail">
    <el-card v-loading="loading">
      <template #header>
        <div class="header">
          <span>工单详情 - {{ ticket?.ticket_id }}</span>
          <el-button @click="$router.push('/tickets')">返回列表</el-button>
        </div>
      </template>

      <template v-if="ticket">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="工单编号">{{ ticket.ticket_id }}</el-descriptions-item>
          <el-descriptions-item label="标题">{{ ticket.title }}</el-descriptions-item>
          <el-descriptions-item label="场景">{{ ticket.scene }}</el-descriptions-item>
          <el-descriptions-item label="申请人">{{ ticket.applicant }}</el-descriptions-item>
          <el-descriptions-item label="审批人">{{ ticket.approver }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusType(ticket.status)">{{ statusText(ticket.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ ticket.created_at }}</el-descriptions-item>
          <el-descriptions-item label="申请原因" :span="2">{{ ticket.reason || '-' }}</el-descriptions-item>
          <el-descriptions-item v-if="ticket.comment" label="审批意见" :span="2">
            {{ ticket.comment }}
          </el-descriptions-item>
        </el-descriptions>

        <div v-if="ticket.status === 'pending' && isApprover" class="actions">
          <h3>审批操作</h3>
          <el-input
            v-model="comment"
            type="textarea"
            :rows="3"
            placeholder="请输入审批意见（拒绝时必填）"
          />
          <div class="buttons">
            <el-button type="success" @click="handleApprove" :loading="submitting">
              同意
            </el-button>
            <el-button type="danger" @click="handleReject" :loading="submitting">
              拒绝
            </el-button>
          </div>
        </div>
      </template>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '../stores/user'
import { ticketApi } from '../api/ticket'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const ticket = ref<any>(null)
const comment = ref('')
const loading = ref(false)
const submitting = ref(false)

const isApprover = computed(() => {
  return ticket.value?.approver === userStore.username
})

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

const loadTicket = async () => {
  loading.value = true
  try {
    const { data } = await ticketApi.get(route.params.id as string)
    ticket.value = data
  } catch (error) {
    ElMessage.error('加载工单失败')
    router.push('/tickets')
  } finally {
    loading.value = false
  }
}

const handleApprove = async () => {
  submitting.value = true
  try {
    await ticketApi.approve(ticket.value.ticket_id, comment.value)
    ElMessage.success('工单已批准')
    router.push('/tickets')
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '操作失败')
  } finally {
    submitting.value = false
  }
}

const handleReject = async () => {
  if (!comment.value) {
    ElMessage.warning('拒绝时必须填写审批意见')
    return
  }

  submitting.value = true
  try {
    await ticketApi.reject(ticket.value.ticket_id, comment.value)
    ElMessage.success('工单已拒绝')
    router.push('/tickets')
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '操作失败')
  } finally {
    submitting.value = false
  }
}

onMounted(loadTicket)
</script>

<style scoped>
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.actions {
  margin-top: 30px;
  padding-top: 20px;
  border-top: 1px solid #eee;
}

.actions h3 {
  margin-bottom: 15px;
}

.buttons {
  margin-top: 15px;
  display: flex;
  gap: 10px;
}
</style>

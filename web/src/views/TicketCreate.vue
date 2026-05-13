<template>
  <div class="ticket-create">
    <el-card>
      <template #header>
        <div class="header">
          <span>创建工单</span>
          <el-button @click="$router.push('/tickets')">返回列表</el-button>
        </div>
      </template>

      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="标题" prop="title">
          <el-input v-model="form.title" placeholder="请输入工单标题" />
        </el-form-item>
        <el-form-item label="业务场景" prop="scene">
          <el-select v-model="form.scene" placeholder="请选择业务场景">
            <el-option label="创建集群" value="创建集群" />
            <el-option label="创建资源" value="创建资源" />
          </el-select>
        </el-form-item>
        <el-form-item label="审批人" prop="approver">
          <el-input v-model="form.approver" placeholder="请输入审批人用户名" />
        </el-form-item>
        <el-form-item label="申请原因" prop="reason">
          <el-input
            v-model="form.reason"
            type="textarea"
            :rows="4"
            placeholder="请输入申请原因"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSubmit" :loading="loading">
            提交
          </el-button>
          <el-button @click="resetForm">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance } from 'element-plus'
import { ticketApi } from '../api/ticket'

const router = useRouter()
const formRef = ref<FormInstance>()
const loading = ref(false)

const form = ref({
  title: '',
  scene: '',
  approver: '',
  reason: ''
})

const rules = {
  title: [{ required: true, message: '请输入工单标题', trigger: 'blur' }],
  scene: [{ required: true, message: '请输入业务场景', trigger: 'blur' }],
  approver: [{ required: true, message: '请输入审批人', trigger: 'blur' }]
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    loading.value = true
    try {
      await ticketApi.create(form.value)
      ElMessage.success('工单创建成功')
      router.push('/tickets')
    } catch (error: any) {
      ElMessage.error(error.response?.data?.error || '创建失败')
    } finally {
      loading.value = false
    }
  })
}

const resetForm = () => {
  if (formRef.value) {
    formRef.value.resetFields()
  }
}
</script>

<style scoped>
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.el-form {
  max-width: 600px;
}
</style>

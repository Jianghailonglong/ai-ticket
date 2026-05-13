<template>
  <div class="login-container">
    <el-card class="login-card">
      <template #header>
        <h2>{{ isRegister ? '注册' : '登录' }}</h2>
      </template>

      <el-form :model="form" label-width="80px">
        <el-form-item label="用户名">
          <el-input v-model="form.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" placeholder="请输入密码" />
        </el-form-item>
        <el-form-item v-if="isRegister" label="显示名">
          <el-input v-model="form.display_name" placeholder="请输入显示名称" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSubmit" :loading="loading">
            {{ isRegister ? '注册' : '登录' }}
          </el-button>
          <el-button link @click="isRegister = !isRegister">
            {{ isRegister ? '已有账号？去登录' : '没有账号？去注册' }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '../stores/user'
import { authApi } from '../api/auth'

const router = useRouter()
const userStore = useUserStore()

const isRegister = ref(false)
const loading = ref(false)
const form = ref({
  username: '',
  password: '',
  display_name: ''
})

const handleSubmit = async () => {
  if (!form.value.username || !form.value.password) {
    ElMessage.warning('请填写用户名和密码')
    return
  }

  loading.value = true

  try {
    if (isRegister.value) {
      await authApi.register(form.value)
      ElMessage.success('注册成功，请登录')
      isRegister.value = false
    } else {
      const { data } = await authApi.login(form.value)
      userStore.setUser(data)
      ElMessage.success('登录成功')
      router.push('/tickets')
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '操作失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: calc(100vh - 200px);
}

.login-card {
  width: 400px;
}

.login-card h2 {
  text-align: center;
  margin: 0;
}
</style>

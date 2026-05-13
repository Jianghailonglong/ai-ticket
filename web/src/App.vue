<template>
  <el-container class="app-container">
    <el-header>
      <div class="header">
        <h1>云平台工单系统</h1>
        <div v-if="userStore.isLoggedIn" class="user-info">
          <span>{{ userStore.displayName || userStore.username }}</span>
          <el-button link @click="handleLogout">退出</el-button>
        </div>
      </div>
    </el-header>
    <el-main>
      <router-view />
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useUserStore } from './stores/user'

const router = useRouter()
const userStore = useUserStore()

const handleLogout = () => {
  userStore.logout()
  router.push('/login')
}
</script>

<style>
.app-container {
  min-height: 100vh;
}

.el-header {
  background: #409eff;
  color: white;
  display: flex;
  align-items: center;
}

.header {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header h1 {
  margin: 0;
  font-size: 20px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-info span {
  color: white;
}

.el-main {
  padding: 20px;
  background: #f5f5f5;
  min-height: calc(100vh - 60px);
}
</style>

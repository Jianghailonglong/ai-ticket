import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '../stores/user'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/Login.vue')
    },
    {
      path: '/',
      redirect: '/tickets'
    },
    {
      path: '/tickets',
      name: 'TicketList',
      component: () => import('../views/TicketList.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/tickets/create',
      name: 'TicketCreate',
      component: () => import('../views/TicketCreate.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/tickets/:id',
      name: 'TicketDetail',
      component: () => import('../views/TicketDetail.vue'),
      meta: { requiresAuth: true }
    }
  ]
})

router.beforeEach((to, from, next) => {
  const userStore = useUserStore()

  if (to.meta.requiresAuth && !userStore.isLoggedIn) {
    next('/login')
  } else {
    next()
  }
})

export default router

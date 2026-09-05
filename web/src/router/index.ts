import { createRouter, createWebHistory } from 'vue-router'

import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      component: () => import('@/layouts/AuthLayout.vue'),
      meta: { guestOnly: true },
      children: [
        {
          path: '',
          name: 'login',
          component: () => import('@/views/LoginView.vue'),
          meta: { title: 'Sign in', guestOnly: true },
        },
      ],
    },
    {
      path: '/register',
      component: () => import('@/layouts/AuthLayout.vue'),
      meta: { guestOnly: true },
      children: [
        {
          path: '',
          name: 'register',
          component: () => import('@/views/RegisterView.vue'),
          meta: { title: 'Create account', guestOnly: true },
        },
      ],
    },
    {
      path: '/',
      component: () => import('@/layouts/AppLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'radar',
          component: () => import('@/views/RadarView.vue'),
          meta: { title: 'Radar', requiresAuth: true },
        },
      ],
    },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (!auth.ready) {
    await auth.hydrate()
  }

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  if (to.meta.guestOnly && auth.isAuthenticated) {
    return { name: 'radar' }
  }

  return true
})

router.afterEach((to) => {
  const title = typeof to.meta.title === 'string' ? to.meta.title : 'FRIDAY'
  document.title = `${title} · FRIDAY`
})

export default router

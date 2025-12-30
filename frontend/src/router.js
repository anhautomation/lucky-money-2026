import { createRouter, createWebHistory } from 'vue-router'
import UserLanding from './pages/UserLanding.vue'
import AdminPanel from './pages/AdminPanel.vue'

const routes = [
  {
    path: '/',
    component: UserLanding,
  },
  {
    path: '/panel',
    component: AdminPanel,
  },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

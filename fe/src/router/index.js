// src/router/index.js
import { createRouter, createWebHistory } from 'vue-router'
import Layout from '@/views/Layout.vue'
import Overview from '@/views/Overview.vue'
import Tasks from '@/views/Tasks.vue'
import Config from '@/views/Config.vue'
import Debug from '@/views/Debug.vue'
import Logs from '@/views/Logs.vue'

const routes = [
  {
    path: '/',
    component: Layout,
    children: [
      {
        path: '',
        redirect: '/overview'
      },
      {
        path: 'overview',
        name: 'Overview',
        component: Overview
      },
      {
        path: 'tasks',
        name: 'Tasks',
        component: Tasks
      },
      {
        path: 'config',
        name: 'Config',
        component: Config
      },
      {
        path: 'debug',
        name: 'Debug',
        component: Debug
      },
      {
        path: 'logs',
        name: 'Logs',
        component: Logs
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes
})

export default router
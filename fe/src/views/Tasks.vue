<!-- src/views/Tasks.vue -->
<template>
  <div class="tasks-page">
    <div class="page-header">
      <h1>任务管理</h1>
      <button class="btn refresh-btn" @click="refreshTasks" :disabled="loadingStates.tasks">
        <span v-if="loadingStates.tasks" class="loading-spinner"></span>
        刷新
      </button>
    </div>
    
    <TasksList 
      :tasks="tasks" 
      :loading-states="loadingStates"
      @refresh-tasks="refreshTasks" />
  </div>
</template>

<script>
import TasksList from '@/components/tasks.vue'
import apiService from '@/services/api.js'

export default {
  name: 'TasksView',
  components: {
    TasksList
  },
  data() {
    return {
      tasks: [],
      loadingStates: {
        tasks: false
      }
    }
  },
  async mounted() {
    await this.refreshTasks()
  },
  methods: {
    async refreshTasks() {
      this.loadingStates.tasks = true
      try {
        // 使用统一的API服务获取数据
        const response = await apiService.getTasks()
        console.log('获取到的任务数据:', response.data.entity)
        this.tasks = response.data.entity.Tasks || []
        console.log('获取到的任务数据:', this.tasks)
      } catch (error) {
        alert('获取任务数据失败: ' + (error.message || '未知错误'))
      } finally {
        this.loadingStates.tasks = false
      }
    }
  }
}
</script>

<style scoped>
.tasks-page {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
  padding-bottom: 15px;
  border-bottom: 1px solid #eee;
}

.page-header h1 {
  margin: 0;
  color: #333;
  font-size: 24px;
}

.refresh-btn {
  background-color: #1890ff;
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  display: inline-flex;
  align-items: center;
  transition: background-color 0.3s;
}

.refresh-btn:hover:not(:disabled) {
  background-color: #40a9ff;
}

.refresh-btn:disabled {
  background-color: #d9d9d9;
  cursor: not-allowed;
}

.loading-spinner {
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  border-top-color: white;
  animation: spin 1s ease-in-out infinite;
  margin-right: 8px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
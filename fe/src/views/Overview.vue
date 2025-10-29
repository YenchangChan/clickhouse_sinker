<!-- src/views/Overview.vue -->
<template>
  <div class="overview-page">
    <div class="page-header">
      <h1>系统概览</h1>
      <button class="btn refresh-btn" @click="refreshOverview" :disabled="loadingStates.overview">
        <span v-if="loadingStates.overview" class="loading-spinner"></span>
        刷新
      </button>
    </div>
    
    <OverviewPage 
      :proc-info="procInfo" 
      :loading-states="loadingStates" />
  </div>
</template>

<script>
import OverviewPage from '@/components/overview.vue'
import apiService from '@/services/api.js'

export default {
  name: 'OverviewView',
  components: {
    OverviewPage
  },
  data() {
    return {
      procInfo: {},
      loadingStates: {
        overview: false
      }
    }
  },
  async mounted() {
    await this.refreshOverview()
  },
  methods: {
    async refreshOverview() {
      this.loadingStates.overview = true
      try {
        // 使用统一的API服务获取数据
        const response = await apiService.getOverview()
        
        this.procInfo = response.data.entity || {}
      } catch (error) {
        alert('获取系统概览数据失败: ' + (error.message || '未知错误'))
      } finally {
        this.loadingStates.overview = false
      }
    }
  }
}
</script>

<style scoped>
.overview-page {
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
<!-- src/views/Config.vue -->
<template>
  <div class="config-page">
    <div class="page-header">
      <h1>配置管理</h1>
      <button class="btn refresh-btn" @click="refreshConfig" :disabled="loadingStates.config || loadingStates.cmdline">
        <span v-if="loadingStates.config || loadingStates.cmdline" class="loading-spinner"></span>
        刷新
      </button>
    </div>
    
    <AppConfig 
      :config="config" 
      :cmdline-config="cmdlineConfig" 
      :loading-states="loadingStates"
      @refresh-config="refreshConfig" />
  </div>
</template>

<script>
import AppConfig from '@/components/config.vue'
import apiService from '@/services/api.js'

export default {
  name: 'ConfigView',
  components: {
    AppConfig
  },
  data() {
    return {
      config: {},
      cmdlineConfig: {},
      loadingStates: {
        config: false,
        cmdline: false
      }
    }
  },
  async mounted() {
    await this.refreshConfig()
  },
  methods: {
    async refreshConfig() {
      this.loadingStates.config = true
      this.loadingStates.cmdline = true
      try {
        // 使用统一的API服务获取数据
        const [configResponse, cmdlineResponse] = await Promise.all([
          apiService.getConfig(),
          apiService.getCmdline()
        ])
        
        // 确保数据正确填充，从entity字段提取数据
        this.config = configResponse.data?.entity || configResponse.data || {}
        this.cmdlineConfig = cmdlineResponse.data?.entity || cmdlineResponse.data || {}
        
        console.log('获取到的配置数据:', this.config)
        console.log('获取到的命令行参数:', this.cmdlineConfig)
      } catch (error) {
        console.error('获取配置信息失败:', error)
        // 显示错误信息
        alert('获取配置数据失败: ' + (error.message || '未知错误'))
      } finally {
        this.loadingStates.config = false
        this.loadingStates.cmdline = false
      }
    }
  }
}
</script>

<style scoped>
.config-page {
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
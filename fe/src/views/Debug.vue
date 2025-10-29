<!-- src/views/Debug.vue -->
<template>
  <div class="debug-page">
    <div class="page-header">
      <h1>调试工具</h1>
    </div>
    
    <DebugPanel 
      :loading-states="loadingStates" 
      :debug-info="debugInfo"
      @update-debug-info="updateDebugInfo" />
  </div>
</template>

<script>
import DebugPanel from '@/components/debug.vue'

export default {
  name: 'DebugView',
  components: {
    DebugPanel
  },
  data() {
    return {
      debugInfo: null,
      loadingStates: {
        debug: false
      }
    }
  },
  methods: {
    // 更新调试信息（由Debug组件触发）
    updateDebugInfo(info) {
      this.debugInfo = info
    }
  }
}
</script>

<style scoped>
.debug-page {
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
</style>
<!-- src/App.vue -->
<template>
  <div id="app">
    <router-view />
  </div>
</template>

<script>
export default {
  name: 'App'
}
</script>

<style>
* {
  box-sizing: border-box;
}

body {
  margin: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
  background-color: #f5f5f5;
  color: #333;
}

#app {
  min-height: 100vh;
}

.btn {
  padding: 6px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.3s;
}

.card {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  padding: 20px;
}

/* 全局样式 */
.metric-grid {
  display: flex;
  flex-direction: column;
  gap: 20px;
  margin-bottom: 20px;
}

.metric-row {
  display: flex;
  gap: 20px;
}

.metric-card {
  flex: 1;
  background: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 8px;
  padding: 20px;
  text-align: center;
  min-width: 0;
}

.metric-value {
  font-size: 24px;
  font-weight: bold;
  color: #212529;
  margin-bottom: 8px;
}

.metric-label {
  font-size: 14px;
  color: #6c757d;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .metric-row {
    flex-direction: column;
  }
  
  .metric-card {
    margin-bottom: 15px;
  }
  
  .metric-card:last-child {
    margin-bottom: 0;
  }
}
</style>
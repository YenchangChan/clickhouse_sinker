<!-- src/views/Layout.vue -->
<template>
  <div class="app-layout">
    <!-- 左侧导航栏 -->
    <aside class="sidebar">
      <div class="sidebar-header">
        <h2>ClickHouse Sinker</h2>
      </div>
      <nav class="sidebar-nav">
        <router-link 
          to="/overview" 
          class="nav-item"
          active-class="active"
        >
          <span class="nav-icon">📊</span>
          <span class="nav-text">系统概览</span>
        </router-link>
        <router-link 
          to="/tasks" 
          class="nav-item"
          active-class="active"
        >
          <span class="nav-icon">📋</span>
          <span class="nav-text">任务管理</span>
        </router-link>
        <router-link 
          to="/config" 
          class="nav-item"
          active-class="active"
        >
          <span class="nav-icon">⚙️</span>
          <span class="nav-text">配置管理</span>
        </router-link>
        <router-link 
          to="/debug" 
          class="nav-item"
          active-class="active"
        >
          <span class="nav-icon">🐛</span>
          <span class="nav-text">调试工具</span>
        </router-link>
        <router-link 
          to="/logs" 
          class="nav-item"
          active-class="active"
        >
          <span class="nav-icon">📝</span>
          <span class="nav-text">查看日志</span>
        </router-link>
      </nav>
    </aside>
    
    <!-- 主内容区域 -->
    <main class="main-content">
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>
  </div>
</template>

<script>
export default {
  name: 'AppLayout'
}
</script>

<style scoped>
.app-layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

/* 左侧导航栏样式 */
.sidebar {
  width: 240px;
  background-color: #001529;
  color: #fff;
  display: flex;
  flex-direction: column;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.1);
}

.sidebar-header {
  padding: 20px;
  border-bottom: 1px solid #1890ff;
}

.sidebar-header h2 {
  margin: 0;
  font-size: 18px;
  color: #fff;
}

.sidebar-nav {
  flex: 1;
  padding: 20px 0;
}

.nav-item {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  color: rgba(255, 255, 255, 0.65);
  text-decoration: none;
  transition: all 0.3s;
  border-left: 3px solid transparent;
}

.nav-item:hover {
  background-color: rgba(255, 255, 255, 0.08);
  color: #fff;
}

.nav-item.active {
  background-color: #1890ff;
  color: #fff;
  border-left-color: #fff;
}

.nav-icon {
  font-size: 18px;
  margin-right: 12px;
}

.nav-text {
  font-size: 14px;
}

/* 主内容区域样式 */
.main-content {
  flex: 1;
  background-color: #f0f2f5;
  overflow-y: auto;
}

/* 页面切换动画 */
.fade-enter-active, .fade-leave-active {
  transition: opacity 0.3s;
}

.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
</style>
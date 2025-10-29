<!-- src/components/Debug.vue -->
<template>
  <div class="debug-view-container">
    <div class="card">
      <div class="tab-container">
        <!-- Tab 头部和操作按钮 -->
        <div class="tab-header-wrapper">
          <div class="tab-header">
            <div class="tab-item" :class="{active: activeTab === 'heap'}" @click="switchTab('heap')">
              Heap信息
            </div>
            <div class="tab-item" :class="{active: activeTab === 'goroutine'}" @click="switchTab('goroutine')">
              Goroutine信息
            </div>
            <div class="tab-item" :class="{active: activeTab === 'metrics'}" @click="switchTab('metrics')">
              指标信息
            </div>
          </div>
          <div class="tab-actions">
            <button class="btn refresh-btn" @click="refreshCurrentTab">
              刷新
            </button>
          </div>
        </div>
        
        <!-- Tab 内容 - Heap -->
        <div v-if="activeTab === 'heap'" class="tab-content">
          <div class="content-header">
            <span class="content-title">Heap信息</span>
            <div class="copy-icon-wrapper" @click="copyToClipboard(heapInfo, 'heap')">
              <i class="copy-icon" :class="{ copied: copiedStates.heap }" title="复制内容"></i>
              <span v-if="copiedStates.heap" class="copy-tooltip">已复制!</span>
            </div>
          </div>
          <div class="dark-textarea" v-html="formatText(heapInfo)"></div>
        </div>
        
        <!-- Tab 内容 - Goroutine -->
        <div v-if="activeTab === 'goroutine'" class="tab-content">
          <div class="content-header">
            <span class="content-title">Goroutine信息</span>
            <div class="copy-icon-wrapper" @click="copyToClipboard(goroutineInfo, 'goroutine')">
              <i class="copy-icon" :class="{ copied: copiedStates.goroutine }" title="复制内容"></i>
              <span v-if="copiedStates.goroutine" class="copy-tooltip">已复制!</span>
            </div>
          </div>
          <div class="dark-textarea" v-html="formatText(goroutineInfo)"></div>
        </div>
        
        <!-- Tab 内容 - Metrics -->
        <div v-if="activeTab === 'metrics'" class="tab-content">
          <div class="filter-container">
            <input 
              type="text" 
              v-model="metricsFilter" 
              placeholder="输入关键词筛选指标..." 
              class="filter-input"
              @input="filterMetrics"
            />
            <button class="btn clear-filter-btn" @click="clearMetricsFilter" v-if="metricsFilter">
              清除筛选
            </button>
          </div>
          <div class="dark-textarea" v-html="formatText(filteredMetricsInfo)"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import apiService from '@/services/api.js'
import hljs from 'highlight.js/lib/core';
import bash from 'highlight.js/lib/languages/bash';
import 'highlight.js/styles/vs2015.css';

// 注册bash语言
hljs.registerLanguage('bash', bash);

export default {
  name: 'DebugPanel',
  props: {
    loadingStates: {
      type: Object,
      default: () => ({})
    },
    debugInfo: {
      type: [String, Object],
      default: ''
    }
  },
  emits: ['update-debug-info'],
  watch: {
    // 保持对旧的debugInfo属性的兼容性
    debugInfo: {
      handler(newVal) {
        if (newVal && this.activeTab === 'heap') {
          this.heapInfo = newVal;
        }
      },
      immediate: true
    }
  },
  data() {
    return {
      activeTab: 'heap',
      heapInfo: '点击刷新按钮获取Heap信息...',
      goroutineInfo: '点击刷新按钮获取Goroutine信息...',
      metricsInfo: '点击刷新按钮获取指标信息...',
      metricsFilter: '',
      filteredMetricsInfo: '点击刷新按钮获取指标信息...',
      copiedStates: {
        heap: false,
        goroutine: false
      }
    };
  },
  methods: {
    switchTab(tab) {
      this.activeTab = tab;
      // 切换tab时自动刷新数据
      if (tab === 'heap' && this.heapInfo === '点击刷新按钮获取Heap信息...') {
        this.fetchHeapInfo();
      } else if (tab === 'goroutine' && this.goroutineInfo === '点击刷新按钮获取Goroutine信息...') {
        this.fetchGoroutineInfo();
      } else if (tab === 'metrics' && this.metricsInfo === '点击刷新按钮获取指标信息...') {
        this.fetchMetricsInfo();
      }
    },
    
    async fetchHeapInfo() {
      try {
        this.heapInfo = '加载中...';
        const response = await apiService.getHeapInfo();
        this.heapInfo = response.data;
      } catch (error) {
        const errorMsg = error.response?.data || error.message || '未知错误';
        this.heapInfo = `❌ 获取Heap信息失败: ${errorMsg}`;
        console.error('Heap信息获取错误:', error);
      }
    },
    
    async fetchGoroutineInfo() {
      try {
        this.goroutineInfo = '加载中...';
        const response = await apiService.getGoroutineInfo();
        this.goroutineInfo = response.data;
      } catch (error) {
        const errorMsg = error.response?.data || error.message || '未知错误';
        this.goroutineInfo = `❌ 获取Goroutine信息失败: ${errorMsg}`;
        console.error('Goroutine信息获取错误:', error);
      }
    },
    
    async fetchMetricsInfo() {
      try {
        this.metricsInfo = '加载中...';
        this.filteredMetricsInfo = '加载中...';
        const response = await apiService.getMetricsInfo();
        this.metricsInfo = response.data;
        this.filterMetrics();
      } catch (error) {
        const errorMsg = error.response?.data || error.message || '未知错误';
        this.metricsInfo = `❌ 获取指标信息失败: ${errorMsg}`;
        this.filteredMetricsInfo = this.metricsInfo;
        console.error('指标信息获取错误:', error);
      }
    },
    
    filterMetrics() {
      if (!this.metricsInfo || this.metricsInfo === '加载中...' || this.metricsInfo.startsWith('❌')) {
        this.filteredMetricsInfo = this.metricsInfo;
        return;
      }
      
      if (!this.metricsFilter.trim()) {
        this.filteredMetricsInfo = this.metricsInfo;
        return;
      }
      
      const filterText = this.metricsFilter.toLowerCase();
      const lines = this.metricsInfo.split('\n');
      const filteredLines = [];
      let includeCurrentBlock = false;
      
      for (let i = 0; i < lines.length; i++) {
        const line = lines[i];
        
        // 检查指标名称行或注释行
        if (line.startsWith('#') || (line.trim() && !line.startsWith('#') && !line.startsWith(' ') && !line.endsWith('}'))) {
          // 如果行包含筛选文本，标记当前块需要包含
          if (line.toLowerCase().includes(filterText)) {
            includeCurrentBlock = true;
            filteredLines.push(line);
          } else {
            // 否则检查下一行是否可能是指标值行
            const nextLine = lines[i + 1] || '';
            if (nextLine.trim() && (nextLine.startsWith(' ') || nextLine.endsWith('}')) && 
                (line.toLowerCase().includes(filterText) || nextLine.toLowerCase().includes(filterText))) {
              includeCurrentBlock = true;
              filteredLines.push(line);
            } else {
              includeCurrentBlock = false;
            }
          }
        } else if (includeCurrentBlock || line.toLowerCase().includes(filterText)) {
          // 如果当前在包含块中，或者行本身包含筛选文本，包含该行
          filteredLines.push(line);
          
          // 如果是空行，重置块状态
          if (!line.trim()) {
            includeCurrentBlock = false;
          }
        }
      }
      
      this.filteredMetricsInfo = filteredLines.length > 0 ? filteredLines.join('\n') : `没有找到包含 "${this.metricsFilter}" 的指标信息`;
    },
    
    clearMetricsFilter() {
      this.metricsFilter = '';
      this.filteredMetricsInfo = this.metricsInfo;
    },
    
    // 复制内容到剪贴板
    copyToClipboard(content, type) {
      // 跳过加载中和错误状态
      if (!content || content === '加载中...' || content.startsWith('❌')) {
        return;
      }
      
      // 创建临时文本区域
      const textarea = document.createElement('textarea');
      textarea.value = content;
      textarea.style.position = 'fixed';
      textarea.style.opacity = '0';
      document.body.appendChild(textarea);
      
      // 选择并复制文本
      textarea.select();
      document.execCommand('copy');
      
      // 清理临时元素
      document.body.removeChild(textarea);
      
      // 更新复制状态并显示提示
      this.copiedStates[type] = true;
      
      // 2秒后重置复制状态
      setTimeout(() => {
        this.copiedStates[type] = false;
      }, 2000);
    },
    
    // 刷新当前选中的标签页数据
    refreshCurrentTab() {
      if (this.activeTab === 'heap') {
        this.fetchHeapInfo();
      } else if (this.activeTab === 'goroutine') {
        this.fetchGoroutineInfo();
      } else if (this.activeTab === 'metrics') {
        this.fetchMetricsInfo();
      }
    },
    
    formatText(text) {
      if (!text) return '';
      
      // 尝试使用bash语法高亮
      try {
        const highlighted = hljs.highlight(text, { language: 'bash' }).value;
        return highlighted;
      } catch (error) {
        // 如果高亮失败，回退到简单格式化
        console.warn('代码高亮失败，使用回退格式:', error);
        return text.replace(/\n/g, '<br>').replace(/ /g, '&nbsp;');
      }
    }
  },
  mounted() {
    // 初始化时自动加载当前选中的tab数据
    if (this.activeTab === 'heap') {
      this.fetchHeapInfo();
    }
  }
};
</script>

<style scoped>
.debug-view-container {
  padding: 20px;
}

.tab-container {
  margin-top: 20px;
}

.tab-header-wrapper {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 2px solid #eee;
  margin-bottom: 15px;
  padding-bottom: 10px;
}

.tab-actions {
  display: flex;
  gap: 10px;
}

.refresh-btn {
  background-color: #1890ff;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.3s;
}

.refresh-btn:hover {
  background-color: #40a9ff;
}

.tab-header {
  display: flex;
  border-bottom: none;
  margin-bottom: 0;
}

.tab-item {
  padding: 10px 20px;
  cursor: pointer;
  margin-right: 5px;
  border-bottom: 3px solid transparent;
  font-weight: 500;
  transition: all 0.3s ease;
}

.tab-item:hover {
  background-color: #f8f9fa;
}

.tab-item.active {
  color: #007bff;
  border-bottom-color: #007bff;
  background-color: #f8f9fa;
}

.tab-actions {
  margin-bottom: 0;
}

.tab-content {
  padding: 15px;
  background-color: #f8f9fa;
  border-radius: 4px;
}

.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.content-title {
  font-weight: 500;
  color: #333;
  font-size: 14px;
}

.copy-icon-wrapper {
  position: relative;
  cursor: pointer;
  padding: 4px;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
}

.copy-icon-wrapper:hover {
  background-color: rgba(0, 0, 0, 0.05);
}

.copy-icon {
  width: 16px;
  height: 16px;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%23666'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z'/%3E%3C/svg%3E");
  background-size: contain;
  background-repeat: no-repeat;
  transition: all 0.2s ease;
}

.copy-icon.copied {
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%2352c41a'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M5 13l4 4L19 7'/%3E%3C/svg%3E");
}

.copy-tooltip {
  position: absolute;
  right: 100%;
  top: 50%;
  transform: translateY(-50%);
  margin-right: 8px;
  padding: 4px 8px;
  background-color: #333;
  color: white;
  font-size: 12px;
  border-radius: 4px;
  white-space: nowrap;
  pointer-events: none;
  z-index: 1000;
}

.copy-tooltip::after {
  content: '';
  position: absolute;
  left: 100%;
  top: 50%;
  transform: translateY(-50%);
  border: 4px solid transparent;
  border-left-color: #333;
}

.filter-container {
  display: flex;
  gap: 10px;
  margin-bottom: 15px;
}

.filter-input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  max-width: 400px;
}

.clear-filter-btn {
  background-color: #6c757d;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.3s;
}

.clear-filter-btn:hover {
  background-color: #5a6268;
}

.dark-textarea {
  width: 100%;
  height: 400px;
  background-color: #1e1e1e;
  color: #d4d4d4;
  border: 1px solid #3c3c3c;
  border-radius: 4px;
  padding: 15px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.6;
  resize: vertical;
  white-space: pre-wrap;
  overflow-wrap: break-word;
  overflow-y: auto;
  /* 确保代码高亮样式正确应用 */
}

/* 确保highlight.js样式正确应用 */
.dark-textarea :deep(pre) {
  margin: 0;
  background: transparent;
  font-family: inherit;
  font-size: inherit;
  line-height: inherit;
}

.dark-textarea :deep(code) {
  background: transparent;
  font-family: inherit;
  font-size: inherit;
  line-height: inherit;
  white-space: pre-wrap;
  overflow-wrap: break-word;
}

.refresh-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 8px 16px;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.3s ease;
}

.refresh-button:hover {
  background-color: #0056b3;
}

.loading-spinner {
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 2px solid #f3f3f3;
  border-top: 2px solid #3498db;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-right: 5px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}
</style>
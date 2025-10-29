<!-- src/views/Logs.vue -->
<template>
  <div class="logs-container">
    <div class="card">
      <div class="card-header">
        <h2>应用日志</h2>
      </div>
      <div class="card-body">
        <div v-if="loading" class="loading-state">
          <div class="loading-spinner"></div>
          <span>加载中...</span>
        </div>
        <div v-else-if="error" class="error-state">
          <p>{{ error }}</p>
          <button class="btn retry-btn" @click="fetchLogs">重试</button>
        </div>
        <div v-else>
          <!-- 控制选项 -->
          <div class="logs-controls">
            <div class="control-group">
              <label class="control-item">
                <input type="checkbox" v-model="formatJson">
                <span>格式化JSON</span>
              </label>
              <label class="control-item">
                <input type="checkbox" v-model="enableWrap">
                <span>换行显示</span>
              </label>
              <label class="control-item">
                <input type="checkbox" v-model="showErrorLogs">
                <span>仅显示错误日志</span>
              </label>
            </div>
            <div class="control-actions">
              <button class="btn refresh-btn" @click="refreshLogs" :disabled="loading">
                <span v-if="!loading">刷新</span>
                <span v-else>刷新中...</span>
              </button>
            </div>
          </div>
          
          <div class="logs-content">
            <div v-if="logs.length === 0" class="empty-state">
              暂无日志数据
            </div>
            <div v-else class="code-container">
              <div class="line-numbers">
                <div v-for="i in logs.length" :key="i" class="line-number">{{ i }}</div>
              </div>
              <pre class="logs-pre" :class="{ 'nowrap': !enableWrap }" v-html="formatLogs(logs)"></pre>
            </div>
          </div>
          
          <!-- 分页控件 -->
          <div class="pagination">
            <button 
              class="page-btn" 
              :disabled="currentPage === 1"
              @click="goToPage(1)"
            >
              首页
            </button>
            <button 
              class="page-btn" 
              :disabled="currentPage === 1"
              @click="goToPage(currentPage - 1)"
            >
              上一页
            </button>
            <span class="page-info">
              第 {{ currentPage }} / {{ totalPages }} 页，共 {{ total }} 条记录
            </span>
            <button 
              class="page-btn" 
              :disabled="currentPage === totalPages"
              @click="goToPage(currentPage + 1)"
            >
              下一页
            </button>
            <button 
              class="page-btn" 
              :disabled="currentPage === totalPages"
              @click="goToPage(totalPages)"
            >
              末页
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import apiService from '@/services/api.js'

export default {
  name: 'LogsView',
  data() {
    return {
      logs: [],
      total: 0,
      currentPage: 1,
      pageSize: 100,
      loading: false,
      error: null,
        formatJson: false,
        enableWrap: true,
        showErrorLogs: false
    }
  },
  computed: {
    totalPages() {
      return Math.ceil(this.total / this.pageSize);
    }
  },
  watch: {
    // 监听仅显示错误日志选项的变化，自动刷新日志
    showErrorLogs() {
      // 切换选项时重置到第一页
      this.currentPage = 1;
      this.fetchLogs();
    }
  },
  mounted() {
    this.fetchLogs();
  },
  methods: {
    async fetchLogs() {
      this.loading = true;
      this.error = null;
      
      const from = (this.currentPage - 1) * this.pageSize;
      
      try {
        // 直接传递参数，不再嵌套在对象中
        const response = await apiService.getLog(from, this.showErrorLogs ? true : undefined);
        console.log('获取日志成功:', response.data);
        this.logs = response.data.entity.Lines || [];
        this.total = response.data.entity.Total || 0;
      } catch (error) {
        this.error = `获取日志失败: ${error.response?.data || error.message || '未知错误'}`;
        console.error('日志获取错误:', error);
      } finally {
        this.loading = false;
      }
    },
    
    formatLogs(logArray) {
      if (!Array.isArray(logArray) || logArray.length === 0) {
        return '';
      }
      
      return logArray.map(log => {
        try {
          // 尝试JSON格式化
          if (this.formatJson) {
            // 检查是否为JSON字符串（更健壮的检测）
            let parsedLog;
            if (typeof log === 'string') {
              // 清理可能的前后空白字符
              const trimmedLog = log.trim();
              // 检查是否符合JSON格式
              if ((trimmedLog.startsWith('{') && trimmedLog.endsWith('}')) ||
                  (trimmedLog.startsWith('[') && trimmedLog.endsWith(']'))) {
                // 尝试解析JSON
                parsedLog = JSON.parse(trimmedLog);
                // 返回高亮格式化的JSON
                return this.highlightJson(JSON.stringify(parsedLog, null, 2));
              }
            }
          }
          // 普通文本显示
          return this.escapeHtml(log);
        } catch (e) {
          // JSON解析失败，返回原文本
          return this.escapeHtml(log);
        }
      }).join('\n');
    },
    
    // 转义HTML字符
    escapeHtml(text) {
      const map = {
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;',
        "'": '&#039;'
      };
      return text.replace(/[&<>"']/g, m => map[m]);
    },
    
    // JSON语法高亮
    highlightJson(json) {
      return json
        .replace(/"([^"\\]|\\.)*"/g, '<span class="json-string">$&</span>') // 字符串
        .replace(/\b(true|false|null)\b/g, '<span class="json-keyword">$&</span>') // 关键字
        .replace(/\b\d+(\.\d+)?\b/g, '<span class="json-number">$&</span>') // 数字
        .replace(/([{}[\]])/g, '<span class="json-punctuation">$&</span>') // 标点符号
        .replace(/([:,])/g, '<span class="json-separator">$&</span>'); // 分隔符
    },
    
    goToPage(page) {
      if (page >= 1 && page <= this.totalPages && page !== this.currentPage) {
        this.currentPage = page;
        this.fetchLogs();
      }
    },
    
    // 刷新日志（重新加载当前页）
    refreshLogs() {
      this.fetchLogs();
    }
  }
}
</script>

<style scoped>
.logs-container {
  padding: 20px;
  min-height: calc(100vh - 40px);
  overflow: visible !important;
}

.card {
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  overflow: visible !important;
  min-height: calc(100vh - 80px);
  display: flex;
  flex-direction: column;
}

.card-header {
  display: flex;
  align-items: center;
  padding: 16px 20px;
  background-color: #ffffff;
  border-bottom: 1px solid #e8e8e8;
}

.card-header h2 {
  margin: 0;
  font-size: 18px;
  color: #000000;
  font-weight: 500;
}

.card-body {
  padding: 20px;
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: visible !important;
}

.retry-btn {
  background-color: #fff;
  color: #1890ff;
  border: 1px solid #1890ff;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s;
}

.retry-btn:hover {
  background-color: #1890ff;
  color: #fff;
}

.loading-state,
.empty-state,
.error-state {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 300px;
  color: #666;
}

.loading-spinner {
  width: 32px;
  height: 32px;
  border: 3px solid #f3f3f3;
  border-top: 3px solid #1890ff;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.logs-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  padding: 10px;
  background-color: #f5f5f5;
  border-radius: 4px;
  gap: 20px;
}

.control-group {
  display: flex;
  gap: 20px;
}

.control-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  cursor: pointer;
}

.control-item input[type="checkbox"] {
  cursor: pointer;
}

.control-actions {
  display: flex;
  gap: 10px;
}

/* 按钮样式 */
.btn {
  padding: 6px 16px;
  border: 1px solid #ddd;
  background-color: #fff;
  color: #333;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s ease;
}

.btn:hover {
  background-color: #f5f5f5;
  border-color: #bbb;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.refresh-btn {
  background-color: #3498db;
  border-color: #3498db;
  color: white;
}

.refresh-btn:hover:not(:disabled) {
  background-color: #2980b9;
  border-color: #2980b9;
}

.logs-content {
  overflow: auto;
  margin-bottom: 20px;
  border: 1px solid #e8e8e8;
  border-radius: 4px;
  height: calc(100vh - 280px);
  min-height: 300px;
  position: relative;
}

.logs-pre {
  background-color: #f5f5f5;
  color: #333;
  padding: 16px;
  margin: 0;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-wrap: break-word;
  flex: 1;
  overflow: auto;
}

.logs-pre.nowrap {
  white-space: pre;
  word-wrap: normal;
  overflow-x: auto;
  overflow-y: auto;
}

/* 代码容器与行号样式 */
.code-container {
  display: flex;
  position: relative;
  height: 100%;
  overflow: hidden;
}

.line-numbers {
  background-color: #f1f1f1;
  border-right: 1px solid #ddd;
  padding: 16px 8px;
  text-align: right;
  user-select: none;
  overflow: hidden;
  min-width: 40px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.5;
  color: #888;
}

.line-number {
  height: 21px;
}

/* JSON高亮样式 - 全局样式以确保正确应用 */
:deep(.json-string) {
  color: #a31515 !important;
  font-weight: 500 !important;
}

:deep(.json-keyword) {
  color: #0000ff !important;
  font-weight: bold !important;
}

:deep(.json-number) {
  color: #098658 !important;
  font-weight: 500 !important;
}

:deep(.json-punctuation) {
  color: #000000 !important;
  font-weight: bold !important;
}

:deep(.json-separator) {
  color: #666666 !important;
  font-weight: 500 !important;
}

.pagination {
  display: flex !important;
  justify-content: flex-end !important;
  align-items: center !important;
  gap: 10px;
  margin-top: 10px;
  padding: 10px;
  background-color: #fff !important;
  border: 1px solid #e8e8e8 !important;
  border-radius: 4px;
  flex-wrap: wrap;
  min-height: 40px;
  width: 100% !important;
  box-sizing: border-box !important;
  position: sticky;
  bottom: 0;
  z-index: 1000;
  visibility: visible !important;
  opacity: 1 !important;
  clear: both;
}

.page-btn {
  background-color: #fff;
  border: 1px solid #1890ff;
  color: #1890ff;
  padding: 6px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s;
}

.page-btn:hover:not(:disabled) {
  background-color: #1890ff;
  color: #fff;
}

.page-btn:disabled {
  cursor: not-allowed;
  opacity: 0.3;
  background-color: #f5f5f5;
  border-color: #d9d9d9;
  color: #d9d9d9;
}

.page-info {
  color: #1890ff;
  font-size: 14px;
  margin: 0 10px;
  font-weight: 500;
}
</style>
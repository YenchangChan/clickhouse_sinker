<!-- src/components/Tasks.vue -->
<template>
  <div>
    <div class="card">
      <!-- 任务控制区域 -->
      <div class="task-controls">
        <div class="filter-section">
          <input 
            type="text" 
            placeholder="任务名称..." 
            class="filter-input"
            v-model="taskFilter.name"
            @input="applyFilters"
          />
          <select 
            class="filter-select"
            v-model="taskFilter.type"
            @change="applyFilters"
          >
            <option value="">所有类型</option>
            <option value="metric">指标</option>
            <option value="log">日志</option>
          </select>
          <select 
            class="filter-select"
            v-model="taskFilter.status"
            @change="applyFilters"
          >
            <option value="">全部状态</option>
            <option value="Stable">Stable</option>
            <option value="Empty">Empty</option>
            <option value="Dead">Dead</option>
            <option value="Rebalancing">Rebalancing</option>
          </select>
        </div>
        <div class="action-section">
          <span class="task-count">共 {{ filteredTasks.length }} 个任务</span>
          <select 
            class="page-size-select"
            v-model="pageSize"
            @change="handlePageSizeChange"
          >
            <option value="10">10条/页</option>
            <option value="20">20条/页</option>
            <option value="50">50条/页</option>
          </select>
        </div>
      </div>

      <!-- 任务表格 -->
      <div v-if="paginatedTasks.length > 0" class="table-container">
        <table class="task-table">
          <thead>
            <tr>
              <th @click="sortBy('name')" class="sortable">
                任务名 <span class="sort-indicator" v-if="sortField === 'name'">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th>集群</th>
              <th>表名</th>
              <th>topic</th>
              <th>消费者组</th>
              <th @click="sortBy('type')" class="sortable">
                类型 <span class="sort-indicator" v-if="sortField === 'type'">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th @click="sortBy('status')" class="sortable">
                状态 <span class="sort-indicator" v-if="sortField === 'status'"> {{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th @click="sortBy('lag')" class="sortable">
                kafka Lag <span class="sort-indicator" v-if="sortField === 'lag'"> {{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th @click="sortBy('writeSpeed')" class="sortable">
                写入速度 <span class="sort-indicator" v-if="sortField === 'writeSpeed'">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="task in paginatedTasks" :key="task.Name">
              <td><span class="task-name">{{ task.Name }}</span></td>
              <td>{{ task.Cluster || 'N/A' }}</td>
              <td>{{ task.Table || 'N/A' }}</td>
              <td>{{ task.Topic || 'N/A' }}</td>
              <td>{{ task.ConsumerGroup || 'N/A' }}</td>
              <td>
                <span :class="['task-type', task.Type]">{{ task.Type === 'metric' ? '指标' : '日志' }}</span>
              </td>
              <td>
                <span :class="['status-badge', `status-${task.Status.toLowerCase()}`]">
                  {{ task.Status }}
                </span>
              </td>
              <td :class="{ 'lag-critical': task.Lag > 10000, 'lag-warning': task.Lag > 1000 && task.Lag <= 10000 }">
                {{ task.Lag || 0 }}
              </td>
              <td>{{ task.Rate || 'N/A' }}</td>
              <td>
                <button class="btn-small" @click="viewTaskConfig(task.Name)">配置</button>
                <button class="btn-small" @click="viewTaskDbKey(task.Name)">dbkey</button>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- 分页 -->
        <div class="pagination">
          <button class="btn-small" :disabled="currentPage === 1" @click="prevPage">上一页</button>
          <span class="page-info">{{ currentPage }} / {{ totalPages }}</span>
          <button class="btn-small" :disabled="currentPage === totalPages" @click="nextPage">下一页</button>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-else class="empty-state">
        暂无任务数据
      </div>
    </div>

    <!-- 任务详情弹窗 -->
    <!-- 配置弹窗 -->
    <div v-if="showConfigModal" class="modal-overlay" @click.self="closeConfigModal">
      <div class="modal-content">
        <div class="modal-header">
          <h3 class="modal-title">任务配置 - {{ currentTaskName }}</h3>
          <button class="modal-close" @click="closeConfigModal">&times;</button>
        </div>
        <div class="modal-body">
          <div v-if="loadingConfig" class="loading">加载中...</div>
          <div v-else class="json-display-container">
            <pre class="json-display" v-html="highlightJson(taskConfig)"></pre>
          </div>
        </div>
      </div>
    </div>

    <!-- dbkey弹窗 -->
    <div v-if="showDbKeyModal" class="modal-overlay" @click.self="closeDbKeyModal">
      <div class="modal-content">
        <div class="modal-header">
          <h3 class="modal-title">任务dbkey - {{ currentTaskName }}</h3>
          <button class="modal-close" @click="closeDbKeyModal">&times;</button>
        </div>
        <div class="modal-body">
          <div v-if="loadingDbKey" class="loading">加载中...</div>
          <div v-else>
            <!-- 添加搜索筛选功能 -->
            <div class="search-container" style="margin-bottom: 16px;">
              <input 
                type="text" 
                v-model="dbKeySearchKeyword"
                placeholder="搜索数据库名..."
                style="width: 250px; padding: 6px 12px; border: 1px solid #ddd; border-radius: 4px; font-size: 14px;"
              >
              <span style="margin-left: 8px; color: #666; font-size: 14px;">
                共 {{ Array.isArray(taskDbKey) ? taskDbKey.length : 0 }} 条记录
                <template v-if="dbKeySearchKeyword">, 筛选后 {{ filteredDbKeyData.length }} 条</template>
              </span>
            </div>
            
            <!-- 直接渲染表格，不使用复杂的条件判断 -->
            <table class="dbkey-table">
              <thead>
                <tr>
                  <th>数据库名</th>
                  <th>insertSQL</th>
                  <th>series表SQL</th>
                  <th>处理数据</th>
                  <th style="min-width: 100px; width: 100px;">字段个数</th>
                  <th>series下标</th>
                </tr>
              </thead>
              <tbody>
                <!-- 确保taskDbKey是数组并应用筛选 -->
                <tr v-for="(item, index) in filteredDbKeyData" :key="index">
                  <td>{{ item.name || '-' }}</td>
                  <td 
                    class="sql-cell" 
                    :title="item.prepareSQL"
                    @dblclick="copyToClipboard(item.prepareSQL)"
                  >{{ item.prepareSQL || '-' }}</td>
                  <td 
                    class="sql-cell" 
                    :title="item.promSerSQL"
                    @dblclick="copyToClipboard(item.promSerSQL)"
                  >{{ item.promSerSQL || '-' }}</td>
                  <td>{{ item.processed || '-' }}</td>
                  <td>{{ item.numDims || '-' }}</td>
                  <td>{{ item.idxSerId || '-' }}</td>
                </tr>
                <!-- 如果没有数据，显示提示行 -->
                <tr v-if="filteredDbKeyData.length === 0">
                  <td colspan="6" style="text-align: center; color: #666; padding: 20px;">
                    {{ dbKeySearchKeyword ? '没有找到匹配的数据库名' : '暂无数据' }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import api from '../services/api';

export default {
  name: 'TasksList',
  props: {
    tasks: {
      type: Array,
      default: () => []
    },
    loadingStates: {
      type: Object,
      default: () => ({})
    }
  },
  emits: ['refresh-tasks'],
  data() {
    return {
      taskFilter: {
        name: '',
        type: '',
        status: ''
      },
      sortField: 'name',
      sortOrder: 'asc',
      currentPage: 1,
      pageSize: 10,
      selectedTask: null,
      showConfigModal: false,
      showDbKeyModal: false,
      currentTaskName: '',
      taskConfig: null,
      taskDbKey: null,
      loadingConfig: false,
      loadingDbKey: false,
      dbKeySearchKeyword: ''
    };
  },
  computed: {
    // 筛选dbkey数据
    filteredDbKeyData() {
      if (!Array.isArray(this.taskDbKey)) {
        return [];
      }
      
      const keyword = this.dbKeySearchKeyword.toLowerCase().trim();
      if (!keyword) {
        return this.taskDbKey;
      }
      
      return this.taskDbKey.filter(item => {
        return item.name && item.name.toLowerCase().includes(keyword);
      });
    },
      filteredTasks() {
      // 确保tasks是数组类型
      let result = Array.isArray(this.tasks) ? [...this.tasks] : [];
      
      // 应用过滤
      if (this.taskFilter.name) {
        result = result.filter(task => 
          task && task.Name && task.Name.toLowerCase().includes(this.taskFilter.name.toLowerCase())
        );
      }
      if (this.taskFilter.type) {
        result = result.filter(task => task && task.Type === this.taskFilter.type);
      }
      if (this.taskFilter.status) {
        result = result.filter(task => task && task.Status && task.Status.toLowerCase() === this.taskFilter.status.toLowerCase());
      }
      
      // 应用排序
      result.sort((a, b) => {
        // 确保比较的属性存在
        if (!a || !b) return 0;
        
        // 映射排序字段到API返回的字段名
        const fieldMap = {
          'name': 'Name',
          'type': 'Type',
          'status': 'Status',
          'lag': 'Lag',
          'writeSpeed': 'Rate'
        };
        
        const actualField = fieldMap[this.sortField] || this.sortField;
        let aVal = a[actualField];
        let bVal = b[actualField];
        
        if (typeof aVal === 'string' && typeof bVal === 'string') {
          aVal = aVal.toLowerCase();
          bVal = bVal.toLowerCase();
        }
        
        if (aVal < bVal) return this.sortOrder === 'asc' ? -1 : 1;
        if (aVal > bVal) return this.sortOrder === 'asc' ? 1 : -1;
        return 0;
      });
      
      return result;
    },
    formattedTaskData() {
      if (!this.selectedTask) return '';
      return JSON.stringify(this.selectedTask, null, 2);
    },
    paginatedTasks() {
      const start = (this.currentPage - 1) * this.pageSize;
      const end = start + this.pageSize;
      return this.filteredTasks.slice(start, end);
    },
    totalPages() {
      return Math.ceil(this.filteredTasks.length / this.pageSize);
    }
  },
  methods: {
    applyFilters() {
      this.currentPage = 1;
    },
    sortBy(field) {
      if (this.sortField === field) {
        this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc';
      } else {
        this.sortField = field;
        this.sortOrder = 'asc';
      }
    },
    prevPage() {
      if (this.currentPage > 1) {
        this.currentPage--;
      }
    },
    nextPage() {
      if (this.currentPage < this.totalPages) {
        this.currentPage++;
      }
    },
    handlePageSizeChange() {
      this.currentPage = 1;
    },
    refreshTasks() {
      this.$emit('refresh-tasks');
    },
    async viewTaskConfig(taskName) {
      this.currentTaskName = taskName;
      this.loadingConfig = true;
      try {
        const response = await api.getTaskConfig(taskName);
        this.taskConfig = JSON.stringify(response.data.entity, null, 2);
        this.showConfigModal = true;
      } catch (error) {
        console.error('获取任务配置失败:', error);
        alert('获取任务配置失败');
      } finally {
        this.loadingConfig = false;
      }
    },
    async viewTaskDbKey(taskName) {
      this.currentTaskName = taskName;
      this.loadingDbKey = true;
      try {
        const response = await api.getTaskDbKey(taskName);
        
        // 明确处理entity字段中的数据
        if (response.data && response.data.entity) {
          // 确保是数组格式
          const rawData = Array.isArray(response.data.entity) ? response.data.entity : [response.data.entity];
          
          // 转换数据字段名：将大写开头的字段映射为小写开头的字段，确保与表格绑定匹配
          this.taskDbKey = rawData.map(item => ({
            name: item.Name || '-',
            prepareSQL: item.PrepareSQL || '-',
            promSerSQL: item.PromSerSQL || '-',
            processed: item.Processed || '-',
            numDims: item.NumDims || '-',
            idxSerId: item.IdxSerID || '-'
          }));
        } else {
          // 默认情况: 空数组
          this.taskDbKey = [];
        }
        
        // 强制显示弹窗，确保数据加载后弹窗可见
        this.showDbKeyModal = true;
      } catch (error) {
        console.error('获取任务dbkey失败:', error);
        alert('获取任务dbkey失败');
      } finally {
        this.loadingDbKey = false;
      }
    },
    highlightJson(jsonStr) {
      if (!jsonStr) return '';
      
      // 使用更可靠的JSON高亮方法，避免重复替换
      try {
        // 首先替换布尔值和null
        let highlighted = jsonStr
          .replace(/\b(true|false)\b/g, '<span class="boolean">$1</span>')
          .replace(/\b(null)\b/g, '<span class="null">$1</span>')
          .replace(/\b(\d+(?:\.\d+)?)\b/g, '<span class="number">$1</span>');
        
        // 使用更精确的方式处理键和字符串值，避免重复替换
        // 先匹配键值对中的键
        highlighted = highlighted.replace(/("[^"\\]*(?:\\.[^"\\]*)*")\s*:/g, '<span class="key">$1</span>:');
        // 然后匹配剩余的字符串值（不包括已经被处理过的键）
        highlighted = highlighted.replace(/(?::\s*)("[^"\\]*(?:\\.[^"\\]*)*")/g, ': <span class="string">$1</span>');
        
        return highlighted;
      } catch (e) {
        console.error('JSON高亮失败:', e);
        return jsonStr; // 如果处理失败，返回原始字符串
      }
    },
    
    // 复制文本到剪贴板
    copyToClipboard(text) {
      if (!text) return;
      
      // 创建临时元素
      const textarea = document.createElement('textarea');
      textarea.value = text;
      textarea.style.position = 'fixed';
      textarea.style.opacity = '0';
      document.body.appendChild(textarea);
      
      // 选择并复制
      textarea.select();
      document.execCommand('copy');
      
      // 清理
      document.body.removeChild(textarea);
      
      // 显示复制成功提示
      const originalTitle = document.title;
      document.title = '复制成功！';
      setTimeout(() => {
        document.title = originalTitle;
      }, 1000);
    },
    closeConfigModal() {
      this.showConfigModal = false;
      this.taskConfig = null;
    },
    closeDbKeyModal() {
      this.showDbKeyModal = false;
      this.taskDbKey = null;
    }
  }
};
</script>

<style scoped>
.task-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding: 12px;
  background-color: #f5f5f5;
  border-radius: 4px;
}

.filter-section {
  display: flex;
  gap: 12px;
  align-items: center;
}

.filter-input,
.filter-select,
.page-size-select {
  padding: 6px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.filter-input {
  width: 200px;
}

.filter-select,
.page-size-select {
  width: auto;
  min-width: 120px;
}

.btn,
.btn-small {
  padding: 6px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.3s;
}

.btn {
  background-color: #1890ff;
  color: white;
}

.btn:hover {
  background-color: #40a9ff;
}

.btn-small {
  padding: 4px 6px;
  font-size: 11px;
  margin: 0 3px;
  cursor: pointer;
  border: none;
  border-radius: 3px;
  transition: background-color 0.3s;
  background-color: #1890ff;
  color: white;
  display: inline-block;
  min-width: 40px;
  text-align: center;
}

.task-table td:last-child {
  display: flex;
  justify-content: center;
  align-items: center;
  flex-wrap: nowrap;
}

.btn-small:hover {
  background-color: #e6f7ff;
}

.btn:disabled,
.btn-small:disabled {
  background-color: #d9d9d9;
  color: #bfbfbf;
  cursor: not-allowed;
}

.action-section {
  display: flex;
  gap: 16px;
  align-items: center;
}

.task-count {
  font-size: 14px;
  color: #666;
}

.empty-state {
  text-align: center;
  padding: 40px;
  color: #999;
  font-size: 14px;
}

.table-container {
  overflow-x: auto;
  margin-top: 20px;
  min-width: 1000px;
}

.task-table {
  width: 100%;
  min-width: 900px;
  border-collapse: collapse;
  font-size: 14px;
}

.task-table th,
.task-table td {
  padding: 8px 6px;
  text-align: left;
  border-bottom: 1px solid #f0f0f0;
  white-space: nowrap;
}

.task-table th {
    background-color: #fafafa;
    font-weight: 600;
    color: #333;
    font-size: 14px;
  }

.task-table th:nth-child(1),
.task-table td:nth-child(1) {
  width: 100px;
}

.task-table th:nth-child(2),
.task-table td:nth-child(2) {
  width: 70px;
}

.task-table th:nth-child(3),
.task-table td:nth-child(3) {
  width: 80px;
}

.task-table th:nth-child(4),
.task-table td:nth-child(4) {
  width: 120px;
}

.task-table th:nth-child(5),
.task-table td:nth-child(5) {
  width: 100px;
}

.task-table th:nth-child(6),
.task-table td:nth-child(6) {
  width: 50px;
}

.task-table th:nth-child(7),
.task-table td:nth-child(7) {
  width: 70px;
}

.task-table th:nth-child(8),
.task-table td:nth-child(8) {
  width: 80px;
}

.task-table th:nth-child(9),
.task-table td:nth-child(9) {
  width: 80px;
}

.task-table th:nth-child(10),
.task-table td:nth-child(10) {
  width: 120px;
  text-align: center;
}

.sortable {
  cursor: pointer;
  user-select: none;
}

.sortable:hover {
  background-color: #f0f0f0;
}

.sort-indicator {
  margin-left: 4px;
  font-size: 12px;
}

.task-name {
  font-weight: 500;
  color: #262626;
}

.task-type {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 12px;
}

.task-type.metric {
  background-color: #e6f7ff;
  color: #1890ff;
}

.task-type.log {
  background-color: #f6ffed;
  color: #52c41a;
}

.status-badge {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.status-running {
  background-color: #f6ffed;
  color: #52c41a;
  border: 1px solid #b7eb8f;
}

.status-stopped {
  background-color: #fff2e8;
  color: #fa8c16;
  border: 1px solid #ffd591;
}

.status-error {
  background-color: #fff1f0;
  color: #ff4d4f;
  border: 1px solid #ffccc7;
}

.lag-critical {
  color: #ff4d4f;
  font-weight: 600;
}

.lag-warning {
  color: #fa8c16;
  font-weight: 600;
}

.pagination {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 8px;
  margin-top: 16px;
  padding: 12px;
  background-color: #fafafa;
}

.page-info {
  font-size: 14px;
  color: #666;
  margin: 0 8px;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 4px;
  width: 90%;
  max-width: 1200px;
  max-height: 80vh;
  overflow-y: auto;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.modal-header {
  padding: 16px 24px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-title {
  font-size: 16px;
  font-weight: 500;
}

.modal-close {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  color: #999;
}

.modal-body {
  padding: 24px;
}

.json-display-container {
  background-color: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 4px;
  padding: 16px;
  max-height: 500px;
  overflow-y: auto;
}

.json-display {
  background: #f5f5f5;
  padding: 15px;
  border-radius: 5px;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
  font-family: 'Courier New', Courier, monospace;
}

/* JSON高亮样式 */
.json-display .key {
  color: #0000ff;
  font-weight: bold;
}

.json-display .string {
  color: #008000;
}

.json-display .boolean,
.json-display .number {
  color: #ff00ff;
}

.json-display .null {
  color: #800080;
}

/* dbkey表格样式 */
.dbkey-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.dbkey-table th,
.dbkey-table td {
  padding: 8px 12px;
  border: 1px solid #ddd;
  text-align: left;
}

.dbkey-table th {
  background-color: #f5f5f5;
  font-weight: bold;
}

.dbkey-table tr:nth-child(even) {
  background-color: #f9f9f9;
}

.dbkey-table td {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* SQL单元格样式 */
.sql-cell {
  cursor: pointer;
  border-bottom: 1px dashed #999;
  position: relative;
}
.sql-cell:hover {
  background-color: #f0f8ff;
  position: relative;
  z-index: 10;
}
.sql-cell:active {
  background-color: #e0e8f0;
}

/* 加载状态和无数据提示 */
.loading,
.no-data {
  text-align: center;
  padding: 20px;
  color: #666;
}

.card {
  background: white;
  border-radius: 4px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  padding: 16px;
  margin-bottom: 16px;
}
</style>

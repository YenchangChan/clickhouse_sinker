<!-- src/components/Config.vue -->
<template>
  <div class="config-view-container">
    <!-- 配置文件来源显示 -->
    <div class="card config-source-card">
      <h4>配置文件来源</h4>
      <div class="config-source">
        <div class="config-source-row">
          <div class="config-source-label">来源：</div>
          <div class="config-source-value">
            <span v-if="!cmdlineConfig.NacosCommonDataID || cmdlineConfig.NacosCommonDataID.trim() === ''" class="source-local">本地配置</span>
            <span v-else class="source-remote">远程配置 (Nacos)</span>
          </div>
        </div>
        <div v-if="!cmdlineConfig.NacosCommonDataID || cmdlineConfig.NacosCommonDataID.trim() === ''" class="config-file-path">
          <div class="config-source-row">
            <div class="config-source-label">本地配置文件：</div>
            <div class="config-source-value">{{ cmdlineConfig.LocalCfgFile || '未配置' }}</div>
          </div>
        </div>
        <div v-else class="nacos-info">
          <div class="config-source-row">
            <div class="config-source-label">Nacos 服务：</div>
            <div class="config-source-value">{{ cmdlineConfig.NacosAddr || '未配置' }}</div>
          </div>
          <div class="config-source-row">
            <div class="config-source-label">Nacos 命名空间：</div>
            <div class="config-source-value">{{ cmdlineConfig.NacosNamespaceID || '未配置' }}</div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Tab 切换 -->
    <div class="tab-container">
      <div class="tab-header">
        <div class="tab-item" :class="{active: activeTab === 'cmdline'}" @click="activeTab = 'cmdline'">命令行参数</div>
        <div class="tab-item" :class="{active: activeTab === 'config'}" @click="activeTab = 'config'">系统配置</div>
      </div>
      
      <!-- 命令行参数 Tab -->
      <div v-if="activeTab === 'cmdline'" class="tab-content">
        <div class="card">
          <form class="config-form">
            <div class="form-group">
              <label class="form-label">日志级别</label>
              <div class="form-value">{{ cmdlineConfig.LogLevel || '未设置' }}</div>
            </div>
            <div class="form-group">
              <label class="form-label">日志路径</label>
              <div class="form-value">{{ cmdlineConfig.LogPaths || '未设置' }}</div>
            </div>
            <div v-if="!cmdlineConfig.NacosCommonDataID || cmdlineConfig.NacosCommonDataID.trim() === ''"  
            class="form-group">
              <label class="form-label">本地配置文件</label>
              <div class="form-value">{{ cmdlineConfig.LocalCfgFile || '未设置' }}</div>
            </div>
            <div class="form-group">
              <label class="form-label">Nacos地址</label>
              <div class="form-value">{{ cmdlineConfig.NacosAddr || '未设置' }}</div>
            </div>
            <div class="form-group">
              <label class="form-label">Nacos命名空间ID</label>
              <div class="form-value">{{ cmdlineConfig.NacosNamespaceID || '未设置' }}</div>
            </div>
            <div class="form-group">
              <label class="form-label">Nacos组</label>
              <div class="form-value">{{ cmdlineConfig.NacosGroup || '未设置' }}</div>
            </div>
            <div class="form-group">
              <label class="form-label">Nacos用户名</label>
              <div class="form-value">{{ cmdlineConfig.NacosUsername || '未设置' }}</div>
            </div>
            <div class="form-group">
              <label class="form-label">Nacos密码</label>
              <div class="form-value">{{ cmdlineConfig.NacosPassword || '未设置' }}</div>
            </div>
            <div class="form-group">
              <label class="form-label">Nacos DataID</label>
              <div class="form-value">{{ cmdlineConfig.NacosDataID || '未设置' }}</div>
            </div>
            <div class="form-group">
              <label class="form-label">Nacos服务名</label>
              <div class="form-value">{{ cmdlineConfig.NacosServiceName || '未设置' }}</div>
            </div>
            <div class="form-group">
              <label class="form-label">Nacos公共命名空间ID</label>
              <div class="form-value">{{ cmdlineConfig.NacosCommonNamespaceID || '未设置' }}</div>
            </div>
            <div class="form-group">
              <label class="form-label">Nacos公共组</label>
              <div class="form-value">{{ cmdlineConfig.NacosCommonGroup || '未设置' }}</div>
            </div>
            <div class="form-group">
              <label class="form-label">Nacos公共DataID</label>
              <div class="form-value">{{ cmdlineConfig.NacosCommonDataID || '未设置' }}</div>
            </div>
          </form>
        </div>
      </div>
      
      <!-- 系统配置 Tab -->
      <div v-if="activeTab === 'config'" class="tab-content">
        <div class="card">
          <form class="config-form">
            <div v-for="(value, key) in config" :key="key" class="form-group">
              <label class="form-label">{{ formatKey(key) }}</label>
              <div class="form-value" v-if="typeof value === 'object' && value !== null">
                <div v-if="Array.isArray(value)" class="array-container">
                  <div v-for="(item, index) in value" :key="index" class="array-item">
                    <span class="array-index">{{ index + 1 }}.</span>
                    <span class="array-value">{{ formatValue(item) }}</span>
                  </div>
                </div>
                <div v-else>
                  <div class="nested-object">
                    <div v-for="(nestedValue, nestedKey) in value" :key="nestedKey" class="nested-item">
                      <span class="nested-key">{{ formatKey(nestedKey) }}:</span>
                      <span class="nested-value">{{ formatValue(nestedValue) }}</span>
                    </div>
                  </div>
                </div>
              </div>
              <div class="form-value" v-else>
                {{ formatValue(value) }}
              </div>
            </div>
          </form>
        </div>
      </div>
    </div>
    
    <button class="btn refresh-button" @click="$emit('refresh-config')"><i>↻</i> 刷新配置</button>
  </div>
</template>

<script>
export default {
  name: 'AppConfig',
  props: {
    config: {
      type: Object,
      default: () => ({})
    },
    cmdlineConfig: {
      type: Object,
      default: () => ({})
    },
    loadingStates: {
      type: Object,
      default: () => ({})
    }
  },
  emits: ['refresh-config'],
  data() {
    return {
      activeTab: 'cmdline'
    };
  },
  methods: {
    formatKey(key) {
      // 将驼峰命名转换为可读格式
      return key;
    },
    formatValue(value) {
      if (value === null) return 'null';
      if (value === undefined) return 'undefined';
      if (typeof value === 'object') {
        return JSON.stringify(value, null, 2);
      }
      return String(value);
    }
  }
}
</script>

<style scoped>
/* 配置页面主容器 */
.config-view-container {
  animation: fadeIn 0.5s ease-in-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

/* 配置文件来源卡片美化 */
.config-source-card {
  background: linear-gradient(135deg, #f5f7fa 0%, #e4eaf1 100%);
  border: none;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
}

.config-source-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
  transform: translateY(-2px);
}

.config-source-card h4 {
  color: #333;
  font-size: 16px;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid #e0e0e0;
}

.config-source {
  padding: 15px 0;
}

.config-source-row {
  margin-bottom: 12px;
  display: flex;
  align-items: center;
}

.config-source-row:last-child {
  margin-bottom: 0;
}

.config-source-label {
  display: inline-block;
  width: 120px;
  font-weight: 600;
  color: #555;
  font-size: 14px;
  flex-shrink: 0;
}

.config-source-value {
  display: inline-block;
  font-size: 14px;
}

.source-local {
  color: #1890ff;
  font-weight: 500;
  padding: 4px 8px;
  background: rgba(184, 134, 11, 0.1);
  border-radius: 4px;
}

.source-remote {
  color: #1890ff;
  font-weight: 500;
  padding: 4px 8px;
  background: rgba(184, 134, 11, 0.1);
  border-radius: 4px;
}

.config-file-path,
.nacos-info {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px dashed #e0e0e0;
}

/* Tab 容器美化 */
.tab-container {
  margin-bottom: 24px;
  border-radius: 8px;
  background: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  overflow: hidden;
}

.tab-header {
  display: flex;
  border-bottom: 2px solid #f0f0f0;
  background: #f8f9fa;
}

.tab-item {
  padding: 14px 24px;
  cursor: pointer;
  margin-right: 2px;
  border-bottom: 3px solid transparent;
  color: #666;
  font-weight: 500;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.tab-item::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  width: 0;
  height: 3px;
  background: #1890ff;
  transition: width 0.3s ease;
}

.tab-item:hover {
  color: #1890ff;
  background: rgba(24, 144, 255, 0.05);
}

.tab-item.active {
  color: #1890ff;
  background: white;
  border-bottom: 3px solid #1890ff;
}

.tab-item.active::after {
  width: 100%;
}

.tab-content {
  padding: 20px;
  animation: tabFadeIn 0.3s ease-in-out;
}

@keyframes tabFadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

/* 表单美化 */
.config-form {
  width: 100%;
}

.form-group {
  margin-bottom: 20px;
  padding: 15px;
  background: #fafbfc;
  border-radius: 6px;
  border-left: 3px solid transparent;
  transition: all 0.3s ease;
}

.form-group:hover {
  border-left: 3px solid #1890ff;
  background: #f5f8fa;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-label {
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
  color: #444;
  font-size: 14px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.form-value {
  color: #333;
  font-size: 14px;
  line-height: 1.6;
  word-break: break-all;
  padding: 8px 12px;
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  min-height: 36px;
  display: flex;
  align-items: center;
}

/* 数组和嵌套对象样式优化 */
.array-container {
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  padding: 12px;
}

.array-item {
  margin-left: 0;
  margin-bottom: 8px;
  padding: 8px 12px;
  background: #f8f9fa;
  border-left: 3px solid #1890ff;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.array-item:hover {
  background: #e9ecef;
  transform: translateX(3px);
}

.array-item:last-child {
  margin-bottom: 0;
}

.array-index {
  color: #1890ff;
  font-weight: 600;
  margin-right: 8px;
  min-width: 25px;
  display: inline-block;
}

.array-value {
  color: #333;
  font-size: 13px;
}

.nested-object {
  margin-left: 0;
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  padding: 12px;
}

.nested-item {
  margin-bottom: 8px;
  padding: 8px 12px;
  background: #f8f9fa;
  border-radius: 4px;
  display: flex;
  align-items: flex-start;
  flex-wrap: wrap;
}

.nested-item:hover {
  background: #e9ecef;
}

.nested-item:last-child {
  margin-bottom: 0;
}

.nested-key {
  color: #666;
  font-weight: 600;
  margin-right: 8px;
  min-width: 120px;
  flex-shrink: 0;
}

.nested-value {
  color: #333;
  flex: 1;
  min-width: 200px;
}

/* 刷新按钮美化 */
.refresh-button {
  background: linear-gradient(135deg, #1890ff 0%, #40a9ff 100%);
  color: white;
  border: none;
  padding: 12px 24px;
  font-size: 14px;
  font-weight: 500;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 2px 8px rgba(24, 144, 255, 0.3);
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.refresh-button:hover {
  background: linear-gradient(135deg, #096dd9 0%, #1890ff 100%);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(24, 144, 255, 0.4);
}

.refresh-button:active {
  transform: translateY(0);
}

.refresh-button i {
  margin-right: 8px;
}

/* 加载状态美化 */
.loading-spinner {
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 2px solid rgba(24, 144, 255, 0.3);
  border-radius: 50%;
  border-top-color: #1890ff;
  animation: spin 1s ease-in-out infinite;
  margin-right: 8px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* 响应式设计优化 */
@media (max-width: 768px) {
  .config-source-row {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .config-source-label {
    width: 100%;
    margin-bottom: 4px;
  }
  
  .tab-header {
    flex-direction: column;
  }
  
  .tab-item {
    margin-right: 0;
    border-bottom: 1px solid #e0e0e0;
  }
  
  .form-group {
    padding: 12px;
  }
  
  .nested-key {
    min-width: auto;
    width: 100%;
    margin-bottom: 4px;
  }
  
  .nested-value {
    min-width: auto;
    width: 100%;
  }
}
</style>
<template>
<div class="config-view-container">
    <h3>配置管理 <span v-if="loadingStates.config || loadingStates.cmdline" style="font-size: 14px; color: #666;"><span class="loading-spinner"></span>加载中...</span></h3>
    
    <!-- 配置文件来源显示 -->
    <div class="card config-source-card">
        <h4>配置文件来源</h4>
        <div class="config-source">
            <div class="config-source-row">
                <div class="config-source-label">来源：</div>
                <div class="config-source-value">
                    <span v-if="cmdlineConfig.LocalCfgFile && cmdlineConfig.LocalCfgFile.trim() !== ''" class="source-local">本地配置文件</span>
                    <span v-else class="source-remote">远程配置 (Nacos)</span>
                </div>
            </div>
            <div v-if="cmdlineConfig.LocalCfgFile && cmdlineConfig.LocalCfgFile.trim() !== ''" class="config-file-path">
                <div class="config-source-row">
                    <div class="config-source-label">路径：</div>
                    <div class="config-source-value">{{ cmdlineConfig.LocalCfgFile }}</div>
                </div>
            </div>
            <div v-else class="nacos-info">
                <div class="config-source-row">
                    <div class="config-source-label">Nacos 服务：</div>
                    <div class="config-source-value">{{ cmdlineConfig.NacosUrl || '未配置' }}</div>
                </div>
                <div class="config-source-row">
                    <div class="config-source-label">Nacos 命名空间：</div>
                    <div class="config-source-value">{{ cmdlineConfig.NacosNamespace || '未配置' }}</div>
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
                    <div class="form-group">
                        <label class="form-label">配置文件路径</label>
                        <div class="form-value">{{ cmdlineConfig.LocalCfgFile || '未设置' }}</div>
                    </div>
                    <div class="form-group">
                        <label class="form-label">Nacos 服务地址</label>
                        <div class="form-value">{{ cmdlineConfig.NacosUrl || '未设置' }}</div>
                    </div>
                    <div class="form-group">
                        <label class="form-label">Nacos 用户名</label>
                        <div class="form-value">{{ cmdlineConfig.NacosUsername || '未设置' }}</div>
                    </div>
                    <div class="form-group">
                        <label class="form-label">Nacos 密码</label>
                        <div class="form-value">{{ cmdlineConfig.NacosPassword || '未设置' }}</div>
                    </div>
                    <div class="form-group">
                        <label class="form-label">Nacos 命名空间ID</label>
                        <div class="form-value">{{ cmdlineConfig.NacosNamespace || '' }}</div>
                    </div>
                    <div class="form-group">
                        <label class="form-label">Nacos 组</label>
                        <div class="form-value">{{ cmdlineConfig.NacosGroup || '未设置' }}</div>
                    </div>
                    <div class="form-group">
                        <label class="form-label">Nacos DataId</label>
                        <div class="form-value">{{ cmdlineConfig.NacosDataId || '未设置' }}</div>
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
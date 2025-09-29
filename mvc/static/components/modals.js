// 模态框组件
const TaskDetailModal = {
    template: `
        <div v-if="visible" class="modal-overlay" @click="$emit('close')">
            <div class="modal-content" @click.stop>
                <div class="modal-header">
                    <h4>任务详情</h4>
                    <button class="modal-close" @click="$emit('close')">&times;</button>
                </div>
                <div class="modal-body">
                    <pre v-if="task">{{ JSON.stringify(task, null, 2) }}</pre>
                </div>
            </div>
        </div>
    `,
    props: ['visible', 'task'],
    emits: ['close']
};

const TaskMetricsModal = {
    template: `
        <div v-if="visible" class="modal-overlay" @click="$emit('close')">
            <div class="modal-content" @click.stop>
                <div class="modal-header">
                    <h4>任务指标</h4>
                    <button class="modal-close" @click="$emit('close')">&times;</button>
                </div>
                <div class="modal-body">
                    <div v-if="task">
                        <p><strong>任务名:</strong> {{ task.Name }}</p>
                        <p><strong>Topic:</strong> {{ task.Topic }}</p>
                        <p><strong>消费者组:</strong> {{ task.ConsumerGroup }}</p>
                        <p><strong>表名:</strong> {{ task.TableName }}</p>
                        <p><strong>类型:</strong> {{ task.PrometheusSchema ? '指标' : '日志' }}</p>
                        <hr>
                        <p><em>详细指标数据功能待实现...</em></p>
                    </div>
                </div>
            </div>
        </div>
    `,
    props: ['visible', 'task'],
    emits: ['close']
};

const TaskStatusModal = {
    template: `
        <div v-if="visible" class="modal-overlay" @click="$emit('close')">
            <div class="modal-content" @click.stop>
                <div class="modal-header">
                    <h4>任务运行状态</h4>
                    <button class="modal-close" @click="$emit('close')">&times;</button>
                </div>
                <div class="modal-body">
                    <div v-if="task && taskStatus">
                        <div class="status-grid">
                            <div class="status-item">
                                <label>任务名:</label>
                                <span>{{ task.Name }}</span>
                            </div>
                            <div class="status-item">
                                <label>运行状态:</label>
                                <span :class="['status-badge', getStatusClass(taskStatus)]">
                                    {{ getStatusText(taskStatus) }}
                                </span>
                            </div>
                            <div class="status-item">
                                <label>启动时间:</label>
                                <span>{{ formatTime(taskStatus.startTime) }}</span>
                            </div>
                            <div class="status-item">
                                <label>最后活跃:</label>
                                <span>{{ formatTime(taskStatus.lastActiveTime) }}</span>
                            </div>
                            <div class="status-item">
                                <label>处理消息总数:</label>
                                <span>{{ formatNumber(taskStatus.messagesTotal) }}</span>
                            </div>
                            <div class="status-item">
                                <label>消息处理速率:</label>
                                <span>{{ formatRate(taskStatus.messagesPerSec) }}</span>
                            </div>
                            <div class="status-item">
                                <label>Kafka Lag:</label>
                                <span :class="getLagClass(taskStatus)">{{ formatLag(taskStatus.kafkaLag) }}</span>
                            </div>
                            <div class="status-item">
                                <label>错误次数:</label>
                                <span :class="taskStatus.errorCount > 0 ? 'error-count' : ''">{{ taskStatus.errorCount }}</span>
                            </div>
                            <div v-if="taskStatus.lastError" class="status-item full-width">
                                <label>最后错误:</label>
                                <span class="error-message">{{ taskStatus.lastError }}</span>
                            </div>
                        </div>
                        <div class="status-actions">
                            <button class="btn-small" @click="refreshStatus">刷新状态</button>
                        </div>
                    </div>
                    <div v-else-if="task">
                        <p>正在获取任务状态...</p>
                    </div>
                    <div v-else>
                        <p>无任务信息</p>
                    </div>
                </div>
            </div>
        </div>
    `,
    props: ['visible', 'task'],
    emits: ['close'],
    data() {
        return {
            taskStatus: null,
            refreshInterval: null
        };
    },
    watch: {
        visible(newVal) {
            if (newVal && this.task) {
                this.fetchTaskStatus();
                this.startAutoRefresh();
            } else {
                this.stopAutoRefresh();
            }
        }
    },
    methods: {
        async fetchTaskStatus() {
            if (!this.task) return;
            
            try {
                const response = await fetch(`/api/v1/tasks/${this.task.Name}/status`);
                const data = await response.json();
                
                if (data.retCode === '0000' && data.entity) {
                    this.taskStatus = data.entity;
                } else {
                    console.warn('⚠️ Task status API returned error:', data.retMsg);
                }
            } catch (error) {
                console.error('❌ Error fetching task status:', error);
            }
        },
        
        refreshStatus() {
            this.fetchTaskStatus();
        },
        
        startAutoRefresh() {
            this.stopAutoRefresh();
            this.refreshInterval = setInterval(() => {
                this.fetchTaskStatus();
            }, 5000);
        },
        
        stopAutoRefresh() {
            if (this.refreshInterval) {
                clearInterval(this.refreshInterval);
                this.refreshInterval = null;
            }
        },
        
        getStatusClass(status) {
            if (!status) return 'unknown';
            switch (status.status) {
                case 'running': return 'running';
                case 'stopped': return 'stopped';
                case 'error': return 'error';
                default: return 'unknown';
            }
        },
        
        getStatusText(status) {
            if (!status) return '未知';
            switch (status.status) {
                case 'running': return '运行中';
                case 'stopped': return '已停止';
                case 'error': return '错误';
                default: return '未知';
            }
        },
        
        formatTime(timeStr) {
            if (!timeStr) return '-';
            const date = new Date(timeStr);
            return date.toLocaleString('zh-CN');
        },
        
        formatNumber(num) {
            if (num === undefined || num === null) return '0';
            return num.toLocaleString();
        },
        
        formatRate(rate) {
            if (!rate) return '0/s';
            if (rate >= 1000) {
                return (rate / 1000).toFixed(1) + 'k/s';
            }
            return Math.round(rate) + '/s';
        },
        
        formatLag(lag) {
            if (lag === undefined || lag === null) return '-';
            if (lag >= 1000000) {
                return (lag / 1000000).toFixed(1) + 'M';
            } else if (lag >= 1000) {
                return (lag / 1000).toFixed(1) + 'K';
            }
            return lag.toString();
        },
        
        getLagClass(status) {
            if (!status || status.kafkaLag === undefined) return '';
            const lag = status.kafkaLag;
            if (lag > 10000) return 'lag-high';
            if (lag > 1000) return 'lag-medium';
            return 'lag-low';
        }
    },
    
    beforeUnmount() {
        this.stopAutoRefresh();
    }
};
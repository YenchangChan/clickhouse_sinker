// 任务管理页面组件
const TasksView = {
    template: `
        <div>
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
                    background-color: #f0f0f0;
                    color: #333;
                    padding: 4px 12px;
                    font-size: 12px;
                    margin-right: 4px;
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
                
                .task-table {
                    width: 100%;
                    border-collapse: collapse;
                    font-size: 14px;
                }
                
                .task-table th,
                .task-table td {
                    padding: 12px;
                    text-align: left;
                    border-bottom: 1px solid #f0f0f0;
                }
                
                .task-table th {
                    background-color: #fafafa;
                    font-weight: 600;
                    color: #333;
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
            </style>
            
            <h3>任务管理 <span v-if="loadingStates.tasks" style="font-size: 14px; color: #666;">加载中...</span></h3>
            <div class="card">
                <!-- 筛选和操作区域 -->
                <div class="task-controls">
                    <div class="filter-section">
                        <input 
                            type="text" 
                            v-model="taskFilter.name" 
                            placeholder="按任务名筛选..." 
                            class="filter-input"
                        >
                        <select v-model="taskFilter.type" class="filter-select">
                            <option value="">全部类型</option>
                            <option value="metric">metric</option>
                            <option value="log">log</option>
                        </select>
                        <select v-model="taskFilter.status" class="filter-select">
                            <option value="">全部状态</option>
                            <option value="running">Stable</option>
                            <option value="stopped">Empty</option>
                            <option value="error">Dead</option>
                            <option value="error">Rebalancing</option>
                        </select>
                        <button class="btn-small" @click="clearFilters">清空筛选</button>
                    </div>
                    <div class="action-section">
                        <span class="task-count">共 {{ filteredTasks.length }} / {{ tasks.length }} 个任务</span>
                        <button class="btn" @click="$emit('refresh-tasks')">刷新任务</button>
                    </div>
                </div>

                <table class="task-table">
                        <thead>
                            <tr>
                                <th class="sortable" @click="sortTasks('Name')">
                                    任务名 
                                    <span class="sort-indicator" v-if="sortField === 'Name'">
                                        {{ sortOrder === 'asc' ? '↑' : '↓' }}
                                    </span>
                                </th>
                                <th class="sortable" @click="sortTasks('Cluster')">
                                    集群
                                    <span class="sort-indicator" v-if="sortField === 'Cluster'">
                                        {{ sortOrder === 'asc' ? '↑' : '↓' }}
                                    </span>
                                </th>
                                <th class="sortable" @click="sortTasks('Table')">
                                    表名
                                    <span class="sort-indicator" v-if="sortField === 'Table'">
                                        {{ sortOrder === 'asc' ? '↑' : '↓' }}
                                    </span>
                                </th>
                                <th class="sortable" @click="sortTasks('Topic')">
                                    Topic
                                    <span class="sort-indicator" v-if="sortField === 'Topic'">
                                        {{ sortOrder === 'asc' ? '↑' : '↓' }}
                                    </span>
                                </th>
                                <th class="sortable" @click="sortTasks('ConsumerGroup')">
                                    消费者组
                                    <span class="sort-indicator" v-if="sortField === 'ConsumerGroup'">
                                        {{ sortOrder === 'asc' ? '↑' : '↓' }}
                                    </span>
                                </th>
                                <th class="sortable" @click="sortTasks('Type')">
                                    类型
                                    <span class="sort-indicator" v-if="sortField === 'Type'">
                                        {{ sortOrder === 'asc' ? '↑' : '↓' }}
                                    </span>
                                </th>
                                <th class="sortable" @click="sortTasks('Status')">
                                    状态
                                    <span class="sort-indicator" v-if="sortField === 'Status'">
                                        {{ sortOrder === 'asc' ? '↑' : '↓' }}
                                    </span>
                                </th>
                                <th class="sortable" @click="sortTasks('Lag')">
                                    Kafka Lag
                                    <span class="sort-indicator" v-if="sortField === 'Lag'">
                                        {{ sortOrder === 'asc' ? '↑' : '↓' }}
                                    </span>
                                </th>
                                <th class="sortable" @click="sortTasks('WriteSpeed')">
                                    写入速度
                                    <span class="sort-indicator" v-if="sortField === 'WriteSpeed'">
                                        {{ sortOrder === 'asc' ? '↑' : '↓' }}
                                    </span>
                                </th>
                                <th class="sortable" @click="sortTasks('LastUpdate')">
                                    最后更新
                                    <span class="sort-indicator" v-if="sortField === 'LastUpdate'">
                                        {{ sortOrder === 'asc' ? '↑' : '↓' }}
                                    </span>
                                </th>
                            </tr>
                        </thead>
                        <tbody>
                            <template v-if="filteredTasks.length === 0">
                                <tr>
                                    <td colspan="10" class="empty-state">
                                        {{ tasks.length === 0 ? '暂无任务数据' : '没有符合筛选条件的任务' }}
                                    </td>
                                </tr>
                            </template>
                            <tr v-else v-for="task in paginatedTasks" :key="task.Name">
                                <td class="task-name">{{ task.Name }}</td>
                                <td>{{ task.Cluster }}</td>
                                <td>{{ task.Table }}</td>
                                <td>{{ task.Topic }}</td>
                                <td>{{ task.ConsumerGroup }}</td>
                                <td>
                                    <span :class="['task-type', task.Type]">
                                        {{ task.Type }}
                                    </span>
                                </td>
                                <td>
                                    <span :class="['status-badge', getStatusClass(task.Status)]">
                                        {{ getStatusText(task.Status) }}
                                    </span>
                                </td>
                                <td :class="getLagClass(task.Lag)">
                                    {{ formatNumber(task.Lag) }}
                                </td>
                                <td>{{ task.WriteSpeed || '-' }}</td>
                                <td>{{ formatDateTime(task.LastUpdate) || '-' }}</td>
                            </tr>
                        </tbody>
                    </table>

                    <!-- 分页控件 -->
                    <div class="pagination" v-if="filteredTasks.length > 0 && totalPages > 1">
                        <button 
                            class="btn-small" 
                            @click="currentPage = 1" 
                            :disabled="currentPage === 1"
                        >
                            首页
                        </button>
                        <button 
                            class="btn-small" 
                            @click="currentPage--" 
                            :disabled="currentPage === 1"
                        >
                            上一页
                        </button>
                        <span class="page-info">
                            第 {{ currentPage }} / {{ totalPages }} 页
                        </span>
                        <button 
                            class="btn-small" 
                            @click="currentPage++" 
                            :disabled="currentPage === totalPages"
                        >
                            下一页
                        </button>
                        <button 
                            class="btn-small" 
                            @click="currentPage = totalPages" 
                            :disabled="currentPage === totalPages"
                        >
                            末页
                        </button>
                        <select v-model="pageSize" class="page-size-select" @change="currentPage = 1">
                            <option value="10">10条/页</option>
                            <option value="20">20条/页</option>
                            <option value="50">50条/页</option>
                            <option value="100">100条/页</option>
                        </select>
                    </div>
                </div>
            </div>
        </div>
    `,
    props: ['tasks', 'loadingStates'],
    emits: ['refresh-tasks'],
    data() {
        return {
            taskFilter: {
                name: '',
                type: '',
                status: ''
            },
            sortField: '',
            sortOrder: 'asc',
            currentPage: 1,
            pageSize: 20
        };
    },
    computed: {
        filteredTasks() {
            let filtered = this.tasks;

            // 按任务名筛选
            if (this.taskFilter.name) {
                filtered = filtered.filter(task =>
                    task.Name.toLowerCase().includes(this.taskFilter.name.toLowerCase())
                );
            }

            // 按类型筛选
            if (this.taskFilter.type) {
                filtered = filtered.filter(task => task.Type === this.taskFilter.type);
            }

            // 按状态筛选
            if (this.taskFilter.status) {
                filtered = filtered.filter(task => task.Status === this.taskFilter.status);
            }

            // 排序
            if (this.sortField) {
                filtered.sort((a, b) => {
                    let aVal = a[this.sortField];
                    let bVal = b[this.sortField];

                    // 处理布尔值排序（类型字段）
                    if (this.sortField === 'PrometheusSchema') {
                        aVal = aVal ? 1 : 0;
                        bVal = bVal ? 1 : 0;
                    }

                    // 字符串排序
                    if (typeof aVal === 'string') {
                        aVal = aVal.toLowerCase();
                        bVal = bVal.toLowerCase();
                    }

                    if (aVal < bVal) return this.sortOrder === 'asc' ? -1 : 1;
                    if (aVal > bVal) return this.sortOrder === 'asc' ? 1 : -1;
                    return 0;
                });
            }

            return filtered;
        },

        totalPages() {
            return Math.ceil(this.filteredTasks.length / this.pageSize);
        },

        paginatedTasks() {
            const start = (this.currentPage - 1) * this.pageSize;
            const end = start + parseInt(this.pageSize);
            return this.filteredTasks.slice(start, end);
        }
    },
    methods: {
        clearFilters() {
            this.taskFilter.name = '';
            this.taskFilter.type = '';
            this.taskFilter.status = '';
            this.currentPage = 1;
        },

        getStatusClass(status) {
            switch (status) {
                case 'running':
                    return 'status-running';
                case 'stopped':
                    return 'status-stopped';
                case 'error':
                    return 'status-error';
                default:
                    return '';
            }
        },

        getStatusText(status) {
            switch (status) {
                case 'running':
                    return '运行中';
                case 'stopped':
                    return '已停止';
                case 'error':
                    return '错误';
                default:
                    return status;
            }
        },

        getLagClass(lag) {
            if (lag > 10000) {
                return 'lag-critical';
            } else if (lag > 1000) {
                return 'lag-warning';
            }
            return '';
        },

        formatNumber(num) {
            return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
        },
        
        formatDateTime(timestamp) {
            if (!timestamp) return '';
            const date = new Date(timestamp);
            const year = date.getFullYear();
            const month = String(date.getMonth() + 1).padStart(2, '0');
            const day = String(date.getDate()).padStart(2, '0');
            const hours = String(date.getHours()).padStart(2, '0');
            const minutes = String(date.getMinutes()).padStart(2, '0');
            const seconds = String(date.getSeconds()).padStart(2, '0');
            return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
        },

        sortTasks(field) {
            if (this.sortField === field) {
                this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc';
            } else {
                this.sortField = field;
                this.sortOrder = 'asc';
            }
            this.currentPage = 1;
        }
    }
};
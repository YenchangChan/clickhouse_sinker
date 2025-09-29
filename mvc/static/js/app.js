// 主应用逻辑
console.log('🚀 Starting ClickHouse Sinker Web UI...');

window.addEventListener('load', function () {
    if (typeof Vue === 'undefined') {
        document.getElementById('app').innerHTML = '<div class="error-state">❌ Vue.js 加载失败</div>';
        return;
    }

    const { createApp } = Vue;

    createApp({
        components: {
            'overview-view': OverviewView,
            'tasks-view': TasksView,
            'config-view': ConfigView,
            'debug-view': DebugView,
            'task-detail-modal': TaskDetailModal,
            'task-metrics-modal': TaskMetricsModal,
            'task-status-modal': TaskStatusModal
        },
        
        data() {
            return {
                activeTab: 'overview',
                procInfo: {},
                tasks: [],
                config: {},
                cmdlineConfig: {},
                debugInfo: '',
                loading: true,
                loadingStates: {
                    overview: false,
                    tasks: false,
                    config: false,
                    cmdline: false
                },
                selectedTask: null,
                taskDetailVisible: false,
                taskMetricsVisible: false,
                taskStatusVisible: false
            };
        },

        computed: {
            currentView() {
                const viewMap = {
                    'overview': 'overview-view',
                    'tasks': 'tasks-view',
                    'config': 'config-view',
                    'debug': 'debug-view'
                };
                return viewMap[this.activeTab] || 'overview-view';
            }
        },

        mounted() {
            console.log('✅ Vue app mounted successfully!');
            this.loadData();

            // 定时刷新概览数据
            setInterval(() => {
                if (this.activeTab === 'overview') {
                    this.fetchProcInfo();
                }
            }, 30000);
        },

        methods: {
            setActiveTab(tab) {
                console.log('📍 Switching to tab:', tab);
                this.activeTab = tab;
                
                // 根据切换的标签页调用相应的API
                switch (tab) {
                    case 'overview':
                        this.fetchProcInfo();
                        break;
                    case 'tasks':
                        this.fetchTasks();
                        break;
                    case 'config':
                        this.fetchConfig();
                        this.fetchCmdline();
                        break;
                    case 'debug':
                        break;
                }
            },

            async loadData() {
                console.log('📡 Loading initial data...');
                this.loading = true;

                try {
                    await Promise.all([
                        this.fetchProcInfo(),
                        this.fetchConfig(),
                        this.fetchCmdline()
                    ]);
                    console.log('✅ All data loaded successfully');
                } catch (error) {
                    console.error('❌ Error loading data:', error);
                } finally {
                    this.loading = false;
                }
            },

            async fetchProcInfo() {
                this.loadingStates.overview = true;
                try {
                    console.log('📊 Fetching proc info...');
                    const response = await fetch('/api/v1/metrics/procinfo');
                    const data = await response.json();

                    if (data.retCode === '0000' && data.entity) {
                        this.procInfo = data.entity;
                        console.log('✅ Proc info updated:', data.entity);
                    } else {
                        console.warn('⚠️ Proc info API returned error:', data.retMsg);
                    }
                } catch (error) {
                    console.error('❌ Error fetching proc info:', error);
                } finally {
                    this.loadingStates.overview = false;
                }
            },

            async fetchTasks() {
                this.loadingStates.tasks = true;
                try {
                    console.log('📋 Fetching tasks...');
                    const response = await fetch('/api/v1/tasks');
                    const data = await response.json();

                    if (data.retCode === '0000' && data.entity) {
                        this.tasks = data.entity.Tasks || [];
                        console.log('✅ Tasks updated:', this.tasks.length, 'tasks');
                    } else {
                        console.warn('⚠️ Tasks API returned error:', data.retMsg);
                        this.tasks = [];
                    }
                } catch (error) {
                    console.error('❌ Error fetching tasks:', error);
                    this.tasks = [];
                } finally {
                    this.loadingStates.tasks = false;
                }
            },

            async fetchConfig() {
                this.loadingStates.config = true;
                try {
                    console.log('⚙️ Fetching config...');
                    const response = await fetch('/api/v1/config');
                    const data = await response.json();

                    if (data.retCode === '0000' && data.entity) {
                        this.config = data.entity;
                        console.log('✅ Config updated');
                    } else {
                        console.warn('⚠️ Config API returned error:', data.retMsg);
                    }
                } catch (error) {
                    console.error('❌ Error fetching config:', error);
                } finally {
                    this.loadingStates.config = false;
                }
            },

            async fetchCmdline() {
                this.loadingStates.cmdline = true;
                try {
                    console.log('💻 Fetching cmdline config...');
                    const response = await fetch('/api/v1/cmdline');
                    const data = await response.json();

                    if (data.retCode === '0000' && data.entity) {
                        this.cmdlineConfig = data.entity;
                        console.log('✅ Cmdline config updated');
                    } else {
                        console.warn('⚠️ Cmdline API returned error:', data.retMsg);
                    }
                } catch (error) {
                    console.error('❌ Error fetching cmdline:', error);
                } finally {
                    this.loadingStates.cmdline = false;
                }
            },

            refreshConfig() {
                this.fetchConfig();
                this.fetchCmdline();
            },

            showTaskDetail(task) {
                this.selectedTask = task;
                this.taskDetailVisible = true;
            },

            showTaskMetrics(task) {
                this.selectedTask = task;
                this.taskMetricsVisible = true;
            },

            showTaskStatus(task) {
                this.selectedTask = task;
                this.taskStatusVisible = true;
            },

            updateDebugInfo(info) {
                this.debugInfo = info;
            }
        }
    }).mount('#app');
});
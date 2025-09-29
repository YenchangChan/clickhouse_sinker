// 调试工具页面组件
const DebugView = {
    template: `
        <div class="debug-view-container">
            <h3>调试工具</h3>
            
            <!-- 添加样式 -->
            <style>
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
                
                .dark-textarea {
                    width: 100%;
                    height: 400px;
                    background-color: #2d2d2d;
                    color: #e0e0e0;
                    border: 1px solid #444;
                    border-radius: 4px;
                    padding: 15px;
                    font-family: 'Courier New', Courier, monospace;
                    font-size: 14px;
                    line-height: 1.5;
                    resize: vertical;
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
                    </div>
                    
                    <!-- Tab 内容 - Heap -->
                    <div v-if="activeTab === 'heap'" class="tab-content">
                        <div class="dark-textarea" v-html="formatText(heapInfo)"></div>
                    </div>
                    
                    <!-- Tab 内容 - Goroutine -->
                    <div v-if="activeTab === 'goroutine'" class="tab-content">
                        <div class="dark-textarea" v-html="formatText(goroutineInfo)"></div>
                    </div>
                    
                    <!-- Tab 内容 - Metrics -->
                    <div v-if="activeTab === 'metrics'" class="tab-content">
                        <div class="dark-textarea" v-html="formatText(metricsInfo)"></div>
                    </div>
                </div>
            </div>
        </div>
    `,
    props: ['loadingStates', 'debugInfo'],
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
                const response = await fetch('/debug/pprof/heap?debug=1');
                const text = await response.text();
                this.heapInfo = text;
            } catch (error) {
                this.heapInfo = `❌ 获取Heap信息失败: ${error.message}`;
            }
        },
        
        async fetchGoroutineInfo() {
            try {
                this.goroutineInfo = '加载中...';
                const response = await fetch('/debug/pprof/goroutine?debug=1');
                const text = await response.text();
                this.goroutineInfo = text;
            } catch (error) {
                this.goroutineInfo = `❌ 获取Goroutine信息失败: ${error.message}`;
            }
        },
        
        async fetchMetricsInfo() {
            try {
                this.metricsInfo = '加载中...';
                const response = await fetch('/metrics');
                const text = await response.text();
                this.metricsInfo = text;
            } catch (error) {
                this.metricsInfo = `❌ 获取指标信息失败: ${error.message}`;
            }
        },
        
        formatText(text) {
            // 简单的文本格式化，将替换为<br>，将空格替换为&nbsp;
            return text ? text.replace(/\n/g, '<br>').replace(/ /g, '&nbsp;') : '';
        }
    },
    mounted() {
        // 初始化时自动加载当前选中的tab数据
        if (this.activeTab === 'heap') {
            this.fetchHeapInfo();
        }
    }
};
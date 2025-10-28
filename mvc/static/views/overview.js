// 概览页面组件
const OverviewView = {
    template: `
        <div>
            <h3>系统概览 <span v-if="loadingStates.overview" style="font-size: 14px; color: #666;">刷新中...</span></h3>
            <div class="metric-grid">
                <div class="metric-card">
                    <div class="metric-value">{{ procInfo.Version || 'N/A' }}</div>
                    <div class="metric-label">Sinker版本</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value">{{ procInfo.Goroutines || 0 }}</div>
                    <div class="metric-label">Goroutines</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value">{{ (procInfo.CPU || 0).toFixed(1) }}%</div>
                    <div class="metric-label">CPU使用率</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value">{{ formatMemory(procInfo.Memory || 0) }}</div>
                    <div class="metric-label">内存使用</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value">{{ procInfo.Tasks || 0 }}</div>
                    <div class="metric-label">任务个数</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value">{{ procInfo.RecordPoolSize || 0 }}</div>
                    <div class="metric-label">消息池大小</div>
                </div>
            </div>
            
            <div class="card">
                <h4>系统状态</h4>
                <p><span class="status-indicator"></span>服务运行正常</p>
                <p>运行时间: <strong>{{ formatUptime(procInfo.Uptime || 0) }}</strong></p>
                <p>启动时间: <strong>{{ formatStartTime(procInfo.StartTime || 0) }}</strong></p>
                <p>Go版本: <strong>{{ procInfo.GoVersion || 'N/A' }}</strong></p>
                <p>构建时间: <strong>{{ procInfo.BuildTime || 'N/A' }}</strong></p>
                <p>Commit: <strong>{{ formatCommit(procInfo.Commit) || 'N/A' }}</strong></p>
            </div>
        </div>
    `,
    props: ['procInfo', 'loadingStates'],
    methods: {
        formatMemory(bytes) {
            if (!bytes || bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        },
        
        formatUptime(seconds) {
            if (!seconds || seconds === 0) return '0s';
            const days = Math.floor(seconds / 86400);
            const hours = Math.floor((seconds % 86400) / 3600);
            const minutes = Math.floor((seconds % 3600) / 60);
            const secs = seconds % 60;

            let result = '';
            if (days > 0) result += `${days}d `;
            if (hours > 0) result += `${hours}h `;
            if (minutes > 0) result += `${minutes}m `;
            if (secs > 0) result += `${secs}s`;

            return result.trim() || '0s';
        },

        formatStartTime(timestamp) {
            if (!timestamp || timestamp === 0) return 'N/A';
            const date = new Date(timestamp * 1000);
            return date.toLocaleString();
        },

        formatCommit(commit) {
            if (!commit || commit === '') return 'N/A';
            return commit.length > 8 ? commit.substring(0, 8) : commit;
        }
    }
};
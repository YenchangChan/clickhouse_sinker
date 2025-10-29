// src/services/api.js
import axios from 'axios'

const api = axios.create({})

export default {
  // 获取系统概览信息
  getOverview() {
    return api.get('/api/v1/metrics/procinfo')
  },
  
  // 获取任务列表
  getTasks() {
    return api.get('/api/v1/tasks')
  },
  
  // 获取配置信息
  getConfig() {
    return api.get('/api/v1/config')
  },
  
  // 获取命令行参数
  getCmdline() {
    return api.get('/api/v1/cmdline')
  },
  
  // 刷新任务
  refreshTasks() {
    return api.post('/api/v1/tasks/refresh')
  },
  
  // 刷新配置
  refreshConfig() {
    return api.post('/api/v1/config/refresh')
  },
  
  // Debug相关接口
  // 获取Heap信息
  getHeapInfo() {
    return api.get('debug/pprof/heap?debug=1', {
      responseType: 'text'
    });
  },
  
  // 获取Goroutine信息
  getGoroutineInfo() {
    return api.get('debug/pprof/goroutine?debug=1', {
      responseType: 'text'
    });
  },
  
  // 获取Metrics信息
  getMetricsInfo() {
    return api.get('metrics', {
      responseType: 'text'
    });
  },

  // 获取日志
  getLog(from = 0, error = undefined) {
    const params = { from };
    if (error !== undefined) {
      params.error = error;
    }
    return api.get('/api/v1/log', {
      params
    })
  },

  // 获取任务配置
  getTaskConfig(taskName) {
    return api.get(`/api/v1/task/${taskName}`)
  },

  // 获取任务dbkey
  getTaskDbKey(taskName) {
    return api.get(`/api/v1/dbkey/${taskName}`)
  }
}
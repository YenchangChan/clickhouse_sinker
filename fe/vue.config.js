// vue.config.js
module.exports = {
  publicPath: './',
  outputDir: '../mvc/static', // 输出到你的Go项目的静态文件目录
  assetsDir: 'assets',
  devServer: {
    proxy: {
      '/api': {
        target: 'http://192.168.101.95:20088', // 你的后端服务地址
        changeOrigin: true
      },
      '/debug': {
        target: 'http://192.168.101.95:20088', // 你的后端服务地址
        changeOrigin: true
      },
      '/metrics': {
        target: 'http://192.168.101.95:20088', // 你的后端服务地址
        changeOrigin: true
      }
    }
  }
}
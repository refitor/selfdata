module.exports = {
    // 将构建好的文件输出到哪里（或者说将编译的文件）
    outputDir: '../datas/web',

    // 放置静态资源的地方 (js/css/img/font/...)
    assetsDir: 'assets',

    devServer: {
        overlay: { // 让浏览器 overlay 同时显示警告和错误
            warnings: true,
            errors: true
        },
        port: 8000, // 端口号
        https: false, // https:{type:Boolean}
        open: false, //配置自动启动浏览器
        hot: true, // 热更新
        proxy: { //配置多个跨域
        "/api": {
            target: 'http://localhost:7351',
            changeOrigin: false,
            secure: false,
        }
      }
    },
    // 打包时不生成.map文件
    productionSourceMap: false,
  }
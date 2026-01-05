<template>
  <footer class="py-16 border-t border-slate-800 bg-slate-900">
    <div class="max-w-7xl mx-auto px-4">
      <div class="grid grid-cols-2 md:grid-cols-4 gap-12 mb-12">
        <div class="col-span-2">
          <div class="flex items-center gap-2 mb-6">
            <div class="w-8 h-8 bg-cyan-500 rounded-lg flex items-center justify-center font-bold text-slate-900">M</div>
            <span class="text-xl font-bold text-white">BotMatrix</span>
          </div>
          <p class="text-slate-500 max-w-sm leading-relaxed">
            AI 原生智能体操作系统。我们致力于通过 Global Agent Mesh 协议打破企业数据孤岛，构建可信、安全、高效的数字员工协作生态。
          </p>
        </div>
        <div>
          <h4 class="font-bold mb-6 text-white">资源</h4>
          <ul class="space-y-4 text-slate-400 text-sm">
            <li><router-link to="/docs" class="hover:text-cyan-400 transition-colors">使用文档</router-link></li>
            <li><a href="#" class="hover:text-cyan-400 transition-colors">API 参考</a></li>
            <li><a href="#" class="hover:text-cyan-400 transition-colors">插件市场</a></li>
            <li><router-link to="/pricing" class="hover:text-cyan-400 transition-colors">版本对比</router-link></li>
          </ul>
        </div>
        <div>
          <h4 class="font-bold mb-6 text-white">关于</h4>
          <ul class="space-y-4 text-slate-400 text-sm">
            <li><router-link to="/about" class="hover:text-cyan-400 transition-colors">项目简介</router-link></li>
            <li><a href="https://github.com/changliaotong/BotMatrix" target="_blank" class="hover:text-cyan-400 transition-colors">GitHub 源码</a></li>
            <li><a href="#" class="hover:text-cyan-400 transition-colors">加入我们</a></li>
            <li><a href="#" class="hover:text-cyan-400 transition-colors">隐私政策</a></li>
          </ul>
        </div>
      </div>
      <div class="pt-8 border-t border-slate-800 flex flex-col md:flex-row justify-between items-center gap-6 text-slate-500 text-sm">
        <p>© 2026 BotMatrix Open Source Project. All rights reserved.</p>
        <div class="flex gap-8">
          <span>Made with ❤️ for Developers</span>
        </div>
      </div>
    </div>
  </footer>
</template>

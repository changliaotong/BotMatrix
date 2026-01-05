<template>
  <div class="min-h-screen bg-slate-900 text-white selection:bg-cyan-500/30">
    <PortalHeader />

    <!-- Hero Section -->
    <header class="pt-32 pb-16 px-4 relative overflow-hidden">
      <div class="absolute top-0 left-1/2 -translate-x-1/2 w-full h-full bg-[radial-gradient(circle_at_center,rgba(6,182,212,0.1)_0%,transparent_70%)] -z-10"></div>
      <div class="max-w-7xl mx-auto text-center">
        <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-cyan-500/10 border border-cyan-500/20 text-cyan-400 text-xs font-bold mb-6 animate-fade-in">
          <Sparkles class="w-3 h-3" />
          v2.0 焕新发布，支持更多 AI 模型
        </div>
        <h1 class="text-6xl md:text-8xl font-black mb-8 tracking-tighter leading-none">
          构建你的 <br/>
          <span class="bg-clip-text text-transparent bg-gradient-to-r from-cyan-400 via-blue-500 to-purple-600 animate-gradient-x">
            企业机器人矩阵
          </span>
        </h1>
        <p class="text-xl md:text-2xl text-slate-400 mb-12 max-w-3xl mx-auto font-light leading-relaxed">
          从简单的消息响应，到具备专业技能的机器人，再到自主工作的数字员工。
          BotMatrix 为您提供一站式的智能自动化管理方案。
        </p>
        <div class="flex flex-col sm:flex-row gap-4 justify-center items-center">
          <router-link :to="authStore.isAuthenticated ? '/console' : '/login'" class="w-full sm:w-auto px-8 py-4 bg-cyan-500 hover:bg-cyan-400 text-slate-900 rounded-xl text-lg font-bold transition-all shadow-lg shadow-cyan-500/25 flex items-center justify-center gap-2">
            立即部署 <ArrowRight class="w-5 h-5" />
          </router-link>
          <a href="https://github.com/changliaotong/BotMatrix" target="_blank" class="w-full sm:w-auto px-8 py-4 bg-slate-800 hover:bg-slate-700 text-white rounded-xl text-lg font-bold transition-all border border-slate-700 flex items-center justify-center gap-2">
            <Github class="w-5 h-5" /> GitHub 源码
          </a>
        </div>

        <!-- Quick Stats -->
        <div class="mt-20 grid grid-cols-2 md:grid-cols-4 gap-8 max-w-4xl mx-auto py-8 border-y border-slate-800/50">
          <div v-for="stat in stats" :key="stat.label" class="text-center">
            <div class="text-3xl font-bold text-white mb-1">{{ stat.value }}</div>
            <div class="text-sm text-slate-500 uppercase tracking-widest font-medium">{{ stat.label }}</div>
          </div>
        </div>
      </div>
    </header>

    <!-- Core Concepts Section -->
    <section class="py-24 bg-slate-900 relative">
      <div class="max-w-7xl mx-auto px-4">
        <div class="text-center mb-16">
          <h2 class="text-3xl md:text-5xl font-bold mb-6 text-white">从机器人到数字员工的进化</h2>
          <p class="text-slate-400 max-w-2xl mx-auto text-lg">
            通过 BotMatrix，您可以轻松管理大规模机器人集群，并逐步将它们培养成为企业创造价值的数字员工。
          </p>
        </div>

        <div class="grid grid-cols-1 lg:grid-cols-3 gap-8 relative">
          <!-- Connection Lines (Desktop) -->
          <div class="hidden lg:block absolute top-1/2 left-[30%] right-[30%] h-px bg-gradient-to-r from-cyan-500/50 via-blue-500/50 to-purple-500/50 -z-10"></div>

          <!-- Concept 1: Robot Matrix -->
          <div class="bg-slate-800/40 border border-slate-700 p-10 rounded-[3rem] hover:border-cyan-500/50 transition-all group">
            <div class="w-16 h-16 bg-cyan-500/10 text-cyan-500 rounded-2xl flex items-center justify-center mb-8 group-hover:scale-110 transition-transform">
              <Layers class="w-8 h-8" />
            </div>
            <h3 class="text-2xl font-bold mb-4 text-white">机器人矩阵</h3>
            <p class="text-slate-400 leading-relaxed mb-6">
              <b>基础设施层</b>：像管理服务器集群一样管理你的机器人。支持成千上万个机器人的统一调度、状态监控与分布式部署。
            </p>
            <ul class="space-y-3 text-sm text-slate-500 font-medium">
              <li class="flex items-center gap-2"><div class="w-1.5 h-1.5 rounded-full bg-cyan-500"></div> 多平台接入支持</li>
              <li class="flex items-center gap-2"><div class="w-1.5 h-1.5 rounded-full bg-cyan-500"></div> 分布式节点扩容</li>
              <li class="flex items-center gap-2"><div class="w-1.5 h-1.5 rounded-full bg-cyan-500"></div> 统一后台管控</li>
            </ul>
          </div>

          <!-- Concept 2: Skill Robots -->
          <div class="bg-slate-800/40 border border-slate-700 p-10 rounded-[3rem] hover:border-blue-500/50 transition-all group lg:translate-y-12">
            <div class="w-16 h-16 bg-blue-500/10 text-blue-500 rounded-2xl flex items-center justify-center mb-8 group-hover:scale-110 transition-transform">
              <Zap class="w-8 h-8" />
            </div>
            <h3 class="text-2xl font-bold mb-4 text-white">技能机器人</h3>
            <p class="text-slate-400 leading-relaxed mb-6">
              <b>插件化能力</b>：通过加载不同的功能插件，赋予机器人专业技能。无论是社群管理、安全防护还是业务自动化，随心组合。
            </p>
            <ul class="space-y-3 text-sm text-slate-500 font-medium">
              <li class="flex items-center gap-2"><div class="w-1.5 h-1.5 rounded-full bg-blue-500"></div> 丰富的插件生态</li>
              <li class="flex items-center gap-2"><div class="w-1.5 h-1.5 rounded-full bg-blue-500"></div> 灵活的功能配置</li>
              <li class="flex items-center gap-2"><div class="w-1.5 h-1.5 rounded-full bg-blue-500"></div> 秒级技能切换</li>
            </ul>
          </div>

          <!-- Concept 3: Digital Employees -->
          <div class="bg-slate-800/40 border border-slate-700 p-10 rounded-[3rem] hover:border-purple-500/50 transition-all group">
            <div class="w-16 h-16 bg-purple-500/10 text-purple-500 rounded-2xl flex items-center justify-center mb-8 group-hover:scale-110 transition-transform">
              <Users class="w-8 h-8" />
            </div>
            <h3 class="text-2xl font-bold mb-4 text-white">数字员工</h3>
            <p class="text-slate-400 leading-relaxed mb-6">
              <b>智能化终态</b>：为机器人赋予身份、职位与 KPI。它们不再是简单的工具，而是能够自主思考、解决问题并达成目标的虚拟雇员。
            </p>
            <ul class="space-y-3 text-sm text-slate-500 font-medium">
              <li class="flex items-center gap-2"><div class="w-1.5 h-1.5 rounded-full bg-purple-500"></div> 长期记忆与认知</li>
              <li class="flex items-center gap-2"><div class="w-1.5 h-1.5 rounded-full bg-purple-500"></div> 自主业务闭环</li>
              <li class="flex items-center gap-2"><div class="w-1.5 h-1.5 rounded-full bg-purple-500"></div> 绩效考核体系</li>
            </ul>
          </div>
        </div>
      </div>
    </section>

    <!-- Official Bots Showcase -->
    <section class="py-24 bg-slate-900">
      <div class="max-w-7xl mx-auto px-4">
        <div class="flex flex-col md:flex-row justify-between items-end mb-12 gap-6">
          <div>
            <h2 class="text-3xl md:text-4xl font-bold mb-4">官方机器人展示</h2>
            <p class="text-slate-400 max-w-xl">我们基于 BotMatrix 内核开发了一系列功能强大的官方机器人，开箱即用。</p>
          </div>
          <router-link to="/bots/early-meow" class="text-cyan-400 font-bold hover:text-cyan-300 transition-colors flex items-center gap-2">
            查看全部 <ChevronRight class="w-5 h-5" />
          </router-link>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          <!-- Early Meow Card -->
          <div class="group relative bg-slate-800/50 border border-slate-700 rounded-3xl overflow-hidden hover:border-pink-500/50 transition-all duration-500">
            <div class="absolute inset-0 bg-gradient-to-br from-pink-500/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity"></div>
            <div class="p-8 relative">
              <div class="w-16 h-16 bg-pink-500 rounded-2xl flex items-center justify-center mb-6 shadow-lg shadow-pink-500/20 group-hover:scale-110 transition-transform duration-500">
                <Cat class="w-10 h-10 text-white" />
              </div>
              <h3 class="text-2xl font-bold mb-3">早喵机器人</h3>
              <p class="text-slate-400 mb-6 line-clamp-2">专为社群管理与温馨陪伴设计的智能机器人，支持天气播报、入群审核及 AI 陪聊。</p>
              <router-link to="/bots/early-meow" class="inline-flex items-center gap-2 text-pink-400 font-bold group-hover:gap-3 transition-all">
                了解更多 <ArrowRight class="w-4 h-4" />
              </router-link>
            </div>
          </div>

          <!-- Nexus Guard Card -->
          <div class="group relative bg-slate-800/50 border border-slate-700 rounded-3xl overflow-hidden hover:border-blue-500/50 transition-all duration-500">
            <div class="absolute inset-0 bg-gradient-to-br from-blue-500/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity"></div>
            <div class="p-8 relative">
              <div class="w-16 h-16 bg-blue-600 rounded-2xl flex items-center justify-center mb-6 shadow-lg shadow-blue-500/20 group-hover:scale-110 transition-transform duration-500">
                <ShieldAlert class="w-10 h-10 text-white" />
              </div>
              <h3 class="text-2xl font-bold mb-3">Nexus Guard</h3>
              <p class="text-slate-400 mb-6 line-clamp-2">企业级安全防护机器人，提供防红检测、群监控及敏感词过滤等安全核心功能。</p>
              <router-link to="/bots/nexus-guard" class="inline-flex items-center gap-2 text-blue-400 font-bold group-hover:gap-3 transition-all">
                了解更多 <ArrowRight class="w-4 h-4" />
              </router-link>
            </div>
          </div>

          <!-- Digital Employee Card (The Ultimate State) -->
          <div class="group relative bg-slate-800/50 border border-slate-700 rounded-3xl overflow-hidden hover:border-purple-500/50 transition-all duration-500 lg:scale-105 lg:z-10 shadow-2xl shadow-purple-500/10">
            <div class="absolute inset-0 bg-gradient-to-br from-purple-500/20 via-blue-500/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity"></div>
            <div class="p-8 relative">
              <div class="inline-flex items-center gap-2 px-2 py-1 rounded-md bg-purple-500/20 text-purple-400 text-[10px] font-bold mb-4 uppercase tracking-widest">Ultimate State</div>
              <div class="w-16 h-16 bg-gradient-to-tr from-purple-600 to-blue-600 rounded-2xl flex items-center justify-center mb-6 shadow-lg shadow-purple-500/40 group-hover:rotate-[360deg] transition-transform duration-1000">
                <BrainCircuit class="w-10 h-10 text-white" />
              </div>
              <h3 class="text-2xl font-bold mb-3 bg-clip-text text-transparent bg-gradient-to-r from-white to-purple-400">数字员工 (AI Worker)</h3>
              <p class="text-slate-400 mb-6">矩阵革命的终极形态。不再是简单的回复工具，而是具备自主意识、绩效目标与生产关系的数字化雇员。</p>
              <router-link to="/bots/digital-employee" class="inline-flex items-center gap-2 text-purple-400 font-bold group-hover:gap-3 transition-all">
                探索 AI 未来 <ArrowRight class="w-4 h-4" />
              </router-link>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Features Grid -->
    <section id="features" class="py-24 bg-slate-800/30 relative">
      <div class="max-w-7xl mx-auto px-4">
        <div class="text-center mb-20">
          <h2 class="text-3xl md:text-5xl font-bold mb-6">核心技术栈：支撑矩阵运转</h2>
          <p class="text-slate-400 max-w-2xl mx-auto">基于现代化架构，为每一位数字员工提供稳健的底层支撑。</p>
        </div>
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-12">
          <div v-for="feature in features" :key="feature.title" class="relative group">
            <div class="absolute -inset-4 bg-gradient-to-r from-cyan-500/10 to-blue-500/10 rounded-2xl opacity-0 group-hover:opacity-100 transition-opacity"></div>
            <div class="relative">
              <div :class="['w-14 h-14 rounded-2xl flex items-center justify-center mb-6 transition-all group-hover:scale-110', feature.bg, feature.color]">
                <component :is="feature.icon" class="w-7 h-7" />
              </div>
              <h3 class="text-2xl font-bold mb-4">{{ feature.title }}</h3>
              <p class="text-slate-400 leading-relaxed text-sm">{{ feature.desc }}</p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Global Agent Mesh Visualization -->
    <section class="py-24 bg-slate-950 overflow-hidden relative">
      <div class="absolute inset-0 bg-[radial-gradient(circle_at_center,rgba(168,85,247,0.05)_0%,transparent_70%)]"></div>
      <div class="max-w-7xl mx-auto px-4">
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-24 items-center">
          <div class="relative order-2 lg:order-1">
            <div class="aspect-square bg-slate-900 rounded-[3rem] border border-slate-800 p-12 flex items-center justify-center relative overflow-hidden">
               <div class="absolute inset-0 opacity-10">
                 <div class="absolute inset-0 bg-[grid-slate-700] [mask-image:radial-gradient(white,transparent)]"></div>
               </div>
               <!-- Animated Mesh Nodes -->
               <div class="relative z-10 w-full h-full">
                 <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-32 h-32 bg-purple-600/20 rounded-full blur-2xl animate-pulse"></div>
                 <div class="absolute top-0 left-1/4 w-12 h-12 bg-slate-800 border border-slate-700 rounded-xl flex items-center justify-center shadow-xl animate-bounce-slow">
                   <Building2 class="w-6 h-6 text-cyan-400" />
                 </div>
                 <div class="absolute bottom-1/4 right-0 w-12 h-12 bg-slate-800 border border-slate-700 rounded-xl flex items-center justify-center shadow-xl animate-bounce-slow" style="animation-delay: 1s">
                   <Building2 class="w-6 h-6 text-purple-400" />
                 </div>
                 <div class="absolute top-1/2 right-1/4 w-12 h-12 bg-slate-800 border border-slate-700 rounded-xl flex items-center justify-center shadow-xl animate-bounce-slow" style="animation-delay: 2s">
                   <Globe class="w-6 h-6 text-blue-400" />
                 </div>
                 <!-- Connection Lines (SVG) -->
                 <svg class="absolute inset-0 w-full h-full pointer-events-none">
                   <line x1="25%" y1="10%" x2="75%" y2="50%" stroke="currentColor" class="text-slate-800" stroke-width="1" stroke-dasharray="4 4" />
                   <line x1="75%" y1="50%" x2="85%" y2="75%" stroke="currentColor" class="text-slate-800" stroke-width="1" stroke-dasharray="4 4" />
                   <line x1="25%" y1="10%" x2="85%" y2="75%" stroke="currentColor" class="text-slate-800" stroke-width="1" stroke-dasharray="4 4" />
                 </svg>
               </div>
            </div>
          </div>
          <div class="order-1 lg:order-2">
            <h2 class="text-4xl md:text-5xl font-black mb-8 leading-tight">Global Agent Mesh <br/><span class="text-purple-500">全球协作协议</span></h2>
            <p class="text-xl text-slate-400 mb-10 leading-relaxed">
              BotMatrix 独有的跨企业协作协议。通过安全握手，不同企业的数字员工可以建立互信连接，实现身份委派、技能共享与任务协同。
            </p>
            <div class="space-y-6">
              <div v-for="item in meshBenefits" :key="item.title" class="flex gap-4">
                <div class="shrink-0 w-6 h-6 rounded-full bg-purple-500/20 flex items-center justify-center mt-1">
                  <Check class="w-4 h-4 text-purple-500" />
                </div>
                <div>
                  <h4 class="font-bold text-lg mb-1">{{ item.title }}</h4>
                  <p class="text-slate-400 text-sm leading-relaxed">{{ item.desc }}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Working Principle Section -->
    <section class="py-24 bg-slate-900 overflow-hidden">
      <div class="max-w-7xl mx-auto px-4">
        <div class="flex flex-col lg:flex-row items-center gap-16">
          <div class="flex-1">
            <h2 class="text-3xl md:text-5xl font-bold mb-8">它是如何运作的？</h2>
            <div class="space-y-8">
              <div v-for="(step, index) in workingSteps" :key="index" class="flex gap-6 group">
                <div class="flex-shrink-0 w-12 h-12 rounded-2xl bg-slate-800 border border-slate-700 flex items-center justify-center text-xl font-bold text-cyan-400 group-hover:bg-cyan-500 group-hover:text-slate-900 transition-all">
                  {{ index + 1 }}
                </div>
                <div>
                  <h4 class="text-xl font-bold mb-2 group-hover:text-cyan-400 transition-colors">{{ step.title }}</h4>
                  <p class="text-slate-400 leading-relaxed">{{ step.desc }}</p>
                </div>
              </div>
            </div>
          </div>
          
          <div class="flex-1 relative">
            <!-- Visual Representation of Architecture -->
            <div class="relative bg-slate-800/50 rounded-[2rem] p-8 border border-slate-700 aspect-square flex items-center justify-center">
              <div class="absolute inset-0 bg-[radial-gradient(circle_at_center,rgba(6,182,212,0.1)_0%,transparent_70%)]"></div>
              
              <!-- Central Hub (BotNexus) -->
              <div class="relative z-10 w-32 h-32 bg-cyan-500 rounded-3xl flex flex-col items-center justify-center shadow-2xl shadow-cyan-500/50 animate-pulse-slow">
                <Cpu class="w-12 h-12 text-slate-900" />
                <span class="text-xs font-bold text-slate-900 mt-2">BotNexus</span>
              </div>
              
              <!-- Orbiting Elements -->
              <div class="absolute inset-0 flex items-center justify-center">
                <div class="w-full h-full border border-slate-700/50 rounded-full animate-spin-slow"></div>
                <div class="absolute top-10 left-1/2 -translate-x-1/2 w-16 h-16 bg-slate-800 border border-slate-700 rounded-2xl flex flex-col items-center justify-center shadow-xl">
                  <Globe class="w-6 h-6 text-blue-400" />
                  <span class="text-[10px] text-slate-500">WebUI</span>
                </div>
                <div class="absolute bottom-10 left-1/2 -translate-x-1/2 w-16 h-16 bg-slate-800 border border-slate-700 rounded-2xl flex flex-col items-center justify-center shadow-xl">
                  <Box class="w-6 h-6 text-purple-400" />
                  <span class="text-[10px] text-slate-500">Worker</span>
                </div>
                <div class="absolute left-10 top-1/2 -translate-y-1/2 w-16 h-16 bg-slate-800 border border-slate-700 rounded-2xl flex flex-col items-center justify-center shadow-xl">
                  <MessageSquare class="w-6 h-6 text-green-400" />
                  <span class="text-[10px] text-slate-500">Adapter</span>
                </div>
                <div class="absolute right-10 top-1/2 -translate-y-1/2 w-16 h-16 bg-slate-800 border border-slate-700 rounded-2xl flex flex-col items-center justify-center shadow-xl">
                  <Shield class="w-6 h-6 text-orange-400" />
                  <span class="text-[10px] text-slate-500">Auth</span>
                </div>
              </div>
              
              <!-- Connecting Lines (Decorative) -->
              <div class="absolute inset-0 overflow-hidden pointer-events-none opacity-20">
                <div class="absolute top-1/2 left-0 w-full h-px bg-gradient-to-r from-transparent via-cyan-500 to-transparent"></div>
                <div class="absolute left-1/2 top-0 h-full w-px bg-gradient-to-b from-transparent via-cyan-500 to-transparent"></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Ecosystem Section -->
    <section id="ecosystem" class="py-24">
      <div class="max-w-7xl mx-auto px-4">
        <div class="bg-gradient-to-br from-slate-800 to-slate-900 rounded-[3rem] p-8 md:p-16 border border-slate-700 relative overflow-hidden">
          <div class="absolute top-0 right-0 w-96 h-96 bg-cyan-500/5 blur-[100px] -z-10"></div>
          <div class="flex flex-col lg:flex-row gap-16 items-center">
            <div class="flex-1">
              <h2 class="text-4xl md:text-5xl font-bold mb-8 leading-tight">强大的扩展性 <br/><span class="text-cyan-400">无限的可能</span></h2>
              <div class="space-y-6">
                <div v-for="item in ecosystemItems" :key="item.title" class="flex gap-4">
                  <div class="flex-shrink-0 w-6 h-6 rounded-full bg-cyan-500/20 flex items-center justify-center mt-1">
                    <Check class="w-4 h-4 text-cyan-500" />
                  </div>
                  <div>
                    <h4 class="font-bold text-lg mb-1">{{ item.title }}</h4>
                    <p class="text-slate-400 text-sm leading-relaxed">{{ item.desc }}</p>
                  </div>
                </div>
              </div>
            </div>
            <div class="flex-1 w-full grid grid-cols-2 gap-4">
              <div v-for="i in 4" :key="i" class="aspect-square bg-slate-800/50 rounded-2xl border border-slate-700 flex items-center justify-center p-6 hover:bg-slate-800 transition-colors group">
                <component :is="ecosystemIcons[i-1]" class="w-12 h-12 text-slate-600 group-hover:text-cyan-400 transition-colors" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- CTA Section -->
    <section class="py-24 text-center px-4">
      <h2 class="text-4xl md:text-6xl font-black mb-8">准备好开启你的机器人矩阵了吗？</h2>
      <p class="text-slate-400 mb-10 max-w-xl mx-auto">立即加入我们，体验前所未有的智能自动化管理。</p>
      <div class="flex flex-col sm:flex-row gap-4 justify-center">
        <router-link :to="authStore.isAuthenticated ? '/console' : '/login'" class="px-10 py-5 bg-cyan-500 hover:bg-cyan-400 text-slate-900 rounded-2xl text-xl font-bold transition-all shadow-xl shadow-cyan-500/20">
          免费开始使用
        </router-link>
        <button class="px-10 py-5 bg-slate-800 hover:bg-slate-700 text-white rounded-2xl text-xl font-bold transition-all border border-slate-700">
          联系我们
        </button>
      </div>
    </section>

    <!-- Roadmap -->
    <section class="py-24 relative overflow-hidden">
      <div class="max-w-7xl mx-auto px-4">
        <div class="text-center mb-16">
          <h2 class="text-4xl font-bold mb-4">技术蓝图</h2>
          <p class="text-slate-400">我们不仅仅是在做工具，我们正在定义智能体时代的底层协议。</p>
        </div>
        <div class="grid grid-cols-1 md:grid-cols-4 gap-8">
          <div v-for="(step, index) in roadmap" :key="step.title" class="relative">
            <div v-if="index < roadmap.length - 1" class="hidden md:block absolute top-10 left-full w-full h-px bg-gradient-to-r from-cyan-500/50 to-transparent -z-10"></div>
            <div class="bg-slate-800/50 p-8 rounded-3xl border border-slate-700 hover:border-cyan-500/50 transition-all group">
              <div class="text-cyan-500 font-mono text-sm mb-4">PHASE 0{{ index + 1 }}</div>
              <h3 class="text-xl font-bold mb-3 group-hover:text-cyan-400 transition-colors">{{ step.title }}</h3>
              <p class="text-slate-400 text-sm leading-relaxed">{{ step.desc }}</p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <PortalFooter />

    <!-- Floating Action Button -->
    <div class="fixed bottom-8 right-8 z-50 flex flex-col gap-4">
      <button @click="showSupport = true" class="w-14 h-14 bg-cyan-500 text-slate-900 rounded-2xl flex items-center justify-center shadow-2xl shadow-cyan-500/20 hover:scale-110 hover:rotate-12 transition-all group relative">
        <MessageCircle class="w-7 h-7" />
        <div class="absolute right-full mr-4 px-3 py-1 bg-slate-800 text-white text-xs font-bold rounded-lg border border-slate-700 opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none">
          在线咨询
        </div>
      </button>
      <button @click="scrollToTop" v-show="y > 400" class="w-14 h-14 bg-slate-800 text-slate-400 rounded-2xl flex items-center justify-center border border-slate-700 shadow-2xl hover:bg-slate-700 transition-all group relative">
        <ArrowUp class="w-6 h-6" />
      </button>
    </div>

    <!-- Support Modal -->
    <div v-if="showSupport" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-slate-950/80 backdrop-blur-sm" @click="showSupport = false"></div>
      <div class="relative w-full max-w-md bg-slate-900 border border-slate-700 rounded-[2rem] p-8 shadow-2xl overflow-hidden">
        <div class="absolute top-0 left-0 w-full h-1 bg-cyan-500"></div>
        <button @click="showSupport = false" class="absolute top-6 right-6 p-2 hover:bg-slate-800 rounded-full transition-colors">
          <X class="w-5 h-5" />
        </button>
        
        <div class="text-center">
          <div class="w-16 h-16 bg-cyan-500/10 text-cyan-500 rounded-2xl flex items-center justify-center mx-auto mb-6">
            <Users class="w-8 h-8" />
          </div>
          <h3 class="text-2xl font-bold mb-4">加入我们的社区</h3>
          <p class="text-slate-400 mb-8 leading-relaxed">与数千名开发者共同交流机器人开发心得，获取最新的技术支持与插件资源。</p>
          
          <div class="grid grid-cols-2 gap-4 mb-8">
            <div class="bg-slate-800/50 rounded-2xl p-6 border border-slate-700 flex flex-col items-center gap-3">
              <div class="w-10 h-10 bg-green-500/10 text-green-500 rounded-xl flex items-center justify-center">
                <MessageCircle class="w-6 h-6" />
              </div>
              <div class="text-xs font-bold">微信交流群</div>
            </div>
            <div class="bg-slate-800/50 rounded-2xl p-6 border border-slate-700 flex flex-col items-center gap-3">
              <div class="w-10 h-10 bg-blue-500/10 text-blue-500 rounded-xl flex items-center justify-center">
                <Send class="w-6 h-6" />
              </div>
              <div class="text-xs font-bold">Telegram 频道</div>
            </div>
          </div>

          <button @click="showSupport = false" class="w-full py-4 bg-cyan-500 hover:bg-cyan-400 text-slate-900 rounded-xl font-bold transition-all">
            我知道了
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import PortalHeader from '@/components/layout/PortalHeader.vue';
import PortalFooter from '@/components/layout/PortalFooter.vue';
import {
  Rocket,
  ShieldCheck,
  Cpu,
  Globe,
  Zap,
  Bot,
  ArrowRight,
  Sparkles,
  MessageCircle,
  Network,
  Lock,
  Workflow,
  Check,
  Building2,
  ChevronRight,
  Plus,
  Layers,
  Shield,
  MessageSquare,
  Box,
  Code,
  ArrowUp,
  X,
  Users,
  Send,
  BrainCircuit,
  Headphones,
  Search,
  Activity,
  Calculator,
  Calendar,
  LineChart,
  ShieldAlert
} from 'lucide-vue-next';
import { useAuthStore } from '@/stores/auth';
import { ref, onMounted, onUnmounted } from 'vue';

const authStore = useAuthStore();
const showSupport = ref(false);
const y = ref(0);

const handleScroll = () => {
  y.value = window.scrollY;
};

onMounted(() => {
  window.addEventListener('scroll', handleScroll, { passive: true });
});

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll);
});

const scrollToTop = () => {
  window.scrollTo({ top: 0, behavior: 'smooth' });
};

const stats = [
  { label: '累计消息', value: '1.2M+' },
  { label: '活跃机器人', value: '500+' },
  { label: '分布式节点', value: '120+' },
  { label: '开源 Star', value: '2.5k+' }
];

const features = [
  {
    icon: Network,
    title: '分布式 Worker 架构',
    desc: 'Nexus 节点负责全局调度，Worker 节点分布在不同地理位置执行具体任务，实现高可用与低延迟。',
    color: 'text-cyan-500',
    bg: 'bg-cyan-500/10'
  },
  {
    icon: Lock,
    title: 'Privacy Bastion 隐私堡垒',
    desc: '内置数据脱敏引擎，确保企业敏感信息在进入公有云大模型前被安全过滤，合规性 100%。',
    color: 'text-purple-500',
    bg: 'bg-purple-500/10'
  },
  {
    icon: Workflow,
    title: 'MCP 技能集集成',
    desc: '支持 Model Context Protocol 标准，无缝对接上万个开源工具，让数字员工具备无限的操作可能。',
    color: 'text-blue-500',
    bg: 'bg-blue-500/10'
  },
  {
    icon: Zap,
    title: '认知记忆引擎',
    desc: '基于向量数据库的长期记忆 system，支持事实提取、知识沉淀与自主学习，AI 越用越聪明。',
    color: 'text-amber-500',
    bg: 'bg-amber-500/10'
  },
  {
    icon: ShieldCheck,
    title: 'HITL 安全边界',
    desc: '独有的 Human-in-the-loop 机制，确保高风险操作必须经过人类确认，安全边界清晰可见。',
    color: 'text-emerald-500',
    bg: 'bg-emerald-500/10'
  },
  {
    icon: Globe,
    title: 'Global Agent Mesh',
    desc: '跨企业协作协议，打破数据孤岛，实现全球范围内智能体的安全发现与能力委派。',
    color: 'text-indigo-500',
    bg: 'bg-indigo-500/10'
  }
];

const meshBenefits = [
  {
    title: '跨域身份互认 (DID)',
    desc: '基于去中心化身份协议，确保不同矩阵间的数字员工身份可信、权限受控。'
  },
  {
    title: '技能委派机制',
    desc: '当本地数字员工无法解决问题时，可安全地将任务委派给具备相应技能的外部专家 AI。'
  },
  {
    title: '资源动态平衡',
    desc: '矩阵间的算力与存储资源可根据负载情况进行动态共享与调度，极大提升利用率。'
  }
];

const evolution = [
  {
    title: '单体机器人',
    desc: '传统的规则驱动或简单对话机器人。',
    tags: ['固定指令', '无记忆', '孤岛运行'],
    color: 'text-slate-400'
  },
  {
    title: '技能机器人 (Agent)',
    desc: '具备工具调用能力，能解决特定领域的复杂任务。',
    tags: ['MCP 插件', '短期记忆', '意图识别'],
    color: 'text-cyan-400'
  },
  {
    title: '数字员工 (Digital Employee)',
    desc: '具备身份、职责、KPI 考核与长期记忆的虚拟劳动力。',
    tags: ['KPI 驱动', '认知进化', '身份认证'],
    color: 'text-purple-400'
  },
  {
    title: '机器人矩阵 (Matrix)',
    desc: '由 Global Agent Mesh 连接的全球智能网格，生产力的终极革命。',
    tags: ['全球协作', '算力共享', '自主进化'],
    color: 'text-white'
  }
];

const workingSteps = [
  {
    title: '部署矩阵中枢',
    desc: '搭建 BotNexus 管理后台，这是您所有机器人的指挥中心。'
  },
  {
    title: '接入通讯账号',
    desc: '快速接入微信、QQ、飞书等账号，将其纳入矩阵，实现统一管控。'
  },
  {
    title: '注入专业知识',
    desc: '上传业务文档或接入数据库，通过 RAG 技术让 AI 快速掌握行业背景。'
  },
  {
    title: '开启数字办公',
    desc: '为机器人设定工作目标，让它们像正式员工一样开始为您处理业务。'
  }
];

const ecosystemItems = [
  {
    title: '多平台适配器',
    desc: '原生支持微信、飞书、钉钉、Telegram 及私有部署的 IM 系统。'
  },
  {
    title: '插件市场 (MCP)',
    desc: '一键安装各种专业技能，从财务报销到代码审计，应有尽有。'
  },
  {
    title: '开发者 SDK',
    desc: '使用 Go/TS 轻松编写自定义插件，接入企业内部核心系统。'
  }
];

const ecosystemIcons = [Layers, Zap, Code, Shield];

const roadmap = [
  {
    title: '矩阵起步',
    desc: '实现多平台账号接入与基础消息调度，搭建稳健的底层管理架构。'
  },
  {
    title: '技能进化',
    desc: '推出官方插件市场，引入 AI 语义感知能力，让机器人具备专业技能。'
  },
  {
    title: '员工养成',
    desc: '上线数字员工管理系统，引入工号、职责及绩效考核，实现业务闭环。'
  },
  {
    title: '智能生态',
    desc: '构建全球智能体协作网络，实现跨组织的技能共享与自主协作。'
  }
];
</script>

<style scoped>
@keyframes fade-in {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.animate-fade-in {
  animation: fade-in 0.8s ease-out forwards;
}

@keyframes spin-slow {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@keyframes pulse-slow {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.8; transform: scale(0.95); }
}

.animate-spin-slow {
  animation: spin-slow 20s linear infinite;
}

.animate-pulse-slow {
  animation: pulse-slow 3s ease-in-out infinite;
}

.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>


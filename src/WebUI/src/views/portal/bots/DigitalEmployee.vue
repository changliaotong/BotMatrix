<template>
  <div class="min-h-screen bg-slate-950 text-white selection:bg-purple-500/30">
    <!-- Cyberpunk Header -->
    <nav class="fixed top-0 w-full z-50 bg-slate-950/80 backdrop-blur-xl border-b border-purple-500/10">
      <div class="max-w-7xl mx-auto px-4 h-20 flex justify-between items-center">
        <div class="flex items-center gap-3 group cursor-pointer" @click="scrollTo('hero')">
          <div class="w-10 h-10 bg-gradient-to-tr from-purple-600 to-blue-600 rounded-xl flex items-center justify-center shadow-lg shadow-purple-500/20 group-hover:rotate-[360deg] transition-transform duration-1000">
            <BrainCircuit class="w-6 h-6 text-white" />
          </div>
          <div class="flex flex-col">
            <span class="text-xl font-black tracking-tighter text-transparent bg-clip-text bg-gradient-to-r from-white to-purple-400">AI WORKER</span>
            <span class="text-[10px] text-purple-500 font-bold tracking-[0.2em] uppercase leading-none">Matrix Revolution</span>
          </div>
        </div>
        
        <div class="hidden md:flex items-center gap-8 text-sm font-bold text-slate-400">
          <button @click="scrollTo('vision')" class="hover:text-purple-400 transition-colors">终极愿景</button>
          <button @click="scrollTo('revolution')" class="hover:text-purple-400 transition-colors">矩阵革命</button>
          <button @click="scrollTo('kpi')" class="hover:text-purple-400 transition-colors">生产关系</button>
          <router-link to="/" class="px-4 py-2 rounded-lg bg-slate-900 text-slate-400 hover:text-white transition-all border border-slate-800">
            返回矩阵中枢
          </router-link>
        </div>
      </div>
    </nav>

    <!-- Hero Section -->
    <header id="hero" class="pt-40 pb-32 px-4 relative overflow-hidden">
      <!-- Matrix Code Rain Placeholder / Background -->
      <div class="absolute inset-0 opacity-10 pointer-events-none overflow-hidden">
        <div class="absolute inset-0 bg-[grid-slate-800] [mask-image:radial-gradient(white,transparent)]"></div>
        <div class="matrix-lines absolute inset-0"></div>
      </div>
      
      <div class="absolute top-0 left-1/2 -translate-x-1/2 w-full h-full bg-[radial-gradient(circle_at_center,rgba(168,85,247,0.2)_0%,transparent_70%)] -z-10"></div>
      <div class="max-w-7xl mx-auto text-center relative reveal">
        <div class="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-purple-500/10 border border-purple-500/20 text-purple-400 text-xs font-bold mb-8 animate-pulse shadow-[0_0_15px_rgba(168,85,247,0.2)]">
          <Sparkles class="w-4 h-4" />
          机器人矩阵的终极形态：数字员工 (Agentic Workflow)
        </div>
        <h1 class="text-6xl md:text-9xl font-black mb-12 tracking-tighter leading-none">
          不仅仅是 AI <br/>
          <span class="bg-clip-text text-transparent bg-gradient-to-r from-purple-400 via-blue-500 to-cyan-400 drop-shadow-[0_10px_10px_rgba(168,85,247,0.3)]">
            更是你的数字化雇员
          </span>
        </h1>
        <p class="text-xl md:text-3xl text-slate-400 mb-16 max-w-4xl mx-auto font-light leading-relaxed">
          打破“工具”的边界。在 BotMatrix 的矩阵革命中，AI 进化为具备认知、身份与 KPI 的独立生产力单元，重构人类的生产关系。
        </p>
        <div class="flex flex-col sm:flex-row gap-6 justify-center items-center">
          <button @click="openAction('join')" class="px-10 py-5 bg-purple-600 hover:bg-purple-500 text-white rounded-2xl text-xl font-bold transition-all shadow-[0_0_40px_rgba(168,85,247,0.4)] flex items-center gap-3 group">
            加入矩阵革命 <ArrowRight class="w-6 h-6 group-hover:translate-x-2 transition-transform" />
          </button>
          <button @click="scrollTo('vision')" class="px-10 py-5 bg-slate-900/50 hover:bg-slate-800 text-white rounded-2xl text-xl font-bold transition-all border border-slate-800 backdrop-blur-sm">
            探索未来蓝图
          </button>
        </div>
      </div>
    </header>

    <!-- Vision Section: Production Relations Revolution -->
    <section id="vision" class="py-32 bg-slate-950 relative overflow-hidden">
      <div class="max-w-7xl mx-auto px-4">
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-24 items-center">
          <div>
            <h2 class="text-4xl md:text-6xl font-black mb-8 leading-tight">
              生产关系的 <br/>
              <span class="text-purple-500">终极重构</span>
            </h2>
            <p class="text-xl text-slate-400 mb-12 leading-relaxed">
              在传统视角下，AI 只是提高效率的扳手。但在 BotMatrix 愿景中，数字员工是具备 **工号、职位、晋升路径与虚拟薪资** 的正式成员。
            </p>
            <div class="space-y-8">
              <div v-for="v in visionPoints" :key="v.title" class="flex gap-6 group">
                <div class="w-14 h-14 shrink-0 bg-slate-900 border border-slate-800 rounded-2xl flex items-center justify-center text-purple-500 group-hover:bg-purple-500 group-hover:text-white transition-all">
                  <component :is="v.icon" class="w-7 h-7" />
                </div>
                <div>
                  <h4 class="text-2xl font-bold mb-2 group-hover:text-purple-400 transition-colors">{{ v.title }}</h4>
                  <p class="text-slate-400 leading-relaxed">{{ v.desc }}</p>
                </div>
              </div>
            </div>
          </div>
          <div class="relative">
            <div class="absolute inset-0 bg-purple-500/20 blur-[100px] -z-10"></div>
            <div class="bg-slate-900/50 border border-slate-800 rounded-[3rem] p-12 aspect-square flex items-center justify-center relative overflow-hidden">
              <div class="absolute inset-0 opacity-20">
                <div class="absolute top-0 left-0 w-full h-full bg-[grid-slate-800] [mask-image:radial-gradient(white,transparent)]"></div>
              </div>
              <!-- Central AI Identity -->
              <div class="relative z-10 flex flex-col items-center">
                <div class="w-40 h-40 bg-gradient-to-br from-purple-500 to-blue-600 rounded-full flex items-center justify-center shadow-2xl animate-pulse">
                  <UserCheck class="w-20 h-20 text-white" />
                </div>
                <div class="mt-8 text-center">
                  <div class="text-3xl font-black mb-2 tracking-widest">ID: AX-709</div>
                  <div class="text-purple-500 font-bold uppercase tracking-[0.3em] text-sm">Senior AI Analyst</div>
                </div>
              </div>
              <!-- Orbiting Stats -->
              <div class="absolute top-12 right-12 bg-slate-800/80 backdrop-blur border border-slate-700 p-4 rounded-2xl shadow-xl animate-bounce-slow">
                <div class="text-xs text-slate-500 uppercase font-bold mb-1">Performance</div>
                <div class="text-xl font-black text-green-400">99.8%</div>
              </div>
              <div class="absolute bottom-20 left-12 bg-slate-800/80 backdrop-blur border border-slate-700 p-4 rounded-2xl shadow-xl animate-bounce-slow" style="animation-delay: 1s">
                <div class="text-xs text-slate-500 uppercase font-bold mb-1">Total Token Paid</div>
                <div class="text-xl font-black text-purple-400">12.5M</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Matrix Revolution: The Global Agent Mesh -->
    <section id="revolution" class="py-32 bg-slate-900/30">
      <div class="max-w-7xl mx-auto px-4 text-center">
        <h2 class="text-4xl md:text-7xl font-black mb-12">矩阵革命：无界协作</h2>
        <p class="text-xl text-slate-400 max-w-3xl mx-auto mb-20 leading-relaxed">
          通过 **Global Agent Mesh** 协议，数字员工不再受限于单一组织。它们可以跨企业、跨地域进行安全握手与技能委派，构建一个全球化的分布式智能协作网络。
        </p>
        
        <div class="grid grid-cols-1 md:grid-cols-3 gap-12">
          <div v-for="r in revolutionSteps" :key="r.title" class="p-12 rounded-[3rem] bg-slate-950 border border-slate-800 hover:border-purple-500/50 transition-all group">
            <div class="w-20 h-20 bg-purple-500/10 text-purple-500 rounded-3xl flex items-center justify-center mx-auto mb-8 group-hover:scale-110 group-hover:bg-purple-500 group-hover:text-white transition-all">
              <component :is="r.icon" class="w-10 h-10" />
            </div>
            <h3 class="text-2xl font-bold mb-6 text-white">{{ r.title }}</h3>
            <p class="text-slate-400 leading-relaxed">{{ r.desc }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- KPI & HR Section -->
    <section id="kpi" class="py-32 bg-slate-950">
      <div class="max-w-7xl mx-auto px-4">
        <div class="bg-gradient-to-br from-slate-900 to-purple-900/20 rounded-[4rem] p-12 md:p-24 border border-slate-800 relative overflow-hidden">
          <div class="absolute top-0 right-0 w-full h-full bg-[radial-gradient(circle_at_top_right,rgba(168,85,247,0.1),transparent_50%)]"></div>
          <div class="flex flex-col lg:flex-row gap-20 items-center relative z-10">
            <div class="flex-1">
              <h2 class="text-4xl md:text-5xl font-black mb-8 leading-tight">管理 AI，<br/>像管理你的精英团队</h2>
              <p class="text-xl text-slate-400 mb-10 leading-relaxed">
                我们为数字员工打造了专属的 HR 系统。您可以为每个机器人设定季度 KPI，系统会自动根据其完成度、资源消耗与用户反馈计算绩效，并实时调整其模型算力权重。
              </p>
              <div class="grid grid-cols-2 gap-6">
                <div v-for="stat in hrStats" :key="stat.label" class="p-6 bg-slate-950/50 rounded-2xl border border-slate-800">
                  <div class="text-3xl font-black text-purple-400 mb-2">{{ stat.value }}</div>
                  <div class="text-xs text-slate-500 font-bold uppercase tracking-widest">{{ stat.label }}</div>
                </div>
              </div>
            </div>
            <div class="flex-1 w-full space-y-6">
              <div v-for="role in roles" :key="role.name" class="p-6 bg-slate-950/80 backdrop-blur rounded-3xl border border-slate-800 hover:border-purple-500/50 transition-all">
                <div class="flex justify-between items-center mb-4">
                  <div class="flex items-center gap-4">
                    <div class="w-12 h-12 bg-slate-800 rounded-xl flex items-center justify-center">
                      <component :is="role.icon" class="w-6 h-6 text-purple-500" />
                    </div>
                    <div>
                      <div class="font-bold text-lg">{{ role.name }}</div>
                      <div class="text-xs text-slate-500">{{ role.dept }}</div>
                    </div>
                  </div>
                  <div class="px-3 py-1 bg-green-500/10 text-green-500 text-[10px] font-bold rounded-full uppercase">Active</div>
                </div>
                <div class="mb-2 text-xs text-slate-400 line-clamp-1">核心技能: {{ role.skills }}</div>
                <div class="w-full bg-slate-800 h-1.5 rounded-full overflow-hidden">
                  <div class="bg-purple-500 h-full" :style="{ width: role.progress + '%' }"></div>
                </div>
                <div class="flex justify-between mt-2 text-[10px] font-bold text-slate-500">
                  <span>KPI PROGRESS</span>
                  <span>{{ role.progress }}%</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Industry Roles Showcase -->
    <section class="py-32 bg-slate-950">
      <div class="max-w-7xl mx-auto px-4">
        <div class="text-center mb-20">
          <h2 class="text-4xl md:text-6xl font-black mb-6">全场景岗位覆盖</h2>
          <p class="text-xl text-slate-400 max-w-3xl mx-auto">
            BotMatrix 提供标准化的岗位模板，覆盖企业运营的各个核心环节，实现从“人力密集”到“算力密集”的平滑过渡。
          </p>
        </div>
        
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          <div v-for="job in industryJobs" :key="job.title" class="group bg-slate-900/50 border border-slate-800 rounded-[2.5rem] p-10 hover:border-purple-500/30 transition-all hover:-translate-y-2">
            <div class="w-16 h-16 rounded-2xl bg-purple-500/10 text-purple-500 flex items-center justify-center mb-8 group-hover:bg-purple-500 group-hover:text-white transition-all">
              <component :is="job.icon" class="w-8 h-8" />
            </div>
            <h3 class="text-2xl font-bold mb-4">{{ job.title }}</h3>
            <p class="text-slate-400 mb-8 leading-relaxed text-sm">{{ job.desc }}</p>
            <div class="space-y-4">
              <div class="text-[10px] font-bold text-slate-500 uppercase tracking-widest border-b border-slate-800 pb-2">核心能力 (MCP Tools)</div>
              <div class="flex flex-wrap gap-2">
                <span v-for="skill in job.skills" :key="skill" class="px-3 py-1 bg-slate-800 rounded-full text-xs text-slate-400">{{ skill }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Lifecycle & HITL -->
    <section id="security" class="py-32 bg-slate-900/30">
      <div class="max-w-7xl mx-auto px-4">
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-24 items-center">
          <div>
            <h2 class="text-4xl md:text-5xl font-black mb-8 leading-tight">
              Human-in-the-loop <br/>
              <span class="text-purple-500">人机协作的安全边界</span>
            </h2>
            <p class="text-xl text-slate-400 mb-12 leading-relaxed">
              数字员工并非完全脱离监管。BotMatrix 内置了严格的 HITL (人工干预) 机制，确保高风险操作与知识盲区始终在人类的掌控之下。
            </p>
            <div class="space-y-6">
              <div v-for="feature in securityFeatures" :key="feature.title" class="flex gap-4 p-6 bg-slate-950 rounded-2xl border border-slate-800">
                <div class="shrink-0 w-10 h-10 bg-purple-500/10 text-purple-500 rounded-lg flex items-center justify-center">
                  <component :is="feature.icon" class="w-5 h-5" />
                </div>
                <div>
                  <h4 class="font-bold mb-1">{{ feature.title }}</h4>
                  <p class="text-sm text-slate-400 leading-relaxed">{{ feature.desc }}</p>
                </div>
              </div>
            </div>
          </div>
          <div class="relative bg-slate-950 border border-slate-800 rounded-[3rem] p-8 md:p-12 overflow-hidden shadow-2xl">
            <div class="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-purple-500 via-blue-500 to-cyan-500"></div>
            <h3 class="text-xl font-bold mb-8 flex items-center gap-3">
              <Activity class="w-5 h-5 text-purple-500" /> 运行全生命周期 (Operational Lifecycle)
            </h3>
            <div class="relative space-y-8 before:absolute before:left-[1.2rem] before:top-2 before:bottom-2 before:w-px before:bg-slate-800">
              <div v-for="(step, i) in lifecycleSteps" :key="step.title" class="relative pl-12">
                <div class="absolute left-0 top-1 w-10 h-10 bg-slate-900 border border-slate-800 rounded-full flex items-center justify-center text-xs font-bold text-purple-400 z-10">
                  {{ i + 1 }}
                </div>
                <h4 class="font-bold text-white mb-1">{{ step.title }}</h4>
                <p class="text-xs text-slate-500 leading-relaxed">{{ step.desc }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Agent Collaboration Bridge -->
    <section id="collaboration" class="py-32 bg-slate-950">
      <div class="max-w-7xl mx-auto px-4">
        <div class="text-center mb-20">
          <h2 class="text-4xl md:text-6xl font-black mb-6">智能体协作网桥</h2>
          <p class="text-xl text-slate-400 max-w-3xl mx-auto">
            打破单一 AI 的能力上限。BotMatrix 协作网桥允许数字员工之间像人类同事一样相互咨询、委派任务与汇报进度。
          </p>
        </div>

        <div class="grid grid-cols-1 lg:grid-cols-3 gap-8 mb-20">
          <div v-for="feature in collaborationFeatures" :key="feature.title" class="bg-slate-900/50 border border-slate-800 p-10 rounded-[2.5rem] hover:border-purple-500/30 transition-all">
            <div class="w-14 h-14 bg-purple-500/10 text-purple-500 rounded-2xl flex items-center justify-center mb-8">
              <component :is="feature.icon" class="w-7 h-7" />
            </div>
            <h3 class="text-2xl font-bold mb-4">{{ feature.title }}</h3>
            <p class="text-slate-400 leading-relaxed">{{ feature.desc }}</p>
          </div>
        </div>

        <!-- Trace Mockup -->
        <div class="bg-slate-900/50 border border-slate-800 rounded-[3rem] p-8 md:p-12">
          <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-6 mb-12">
            <div>
              <h3 class="text-2xl font-bold mb-2 flex items-center gap-3">
                <ListTree class="w-6 h-6 text-purple-500" />
                协作链路追踪 (A2A Collaboration Trace)
                <span class="flex h-2 w-2 relative ml-2">
                  <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-purple-400 opacity-75"></span>
                  <span class="relative inline-flex rounded-full h-2 w-2 bg-purple-500"></span>
                </span>
              </h3>
              <p class="text-slate-500 text-sm">实时监控矩阵中数字员工之间的任务流转与指令交互 (Real-time Mesh Protocol)</p>
            </div>
            <div class="flex items-center gap-4 bg-slate-950 px-4 py-2 rounded-xl border border-slate-800 w-full md:w-auto">
              <LucideSearch class="w-4 h-4 text-slate-500" />
              <input type="text" placeholder="搜索 Trace ID / 参与者..." class="bg-transparent border-none text-sm focus:outline-none text-slate-300 w-full">
            </div>
          </div>

          <div class="space-y-4">
            <div v-for="trace in collaborationTraces" :key="trace.id" class="flex flex-col lg:flex-row items-start lg:items-center gap-6 p-6 bg-slate-950/80 rounded-2xl border border-slate-800 hover:border-purple-500/30 transition-all group">
              <div class="flex items-center gap-4 min-w-[120px]">
                <div class="w-2 h-2 rounded-full bg-purple-500 shadow-[0_0_10px_rgba(168,85,247,0.5)]"></div>
                <span class="font-mono text-xs text-slate-500">{{ trace.id }}</span>
              </div>
              <div class="flex-1 flex flex-col md:flex-row items-start md:items-center gap-4 md:gap-12">
                <div class="flex items-center gap-3">
                  <div class="px-3 py-1 bg-slate-900 rounded-lg border border-slate-800 text-xs font-bold">{{ trace.from }}</div>
                  <ArrowRight class="w-4 h-4 text-slate-600" />
                  <div class="px-3 py-1 bg-slate-900 rounded-lg border border-slate-800 text-xs font-bold">{{ trace.to }}</div>
                </div>
                <div class="text-sm font-medium text-slate-300 group-hover:text-purple-400 transition-colors">{{ trace.action }}</div>
              </div>
              <div class="flex items-center gap-6 min-w-[200px] justify-end w-full lg:w-auto">
                <div class="flex items-center gap-2" :class="trace.color">
                  <component :is="trace.icon" class="w-4 h-4" />
                  <span class="text-[10px] font-black uppercase tracking-widest">{{ trace.status }}</span>
                </div>
                <div class="text-[10px] font-bold text-slate-600 uppercase">{{ trace.time }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Futuristic Vision (Computer Use & Swarm) -->
    <section class="py-32 bg-gradient-to-b from-slate-950 to-purple-950/20">
      <div class="max-w-7xl mx-auto px-4">
        <div class="flex flex-col lg:flex-row items-center gap-24">
          <div class="flex-1">
            <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-purple-500/10 text-purple-400 text-xs font-bold mb-6 border border-purple-500/20">
              <Sparkles class="w-4 h-4" /> 未来实验室 (Future Lab)
            </div>
            <h2 class="text-4xl md:text-5xl font-black mb-8 leading-tight">迈向 AGI 的 <br/><span class="text-purple-500">终极演进</span></h2>
            <p class="text-xl text-slate-400 mb-12">
              我们不仅在构建工具，更在定义未来。Vision 3.0 计划将赋予数字员工跨越软件边界的操作能力与集群化协同智慧。
            </p>
            <div class="space-y-8">
              <div v-for="feature in futuristicFeatures" :key="feature.title" class="group relative p-8 bg-slate-900/50 border border-slate-800 rounded-3xl hover:border-purple-500/50 transition-all">
                <div class="absolute top-4 right-8 px-2 py-1 bg-purple-500/20 text-purple-400 text-[10px] font-bold rounded uppercase tracking-tighter">{{ feature.tag }}</div>
                <div class="flex gap-6">
                  <div class="shrink-0 w-12 h-12 bg-purple-500/10 text-purple-500 rounded-xl flex items-center justify-center">
                    <component :is="feature.icon" class="w-6 h-6" />
                  </div>
                  <div>
                    <h4 class="text-xl font-bold mb-2">{{ feature.title }}</h4>
                    <p class="text-slate-400 text-sm leading-relaxed">{{ feature.desc }}</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div class="flex-1 relative">
            <div class="aspect-video bg-slate-900 rounded-[2.5rem] border border-slate-800 p-2 overflow-hidden shadow-2xl">
              <div class="w-full h-full bg-slate-950 rounded-[2rem] relative flex items-center justify-center">
                 <!-- Visual placeholder for Swarm/Computer Use -->
                 <div class="absolute inset-0 bg-[radial-gradient(circle_at_center,rgba(168,85,247,0.1)_0%,transparent_70%)]"></div>
                 <div class="relative z-10 text-center">
                   <div class="w-20 h-20 bg-purple-500/20 rounded-full flex items-center justify-center mx-auto mb-6 animate-pulse">
                     <BrainCircuit class="w-10 h-10 text-purple-500" />
                   </div>
                   <div class="text-sm font-mono text-purple-400">MATRIX EVOLUTION IN PROGRESS...</div>
                 </div>
              </div>
            </div>
            <!-- Decorative elements -->
            <div class="absolute -top-12 -right-12 w-64 h-64 bg-purple-500/10 blur-[100px] -z-10"></div>
          </div>
        </div>
      </div>
    </section>

    <!-- The Future Vision -->
    <section class="py-40 text-center px-4 relative">
      <div class="absolute inset-0 bg-[radial-gradient(circle_at_center,rgba(168,85,247,0.05)_0%,transparent_70%)]"></div>
      <h2 class="text-5xl md:text-8xl font-black mb-12 tracking-tighter">AI 的未来，<br/>是生产关系的革命。</h2>
      <p class="text-xl md:text-2xl text-slate-400 mb-16 max-w-3xl mx-auto font-light leading-relaxed">
        我们不只是在制造更聪明的机器，我们正在构建一个数字与物理世界无缝交织的、基于可信协议的新型社会组织。
      </p>
      <button @click="openAction('join')" class="px-12 py-6 bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-500 hover:to-blue-500 text-white rounded-3xl text-2xl font-black transition-all shadow-2xl shadow-purple-500/40">
        开启你的数字员工时代
      </button>
    </section>

    <PortalFooter />

    <!-- Action Modal -->
    <div v-if="showModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-slate-950/90 backdrop-blur-md" @click="showModal = false"></div>
      <div class="relative w-full max-w-lg bg-slate-900 border border-slate-800 rounded-[3rem] p-10 shadow-2xl overflow-hidden">
        <div class="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-purple-500 to-blue-500"></div>
        <button @click="showModal = false" class="absolute top-8 right-8 p-2 hover:bg-slate-800 rounded-full transition-colors">
          <X class="w-6 h-6 text-slate-500" />
        </button>
        
        <div class="text-center">
          <div class="w-20 h-20 bg-purple-500/10 text-purple-500 rounded-[2rem] flex items-center justify-center mx-auto mb-8 animate-pulse">
            <Zap class="w-10 h-10" />
          </div>
          <h3 class="text-3xl font-bold mb-6">加入矩阵革命</h3>
          <p class="text-slate-400 mb-10 leading-relaxed text-lg">
            数字员工功能目前处于 **邀请制预览阶段**。如果您代表企业或拥有大规模机器人集群，欢迎提交申请，共同探索 AI 的终极形态。
          </p>
          <div class="bg-slate-950/50 rounded-2xl p-6 mb-10 text-left border border-slate-800">
            <div class="text-xs font-bold text-slate-500 uppercase tracking-widest mb-4">Priority Access</div>
            <div class="flex items-center gap-4 text-purple-400 font-bold">
              <MessageCircle class="w-6 h-6" />
              <span>微信联系：Matrix_Revolution_AI</span>
            </div>
          </div>
          <button @click="showModal = false" class="w-full py-5 bg-purple-600 hover:bg-purple-500 text-white rounded-2xl font-black transition-all text-xl">
            接受邀请
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import PortalFooter from '@/components/layout/PortalFooter.vue';
import { 
  BrainCircuit, 
  Sparkles, 
  ArrowRight, 
  UserCheck, 
  Globe, 
  Zap, 
  MessageCircle, 
  X, 
  Target, 
  ShieldCheck, 
  Cpu, 
  TrendingUp, 
  Briefcase, 
  Search as LucideSearch,
  Activity, 
  Calculator, 
  Calendar, 
  Headphones, 
  LineChart, 
  ShieldAlert, 
  Eye, 
  Lock, 
  Network, 
  Workflow, 
  Monitor, 
  Users, 
  ListTree, 
  Clock, 
  CheckCircle2
} from 'lucide-vue-next';
import { ref } from 'vue';

const showModal = ref(false);

const scrollTo = (id: string) => {
  const el = document.getElementById(id);
  if (el) el.scrollIntoView({ behavior: 'smooth' });
};

const openAction = (type: string) => {
  showModal.value = true;
};

const visionPoints = [
  {
    icon: Target,
    title: 'KPI 驱动',
    desc: '为 AI 设定明确的业务目标，系统根据转化率、满意度等指标自动考核。'
  },
  {
    icon: Briefcase,
    title: '身份与职责',
    desc: '每个数字员工拥有唯一的数字工号，承载特定的业务知识与操作权限。'
  },
  {
    icon: TrendingUp,
    title: '自主进化',
    desc: '基于工作反馈与知识库沉淀，数字员工会持续优化自己的响应逻辑与专业技能。'
  }
];

const revolutionSteps = [
  {
    icon: Globe,
    title: '跨域握手',
    desc: '打破企业孤岛，通过安全网桥协议实现不同矩阵间的智能体发现与能力委派。'
  },
  {
    icon: ShieldCheck,
    title: '可信身份',
    desc: '基于去中心化标识符 (DID)，确保每一个数字员工的身份真实、操作可追溯、权限受控。'
  },
  {
    icon: Cpu,
    title: '算力共享',
    desc: '矩阵内部资源动态平衡，闲置算力自动支援高负载任务，实现生产力最大化。'
  }
];

const hrStats = [
  { label: 'Avg. ROI', value: '420%' },
  { label: 'Token Salary Paid', value: '1.2B' },
  { label: 'Tasks Automated', value: '50M+' },
  { label: 'System Uptime', value: '99.99%' }
];

const roles = [
  { name: 'AX-709', dept: 'Customer Support', icon: Headphones, progress: 98, skills: '意图识别、RAG 知识库、多语言翻译' },
  { name: 'BX-102', dept: 'Finance & Audit', icon: Calculator, progress: 85, skills: '发票识别、ERP 接口、风险预警' },
  { name: 'CX-443', dept: 'Admin Assistant', icon: Calendar, progress: 92, skills: '日程管理、会议预订、差旅审批' }
];

const industryJobs = [
  {
    title: '数字财务初审',
    icon: Calculator,
    desc: '自动化发票报销预审、账单核对及 ERP 凭证录入。确保每一笔账目合规、准确、透明。',
    skills: ['发票识别', '账单审计', 'ERP 对接', '税务核查']
  },
  {
    title: '数字行政助理',
    icon: Calendar,
    desc: '全天候处理日程安排、会议室预订、差旅申请。像真人助理一样理解优先级。',
    skills: ['自然语言排程', '会议纪要', '资产管理', '审批流转']
  },
  {
    title: '数字技术支持',
    icon: Headphones,
    desc: '7x24 小时监控系统日志，发现异常自动报送。辅助代码 Review 与知识库查询。',
    skills: ['日志监控', '故障排查', '代码分析', '文档搜索']
  },
  {
    title: '数字前台分诊',
    icon: UserCheck,
    desc: '多渠道访客接待，精准识别意图并分派给最合适的专业员工。',
    skills: ['意图路由', '访客登记', 'FAQ 应答', '满意度回访']
  },
  {
    title: '数字数据分析师',
    icon: LineChart,
    desc: '每日自动生成经营日报，预测库存波动，发现业务增长的隐藏规律。',
    skills: ['SQL 生成', '可视化报表', '趋势预测', '归因分析']
  },
  {
    title: '数字安全合规官',
    icon: ShieldAlert,
    desc: '实时审计 AI 操作行为，确保数据脱敏，防止敏感信息泄露。',
    skills: ['数据脱敏', '操作审计', '敏感词过滤', '合规检查']
  }
];

const securityFeatures = [
  {
    title: '低置信度拦截',
    icon: Eye,
    desc: '当 AI 对任务理解置信度低于 0.6 时，自动挂起任务并向人类主管推送通知。'
  },
  {
    title: '高风险操作确认',
    icon: Lock,
    desc: '涉及大额支付或核心系统变更时，必须由人类管理员在 App 端点击核准。'
  },
  {
    title: '知识盲区求助',
    icon: Zap,
    desc: 'AI 遇到无法解决的难题时，可主动发出信号由人类导师接管对话。'
  },
  {
    title: '数据脱敏 & 审计',
    icon: ShieldCheck,
    desc: '所有工具调用均经过脱敏处理，并留下完整的 execution_id 全链路审计日志。'
  }
];

const lifecycleSteps = [
  {
    title: '意图激活 (Triggering)',
    desc: '通过 IM 指令、协作委派 or 定时 Cron 任务触发工作流。'
  },
  {
    title: '认知处理循环 (Cognitive Loop)',
    desc: '加载长短期记忆，检索 RAG 知识库，生成多步执行计划。'
  },
  {
    title: '协作与执行 (Execution)',
    desc: '通过 MCP 协议调用工具，或咨询 Global Agent Mesh 中的同事。'
  },
  {
    title: '结果验证与闭环 (Closing)',
    desc: '验证工具输出，提取新事实存入长期记忆，并自动生成 KPI 报表。'
  }
];

const collaborationFeatures = [
  {
    icon: Network,
    title: 'colleague_consult',
    desc: '当数字员工遇到知识盲区时，会自动寻找具备相关技能的“同事”进行实时咨询，获取专业建议。'
  },
  {
    icon: Workflow,
    title: 'task_delegate',
    desc: '支持复杂的任务拆解与下发。主管级 AI 可以将子任务委派给执行级 AI，并实时监控进度。'
  },
  {
    icon: Monitor,
    title: 'traceable_execution',
    desc: '所有协作行为均被记录在区块链式的 Trace 链路中，确保每一项指令的来源与去向可追溯、可审计。'
  }
];

const collaborationTraces = [
  { 
    id: 'TRC-901', 
    from: '财务主管 (AX-709)', 
    to: '行政专员 (BX-102)', 
    action: '委派：月度报表审核', 
    status: 'Completed', 
    time: '2 mins ago',
    icon: CheckCircle2,
    color: 'text-green-400'
  },
  { 
    id: 'TRC-902', 
    from: '行政专员 (BX-102)', 
    to: '知识库助手 (KX-001)', 
    action: '咨询：报销合规性准则', 
    status: 'In Progress', 
    time: 'Just now',
    icon: Clock,
    color: 'text-purple-400'
  },
  { 
    id: 'TRC-903', 
    from: '技术支持 (SX-552)', 
    to: '安全审计员 (QX-990)', 
    action: '拦截：高风险脚本执行尝试', 
    status: 'Blocked', 
    time: '5 mins ago',
    icon: ShieldAlert,
    color: 'text-red-400'
  },
  { 
    id: 'TRC-904', 
    from: '系统路由 (Orchestrator)', 
    to: '数据分析师 (DX-301)', 
    action: '生成：Q4 业绩预测简报', 
    status: 'Completed', 
    time: '12 mins ago',
    icon: CheckCircle2,
    color: 'text-green-400'
  },
  { 
    id: 'TRC-905', 
    from: '数据分析师 (DX-301)', 
    to: '系统路由 (Orchestrator)', 
    action: '汇报：发现 3 个增长异常点', 
    status: 'Completed', 
    time: '11 mins ago',
    icon: CheckCircle2,
    color: 'text-green-400'
  }
];

const futuristicFeatures = [
  {
    icon: Monitor,
    title: 'Computer Use 视觉化操作',
    desc: '不再依赖 API。数字员工可以像人类一样识别 UI 界面，自主操作 ERP、网页及各类办公软件。',
    tag: 'Vision 3.0'
  },
  {
    icon: Users,
    title: 'Agent Swarm 集群协同',
    desc: '引入调度员 (Orchestrator)、执行员 (Executor) 与审计员 (Auditor) 角色，实现真正的集群化作业。',
    tag: 'Alpha'
  }
];
</script>

<style scoped>
@keyframes bounce-slow {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}
.animate-bounce-slow {
  animation: bounce-slow 4s ease-in-out infinite;
}

@keyframes scanline {
  0% { transform: translateY(-100%); }
  100% { transform: translateY(100%); }
}

.matrix-lines::after {
  content: "";
  position: absolute;
  inset: 0;
  background: linear-gradient(to bottom, transparent, rgba(168, 85, 247, 0.1), transparent);
  height: 20%;
  width: 100%;
  animation: scanline 8s linear infinite;
}

.reveal {
  animation: reveal 1s cubic-bezier(0, 0, 0.2, 1) forwards;
}

@keyframes reveal {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

/* Glassmorphism Refinement */
.backdrop-blur-xl {
  backdrop-filter: blur(24px);
}

/* Scroll Snap for smoother experience */
section {
  scroll-margin-top: 80px;
}
</style>

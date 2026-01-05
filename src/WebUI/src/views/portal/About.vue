<template>
  <div class="min-h-screen bg-slate-900 text-white selection:bg-cyan-500/30">
    <PortalHeader />

    <!-- Hero -->
    <header class="pt-32 pb-20 px-4 relative overflow-hidden">
      <div class="absolute top-0 left-1/2 -translate-x-1/2 w-full h-full bg-[radial-gradient(circle_at_center,rgba(6,182,212,0.1)_0%,transparent_70%)] -z-10"></div>
      <div class="max-w-4xl mx-auto text-center">
        <h1 class="text-4xl md:text-6xl font-black mb-8">构建机器人自动化的 <span class="text-cyan-400">未来</span></h1>
        <p class="text-xl text-slate-400 leading-relaxed">
          BotMatrix 诞生于对高效通讯与智能化管理的追求。我们相信，机器人不应只是简单的自动回复工具，而应是能够理解需求、协同工作、并具备自我进化能力的智能节点。
        </p>
      </div>
    </header>

    <!-- Project Vision -->
    <section class="py-24 bg-slate-800/30 relative overflow-hidden">
      <div class="absolute top-0 left-0 w-full h-px bg-gradient-to-r from-transparent via-cyan-500/20 to-transparent"></div>
      <div class="max-w-7xl mx-auto px-4">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-16 items-center">
          <div>
            <h2 class="text-3xl font-bold mb-6">核心愿景</h2>
            <div class="space-y-6">
              <div v-for="vision in visions" :key="vision.title" class="flex gap-4 group">
                <div class="flex-shrink-0 w-12 h-12 bg-cyan-500/10 rounded-xl flex items-center justify-center text-cyan-400 group-hover:bg-cyan-500 group-hover:text-slate-900 transition-all">
                  <component :is="vision.icon" class="w-6 h-6" />
                </div>
                <div>
                  <h4 class="text-xl font-bold mb-2 group-hover:text-cyan-400 transition-colors">{{ vision.title }}</h4>
                  <p class="text-slate-400 leading-relaxed">{{ vision.desc }}</p>
                </div>
              </div>
            </div>
          </div>
          <div class="bg-slate-800 p-8 rounded-[2rem] border border-slate-700 relative overflow-hidden">
            <div class="absolute -top-10 -right-10 w-40 h-40 bg-cyan-500/10 blur-3xl"></div>
            <h3 class="text-2xl font-bold mb-8 flex items-center gap-2">
              <Code class="w-6 h-6 text-cyan-400" />
              技术栈演进
            </h3>
            <div class="space-y-4">
              <div v-for="tech in techStack" :key="tech.name" class="flex items-center justify-between p-4 bg-slate-900 rounded-xl border border-slate-700 hover:border-cyan-500/50 transition-colors group">
                <div class="flex items-center gap-3">
                  <div class="w-2 h-2 rounded-full bg-cyan-500 group-hover:scale-150 transition-transform"></div>
                  <span class="font-bold group-hover:text-cyan-400 transition-colors">{{ tech.name }}</span>
                </div>
                <span class="text-sm text-slate-500">{{ tech.role }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Development Roadmap -->
    <section class="py-24">
      <div class="max-w-7xl mx-auto px-4">
        <div class="text-center mb-16">
          <h2 class="text-3xl md:text-5xl font-bold mb-6">发展路线图</h2>
          <p class="text-slate-400">我们正在稳步推进 BotMatrix 的每一个里程碑。</p>
        </div>
        
        <div class="relative">
          <!-- Timeline line -->
          <div class="absolute left-1/2 -translate-x-1/2 top-0 bottom-0 w-px bg-slate-800 hidden md:block"></div>
          
          <div class="space-y-12">
            <div v-for="(milestone, index) in roadmap" :key="milestone.title" 
                 class="relative flex flex-col md:flex-row gap-8 items-center"
                 :class="index % 2 === 0 ? 'md:flex-row-reverse' : ''">
              <!-- Point -->
              <div class="absolute left-1/2 -translate-x-1/2 w-4 h-4 rounded-full bg-cyan-500 shadow-[0_0_15px_rgba(6,182,212,0.5)] z-10 hidden md:block"></div>
              
              <div class="flex-1 w-full md:w-1/2">
                <div class="bg-slate-800/50 p-6 rounded-2xl border border-slate-700 hover:border-cyan-500/30 transition-all"
                     :class="index % 2 === 0 ? 'text-left md:text-left' : 'text-left md:text-right'">
                  <div class="text-cyan-400 font-mono text-sm mb-2">{{ milestone.date }}</div>
                  <h4 class="text-xl font-bold mb-3">{{ milestone.title }}</h4>
                  <p class="text-slate-400 text-sm leading-relaxed">{{ milestone.desc }}</p>
                </div>
              </div>
              <div class="flex-1 hidden md:block"></div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <PortalFooter />
  </div>
</template>

<script setup lang="ts">
import PortalHeader from '@/components/layout/PortalHeader.vue';
import PortalFooter from '@/components/layout/PortalFooter.vue';
import { Layers, Shield, Code, Globe } from 'lucide-vue-next';

const visions = [
  {
    icon: Globe,
    title: 'Global Agent Mesh',
    desc: '构建去中心化的协作网络，实现跨企业、跨域的技能机器人发现与协作，打破智能孤岛。'
  },
  {
    icon: Layers,
    title: 'MCP & 双栈架构',
    desc: '深度集成 Model Context Protocol，完美兼容传统 IM 协议与现代 AI 协议，实现能力的彻底解耦。'
  },
  {
    icon: Shield,
    title: '隐私堡垒 (Privacy Bastion)',
    desc: '全链路自动脱敏敏感信息，确保私有数据不出域，让企业敢于向 AI 开放核心数据。'
  }
];

const techStack = [
  { name: 'Golang', role: '高性能分布式核心' },
  { name: 'Vue 3 + TS', role: '响应式管理终端' },
  { name: 'PostgreSQL + pgvector', role: '向量认知记忆库' },
  { name: 'Docker + K8s', role: '容器化弹性部署' },
  { name: 'Model Context Protocol', role: '标准化能力扩展' }
];

const roadmap = [
  {
    date: '2025 Q1-Q2 (已达成)',
    title: '语义感知与自主能力',
    desc: '实现认知记忆系统、联邦搜索与 B2B 安全调用，从“规则匹配”进化为语义理解。'
  },
  {
    date: '2026 Q3-Q4 (当前阶段)',
    title: '数字员工 HR 系统',
    desc: '引入工号、职位、KPI 考核与虚拟薪资，将 AI 机器人转化为可管理的标准人力资源。'
  },
  {
    date: '2027',
    title: '智能体经济与协作',
    desc: '建立 BotMatrix Hub 技能市场，支持跨组织的大规模智能体蜂群协作与即时授权结算。'
  },
  {
    date: '2028+',
    title: 'Agent OS',
    desc: 'AI 成为数字世界一等公民，具备全环境 UI 操控能力，连接 IoT 实现对物理世界的智能管控。'
  }
];
</script>

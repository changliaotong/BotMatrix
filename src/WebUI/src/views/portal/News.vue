<template>
  <div class="min-h-screen bg-slate-900 text-white selection:bg-cyan-500/30">
    <PortalHeader />

    <!-- Hero -->
    <header class="pt-32 pb-16 px-4 relative overflow-hidden">
      <div class="absolute top-0 left-1/2 -translate-x-1/2 w-full h-full bg-[radial-gradient(circle_at_center,rgba(6,182,212,0.1)_0%,transparent_70%)] -z-10"></div>
      <div class="max-w-4xl mx-auto text-center">
        <h1 class="text-4xl md:text-6xl font-black mb-8">动态更新</h1>
        <p class="text-xl text-slate-400 leading-relaxed">
          关注 BotMatrix 的最新进展、版本公告与技术分享。
        </p>
      </div>
    </header>

    <!-- News List -->
    <section class="py-12 pb-24 px-4">
      <div class="max-w-4xl mx-auto space-y-12">
        <article v-for="item in newsItems" :key="item.id" class="group relative bg-slate-800/30 rounded-[2rem] border border-slate-700 overflow-hidden hover:border-cyan-500/30 transition-all">
          <div class="flex flex-col md:flex-row">
            <div class="md:w-1/3 aspect-[4/3] bg-slate-800 relative overflow-hidden">
              <div class="absolute inset-0 bg-gradient-to-br from-cyan-500/20 to-purple-500/20 group-hover:scale-110 transition-transform duration-500"></div>
              <div class="absolute inset-0 flex items-center justify-center">
                <component :is="item.icon" class="w-16 h-16 text-cyan-400 opacity-50" />
              </div>
            </div>
            <div class="flex-1 p-8">
              <div class="flex items-center gap-3 mb-4">
                <span class="px-3 py-1 rounded-full bg-cyan-500/10 text-cyan-400 text-xs font-bold">{{ item.category }}</span>
                <span class="text-slate-500 text-xs">{{ item.date }}</span>
              </div>
              <h2 class="text-2xl font-bold mb-4 group-hover:text-cyan-400 transition-colors">{{ item.title }}</h2>
              <p class="text-slate-400 mb-6 leading-relaxed">{{ item.summary }}</p>
              <button class="flex items-center gap-2 text-cyan-400 font-bold hover:gap-3 transition-all">
                阅读更多 <ArrowRight class="w-4 h-4" />
              </button>
            </div>
          </div>
        </article>
      </div>
    </section>

    <PortalFooter />
  </div>
</template>

<script setup lang="ts">
import PortalHeader from '@/components/layout/PortalHeader.vue';
import PortalFooter from '@/components/layout/PortalFooter.vue';
import { 
  ArrowRight, 
  Newspaper, 
  Zap, 
  Rocket, 
  Cpu, 
  ShieldCheck 
} from 'lucide-vue-next';

const newsItems = [
  {
    id: 1,
    category: '技术突破',
    date: '2024-01-04',
    title: 'Global Agent Mesh 协议正式发布：打破企业间 AI 孤岛',
    summary: '我们推出了基于去中心化握手协议的 Agent 协作网络，支持跨企业的数字员工身份委派与技能发现。',
    icon: Rocket
  },
  {
    id: 2,
    category: '产品更新',
    date: '2023-12-28',
    title: '早喵机器人“超级群管”插件 2.0 上线',
    summary: '新增积分经济联动系统与阶梯式违规禁言逻辑，让社群管理从“堵”转变为“疏”，大幅提升活跃度。',
    icon: Zap
  },
  {
    id: 3,
    category: '架构演进',
    date: '2023-12-15',
    title: 'Nexus Guard 引入分布式安全审计引擎',
    summary: '通过边缘节点并行审计技术，Nexus Guard 现在能够处理万人群聊中的瞬时消息洪峰，延迟低于 10ms。',
    icon: ShieldCheck
  },
  {
    id: 4,
    category: '愿景发布',
    date: '2023-12-01',
    title: '从机器人到数字员工：BotMatrix 的下一个十年',
    summary: '我们宣布将产品重心全面转向“AI 原生智能体操作系统”，引入工号、职位及 KPI 考核等企业级人力资源管理概念。',
    icon: Cpu
  }
];
</script>

<template>
  <div class="min-h-screen bg-slate-900 text-white selection:bg-cyan-500/30 relative overflow-hidden">
    <div class="absolute top-0 left-0 w-full h-full bg-[radial-gradient(circle_at_top_right,rgba(168,85,247,0.05)_0%,transparent_50%)] pointer-events-none"></div>
    <div class="absolute bottom-0 right-0 w-full h-full bg-[radial-gradient(circle_at_bottom_left,rgba(6,182,212,0.05)_0%,transparent_50%)] pointer-events-none"></div>
    
    <PortalHeader />

    <header class="pt-32 pb-16 px-4 text-center relative reveal">
      <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-cyan-500/10 text-cyan-400 text-xs font-bold mb-6 border border-cyan-500/20">
        <Sparkles class="w-4 h-4" /> 灵活方案，无限可能
      </div>
      <h1 class="text-4xl md:text-7xl font-black mb-6 tracking-tighter">选择适合您的 <span class="text-cyan-400">方案</span></h1>
      <p class="text-slate-400 max-w-2xl mx-auto text-xl font-light">
        无论您是个人开发者、兴趣社群，还是大中型企业，BotMatrix 都能为您提供最可靠的支撑。
      </p>
    </header>

    <section class="py-20 px-4 relative z-10">
      <div class="max-w-7xl mx-auto grid grid-cols-1 md:grid-cols-3 gap-8">
        <!-- Community Edition -->
        <div class="p-10 rounded-[3rem] bg-slate-800/30 border border-slate-700/50 flex flex-col h-full hover:border-slate-500 transition-all hover:-translate-y-2 group reveal" style="animation-delay: 0.1s">
          <div class="mb-8">
            <h3 class="text-2xl font-bold mb-2 group-hover:text-white transition-colors">社区版 (Community)</h3>
            <p class="text-slate-500 text-sm">适合个人开发者与小型社群</p>
          </div>
          <div class="text-4xl font-black mb-10">免费 <span class="text-sm font-normal text-slate-500">/ 永久开源</span></div>
          <ul class="space-y-5 mb-12 flex-1 text-sm">
            <li v-for="item in communityFeatures" :key="item" class="flex items-center gap-3 text-slate-400 group-hover:text-slate-200 transition-colors">
              <CheckCircle2 class="w-4 h-4 text-cyan-500/50" />
              {{ item }}
            </li>
          </ul>
          <button class="w-full py-5 bg-slate-700/50 hover:bg-slate-700 text-white rounded-2xl font-bold transition-all border border-slate-600">
            立即开始
          </button>
        </div>

        <!-- Business Edition -->
        <div class="p-10 rounded-[3rem] bg-gradient-to-br from-cyan-500/10 to-blue-500/10 border border-cyan-500/30 flex flex-col h-full relative overflow-hidden transform lg:scale-105 shadow-2xl shadow-cyan-500/10 reveal" style="animation-delay: 0.2s">
          <div class="absolute top-8 right-8 px-4 py-1 bg-cyan-500 rounded-full text-[10px] font-bold text-slate-900 uppercase tracking-wider">Most Popular</div>
          <div class="mb-8">
            <h3 class="text-2xl font-bold mb-2 text-cyan-400">商业版 (Business)</h3>
            <p class="text-slate-500 text-sm">适合大中型企业数字化转型</p>
          </div>
          <div class="text-4xl font-black mb-10 text-white">敬请期待 <span class="text-sm font-normal text-slate-500">/ Enterprise Ready</span></div>
          <ul class="space-y-5 mb-12 flex-1 text-sm">
            <li v-for="item in proFeatures" :key="item" class="flex items-center gap-3 text-slate-200">
              <Sparkles class="w-4 h-4 text-cyan-400" />
              {{ item }}
            </li>
          </ul>
          <button @click="openAction('contact', 'business')" class="w-full py-5 bg-cyan-500 hover:bg-cyan-400 text-slate-900 rounded-2xl font-black transition-all shadow-xl shadow-cyan-500/30">
            咨询商业授权
          </button>
        </div>

        <!-- Matrix Edition -->
        <div class="p-10 rounded-[3rem] bg-gradient-to-br from-purple-500/10 to-indigo-500/10 border border-purple-500/30 flex flex-col h-full hover:border-purple-500/50 transition-all hover:-translate-y-2 group reveal shadow-2xl shadow-purple-500/5" style="animation-delay: 0.3s">
          <div class="mb-8">
            <h3 class="text-2xl font-bold mb-2 text-purple-400">矩阵版 (Matrix Ultimate)</h3>
            <p class="text-slate-500 text-sm">开启生产关系革命的终极形态</p>
          </div>
          <div class="text-4xl font-black mb-10">定制 <span class="text-sm font-normal text-slate-500">/ Custom Matrix</span></div>
          <ul class="space-y-5 mb-12 flex-1 text-sm">
            <li v-for="item in matrixFeatures" :key="item" class="flex items-center gap-3 text-slate-300 group-hover:text-white transition-colors">
              <Zap class="w-4 h-4 text-purple-400" />
              {{ item }}
            </li>
          </ul>
          <button @click="openAction('contact', 'matrix')" class="w-full py-5 bg-purple-600 hover:bg-purple-500 text-white rounded-2xl font-black transition-all shadow-xl shadow-purple-500/20">
            开启矩阵革命
          </button>
        </div>
      </div>

      <!-- Comparison Table -->
      <div class="max-w-6xl mx-auto mt-32 px-4">
        <h2 class="text-3xl font-bold mb-12 text-center">方案详细对比</h2>
        <div class="overflow-x-auto">
          <table class="w-full border-collapse">
            <thead>
              <tr class="border-b border-slate-800">
                <th class="py-6 px-4 text-left text-slate-500 font-medium">功能特性</th>
                <th class="py-6 px-4 text-center text-white font-bold">社区版</th>
                <th class="py-6 px-4 text-center text-cyan-400 font-bold">商业版</th>
                <th class="py-6 px-4 text-center text-purple-400 font-bold">矩阵版</th>
              </tr>
            </thead>
            <tbody class="text-sm">
              <tr v-for="row in comparisonRows" :key="row.feature" class="border-b border-slate-800/50 hover:bg-slate-800/20 transition-colors">
                <td class="py-6 px-4 text-slate-300">{{ row.feature }}</td>
                <td class="py-6 px-4 text-center">
                  <Check v-if="row.community" class="w-5 h-5 text-slate-600 mx-auto" />
                  <span v-else class="text-slate-700">-</span>
                </td>
                <td class="py-6 px-4 text-center">
                  <Check v-if="row.pro" class="w-5 h-5 text-cyan-500 mx-auto" />
                  <span v-else class="text-slate-700">-</span>
                </td>
                <td class="py-6 px-4 text-center">
                  <Check v-if="row.matrix" class="w-5 h-5 text-purple-500 mx-auto" />
                  <span v-else class="text-slate-700">-</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Pricing FAQ -->
      <div class="max-w-3xl mx-auto mt-32 px-4 pb-20">
        <h2 class="text-3xl font-bold mb-12 text-center">常见问题</h2>
        <div class="space-y-4">
          <div v-for="(item, index) in faqs" :key="index" 
            class="bg-slate-800/30 border border-slate-700 rounded-2xl overflow-hidden">
            <button @click="activeFaq = activeFaq === index ? -1 : index" 
              class="w-full p-6 text-left flex justify-between items-center hover:bg-slate-800/50 transition-colors">
              <span class="font-bold">{{ item.q }}</span>
              <ChevronDown class="w-5 h-5 transition-transform" :class="{ 'rotate-180': activeFaq === index }" />
            </button>
            <div v-show="activeFaq === index" class="p-6 pt-0 text-slate-400 text-sm leading-relaxed border-t border-slate-800/50">
              {{ item.a }}
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Contact Modal -->
    <div v-if="activeModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-slate-950/80 backdrop-blur-sm" @click="activeModal = null"></div>
      <div class="relative w-full max-w-md bg-slate-900 border border-slate-700 rounded-[2rem] p-8 shadow-2xl overflow-hidden">
        <div class="absolute top-0 left-0 w-full h-1" :class="modalType === 'matrix' ? 'bg-purple-500' : 'bg-cyan-500'"></div>
        <button @click="activeModal = null" class="absolute top-6 right-6 p-2 hover:bg-slate-800 rounded-full transition-colors">
          <X class="w-5 h-5" />
        </button>
        
        <div class="text-center">
          <div class="w-16 h-16 rounded-2xl flex items-center justify-center mx-auto mb-6" :class="modalType === 'matrix' ? 'bg-purple-500/10 text-purple-500' : 'bg-cyan-500/10 text-cyan-500'">
            <Sparkles class="w-8 h-8" />
          </div>
          <h3 class="text-2xl font-bold mb-4">{{ modalType === 'matrix' ? '开启矩阵革命' : '联系商务合作' }}</h3>
          <p class="text-slate-400 mb-8 leading-relaxed">请扫描下方二维码或添加企业微信，我们的客户经理将为您提供一对一的咨询服务。</p>
          
          <div class="bg-slate-800/50 rounded-2xl p-6 mb-8 flex flex-col items-center gap-4 border border-slate-700">
            <div class="w-32 h-32 bg-white rounded-xl p-2 flex items-center justify-center">
              <!-- Placeholder for QR Code -->
              <div class="w-full h-full bg-slate-200 rounded flex items-center justify-center text-slate-400 text-[10px] text-center px-2">
                QR CODE<br/>{{ modalType === 'matrix' ? 'MATRIX' : 'COMMERCIAL' }}
              </div>
            </div>
            <div class="text-sm font-bold tracking-wider" :class="modalType === 'matrix' ? 'text-purple-400' : 'text-cyan-400'">企业微信：BotMatrix_Sales</div>
          </div>

          <button @click="activeModal = null" class="w-full py-4 rounded-xl font-bold transition-all" :class="modalType === 'matrix' ? 'bg-purple-600 hover:bg-purple-500 text-white' : 'bg-cyan-500 hover:bg-cyan-400 text-slate-900'">
            好的
          </button>
        </div>
      </div>
    </div>

    <PortalFooter />
  </div>
</template>

<script setup lang="ts">
import PortalHeader from '@/components/layout/PortalHeader.vue';
import PortalFooter from '@/components/layout/PortalFooter.vue';
import { CheckCircle2, Sparkles, Check, ChevronDown, X, Zap } from 'lucide-vue-next';
import { ref } from 'vue';

const activeFaq = ref(-1);
const activeModal = ref(false);
const modalType = ref('business');

const openAction = (type: string, subType: string = 'business') => {
  if (type === 'contact') {
    modalType.value = subType;
    activeModal.value = true;
  }
};

const comparisonRows = [
  { feature: '机器人矩阵核心架构', community: true, pro: true, matrix: true },
  { feature: '多平台账号接入 (OneBot)', community: true, pro: true, matrix: true },
  { feature: '技能插件市场接入 (MCP)', community: true, pro: true, matrix: true },
  { feature: '基础认知记忆 (RAG)', community: true, pro: true, matrix: true },
  { feature: '数字员工身份与工号体系', community: false, pro: true, matrix: true },
  { feature: 'KPI 绩效考核与薪资体系', community: false, pro: true, matrix: true },
  { feature: 'Privacy Bastion 隐私堡垒', community: false, pro: true, matrix: true },
  { feature: 'Global Agent Mesh 协作协议', community: false, pro: false, matrix: true },
  { feature: 'Multi-Agent Swarm 集群调度', community: false, pro: false, matrix: true },
  { feature: 'Computer Use 视觉化操作', community: false, pro: false, matrix: true },
  { feature: '私有化 Hub 节点授权', community: false, pro: false, matrix: true }
];

const faqs = [
  {
    q: '社区版和商业版在核心功能上有区别吗？',
    a: '社区版包含了 BotMatrix 的所有基础核心功能，足以支持个人开发者和小型社群的日常使用。商业版更多针对企业级场景，提供了多租户、大规模集群支持、高级审计及专属技术支持。'
  },
  {
    q: '矩阵版 (Matrix Ultimate) 的定位是什么？',
    a: '矩阵版是为追求极致自动化与生产力革命的头部企业设计的。它包含了 Global Agent Mesh、智能体集群调度等前沿技术，旨在构建跨组织的智能体网络。'
  },
  {
    q: '支持私有化部署在自己的内网环境吗？',
    a: '支持。BotMatrix 天生为私有化部署而设计。商业版还提供了更完善的离线模型接入方案。'
  },
  {
    q: '专业版是按年付费还是永久授权？',
    a: '我们提供灵活的付费方案，包括按年订阅和针对大型企业的永久买断方案，具体请联系我们的商务经理咨询。'
  }
];

const communityFeatures = [
  '机器人矩阵基础管理',
  '多平台通讯账号接入',
  '基础技能插件库',
  '内置 RAG 长期记忆',
  '分布式节点一键部署'
];

const proFeatures = [
  '数字员工 HR 管理系统',
  'KPI 绩效考核体系',
  '可视化业务流引擎',
  '企业级隐私数据脱敏',
  '多租户权限隔离',
  '7x24 专属技术顾问'
];

const matrixFeatures = [
  'Global Agent Mesh 协议',
  'Multi-Agent Swarm 集群',
  'Computer Use 视觉执行',
  'B2B 联邦身份认证',
  '跨域知识库安全共享',
  '生产关系革命咨询服务'
];
</script>

<style scoped>
.reveal {
  opacity: 0;
  animation: reveal 0.8s cubic-bezier(0, 0, 0.2, 1) forwards;
}

@keyframes reveal {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

/* Glassmorphism Refinement */
.backdrop-blur-sm {
  backdrop-filter: blur(8px);
}

/* Custom Table Styles */
table tr {
  transition: all 0.3s ease;
}

table tr:hover td {
  color: white;
}
</style>

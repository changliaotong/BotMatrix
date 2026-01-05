<template>
  <div class="min-h-screen bg-slate-900 text-white selection:bg-cyan-500/30">
    <PortalHeader />

    <header class="pt-32 pb-16 px-4 relative overflow-hidden">
      <div class="absolute top-0 left-1/2 -translate-x-1/2 w-full h-full bg-[radial-gradient(circle_at_center,rgba(6,182,212,0.08)_0%,transparent_70%)] -z-10"></div>
      <div class="max-w-7xl mx-auto text-center">
        <h1 class="text-4xl md:text-6xl font-black mb-6">文档中心</h1>
        <p class="text-slate-400 max-w-2xl mx-auto">
          从环境搭建到插件开发，我们为您准备了详尽的指南。
        </p>
      </div>
    </header>

    <section class="py-12 px-4">
      <div class="max-w-7xl mx-auto grid grid-cols-1 md:grid-cols-3 gap-8">
        <div v-for="guide in guides" :key="guide.title" class="p-8 bg-slate-800/50 border border-slate-700 rounded-3xl hover:border-cyan-500/50 transition-all group">
          <div class="w-12 h-12 bg-cyan-500/10 text-cyan-500 rounded-xl flex items-center justify-center mb-6 group-hover:bg-cyan-500 group-hover:text-slate-900 transition-colors">
            <component :is="guide.icon" class="w-6 h-6" />
          </div>
          <h3 class="text-xl font-bold mb-4">{{ guide.title }}</h3>
          <p class="text-slate-400 text-sm leading-relaxed mb-6">{{ guide.desc }}</p>
          <a href="#" class="inline-flex items-center gap-2 text-cyan-400 font-bold hover:gap-3 transition-all">
            开始阅读 <ArrowRight class="w-4 h-4" />
          </a>
        </div>
      </div>
    </section>

    <section class="py-20 bg-slate-800/20 border-y border-slate-800">
      <div class="max-w-4xl mx-auto px-4">
        <h2 class="text-2xl font-bold mb-8 flex items-center gap-3">
          <Terminal class="w-6 h-6 text-cyan-500" />
          快速部署指令 (Docker)
        </h2>
        <div class="bg-black/50 rounded-2xl p-6 border border-slate-700 font-mono text-sm group relative">
          <button class="absolute top-4 right-4 p-2 bg-slate-800 hover:bg-slate-700 rounded-lg text-slate-400 transition-colors opacity-0 group-hover:opacity-100">
            <Copy class="w-4 h-4" />
          </button>
          <div class="text-cyan-400"># 克隆项目</div>
          <div class="mb-4 text-slate-300">git clone https://github.com/changliaotong/BotMatrix.git</div>
          <div class="text-cyan-400"># 启动服务</div>
          <div class="mb-4 text-slate-300">cd BotMatrix && docker-compose up -d</div>
          <div class="text-cyan-400"># 访问管理后台</div>
          <div class="text-slate-300">http://localhost:5173</div>
        </div>
      </div>
    </section>

    <!-- Technical Architecture Deep Dive -->
    <section class="py-24 bg-slate-900">
      <div class="max-w-7xl mx-auto px-4">
        <h2 class="text-3xl md:text-4xl font-black mb-16 text-center">核心技术理念</h2>
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-16">
          <div v-for="concept in technicalPhilosophy" :key="concept.title" class="space-y-6">
            <div class="flex items-center gap-4">
              <div class="w-10 h-10 bg-cyan-500/10 text-cyan-500 rounded-lg flex items-center justify-center">
                <component :is="concept.icon" class="w-5 h-5" />
              </div>
              <h3 class="text-2xl font-bold">{{ concept.title }}</h3>
            </div>
            <p class="text-slate-400 leading-relaxed">{{ concept.desc }}</p>
            <div class="grid grid-cols-2 gap-4">
              <div v-for="item in concept.items" :key="item" class="flex items-center gap-2 text-sm text-slate-500">
                <CheckCircle2 class="w-4 h-4 text-cyan-500" /> {{ item }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Roadmap & Future -->
    <section class="py-24 bg-slate-800/20">
      <div class="max-w-7xl mx-auto px-4">
        <div class="bg-slate-900/50 border border-slate-800 rounded-[3rem] p-12 md:p-20 relative overflow-hidden">
          <div class="absolute top-0 right-0 w-96 h-96 bg-purple-500/5 blur-[100px] -z-10"></div>
          <h2 class="text-3xl md:text-5xl font-black mb-12">2026 技术路线图</h2>
          <div class="space-y-12 relative before:absolute before:left-0 before:top-2 before:bottom-2 before:w-px before:bg-slate-800 ml-4 pl-12">
            <div v-for="phase in roadmap" :key="phase.title" class="relative">
              <div class="absolute -left-[3.25rem] top-0 w-4 h-4 rounded-full bg-cyan-500 border-4 border-slate-900 shadow-lg shadow-cyan-500/50"></div>
              <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-cyan-500/10 text-cyan-400 text-xs font-bold mb-4 uppercase tracking-widest">{{ phase.time }}</div>
              <h3 class="text-2xl font-bold mb-4">{{ phase.title }}</h3>
              <p class="text-slate-400 mb-6 max-w-2xl">{{ phase.desc }}</p>
              <div class="flex flex-wrap gap-3">
                <span v-for="tag in phase.tags" :key="tag" class="px-3 py-1 bg-slate-800 border border-slate-700 rounded-lg text-xs text-slate-500 font-medium">{{ tag }}</span>
              </div>
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
import { 
  Rocket, 
  Code2, 
  ArrowRight, 
  Terminal, 
  Copy,
  Globe,
  Users,
  BrainCircuit,
  ShieldCheck,
  Cpu,
  Layers,
  Zap,
  ShieldAlert,
  CheckCircle2,
  Calendar,
  LineChart,
  MessageSquare,
  Network,
  MonitorDot
} from 'lucide-vue-next';

const guides = [
  {
    icon: Globe,
    title: 'Global Agent Mesh',
    desc: '了解如何打破企业数据孤岛，通过安全握手实现跨企业的数字员工身份委派与技能共享。'
  },
  {
    icon: Network,
    title: '智能体集群 (Swarm)',
    desc: '探索如何协调成百上千个微型智能体，通过群体智能解决超大规模的并行任务处理与协同进化。'
  },
  {
    icon: MonitorDot,
    title: '计算机使用 (Computer Use)',
    desc: '赋予数字员工操作桌面、浏览器与终端的能力，打破 API 限制，实现真正的全场景端到端自动化。'
  },
  {
    icon: Users,
    title: '数字员工系统',
    desc: '从“机器人”到“虚拟雇员”：配置工号、职位、KPI 考核与基于 Token 的虚拟薪资系统。'
  },
  {
    icon: BrainCircuit,
    title: '认知记忆与 RAG',
    desc: '深入理解基于向量数据库的长期记忆引擎，实现具备事实提取与自主学习能力的智能体。'
  },
  {
    icon: Code2,
    title: 'MCP 插件开发',
    desc: '使用 Model Context Protocol 标准化你的工具集，让一次编写的技能在全球机器人网络通用。'
  }
];

const technicalPhilosophy = [
  {
    icon: Cpu,
    title: '数字员工的“五感六觉”',
    desc: '为了实现“像真人一样工作”，我们将数字员工划分为身份、感知、思维、记忆、技能、协作六大维度。',
    items: ['IdentityGORM 身份模型', 'Intent Dispatcher 意图分发', 'Cognitive Memory 认知记忆', 'MCP Toolset 技能集', 'Agent Mesh 协作协议', 'Auto-Learning 自主进化']
  },
  {
    icon: Layers,
    title: '认知处理循环 (Cognitive Loop)',
    desc: '每一个任务的处理都经过：环境感知 -> 任务规划 -> 工具执行 -> 结果验证的严密循环。',
    items: ['Short-term Memory 会话上下文', 'Long-term Memory 事实片段', 'RAG 业务知识库检索', 'HITL 人工干预机制']
  },
  {
    icon: Zap,
    title: 'Agent-to-Agent (A2A) 协议',
    desc: '定义了智能体之间进行任务委派、咨询与反馈的标准 JSON 格式，支持跨企业安全握手。',
    items: ['双向 JWT 签名校验', 'B2B Gateway 权限网关', 'DID 去中心化身份', '任务全链路 Trace']
  },
  {
    icon: ShieldAlert,
    title: '安全与伦理防护栏',
    desc: '内置 Privacy Bastion 隐私堡垒，确保在调用公有云 LLM 之前完成数据脱敏。',
    items: ['PII 敏感数据识别', '数据流向审计', '操作权限最小化', 'Ethics Guardrail 伦理检查']
  },
  {
    icon: Rocket,
    title: 'Vision 3.0: 生产力奇点',
    desc: '从“工具”到“生命体”：通过 Swarm 集群与 Computer Use 赋予数字员工跨越数字世界的执行力。',
    items: ['Swarm Intelligence 群体智能', 'Computer Use OS 交互', 'Long-Context 全域协同', 'Self-Optimization 自我优化']
  }
];

const roadmap = [
  {
    time: '2026 Q1',
    title: '矩阵基座增强',
    desc: '完善分布式 Worker 节点的动态扩容机制，支持更复杂的 MCP 工具流组合与 A2A 基础协议。',
    tags: ['Worker v2.0', 'MCP SDK', 'Basic Agent Mesh']
  },
  {
    time: '2026 Q2',
    title: '认知记忆进化',
    desc: '上线自主学习系统 (Auto-Learning)，支持从 PDF/Doc/Excel 中主动摄取并更新认知片段。',
    tags: ['Autonomous Memory', 'RAG v3.0', 'Fact Extraction']
  },
  {
    time: '2026 Q3',
    title: '全球协作网络 (GA Mesh)',
    desc: '发布 B2B 协作网关，支持跨企业数字员工身份映射与技能审批流，构建分布式智能网格。',
    tags: ['B2B Gateway', 'Inter-Enterprise Mesh', 'Skill Market']
  },
  {
    time: '2026 Q4',
    title: 'Vision 3.0 矩阵革命',
    desc: '正式发布智能体集群 (Swarm) 与计算机使用 (Computer Use) 模块，实现生产关系的终极重塑。',
    tags: ['Swarm Intelligence', 'OS Interaction', 'Matrix Revolution']
  }
];
</script>

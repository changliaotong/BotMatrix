<template>
  <div class="min-h-screen bg-slate-900 text-white selection:bg-blue-500/30">
    <!-- Independent Security Header -->
    <nav class="fixed top-0 w-full z-50 bg-slate-900/80 backdrop-blur-md border-b border-blue-500/10">
      <div class="max-w-7xl mx-auto px-4 h-20 flex justify-between items-center">
        <div class="flex items-center gap-3 group cursor-pointer" @click="scrollTo('hero')">
          <div class="w-10 h-10 bg-blue-600 rounded-lg flex items-center justify-center shadow-lg shadow-blue-500/20 group-hover:rotate-12 transition-transform">
            <ShieldAlert class="w-6 h-6 text-white" />
          </div>
          <div class="flex flex-col">
            <span class="text-xl font-black tracking-tighter text-white">NEXUS GUARD</span>
            <span class="text-[10px] text-blue-500 font-bold tracking-[0.2em] uppercase leading-none">Security Protocol</span>
          </div>
        </div>
        
        <div class="hidden md:flex items-center gap-8 text-sm font-bold text-slate-400">
          <button @click="scrollTo('features')" class="hover:text-blue-500 transition-colors">核心防护</button>
          <button @click="scrollTo('dashboard')" class="hover:text-blue-500 transition-colors">态势感知</button>
          <button @click="scrollTo('architecture')" class="hover:text-blue-500 transition-colors">底层架构</button>
          <button @click="scrollTo('faq')" class="hover:text-blue-500 transition-colors">安全 FAQ</button>
          <router-link to="/" class="px-4 py-2 rounded-lg bg-slate-800 text-slate-300 hover:bg-slate-700 transition-all border border-slate-700">
            返回主站
          </router-link>
          <button @click="openAction('deploy')" class="px-6 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg transition-all shadow-lg shadow-blue-500/20">
            立即部署
          </button>
        </div>
      </div>
    </nav>

    <!-- Hero Section -->
    <header id="hero" class="pt-32 pb-20 px-4 relative overflow-hidden">
      <div class="absolute top-0 left-1/2 -translate-x-1/2 w-[1000px] h-[600px] bg-blue-500/10 blur-[120px] rounded-full -z-10"></div>
      <div class="max-w-7xl mx-auto flex flex-col md:flex-row items-center gap-12">
        <div class="flex-1 text-center md:text-left">
          <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-blue-500/10 border border-blue-500/20 text-blue-400 text-xs font-bold mb-6">
            <ShieldAlert class="w-3 h-3" />
            官方安全机器人
          </div>
          <h1 class="text-5xl md:text-7xl font-black mb-6 leading-tight">
            Nexus <span class="text-blue-500">Guard</span>
          </h1>
          <p class="text-xl text-slate-400 mb-10 max-w-xl leading-relaxed">
            专为大型社群与企业环境设计的安全防护机器人。实时监控、智能审计，为您的通讯矩阵保驾护航。
          </p>
          <div class="flex flex-wrap gap-4 justify-center md:justify-start">
            <button @click="openAction('deploy')" class="px-8 py-4 bg-blue-600 hover:bg-blue-500 text-white rounded-xl text-lg font-bold transition-all shadow-lg shadow-blue-500/20 flex items-center gap-2">
              <ShieldCheck class="w-5 h-5" />
              开启守护
            </button>
            <button @click="openAction('whitepaper')" class="px-8 py-4 bg-slate-800 hover:bg-slate-700 text-white rounded-xl text-lg font-bold transition-all border border-slate-700">
              安全白皮书
            </button>
          </div>
        </div>
        <div class="flex-1 relative">
          <div class="w-64 h-64 md:w-80 md:h-80 bg-gradient-to-br from-blue-400 to-indigo-600 rounded-[3rem] -rotate-6 absolute inset-0 blur-2xl opacity-20"></div>
          <div class="relative w-64 h-64 md:w-80 md:h-80 bg-slate-800 rounded-[3rem] border border-slate-700 overflow-hidden flex items-center justify-center group hover:scale-105 transition-transform duration-500">
            <Lock class="w-32 h-32 md:w-40 md:h-40 text-blue-500 group-hover:scale-110 transition-transform duration-500" />
          </div>
        </div>
      </div>
    </header>

    <!-- Core Security Features -->
    <section id="features" class="py-24 bg-slate-800/30">
      <div class="max-w-7xl mx-auto px-4">
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
          <div v-for="feature in securityFeatures" :key="feature.title" class="p-8 bg-slate-900/50 rounded-2xl border border-slate-700 hover:border-blue-500/50 transition-all hover:-translate-y-2 group">
            <div class="w-10 h-10 bg-blue-500/10 text-blue-500 rounded-lg flex items-center justify-center mb-6 group-hover:bg-blue-500 group-hover:text-slate-900 transition-all">
              <component :is="feature.icon" class="w-5 h-5" />
            </div>
            <h3 class="font-bold mb-3 group-hover:text-blue-400 transition-colors">{{ feature.title }}</h3>
            <p class="text-slate-400 text-sm leading-relaxed">{{ feature.desc }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Security Dashboard Preview -->
    <section id="dashboard" class="py-24">
      <div class="max-w-7xl mx-auto px-4">
        <div class="flex flex-col lg:flex-row gap-16 items-center">
          <div class="flex-1">
            <h2 class="text-3xl md:text-5xl font-black mb-8 leading-tight">全方位的 <br/><span class="text-blue-500">安全态势感知</span></h2>
            <div class="space-y-8">
              <div v-for="item in securityMetrics" :key="item.title" class="flex gap-6 p-6 rounded-2xl bg-slate-800/50 border border-slate-700 hover:bg-slate-800 transition-all group">
                <div class="flex-shrink-0 w-12 h-12 bg-blue-500/10 rounded-xl flex items-center justify-center text-blue-500 group-hover:bg-blue-500 group-hover:text-slate-900 transition-all">
                  <component :is="item.icon" class="w-6 h-6" />
                </div>
                <div>
                  <h4 class="text-xl font-bold mb-2">{{ item.title }}</h4>
                  <p class="text-slate-400 text-sm leading-relaxed">{{ item.desc }}</p>
                </div>
              </div>
            </div>
          </div>
          
          <div class="flex-1 w-full">
            <div class="relative bg-slate-800 rounded-3xl border border-slate-700 shadow-2xl overflow-hidden aspect-[4/3] group">
              <div class="absolute inset-0 bg-blue-500/5 group-hover:bg-blue-500/10 transition-colors"></div>
              <!-- Mock UI -->
              <div class="p-6 h-full flex flex-col">
                <div class="flex items-center justify-between mb-8">
                  <div class="flex items-center gap-2">
                    <div class="w-3 h-3 rounded-full bg-red-500 animate-pulse"></div>
                    <span class="text-sm font-bold text-slate-300 uppercase tracking-widest">Live Security Feed</span>
                  </div>
                  <div class="px-3 py-1 bg-blue-500/10 text-blue-400 text-[10px] font-bold rounded-full border border-blue-500/20">
                    PROTECTION ACTIVE
                  </div>
                </div>
                <div class="flex-1 space-y-4">
                  <div v-for="i in 5" :key="i" class="h-12 bg-slate-900/80 rounded-xl border border-slate-700/50 flex items-center px-4 gap-4 animate-pulse" :style="{ animationDelay: `${i * 200}ms` }">
                    <div class="w-2 h-2 rounded-full bg-blue-500"></div>
                    <div class="flex-1 h-2 bg-slate-800 rounded-full"></div>
                    <div class="w-12 h-2 bg-slate-800 rounded-full"></div>
                  </div>
                </div>
                <div class="mt-6 pt-6 border-t border-slate-700 flex justify-between items-end">
                  <div>
                    <div class="text-xs text-slate-500 uppercase mb-1">Total Threats Blocked</div>
                    <div class="text-3xl font-black text-white">12,842</div>
                  </div>
                  <BarChart3 class="w-12 h-12 text-blue-500 opacity-50" />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Advanced Protection Tech -->
    <section id="architecture" class="py-24 bg-slate-800/20 relative overflow-hidden">
      <div class="absolute -bottom-24 -left-24 w-96 h-96 bg-blue-500/10 blur-[120px] rounded-full"></div>
      <div class="max-w-7xl mx-auto px-4">
        <div class="text-center mb-16">
          <h2 class="text-3xl md:text-5xl font-bold mb-4">底层防御架构</h2>
          <p class="text-slate-400">基于多层过滤与分布式审计引擎。</p>
        </div>
        
        <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
          <div v-for="tech in protectionTech" :key="tech.title" class="p-10 bg-slate-900/60 rounded-[2.5rem] border border-slate-700 hover:border-blue-500/30 transition-all text-center">
            <div class="w-16 h-16 bg-blue-500/10 text-blue-500 rounded-2xl flex items-center justify-center mx-auto mb-8">
              <component :is="tech.icon" class="w-8 h-8" />
            </div>
            <h3 class="text-2xl font-bold mb-4">{{ tech.title }}</h3>
            <p class="text-slate-400 leading-relaxed">{{ tech.desc }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Compliance Section -->
    <section class="py-24">
      <div class="max-w-4xl mx-auto px-4 text-center">
        <div class="inline-flex p-4 bg-blue-500/10 rounded-3xl border border-blue-500/20 mb-8">
          <ShieldCheck class="w-12 h-12 text-blue-500" />
        </div>
        <h2 class="text-3xl font-bold mb-6">合规与隐私承诺</h2>
        <p class="text-slate-400 text-lg leading-relaxed mb-10">
          Nexus Guard 遵循严格的隐私保护协议。所有敏感词库与审计逻辑均支持本地化部署，不上传任何聊天内容至云端，确保您的商业机密与用户隐私万无一失。
        </p>
        <div class="flex flex-wrap justify-center gap-8 opacity-50 grayscale hover:grayscale-0 transition-all duration-700">
          <div v-for="i in 4" :key="i" class="flex items-center gap-2">
            <CheckCircle2 class="w-5 h-5" />
            <span class="font-bold tracking-tighter">CERTIFIED SECURE</span>
          </div>
        </div>
      </div>
    </section>

    <!-- Trusted By Section -->
    <section class="py-24 border-y border-slate-800 bg-slate-900/50">
      <div class="max-w-7xl mx-auto px-4">
        <p class="text-center text-xs font-bold text-slate-500 uppercase tracking-[0.3em] mb-12">Trusted by industry leaders</p>
        <div class="grid grid-cols-2 md:grid-cols-4 gap-12 opacity-40 grayscale">
          <div class="flex items-center justify-center gap-2">
            <ShieldCheck class="w-8 h-8" />
            <span class="text-xl font-bold tracking-tighter">SECURE CORP</span>
          </div>
          <div class="flex items-center justify-center gap-2">
            <Zap class="w-8 h-8" />
            <span class="text-xl font-bold tracking-tighter">FAST DATA</span>
          </div>
          <div class="flex items-center justify-center gap-2">
            <Activity class="w-8 h-8" />
            <span class="text-xl font-bold tracking-tighter">BIO LOGIC</span>
          </div>
          <div class="flex items-center justify-center gap-2">
            <Server class="w-8 h-8" />
            <span class="text-xl font-bold tracking-tighter">CLOUD MESH</span>
          </div>
        </div>
      </div>
    </section>

    <!-- FAQ Section -->
    <section id="faq" class="py-24 bg-slate-800/20">
      <div class="max-w-3xl mx-auto px-4">
        <h2 class="text-3xl font-bold mb-12 text-center">安全 FAQ</h2>
        <div class="space-y-4">
          <div v-for="(item, index) in faqs" :key="index" 
            class="bg-slate-900/50 border border-slate-700 rounded-2xl overflow-hidden">
            <button @click="activeFaq = activeFaq === index ? -1 : index" 
              class="w-full p-6 text-left flex justify-between items-center hover:bg-slate-800 transition-colors">
              <span class="font-bold">{{ item.q }}</span>
              <ChevronDown class="w-5 h-5 transition-transform" :class="{ 'rotate-180': activeFaq === index }" />
            </button>
            <div v-show="activeFaq === index" class="p-6 pt-0 text-slate-400 text-sm leading-relaxed border-t border-slate-800">
              {{ item.a }}
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Independent Security Footer -->
    <footer class="py-12 border-t border-slate-800 bg-slate-900">
      <div class="max-w-7xl mx-auto px-4">
        <div class="flex flex-col md:flex-row justify-between items-center gap-8">
          <div class="flex items-center gap-3">
            <div class="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
              <ShieldAlert class="w-5 h-5 text-white" />
            </div>
            <span class="text-lg font-black tracking-tighter">NEXUS GUARD</span>
          </div>
          
          <div class="flex gap-8 text-sm text-slate-500 font-medium">
            <button @click="openAction('privacy')" class="hover:text-blue-500 transition-colors">数据隐私说明</button>
            <button @click="openAction('compliance')" class="hover:text-blue-500 transition-colors">合规性声明</button>
            <button @click="openAction('support')" class="hover:text-blue-500 transition-colors">安全响应中心</button>
          </div>
          
          <div class="text-xs text-slate-600">
            © 2026 Nexus Guard Protocol. Military Grade Security.
          </div>
        </div>
      </div>
    </footer>

    <!-- Modal for Actions -->
    <div v-if="activeModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-slate-950/80 backdrop-blur-sm" @click="activeModal = null"></div>
      <div class="relative w-full max-w-lg bg-slate-900 border border-slate-700 rounded-2xl p-8 shadow-2xl overflow-hidden">
        <div class="absolute top-0 left-0 w-full h-1 bg-blue-600"></div>
        <button @click="activeModal = null" class="absolute top-6 right-6 p-2 hover:bg-slate-800 rounded-full transition-colors">
          <X class="w-5 h-5" />
        </button>
        
        <div class="text-center">
          <div class="w-16 h-16 bg-blue-600/10 text-blue-500 rounded-2xl flex items-center justify-center mx-auto mb-6">
            <component :is="modalContent.icon" class="w-8 h-8" />
          </div>
          <h3 class="text-2xl font-bold mb-4">{{ modalContent.title }}</h3>
          <p class="text-slate-400 mb-8 leading-relaxed text-sm">{{ modalContent.desc }}</p>
          
          <div v-if="modalContent.type === 'deploy'" class="bg-slate-800/50 rounded-xl p-6 mb-8 text-left border border-slate-700">
            <h4 class="text-sm font-bold text-blue-400 mb-4 uppercase tracking-wider">快速部署指令</h4>
            <code class="block bg-slate-950 p-3 rounded text-xs text-blue-300 font-mono mb-4">
              curl -sSL https://nexus-guard.io/install.sh | sh
            </code>
            <p class="text-[10px] text-slate-500 italic">注：需要 Linux 64位环境，支持 Docker/K8s 一键部署。</p>
          </div>

          <button @click="activeModal = null" class="w-full py-4 bg-blue-600 hover:bg-blue-500 text-white rounded-xl font-bold transition-all">
            确认
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { 
  ShieldAlert, 
  ShieldCheck,
  Terminal,
  Activity,
  Lock,
  Eye,
  UserX,
  FileSearch,
  BellRing,
  BarChart3,
  CheckCircle2,
  Zap,
  Server,
  ChevronDown,
  X,
  Info,
  Shield,
  FileText,
  Globe,
  BrainCircuit,
  Fingerprint
} from 'lucide-vue-next';
import { ref, computed } from 'vue';

const activeFaq = ref(-1);
const activeModal = ref<string | null>(null);

const scrollTo = (id: string) => {
  const el = document.getElementById(id);
  if (el) {
    el.scrollIntoView({ behavior: 'smooth' });
  }
};

const openAction = (type: string) => {
  activeModal.value = type;
};

const modalContent = computed(() => {
  switch (activeModal.value) {
    case 'deploy':
      return {
        type: 'deploy',
        icon: Terminal,
        title: '私有化部署方案',
        desc: 'Nexus Guard 支持完全私有化部署，确保所有通讯数据不经过任何第三方服务器。请在您的服务器执行以下脚本：'
      };
    case 'whitepaper':
      return {
        type: 'info',
        icon: FileText,
        title: 'Nexus Guard 安全白皮书',
        desc: '本白皮书详细介绍了我们的三重加密协议、异步审计引擎及分布式拒绝访问防御算法。'
      };
    case 'privacy':
      return {
        type: 'info',
        icon: Lock,
        title: '数据隐私说明',
        desc: 'Nexus Guard 遵循“零知识证明”原则。除必要的路由信息外，所有消息内容均在端侧加密，我们无法查看您的任何敏感数据。'
      };
    case 'compliance':
      return {
        type: 'info',
        icon: Shield,
        title: '合规性声明',
        desc: '系统符合 GDPR、SOC2 及中国等级保护三级（等保三级）的安全技术要求，支持审计日志导出。'
      };
    case 'support':
      return {
        type: 'info',
        icon: Activity,
        title: '安全响应中心 (SRC)',
        desc: '如果您发现了系统漏洞或安全隐患，请立即联系我们的安全响应团队。我们设有专项漏洞奖励计划。'
      };
    default:
      return {
        type: 'info',
        icon: Info,
        title: '提示',
        desc: '模块正在同步中，请稍后再试。'
      };
  }
});

const faqs = [
  {
    q: 'Nexus Guard 会导致消息延迟吗？',
    a: '我们的异步审计引擎在 10ms 内即可完成扫描，配合分布式边缘计算节点，用户几乎感知不到任何延迟。'
  },
  {
    q: '是否支持识别图片和音视频中的违规内容？',
    a: '支持。系统集成了 OCR 识别和音视频指纹技术，可精准拦截违规多媒体内容。'
  },
  {
    q: '私有化部署对硬件有什么要求？',
    a: '最低要求 2核/4G内存。对于日活跃用户超过 10w 的社群，建议使用 8核/16G 以上配置并开启集群模式。'
  },
  {
    q: '支持审计撤回的消息吗？',
    a: '支持。开启“镜像审计”后，所有撤回的消息都会在后台保留完整审计快照。'
  }
];

const securityFeatures = [
  {
    icon: Globe,
    title: 'Global Agent Mesh',
    desc: '基于 B2B 协作网桥，实现跨企业的安全身份委派与分布式威胁情报共享。'
  },
  {
    icon: Lock,
    title: 'Privacy Bastion',
    desc: '端到端加密通信与敏感信息自动脱敏技术，确保核心业务数据绝不出域。'
  },
  {
    icon: Activity,
    title: '实时审计流控',
    desc: '内置熔断机制与全链路 Trace 追踪，监控每一次 AI 调用的权限边界与资源消耗。'
  },
  {
    icon: ShieldCheck,
    title: '双栈权限校验',
    desc: '完美融合传统 IM 权限与现代 AI MCP 权限模型，实现细粒度的角色访问控制。'
  }
];

const securityMetrics = [
  {
    icon: BrainCircuit,
    title: '智能威胁分析',
    desc: '基于语义识别而非关键词匹配，自动识别钓鱼、诈骗及高风险意图指令。'
  },
  {
    icon: Fingerprint,
    title: '双向 JWT 握手',
    desc: '企业间建立连接需经过双向数字签名验证，确保通信双方身份的绝对可信。'
  },
  {
    icon: ShieldAlert,
    title: '自动熔断保护',
    desc: '检测到远程端点异常或异常流量时自动触发熔断，防止系统级雪崩效应。'
  }
];

const protectionTech = [
  {
    icon: Zap,
    title: '流式过滤引擎',
    desc: '基于 Go 高性能并发特性，实现流式消息过滤，不阻塞任何正常的聊天流程。'
  },
  {
    icon: Lock,
    title: '动态加密存储',
    desc: '敏感词库与审计日志采用动态密钥加密存储，即使物理硬盘丢失也无法泄露数据。'
  },
  {
    icon: Server,
    title: '插件化审计器',
    desc: '支持开发者自定义审计插件，可根据特定行业需求定制专属的过滤规则。'
  }
];
</script>

<template>
  <div class="min-h-screen bg-slate-900 text-white selection:bg-pink-500/30">
    <!-- Independent Robot Header -->
    <nav class="fixed top-0 w-full z-50 bg-slate-900/80 backdrop-blur-md border-b border-pink-500/10">
      <div class="max-w-7xl mx-auto px-4 h-20 flex justify-between items-center">
        <div class="flex items-center gap-3 group cursor-pointer" @click="scrollTo('hero')">
          <div class="w-10 h-10 bg-pink-500 rounded-2xl flex items-center justify-center shadow-lg shadow-pink-500/20 group-hover:rotate-12 transition-transform">
            <Cat class="w-6 h-6 text-white" />
          </div>
          <div class="flex flex-col">
            <span class="text-xl font-black tracking-tighter text-white">EARLYMEOW</span>
            <span class="text-[10px] text-pink-500 font-bold tracking-[0.2em] uppercase leading-none">Smart Assistant</span>
          </div>
        </div>
        
        <div class="hidden md:flex items-center gap-8 text-sm font-bold text-slate-400">
          <button @click="scrollTo('features')" class="hover:text-pink-500 transition-colors">核心功能</button>
          <button @click="scrollTo('detailed')" class="hover:text-pink-500 transition-colors">模块详解</button>
          <button @click="scrollTo('pricing')" class="hover:text-pink-500 transition-colors">资费方案</button>
          <button @click="scrollTo('faq')" class="hover:text-pink-500 transition-colors">常见问题</button>
          <router-link to="/botmatrix" class="px-4 py-2 rounded-lg bg-slate-800 text-slate-300 hover:bg-slate-700 transition-all border border-slate-700">
            BotMatrix
          </router-link>
          <button @click="openAction('subscribe')" class="px-6 py-2 bg-pink-500 hover:bg-pink-400 text-white rounded-lg transition-all shadow-lg shadow-pink-500/20">
            立即开启
          </button>
        </div>
      </div>
    </nav>

    <!-- Hero Section -->
    <header id="hero" class="pt-32 pb-20 px-4 relative overflow-hidden">
      <div class="absolute top-0 left-1/2 -translate-x-1/2 w-[1000px] h-[600px] bg-pink-500/10 blur-[120px] rounded-full -z-10"></div>
      <div class="max-w-7xl mx-auto flex flex-col md:flex-row items-center gap-12">
        <div class="flex-1 text-center md:text-left">
          <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-pink-500/10 border border-pink-500/20 text-pink-400 text-xs font-bold mb-6">
            <Sparkles class="w-3 h-3" />
            官方推荐机器人
          </div>
          <h1 class="text-5xl md:text-7xl font-black mb-6 leading-tight">
            早喵机器人 <span class="text-pink-500">EarlyMeow</span>
          </h1>
          <p class="text-xl text-slate-400 mb-10 max-w-xl leading-relaxed">
            专为社群管理与温馨陪伴设计的智能机器人。不仅是你的全能管家，更是群内最温暖的陪伴者。
          </p>
          <div class="flex flex-wrap gap-4 justify-center md:justify-start">
            <button @click="openAction('trial')" class="px-8 py-4 bg-pink-500 hover:bg-pink-400 text-white rounded-xl text-lg font-bold transition-all shadow-lg shadow-pink-500/20 flex items-center gap-2">
              <MessageCircle class="w-5 h-5" />
              立即开启 15 天免费试用
            </button>
            <button @click="openAction('demo')" class="px-8 py-4 bg-slate-800 hover:bg-slate-700 text-white rounded-xl text-lg font-bold transition-all border border-slate-700">
              功能演示
            </button>
          </div>
          <p class="mt-4 text-sm text-slate-500 flex items-center justify-center md:justify-start gap-2">
            <ShieldCheck class="w-4 h-4 text-pink-500/50" />
            支持使用个人小号托管，打造专属私人助手
          </p>
        </div>
        <div class="flex-1 relative">
          <div class="w-64 h-64 md:w-80 md:h-80 bg-gradient-to-br from-pink-400 to-purple-600 rounded-[3rem] rotate-6 absolute inset-0 blur-2xl opacity-20"></div>
          <div class="relative w-64 h-64 md:w-80 md:h-80 bg-slate-800 rounded-[3rem] border border-slate-700 overflow-hidden flex items-center justify-center group hover:scale-105 transition-transform duration-500">
            <Cat class="w-32 h-32 md:w-40 md:h-40 text-pink-500 group-hover:scale-110 transition-transform duration-500" />
          </div>
        </div>
      </div>
    </header>

    <!-- Core Features -->
    <section id="features" class="py-24 bg-slate-800/30 relative">
      <div class="max-w-7xl mx-auto px-4">
        <div class="text-center mb-16">
          <h2 class="text-3xl md:text-4xl font-bold mb-4">核心技能</h2>
          <div class="h-1.5 w-20 bg-pink-500 mx-auto rounded-full"></div>
        </div>
        
        <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
          <div v-for="feature in features" :key="feature.title" class="p-8 bg-slate-900/50 rounded-2xl border border-slate-700 hover:border-pink-500/50 transition-all hover:-translate-y-2">
            <div class="w-12 h-12 bg-pink-500/10 text-pink-500 rounded-xl flex items-center justify-center mb-6">
              <component :is="feature.icon" class="w-6 h-6" />
            </div>
            <h3 class="text-xl font-bold mb-4">{{ feature.title }}</h3>
            <p class="text-slate-400 leading-relaxed">
              {{ feature.desc }}
            </p>
          </div>
        </div>
      </div>
    </section>

    <!-- Usage Scenarios -->
    <section class="py-24">
      <div class="max-w-7xl mx-auto px-4">
        <div class="flex flex-col md:flex-row gap-16 items-center">
          <div class="flex-1 order-2 md:order-1">
            <div class="space-y-6">
              <div v-for="(scene, index) in scenes" :key="index" class="flex gap-6 p-6 rounded-2xl bg-slate-800/50 border border-slate-700 hover:bg-slate-800 transition-colors">
                <div class="flex-shrink-0 w-10 h-10 rounded-full bg-slate-700 flex items-center justify-center font-bold text-pink-400">
                  {{ index + 1 }}
                </div>
                <div>
                  <h4 class="text-lg font-bold mb-2">{{ scene.title }}</h4>
                  <p class="text-slate-400 text-sm leading-relaxed">{{ scene.desc }}</p>
                </div>
              </div>
            </div>
          </div>
          <div class="flex-1 order-1 md:order-2">
            <h2 class="text-4xl font-bold mb-6 leading-tight">让社群 <span class="text-pink-500">充满活力</span></h2>
            <p class="text-slate-400 mb-8 leading-relaxed">
              无论是早安问候、天气播报，还是群组自动审核、违规词拦截，早喵机器人都能游刃有余。基于分布式高可用架构，响应速度极快，稳定性极高。
            </p>
            <ul class="space-y-4">
              <li v-for="item in highlights" :key="item" class="flex items-center gap-3 text-slate-300">
                <CheckCircle2 class="w-5 h-5 text-pink-500" />
                {{ item }}
              </li>
            </ul>
          </div>
        </div>
      </div>
    </section>

    <!-- Interactive Modules (New Section) -->
    <section id="detailed" class="py-24 bg-slate-800/20">
      <div class="max-w-7xl mx-auto px-4">
        <div class="text-center mb-16">
          <h2 class="text-3xl md:text-5xl font-bold mb-4">功能模块详解</h2>
          <p class="text-slate-400">模块化设计，按需开启，打造最适合你的早喵。</p>
        </div>
        
        <div class="grid grid-cols-1 md:grid-cols-2 gap-12">
          <div v-for="module in detailedModules" :key="module.title" class="flex flex-col md:flex-row gap-8 items-start p-8 bg-slate-900/40 rounded-3xl border border-slate-700/50 hover:border-pink-500/30 transition-all group">
            <div class="w-20 h-20 shrink-0 bg-pink-500/10 rounded-2xl flex items-center justify-center group-hover:bg-pink-500 group-hover:text-slate-900 transition-all duration-500">
              <component :is="module.icon" class="w-10 h-10 text-pink-500 group-hover:text-slate-900 transition-colors" />
            </div>
            <div>
              <h4 class="text-2xl font-bold mb-4 group-hover:text-pink-400 transition-colors">{{ module.title }}</h4>
              <p class="text-slate-400 leading-relaxed mb-6">{{ module.desc }}</p>
              <div class="flex flex-wrap gap-2">
                <span v-for="tag in module.tags" :key="tag" class="px-3 py-1 bg-slate-800 rounded-full text-xs text-slate-300 border border-slate-700">{{ tag }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Technical Specs -->
    <section class="py-24">
      <div class="max-w-7xl mx-auto px-4">
        <div class="bg-gradient-to-br from-slate-800 to-slate-900 rounded-[3rem] p-12 border border-slate-700 relative overflow-hidden">
          <div class="absolute top-0 right-0 w-96 h-96 bg-pink-500/5 blur-[100px] -z-10"></div>
          <div class="text-center mb-12">
            <h2 class="text-3xl font-bold mb-4">技术指标</h2>
            <p class="text-slate-500 italic">追求卓越的性能与极致的体验</p>
          </div>
          <div class="grid grid-cols-2 md:grid-cols-4 gap-8">
            <div v-for="spec in technicalSpecs" :key="spec.label" class="text-center">
              <div class="text-4xl font-black text-pink-500 mb-2">{{ spec.value }}</div>
              <div class="text-sm text-slate-400 font-medium uppercase tracking-widest">{{ spec.label }}</div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Pricing Section -->
    <section id="pricing" class="py-24 bg-slate-900 relative overflow-hidden">
      <div class="max-w-7xl mx-auto px-4">
        <div class="text-center mb-16">
          <h2 class="text-4xl md:text-5xl font-bold mb-6">资费方案</h2>
          <p class="text-slate-400 max-w-2xl mx-auto text-lg">
            按群收费，灵活选择。所有方案均包含全量功能，支持 15 天免费深度体验。
          </p>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-8 mb-12">
          <div v-for="plan in pricingPlans" :key="plan.duration" 
            class="p-8 rounded-3xl border transition-all duration-500 flex flex-col group"
            :class="plan.highlight ? 'bg-pink-500/10 border-pink-500/50 scale-105 shadow-2xl shadow-pink-500/10' : 'bg-slate-800/40 border-slate-700 hover:border-pink-500/30'">
            
            <div v-if="plan.highlight" class="absolute -top-4 left-1/2 -translate-x-1/2 px-4 py-1 bg-pink-500 text-white text-xs font-bold rounded-full uppercase tracking-widest">
              最受欢迎
            </div>

            <div class="mb-8">
              <h3 class="text-xl font-bold text-slate-300 mb-2">{{ plan.duration }}</h3>
              <div class="flex items-baseline gap-1">
                <span class="text-4xl font-black text-white">¥{{ plan.price }}</span>
                <span v-if="plan.originalPrice" class="text-slate-500 line-through text-sm ml-2">¥{{ plan.originalPrice }}</span>
              </div>
            </div>

            <ul class="space-y-4 mb-8 flex-1">
              <li v-for="feature in plan.features" :key="feature" class="flex items-center gap-3 text-slate-400 text-sm">
                <CheckCircle2 class="w-4 h-4 text-pink-500" />
                {{ feature }}
              </li>
            </ul>

            <button @click="openAction('subscribe')" class="w-full py-4 rounded-xl font-bold transition-all"
              :class="plan.highlight ? 'bg-pink-500 text-white hover:bg-pink-400' : 'bg-slate-700 text-white hover:bg-slate-600'">
              立即订阅
            </button>
          </div>
        </div>

        <div class="max-w-4xl mx-auto bg-slate-800/50 border border-slate-700 rounded-3xl p-8 text-center">
          <h4 class="text-2xl font-bold mb-4 flex items-center justify-center gap-3">
            <Sparkles class="text-pink-500" />
            永久尊享方案
          </h4>
          <p class="text-slate-400 mb-6">一次付费，终身使用。适合长期稳定运营的深度用户。</p>
          <div class="text-5xl font-black text-pink-500 mb-8">¥498 <span class="text-lg text-slate-500 font-normal">/ 永久</span></div>
          <button @click="openAction('subscribe')" class="px-12 py-4 bg-gradient-to-r from-pink-500 to-purple-600 hover:from-pink-400 hover:to-purple-500 text-white rounded-xl font-bold transition-all shadow-xl shadow-pink-500/20">
            购买永久授权
          </button>
        </div>
      </div>
    </section>

    <!-- Success Stories -->
    <section class="py-24 relative overflow-hidden">
      <div class="max-w-7xl mx-auto px-4">
        <div class="text-center mb-16">
          <h2 class="text-4xl font-black mb-4">她们都在用早喵</h2>
          <p class="text-slate-400">超过 10,000+ 社群主的选择</p>
        </div>
        <div class="grid md:grid-cols-3 gap-8">
          <div v-for="(story, index) in stories" :key="index" 
            class="p-8 rounded-3xl bg-slate-800/30 border border-slate-700/50 hover:border-pink-500/30 transition-all group">
            <div class="flex items-center gap-4 mb-6">
              <img :src="story.avatar" class="w-12 h-12 rounded-full border-2 border-pink-500/20" />
              <div>
                <div class="font-bold text-white">{{ story.name }}</div>
                <div class="text-xs text-pink-500">{{ story.role }}</div>
              </div>
            </div>
            <p class="text-slate-400 text-sm leading-relaxed italic">“{{ story.content }}”</p>
          </div>
        </div>
      </div>
    </section>

    <!-- FAQ Section -->
    <section id="faq" class="py-24 bg-slate-800/20">
      <div class="max-w-3xl mx-auto px-4">
        <h2 class="text-3xl font-bold mb-12 text-center">常见问题</h2>
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

    <!-- Independent Robot Footer -->
    <footer class="py-12 border-t border-slate-800 bg-slate-900">
      <div class="max-w-7xl mx-auto px-4">
        <div class="flex flex-col md:flex-row justify-between items-center gap-8">
          <div class="flex items-center gap-3">
            <div class="w-8 h-8 bg-pink-500 rounded-xl flex items-center justify-center">
              <Cat class="w-5 h-5 text-white" />
            </div>
            <span class="text-lg font-black tracking-tighter">EARLYMEOW</span>
          </div>
          
          <div class="flex gap-8 text-sm text-slate-500 font-medium">
            <button @click="openAction('privacy')" class="hover:text-pink-500 transition-colors">隐私政策</button>
            <button @click="openAction('terms')" class="hover:text-pink-500 transition-colors">服务条款</button>
            <button @click="openAction('support')" class="hover:text-pink-500 transition-colors">联系支持</button>
          </div>
          
          <div class="text-xs text-slate-600">
            © 2026 EarlyMeow Robot. All rights reserved.
          </div>
        </div>
      </div>
    </footer>

    <!-- Modal for Actions -->
    <div v-if="activeModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-slate-950/80 backdrop-blur-sm" @click="activeModal = null"></div>
      <div class="relative w-full max-w-lg bg-slate-900 border border-slate-700 rounded-[2rem] p-8 shadow-2xl overflow-hidden">
        <div class="absolute top-0 left-0 w-full h-1 bg-pink-500"></div>
        <button @click="activeModal = null" class="absolute top-6 right-6 p-2 hover:bg-slate-800 rounded-full transition-colors">
          <X class="w-5 h-5" />
        </button>
        
        <div class="text-center">
          <div class="w-16 h-16 bg-pink-500/10 text-pink-500 rounded-2xl flex items-center justify-center mx-auto mb-6">
            <component :is="modalContent.icon" class="w-8 h-8" />
          </div>
          <h3 class="text-2xl font-bold mb-4">{{ modalContent.title }}</h3>
          <p class="text-slate-400 mb-8 leading-relaxed">{{ modalContent.desc }}</p>
          <div v-if="modalContent.type === 'contact'" class="bg-slate-800/50 rounded-xl p-4 mb-8 text-left border border-slate-700">
            <div class="flex items-center gap-3 mb-3">
              <MessageCircle class="w-4 h-4 text-pink-500" />
              <span class="text-sm font-bold">微信客服：EarlyMeow_Bot</span>
            </div>
            <div class="flex items-center gap-3">
              <Sparkles class="w-4 h-4 text-pink-500" />
              <span class="text-sm font-bold">QQ交流群：123456789</span>
            </div>
          </div>
          <button @click="activeModal = null" class="w-full py-4 bg-pink-500 hover:bg-pink-400 text-white rounded-xl font-bold transition-all">
            我知道了
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { 
  Sparkles, 
  MessageCircle, 
  Cat, 
  Sun, 
  ShieldCheck, 
  Zap, 
  CheckCircle2,
  Heart,
  Coffee,
  Bell,
  ChevronDown,
  X,
  Info,
  Shield,
  FileText,
  Coins,
  UserPlus,
  Layout,
  MessageSquare,
  Cpu
} from 'lucide-vue-next';
import { ref, computed } from 'vue';
import { useAuthStore } from '@/stores/auth';

const authStore = useAuthStore();
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
    case 'trial':
      return {
        type: 'info',
        icon: Sparkles,
        title: '开启 15 天免费试用',
        desc: '只需添加我们的官方客服并发送“申请试用”，即可获得 15 天全功能体验包。支持在您的群聊中进行深度测试。'
      };
    case 'demo':
      return {
        type: 'info',
        icon: MessageCircle,
        title: '查看功能演示',
        desc: '您可以加入我们的官方演示群，实时体验早晚安播报、情感陪聊及群管功能。发送“演示”即可获取入群邀请。'
      };
    case 'subscribe':
      return {
        type: 'contact',
        icon: Zap,
        title: '订阅服务',
        desc: '请联系我们的官方客服进行订阅。支持微信、支付宝支付，支付成功后即刻开通权限。'
      };
    case 'privacy':
      return {
        type: 'info',
        icon: Shield,
        title: '隐私政策',
        desc: '我们极度重视您的隐私。早喵机器人不会主动存储您的聊天记录，所有敏感数据均采用 AES-256 加密，仅用于实现机器人功能。'
      };
    case 'terms':
      return {
        type: 'info',
        icon: FileText,
        title: '服务条款',
        desc: '使用早喵机器人即代表您同意我们的服务条款。请勿将机器人用于任何违法违规用途，一经发现我们将永久封禁授权。'
      };
    case 'support':
      return {
        type: 'contact',
        icon: Heart,
        title: '联系支持',
        desc: '遇到任何配置问题或技术故障？我们的技术支持团队随时为您待命。'
      };
    default:
      return {
        type: 'info',
        icon: Info,
        title: '提示',
        desc: '功能开发中，敬请期待。'
      };
  }
});

const faqs = [
  {
    q: '如何添加机器人到我的群聊？',
    a: '首先添加机器人账号为好友，然后将其拉入群聊。在群内发送“/setup”指令即可开始配置。'
  },
  {
    q: '支持哪些平台的群聊？',
    a: '目前完美支持微信、QQ 及 Telegram。未来我们将支持更多主流通讯平台。'
  },
  {
    q: '我可以自定义机器人的名字吗？',
    a: '当然可以！如果您选择“自建机器人”方案，可以使用您自己的账号并设置任何您喜欢的名字。'
  },
  {
    q: '如果我不再续费，数据会丢失吗？',
    a: '您的配置数据会保留 30 天。在此期间续费即可无缝恢复使用。'
  }
];

const stories = [
  {
    name: '林悦',
    role: '读书会社群主',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=Annie',
    content: '自从用了早喵，群里的早起打卡变得非常有仪式感。它的情感陪聊功能让群氛围变得特别温馨。'
  },
  {
    name: '陈大白',
    role: '游戏公会会长',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=Jack',
    content: '群管功能非常强大，自动踢除广告和欢迎新人的话术都很自然，省去了我大量的管理时间。'
  },
  {
    name: 'Sarah',
    role: '跨境电商运营',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=Sarah',
    content: '跨平台的特性帮了大忙，我同时管理微信和 Telegram 群，早喵都能完美胜任，数据还是同步的。'
  }
];

const features = [
  {
    icon: Shield,
    title: '超级群管系统',
    desc: '集成关键词过滤、正则表达式匹配与阶梯式禁言（5min-24h），实现社群 7x24 小时自动化合规监管。'
  },
  {
    icon: Coins,
    title: '积分经济联动',
    desc: '独创积分惩罚与活跃奖励机制。违规自动扣除积分，每日发言或邀请好友可获得积分奖励，提升群活跃。'
  },
  {
    icon: UserPlus,
    title: '智能成员管理',
    desc: '支持多模板欢迎语、自动艾特、强制修改群名片及邀请人追踪，打造最具仪式感的入群体验。'
  },
  {
    icon: Layout,
    title: '免指令交互面板',
    desc: '管理员通过 /config 即可进入图形化配置模式，通过按钮与数字选项轻松管理群功能，无需记忆复杂指令。'
  },
  {
    icon: MessageSquare,
    title: '温暖陪伴系统',
    desc: '内置早晚安、天气提醒、每日资讯等温情功能，让机器人不仅是管家，更是群友们的贴心伙伴。'
  },
  {
    icon: Cpu,
    title: '自托管支持',
    desc: '支持使用个人小号托管，通过 BotWorker 边缘节点私有化部署，确保聊天数据与隐私完全掌控。'
  }
];

const scenes = [
  {
    title: '社群日常活跃',
    desc: '自动播报早晚安，定期推送行业新闻或趣味知识，显著提升社群留存率。'
  },
  {
    title: '活动自动化',
    desc: '支持群打卡、积分抽奖等活动，通过趣味互动增强成员归属感。'
  },
  {
    title: '企业内部通知',
    desc: '对接工作流，自动转发重要通知、会议提醒，确保信息高效触达。'
  }
];

const highlights = [
  '支持全平台消息同步',
  '极速响应，延迟低于 100ms',
  '零配置快速上手',
  '私有化部署，数据更安全'
];

const detailedModules = [
  {
    icon: Sun,
    title: '晨曦播报系统',
    desc: '集成全球气象数据，支持自定义播报时间。除了天气，还能同步股市行情、早间头条。',
    tags: ['天气API', '定时任务', '行情同步']
  },
  {
    icon: Coffee,
    title: '互动积分体系',
    desc: '通过打卡、发言获得积分。支持积分商城，可兑换群头衔或自定义奖励。',
    tags: ['打卡', '积分商城', '头衔系统']
  },
  {
    icon: Bell,
    title: '多维通知引擎',
    desc: '支持 RSS 订阅转发，Webhook 触发通知。重要消息支持多平台同步推送。',
    tags: ['RSS', 'Webhook', '多端同步']
  },
  {
    icon: Heart,
    title: '情感共鸣 AI',
    desc: '接入多款主流大模型，支持上下文关联记忆。具备更拟人化的表达风格。',
    tags: ['RAG', '多模型适配', '上下文记忆']
  },
  {
    icon: Zap,
    title: '自建机器人支持',
    desc: '不限制官方号，支持使用您自己的个人号、小号作为机器人载体，完美集成行业领先的通信引擎。',
    tags: ['个人号托管', '隐私保护', '专属载体']
  }
];

const technicalSpecs = [
  { label: '平均响应时间', value: '85ms' },
  { label: '支持消息类型', value: '20+' },
  { label: '系统资源占用', value: '<50MB' },
  { label: '并发处理能力', value: '10k/s' }
];

const pricingPlans = [
  {
    duration: '1 个月',
    price: '20',
    features: ['全功能解锁', '15 天免费试用', '技术支持', '社区交流群'],
    highlight: false
  },
  {
    duration: '2 个月',
    price: '35',
    originalPrice: '40',
    features: ['更优性价比', '全功能解锁', '优先技术响应', '赠送专属表情包'],
    highlight: false
  },
  {
    duration: '3 个月',
    price: '50',
    originalPrice: '60',
    features: ['季度优惠', '全功能解锁', '高级功能抢先看', '专属身份标识'],
    highlight: true
  },
  {
    duration: '半年 (6个月)',
    price: '80',
    originalPrice: '120',
    features: ['超值半年付', '全功能解锁', '一对一配置指导', '多群联动优惠'],
    highlight: false
  },
  {
    duration: '一年 (12个月)',
    price: '120',
    originalPrice: '240',
    features: ['年度最佳', '全功能解锁', '定制化功能开发', '年费会员勋章'],
    highlight: false
  }
];
</script>

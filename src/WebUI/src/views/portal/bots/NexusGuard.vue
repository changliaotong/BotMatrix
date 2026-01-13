<template>
  <div class="min-h-screen bg-[var(--bg-body)] text-[var(--text-main)] selection:bg-[var(--matrix-color)]/30" :class="[systemStore.style]">
    <!-- Grid Background -->
    <div class="fixed inset-0 z-0 overflow-hidden pointer-events-none opacity-20">
      <div class="absolute inset-0 bg-[radial-gradient(circle_at_center,_var(--tw-gradient-stops))] from-[var(--matrix-color)]/20 via-[var(--bg-body)] to-[var(--bg-body)]"></div>
      <div class="absolute inset-0" style="background-image: radial-gradient(circle at 2px 2px, var(--matrix-color, rgba(139, 92, 246, 0.05)) 1px, transparent 0); background-size: 40px 40px;"></div>
    </div>

    <PortalHeader />

    <!-- Hero Section -->
    <header id="hero" class="pt-32 pb-20 px-4 relative overflow-hidden z-10">
      <div class="absolute top-0 left-1/2 -translate-x-1/2 w-[1000px] h-[600px] bg-[var(--matrix-color)]/10 blur-[120px] rounded-full -z-10"></div>
      <div class="max-w-7xl mx-auto flex flex-col md:flex-row items-center gap-12">
        <div class="flex-1 text-center md:text-left">
          <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-[var(--matrix-color)]/10 border border-[var(--matrix-color)]/20 text-[var(--matrix-color)] text-xs font-bold mb-6">
            <ShieldAlert class="w-3 h-3" />
            {{ tt('nexus_guard.hero_badge') }}
          </div>
          <h1 class="text-5xl md:text-7xl font-black mb-6 leading-tight tracking-tighter text-[var(--text-main)]">
            {{ tt('nexus_guard.hero_title_1') }} <span class="text-transparent bg-clip-text bg-gradient-to-r from-[var(--matrix-color)] to-[rgba(var(--matrix-color-rgb),0.5)]">{{ tt('nexus_guard.hero_title_2') }}</span>
          </h1>
          <p class="text-xl text-[var(--text-muted)] mb-10 max-w-xl leading-relaxed">
            {{ tt('nexus_guard.hero_desc') }}
          </p>
          <div class="flex flex-wrap gap-4 justify-center md:justify-start">
            <button @click="openAction('deploy')" class="px-8 py-4 bg-[var(--matrix-color)] hover:bg-[var(--matrix-color)]/80 text-white rounded-xl text-lg font-bold transition-all shadow-lg shadow-[var(--matrix-color)]/30 flex items-center gap-2">
              <ShieldCheck class="w-5 h-5" />
              {{ tt('nexus_guard.start_protection') }}
            </button>
            <button @click="openAction('whitepaper')" class="px-8 py-4 bg-[var(--bg-body)]/50 hover:bg-[var(--bg-body)]/80 text-[var(--text-main)] rounded-xl text-lg font-bold transition-all border border-[var(--border-color)] backdrop-blur-sm">
              {{ tt('nexus_guard.whitepaper') }}
            </button>
          </div>
        </div>
        <div class="flex-1 relative">
          <div class="w-64 h-64 md:w-80 md:h-80 bg-gradient-to-br from-[var(--matrix-color)] to-[rgba(var(--matrix-color-rgb),0.5)] rounded-[3rem] -rotate-6 absolute inset-0 blur-2xl opacity-20"></div>
          <div class="relative w-64 h-64 md:w-80 md:h-80 bg-[var(--bg-body)]/50 backdrop-blur-xl rounded-[3rem] border border-[var(--border-color)] overflow-hidden flex items-center justify-center group hover:scale-105 transition-transform duration-500">
            <Lock class="w-32 h-32 md:w-40 md:h-40 text-[var(--matrix-color)] group-hover:scale-110 transition-transform duration-500" />
          </div>
        </div>
      </div>
    </header>

    <!-- Core Security Features -->
    <section id="features" class="py-24 relative z-10">
      <div class="max-w-7xl mx-auto px-4">
        <h2 class="text-3xl font-bold mb-12 text-center text-white">{{ tt('nexus_guard.features_title') }}</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
          <div v-for="feature in securityFeatures" :key="feature.title" class="p-8 bg-[var(--bg-body)]/50 backdrop-blur-sm rounded-2xl border border-[var(--border-color)] hover:border-[var(--matrix-color)]/50 transition-all hover:-translate-y-2 group">
            <div class="w-10 h-10 bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] rounded-lg flex items-center justify-center mb-6 group-hover:bg-[var(--matrix-color)] group-hover:text-white transition-all">
              <component :is="feature.icon" class="w-5 h-5" />
            </div>
            <h3 class="font-bold mb-3 group-hover:text-[var(--matrix-color)] transition-colors">{{ feature.title }}</h3>
            <p class="text-[var(--text-muted)] text-sm leading-relaxed">{{ feature.desc }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Security Dashboard Preview -->
    <section id="dashboard" class="py-24 relative z-10 bg-[var(--bg-body)]/30 backdrop-blur-sm border-y border-[var(--border-color)]">
      <div class="max-w-7xl mx-auto px-4">
        <div class="flex flex-col lg:flex-row gap-16 items-center">
          <div class="flex-1">
            <h2 class="text-3xl md:text-5xl font-black mb-8 leading-tight tracking-tighter text-[var(--text-main)]">
              {{ tt('nexus_guard.dashboard_title_1') }} <br/>
              <span class="text-transparent bg-clip-text bg-gradient-to-r from-[var(--matrix-color)] to-[rgba(var(--matrix-color-rgb),0.6)]">{{ tt('nexus_guard.dashboard_title_2') }}</span>
            </h2>
            <div class="space-y-8">
              <div v-for="item in securityMetrics" :key="item.title" class="flex gap-6 p-6 rounded-2xl bg-[var(--bg-body)]/50 border border-[var(--border-color)] hover:bg-[var(--bg-body)] transition-all group">
                <div class="flex-shrink-0 w-12 h-12 bg-[var(--matrix-color)]/10 rounded-xl flex items-center justify-center text-[var(--matrix-color)] group-hover:bg-[var(--matrix-color)] group-hover:text-white transition-all">
                  <component :is="item.icon" class="w-6 h-6" />
                </div>
                <div>
                  <h4 class="text-xl font-bold mb-2 text-[var(--text-main)]">{{ item.title }}</h4>
                  <p class="text-[var(--text-muted)] text-sm leading-relaxed">{{ item.desc }}</p>
                </div>
              </div>
            </div>
          </div>
          
          <div class="flex-1 w-full">
            <div class="relative bg-[var(--bg-body)] rounded-3xl border border-[var(--border-color)] shadow-2xl overflow-hidden aspect-[4/3] group">
              <div class="absolute inset-0 bg-[var(--matrix-color)]/5 group-hover:bg-[var(--matrix-color)]/10 transition-colors"></div>
              <!-- Mock UI -->
              <div class="p-6 h-full flex flex-col">
                <div class="flex items-center justify-between mb-8">
                  <div class="flex items-center gap-2">
                    <div class="w-3 h-3 rounded-full bg-red-500 animate-pulse"></div>
                    <span class="text-sm font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ tt('nexus_guard.dashboard.live_feed') }}</span>
                  </div>
                  <div class="px-3 py-1 bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] text-[10px] font-bold rounded-full border border-[var(--matrix-color)]/20">
                    {{ tt('nexus_guard.dashboard.protection_active') }}
                  </div>
                </div>
                <div class="flex-1 space-y-4">
                  <div v-for="i in 5" :key="i" class="h-12 bg-[var(--bg-body)]/80 rounded-xl border border-[var(--border-color)] flex items-center px-4 gap-4 animate-pulse" :style="{ animationDelay: `${i * 200}ms` }">
                    <div class="w-2 h-2 rounded-full bg-[var(--matrix-color)]"></div>
                    <div class="flex-1 h-2 bg-[var(--border-color)] rounded-full"></div>
                    <div class="w-12 h-2 bg-[var(--border-color)] rounded-full"></div>
                  </div>
                </div>
                <div class="mt-6 pt-6 border-t border-[var(--border-color)] flex justify-between items-end">
                  <div>
                    <div class="text-xs text-[var(--text-muted)] uppercase mb-1">{{ tt('nexus_guard.dashboard.threats_blocked') }}</div>
                    <div class="text-3xl font-black text-[var(--text-main)]">12,842</div>
                  </div>
                  <BarChart3 class="w-12 h-12 text-[var(--matrix-color)] opacity-50" />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Advanced Protection Tech -->
    <section id="architecture" class="py-24 relative z-10">
      <div class="max-w-7xl mx-auto px-4">
        <div class="text-center mb-16">
          <h2 class="text-3xl md:text-5xl font-bold mb-4 tracking-tighter text-[var(--text-main)]">{{ tt('nexus_guard.architecture_title') }}</h2>
          <p class="text-[var(--text-muted)]">{{ tt('nexus_guard.architecture_desc') }}</p>
        </div>
        
        <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
          <div v-for="tech in protectionTech" :key="tech.title" class="p-10 bg-[var(--bg-body)]/50 backdrop-blur-sm rounded-[2.5rem] border border-[var(--border-color)] hover:border-[var(--matrix-color)]/30 transition-all text-center">
            <div class="w-16 h-16 bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] rounded-2xl flex items-center justify-center mx-auto mb-8">
              <component :is="tech.icon" class="w-8 h-8" />
            </div>
            <h3 class="text-2xl font-bold mb-4 text-[var(--text-main)]">{{ tech.title }}</h3>
            <p class="text-[var(--text-muted)] leading-relaxed">{{ tech.desc }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Compliance Section -->
    <section class="py-24 relative z-10 bg-[var(--bg-body)]/30 backdrop-blur-sm border-y border-[var(--border-color)]">
      <div class="max-w-4xl mx-auto px-4 text-center">
        <div class="inline-flex p-4 bg-[var(--matrix-color)]/10 rounded-3xl border border-[var(--matrix-color)]/20 mb-8">
          <ShieldCheck class="w-12 h-12 text-[var(--matrix-color)]" />
        </div>
        <h2 class="text-3xl font-bold mb-6 tracking-tighter text-[var(--text-main)]">{{ tt('nexus_guard.compliance_title') }}</h2>
        <p class="text-[var(--text-muted)] text-lg leading-relaxed mb-10">
          {{ tt('nexus_guard.compliance_desc') }}
        </p>
        <div class="flex flex-wrap justify-center gap-8 opacity-50 grayscale hover:grayscale-0 transition-all duration-700">
          <div v-for="i in 4" :key="i" class="flex items-center gap-2">
            <CheckCircle2 class="w-5 h-5 text-[var(--matrix-color)]" />
            <span class="font-bold tracking-tighter text-[var(--text-main)]">{{ tt('nexus_guard.certified_secure') }}</span>
          </div>
        </div>
      </div>
    </section>

    <!-- FAQ Section -->
    <section id="faq" class="py-24 relative z-10">
      <div class="max-w-3xl mx-auto px-4">
        <h2 class="text-3xl font-bold mb-12 text-center text-[var(--text-main)] tracking-tighter">{{ tt('nexus_guard.faq_title') }}</h2>
        <div class="space-y-4">
          <div v-for="(item, index) in faqs" :key="index" 
            class="bg-[var(--bg-body)]/50 backdrop-blur-sm border border-[var(--border-color)] rounded-2xl overflow-hidden">
            <button @click="activeFaq = activeFaq === index ? -1 : index" 
              class="w-full p-6 text-left flex justify-between items-center hover:bg-[var(--bg-body)]/80 transition-colors">
              <span class="font-bold text-[var(--text-main)]">{{ item.q }}</span>
              <ChevronDown class="w-5 h-5 transition-transform text-[var(--text-muted)]" :class="{ 'rotate-180': activeFaq === index }" />
            </button>
            <div v-show="activeFaq === index" class="p-6 pt-0 text-[var(--text-muted)] text-sm leading-relaxed border-t border-[var(--border-color)]">
              {{ item.a }}
            </div>
          </div>
        </div>
      </div>
    </section>

    <PortalFooter />

    <!-- Modal for Actions -->
    <transition
      enter-active-class="transition duration-300 ease-out"
      enter-from-class="opacity-0 scale-95"
      enter-to-class="opacity-100 scale-100"
      leave-active-class="transition duration-200 ease-in"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-95"
    >
      <div v-if="activeModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-[var(--bg-body)]/80 backdrop-blur-sm" @click="activeModal = null"></div>
        <div class="relative w-full max-w-lg bg-[var(--bg-body)] border border-[var(--border-color)] rounded-2xl p-8 shadow-2xl overflow-hidden">
          <div class="absolute top-0 left-0 w-full h-1 bg-[var(--matrix-color)]"></div>
          <button @click="activeModal = null" class="absolute top-6 right-6 p-2 hover:bg-[var(--bg-body)]/80 rounded-full transition-colors">
            <X class="w-5 h-5 text-[var(--text-muted)]" />
          </button>
          
          <div class="text-center">
            <div class="w-16 h-16 bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] rounded-2xl flex items-center justify-center mx-auto mb-6">
              <component :is="modalContent.icon" class="w-8 h-8" />
            </div>
            <h3 class="text-2xl font-bold mb-4 text-[var(--text-main)]">{{ modalContent.title }}</h3>
            <p class="text-[var(--text-muted)] mb-8 leading-relaxed text-sm">{{ modalContent.desc }}</p>
            
            <div v-if="modalContent.type === 'deploy'" class="bg-[var(--bg-body)]/50 rounded-xl p-6 mb-8 text-left border border-[var(--border-color)]">
              <h4 class="text-sm font-bold text-[var(--matrix-color)] mb-4 uppercase tracking-wider">{{ tt('guard.quick_deploy_cmd') }}</h4>
              <code class="block bg-[var(--bg-body)] p-3 rounded text-xs text-[var(--matrix-color)]/80 font-mono mb-4">
                curl -sSL https://nexus-guard.io/install.sh | sh
              </code>
              <p class="text-[10px] text-[var(--text-muted)] italic">{{ tt('nexus_guard.modal.deploy.linux_hint') }}</p>
            </div>

            <button @click="activeModal = null" class="w-full py-4 bg-[var(--matrix-color)] hover:bg-[var(--matrix-color)]/80 text-white rounded-xl font-bold transition-all shadow-lg shadow-[var(--matrix-color)]/20">
              {{ tt('common.confirm') }}
            </button>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { 
  ShieldAlert, 
  ShieldCheck,
  Terminal,
  Activity,
  Lock,
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
import { useSystemStore } from '@/stores/system';
import { t } from '@/utils/i18n';
import PortalHeader from '@/components/layout/PortalHeader.vue';
import PortalFooter from '@/components/layout/PortalFooter.vue';

const systemStore = useSystemStore();
const activeFaq = ref(-1);
const activeModal = ref<string | null>(null);

const tt = (key: string, defaultText?: string) => {
  const res = t(key);
  return res === key ? (defaultText || key) : res;
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
        title: tt('nexus_guard.modal.deploy.title'),
        desc: tt('nexus_guard.modal.deploy.desc')
      };
    case 'whitepaper':
      return {
        type: 'info',
        icon: FileText,
        title: tt('nexus_guard.modal.whitepaper.title'),
        desc: tt('nexus_guard.modal.whitepaper.desc')
      };
    case 'privacy':
      return {
        type: 'info',
        icon: Lock,
        title: tt('nexus_guard.modal.privacy.title'),
        desc: tt('nexus_guard.modal.privacy.desc')
      };
    case 'compliance':
      return {
        type: 'info',
        icon: Shield,
        title: tt('nexus_guard.modal.compliance.title'),
        desc: tt('nexus_guard.modal.compliance.desc')
      };
    case 'support':
      return {
        type: 'info',
        icon: Activity,
        title: tt('nexus_guard.modal.support.title'),
        desc: tt('nexus_guard.modal.support.desc')
      };
    default:
      return {
        type: 'info',
        icon: Info,
        title: tt('common.hint'),
        desc: tt('common.module_syncing')
      };
  }
});

const faqs = computed(() => [
  {
    q: tt('nexus_guard.faq.q1'),
    a: tt('nexus_guard.faq.a1')
  },
  {
    q: tt('nexus_guard.faq.q2'),
    a: tt('nexus_guard.faq.a2')
  },
  {
    q: tt('nexus_guard.faq.q3'),
    a: tt('nexus_guard.faq.a3')
  },
  {
    q: tt('nexus_guard.faq.q4'),
    a: tt('nexus_guard.faq.a4')
  }
]);

const securityFeatures = computed(() => [
  {
    icon: Globe,
    title: tt('nexus_guard.feature.gam.title'),
    desc: tt('nexus_guard.feature.gam.desc')
  },
  {
    icon: Lock,
    title: tt('nexus_guard.feature.pb.title'),
    desc: tt('nexus_guard.feature.pb.desc')
  },
  {
    icon: Activity,
    title: tt('nexus_guard.feature.audit.title'),
    desc: tt('nexus_guard.feature.audit.desc')
  },
  {
    icon: ShieldCheck,
    title: tt('nexus_guard.feature.auth.title'),
    desc: tt('nexus_guard.feature.auth.desc')
  }
]);

const securityMetrics = computed(() => [
  {
    icon: BrainCircuit,
    title: tt('nexus_guard.metric.threat.title'),
    desc: tt('nexus_guard.metric.threat.desc')
  },
  {
    icon: Fingerprint,
    title: tt('nexus_guard.metric.jwt.title'),
    desc: tt('nexus_guard.metric.jwt.desc')
  },
  {
    icon: ShieldAlert,
    title: tt('nexus_guard.metric.circuit.title'),
    desc: tt('nexus_guard.metric.circuit.desc')
  }
]);

const protectionTech = computed(() => [
  {
    icon: Zap,
    title: tt('nexus_guard.tech.stream.title'),
    desc: tt('nexus_guard.tech.stream.desc')
  },
  {
    icon: Lock,
    title: tt('nexus_guard.tech.encrypt.title'),
    desc: tt('nexus_guard.tech.encrypt.desc')
  },
  {
    icon: Server,
    title: tt('nexus_guard.tech.plugin.title'),
    desc: tt('nexus_guard.tech.plugin.desc')
  }
]);
</script>

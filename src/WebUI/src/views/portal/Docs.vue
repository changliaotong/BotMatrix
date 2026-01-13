<template>
  <div class="min-h-screen bg-[var(--bg-body)] text-[var(--text-main)] selection:bg-[var(--matrix-color)]/30 relative overflow-x-hidden" :class="[systemStore.style]">
    <!-- Unified Background -->
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute top-0 left-1/2 -translate-x-1/2 w-full h-full bg-[radial-gradient(circle_at_center,var(--matrix-color,rgba(168,85,247,0.05))_0%,transparent_70%)] opacity-20"></div>
      <div class="absolute inset-0 bg-[grid-slate-800] [mask-image:radial-gradient(white,transparent)] opacity-20"></div>
    </div>

    <PortalHeader />

    <header class="pt-48 pb-24 px-6 relative overflow-hidden">
      <div class="max-w-7xl mx-auto text-center">
        <div class="inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-[var(--matrix-color)]/10 border border-[var(--matrix-color)]/20 text-[var(--matrix-color)] text-xs font-medium mb-10">
          <Rocket class="w-4 h-4" /> {{ tt('portal.docs.badge') }}
        </div>
        <h1 class="text-6xl md:text-8xl font-black mb-12 tracking-tighter leading-none uppercase" v-html="tt('portal.docs.title')"></h1>
        <p class="text-[var(--text-muted)] max-w-3xl mx-auto text-xl leading-relaxed font-light">
          {{ tt('portal.docs.subtitle') }}
        </p>
      </div>
    </header>

    <section class="py-24 px-6">
      <div class="max-w-7xl mx-auto grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-10">
        <div v-for="guide in guides" :key="guide.title" class="p-10 bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-[2.5rem] hover:bg-[var(--bg-body)]/80 hover:border-[var(--matrix-color)]/30 transition-all duration-500 group relative overflow-hidden">
          <div class="absolute top-0 left-0 w-1 h-0 bg-[var(--matrix-color)]/50 group-hover:h-full transition-all duration-700"></div>
          <div class="w-16 h-16 bg-[var(--bg-body)] text-[var(--text-muted)] rounded-2xl flex items-center justify-center mb-10 group-hover:bg-[var(--matrix-color)] group-hover:text-white transition-all duration-500 group-hover:shadow-[0_0_30px_var(--matrix-glow)]">
            <component :is="guide.icon" class="w-8 h-8" />
          </div>
          <h3 class="text-2xl font-bold mb-4 uppercase tracking-tight text-[var(--text-main)] group-hover:text-[var(--matrix-color)] transition-colors">{{ guide.title }}</h3>
          <p class="text-[var(--text-muted)] text-sm font-light leading-relaxed mb-10">{{ guide.desc }}</p>
          <router-link :to="'/docs/' + guide.id" class="inline-flex items-center gap-3 text-[var(--matrix-color)] font-bold uppercase tracking-widest text-[10px] hover:gap-5 transition-all">
            {{ tt('portal.docs.read_now') }} <ArrowRight class="w-4 h-4" />
          </router-link>
        </div>
      </div>
    </section>

    <section class="py-32 bg-[var(--bg-body)]/30 border-y border-[var(--border-color)] backdrop-blur-3xl relative overflow-hidden">
      <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-full h-full bg-[radial-gradient(circle_at_center,var(--matrix-color,rgba(168,85,247,0.03))_0%,transparent_70%)] pointer-events-none"></div>
      <div class="max-w-5xl mx-auto px-6 relative z-10">
        <h2 class="text-3xl font-bold mb-12 flex items-center gap-4 uppercase tracking-tight text-[var(--text-main)]">
          <Terminal class="w-8 h-8 text-[var(--matrix-color)]" />
          {{ tt('portal.docs.quick_start') }}
        </h2>
        <div class="bg-[var(--bg-body)] border border-[var(--border-color)] rounded-[2.5rem] p-10 font-mono text-sm group relative shadow-2xl overflow-hidden">
          <div class="absolute top-0 right-0 p-10 opacity-5 pointer-events-none">
            <MonitorDot class="w-48 h-48 text-[var(--matrix-color)]" />
          </div>
          <button class="absolute top-8 right-8 p-3 bg-[var(--bg-body)] hover:bg-[var(--matrix-color)] hover:text-white rounded-xl text-[var(--text-muted)] transition-all group-hover:opacity-100 opacity-0">
            <Copy class="w-5 h-5" />
          </button>
          <div class="space-y-6 relative">
            <div>
              <div class="text-[var(--matrix-color)]/40 mb-2 uppercase tracking-widest text-[10px] font-bold">{{ tt('portal.docs.step1') }}</div>
              <div class="text-[var(--text-main)] font-bold bg-[var(--bg-body)]/50 p-4 rounded-xl border border-[var(--border-color)]">git clone https://github.com/changliaotong/BotMatrix.git</div>
            </div>
            <div>
              <div class="text-[var(--matrix-color)]/40 mb-2 uppercase tracking-widest text-[10px] font-bold">{{ tt('portal.docs.step2') }}</div>
              <div class="text-[var(--text-main)] font-bold bg-[var(--bg-body)]/50 p-4 rounded-xl border border-[var(--border-color)]">cd BotMatrix && docker-compose up -d</div>
            </div>
            <div>
              <div class="text-[var(--matrix-color)]/40 mb-2 uppercase tracking-widest text-[10px] font-bold">{{ tt('portal.docs.step3') }}</div>
              <div class="text-[var(--matrix-color)] font-bold tracking-widest hover:underline cursor-pointer bg-[var(--matrix-color)]/5 p-4 rounded-xl border border-[var(--matrix-color)]/10 inline-block">http://localhost:5173</div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Technical Architecture Deep Dive -->
    <section class="py-48 relative">
      <div class="max-w-7xl mx-auto px-6">
        <h2 class="text-4xl md:text-6xl font-bold mb-24 text-center uppercase tracking-tight text-[var(--text-main)]">
          {{ tt('portal.docs.philosophy_title') }}
        </h2>
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-20">
          <div v-for="concept in technicalPhilosophy" :key="concept.title" class="space-y-10 group">
            <div class="flex items-center gap-6">
              <div class="w-16 h-16 bg-[var(--bg-body)] text-[var(--text-muted)] rounded-2xl flex items-center justify-center group-hover:bg-[var(--matrix-color)]/10 group-hover:text-[var(--matrix-color)] transition-all duration-500">
                <component :is="concept.icon" class="w-8 h-8" />
              </div>
              <h3 class="text-3xl font-bold uppercase tracking-tight text-[var(--text-main)]">{{ concept.title }}</h3>
            </div>
            <p class="text-[var(--text-muted)] font-light leading-relaxed text-lg">{{ concept.desc }}</p>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div v-for="item in concept.items" :key="item" class="flex items-center gap-4 text-xs font-bold uppercase tracking-widest text-[var(--text-muted)] group-hover:text-[var(--text-main)] transition-colors">
                <CheckCircle2 class="w-4 h-4 text-[var(--matrix-color)]" /> {{ item }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Roadmap & Future -->
    <section class="py-48 bg-[var(--bg-body)]/20 border-t border-[var(--border-color)]">
      <div class="max-w-7xl mx-auto px-6">
        <div class="bg-[var(--bg-body)] border border-[var(--border-color)] rounded-[4rem] p-16 md:p-32 relative overflow-hidden shadow-2xl">
          <div class="absolute top-0 right-0 w-[500px] h-[500px] bg-[var(--matrix-color)]/5 blur-[150px] -z-10 animate-pulse"></div>
          <div class="absolute bottom-0 left-0 w-[500px] h-[500px] bg-[var(--matrix-color)]/5 blur-[150px] -z-10 animate-pulse" style="animation-delay: 2s"></div>
          
          <h2 class="text-5xl md:text-7xl font-bold mb-24 uppercase tracking-tight text-[var(--text-main)]">
            {{ tt('portal.docs.roadmap_title') }}
          </h2>
          
          <div class="space-y-24 relative before:absolute before:left-0 before:top-2 before:bottom-2 before:w-px before:bg-[var(--border-color)] ml-4 pl-16">
            <div v-for="phase in roadmap" :key="phase.title" class="relative group">
              <div class="absolute -left-[4.5rem] top-0 w-6 h-6 rounded-full bg-[var(--bg-body)] border-4 border-[var(--border-color)] group-hover:border-[var(--matrix-color)] transition-colors duration-500 shadow-[0_0_20px_rgba(var(--matrix-color-rgb),0.1)] group-hover:shadow-[0_0_30px_rgba(var(--matrix-color-rgb),0.3)]"></div>
              <div class="inline-flex items-center gap-3 px-4 py-1.5 rounded-full bg-[var(--bg-body)] border border-[var(--border-color)] text-[var(--text-muted)] text-[10px] font-bold uppercase tracking-[0.3em] mb-8 group-hover:bg-[var(--matrix-color)] group-hover:text-white transition-all duration-500">
                <Calendar class="w-3 h-3" /> {{ phase.time }}
              </div>
              <h3 class="text-4xl font-bold mb-6 uppercase tracking-tight text-[var(--text-main)] group-hover:text-[var(--matrix-color)] transition-colors">{{ phase.title }}</h3>
              <p class="text-[var(--text-muted)] mb-10 max-w-3xl text-lg font-light leading-relaxed group-hover:text-[var(--text-main)] transition-colors">{{ phase.desc }}</p>
              <div class="flex flex-wrap gap-4">
                <span v-for="tag in phase.tags" :key="tag" class="px-5 py-2 bg-[var(--bg-body)] border border-[var(--border-color)] rounded-xl text-[10px] text-[var(--text-muted)] font-bold uppercase tracking-widest group-hover:border-[var(--matrix-color)]/30 group-hover:text-[var(--matrix-color)] transition-all">{{ tag }}</span>
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
import { computed } from 'vue';
import { useI18n } from '@/utils/i18n';
import { useSystemStore } from '@/stores/system';

const { t: tt } = useI18n();
const systemStore = useSystemStore();

const guides = computed(() => [
  {
    id: 1,
    icon: Globe,
    title: tt('portal.about.vision1_title'),
    desc: tt('portal.docs.guides.mesh_desc')
  },
  {
    id: 2,
    icon: Network,
    title: tt('portal.docs.guides.swarm_title'),
    desc: tt('portal.docs.guides.swarm_desc')
  },
  {
    id: 3,
    icon: MonitorDot,
    title: tt('portal.docs.guides.vision_title'),
    desc: tt('portal.docs.guides.vision_desc')
  },
  {
    id: 4,
    icon: Users,
    title: tt('portal.docs.guides.employee_title'),
    desc: tt('portal.docs.guides.employee_desc')
  },
  {
    id: 5,
    icon: BrainCircuit,
    title: tt('portal.docs.guides.rag_title'),
    desc: tt('portal.docs.guides.rag_desc')
  },
  {
    id: 6,
    icon: Code2,
    title: tt('portal.docs.guides.mcp_title'),
    desc: tt('portal.docs.guides.mcp_desc')
  }
]);

const technicalPhilosophy = computed(() => [
  {
    icon: Cpu,
    title: tt('portal.docs.philosophy.employee_title'),
    desc: tt('portal.docs.philosophy.employee_desc'),
    items: ['IdentityGORM', 'Intent Dispatcher', 'Cognitive Memory', 'MCP Toolset', 'Agent Mesh', 'Auto-Learning']
  },
  {
    icon: Layers,
    title: tt('portal.docs.philosophy.loop_title'),
    desc: tt('portal.docs.philosophy.loop_desc'),
    items: ['Short-term Memory', 'Long-term Memory', 'RAG Retrieval', 'HITL Mechanism']
  },
  {
    icon: Zap,
    title: tt('portal.docs.philosophy.a2a_title'),
    desc: tt('portal.docs.philosophy.a2a_desc'),
    items: ['JWT Signature', 'B2B Gateway', 'DID Identity', 'Traceability']
  },
  {
    icon: ShieldAlert,
    title: tt('portal.docs.philosophy.security_title'),
    desc: tt('portal.docs.philosophy.security_desc'),
    items: ['PII Identification', 'Audit Logging', 'RBAC Control', 'Ethics Guardrail']
  }
]);

const roadmap = computed(() => [
  {
    time: '2026 Q1',
    title: tt('portal.docs.roadmap.q1_title'),
    desc: tt('portal.docs.roadmap.q1_desc'),
    tags: ['Worker v2.0', 'MCP SDK', 'Basic Agent Mesh']
  },
  {
    time: '2026 Q2',
    title: tt('portal.docs.roadmap.q2_title'),
    desc: tt('portal.docs.roadmap.q2_desc'),
    tags: ['Autonomous Memory', 'RAG v3.0', 'Fact Extraction']
  },
  {
    time: '2026 Q3',
    title: tt('portal.docs.roadmap.q3_title'),
    desc: tt('portal.docs.roadmap.q3_desc'),
    tags: ['Swarm Engine', 'Task Decomposer', 'Collective Intelligence']
  }
]);
</script>

<style scoped>
/* No additional scoped styles needed as Tailwind classes handle everything */
</style>

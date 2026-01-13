<script setup lang="ts">
import { 
  ArrowRight, Sparkles, Shield, Zap, Heart, 
  MessageSquare, Bot, Activity, CheckCircle2,
  Users, Coffee, Code, Rocket, Globe, UserCircle, Hash, Building2, Trophy, Wrench, Calendar,
  Cpu, Coins, Server, Brain, FileText, TrendingUp
} from 'lucide-vue-next';
import { useSystemStore } from '@/stores/system';
import { useMeowStore } from '@/stores/earlymeow';
import { useI18n } from '@/utils/i18n';
import { computed } from 'vue';
import { useRouter } from 'vue-router';

const systemStore = useSystemStore();
const meowStore = useMeowStore();
const router = useRouter();
const { t: tt } = useI18n();

const scrollTo = (id: string) => {
  const el = document.getElementById(id);
  if (el) el.scrollIntoView({ behavior: 'smooth' });
};

const features = computed(() => [
  {
    title: tt('earlymeow.home.features.nlp.title'),
    desc: tt('earlymeow.home.features.nlp.desc'),
    icon: MessageSquare,
    color: 'purple'
  },
  {
    title: tt('earlymeow.home.features.workflow.title'),
    desc: tt('earlymeow.home.features.workflow.desc'),
    icon: Zap,
    color: 'purple'
  },
  {
    title: tt('earlymeow.home.features.emotion.title'),
    desc: tt('earlymeow.home.features.emotion.desc'),
    icon: Heart,
    color: 'purple'
  },
  {
    title: tt('earlymeow.home.features.multi_platform.title'),
    desc: tt('earlymeow.home.features.multi_platform.desc'),
    icon: Bot,
    color: 'purple'
  }
]);

const useCases = computed(() => [
  {
    title: tt('earlymeow.home.cases.community.title'),
    desc: tt('earlymeow.home.cases.community.desc'),
    icon: Users,
    image: 'https://images.unsplash.com/photo-1522071820081-009f0129c71c?auto=format&fit=crop&w=800&q=80'
  },
  {
    title: tt('earlymeow.home.cases.assistant.title'),
    desc: tt('earlymeow.home.cases.assistant.desc'),
    icon: Coffee,
    image: 'https://images.unsplash.com/photo-1484417894907-623942c8ee29?auto=format&fit=crop&w=800&q=80'
  },
  {
    title: tt('earlymeow.home.cases.developer.title'),
    desc: tt('earlymeow.home.cases.developer.desc'),
    icon: Code,
    image: 'https://images.unsplash.com/photo-1461749280684-dccba630e2f6?auto=format&fit=crop&w=800&q=80'
  }
]);

const stats = computed(() => [
  { label: tt('earlymeow.home.stats.bots'), value: '12,400+' },
  { label: tt('earlymeow.home.stats.messages'), value: '8.5M+' },
  { label: tt('earlymeow.home.stats.satisfaction'), value: '99.9%' }
]);

const digitalEmployees = computed(() => [
  {
    id: 'EMP-001',
    name: 'MeowDev',
    role: tt('earlymeow.home.employees.role_dev'),
    department: tt('earlymeow.home.employees.dept_rd'),
    kpi: 98,
    efficiency: 95,
    autonomy: 92,
    taskCount: 1240,
    successRate: 99.2,
    status: tt('earlymeow.home.employees.status_online'),
    skills: ['TypeScript', 'Go', 'LLM Tuning'],
    onboarding: '2024-03-12',
    roi: 96,
    tokenUsage: 1250000,
    costSaved: 1250,
    sandbox: 'Docker',
    memory: tt('earlymeow.home.employees.memory_active'),
    sop: 'R&D-v2.1',
    supervisor: tt('earlymeow.home.employees.supervisor_arch'),
    activeTask: tt('earlymeow.home.employees.task_refactor'),
    toolsCount: 24
  },
  {
    id: 'EMP-002',
    name: 'MeowSupport',
    role: tt('earlymeow.home.employees.role_support'),
    department: tt('earlymeow.home.employees.dept_service'),
    kpi: 95,
    efficiency: 98,
    autonomy: 88,
    taskCount: 45200,
    successRate: 98.5,
    status: tt('earlymeow.home.employees.status_online'),
    skills: ['Multi-lang', 'Sentiment Analysis', 'CRM'],
    onboarding: '2024-05-20',
    roi: 94,
    tokenUsage: 8400000,
    costSaved: 8400,
    sandbox: 'None',
    memory: 'RAG',
    sop: 'Support-v4',
    supervisor: tt('earlymeow.home.employees.supervisor_service'),
    activeTask: tt('earlymeow.home.employees.task_ticket'),
    toolsCount: 12
  },
  {
    id: 'EMP-003',
    name: 'MeowAnalyst',
    role: tt('earlymeow.home.employees.role_analyst'),
    department: tt('earlymeow.home.employees.dept_data'),
    kpi: 92,
    efficiency: 89,
    autonomy: 95,
    taskCount: 850,
    successRate: 97.8,
    status: tt('earlymeow.home.employees.status_busy'),
    skills: ['Python', 'SQL', 'Predictive Modeling'],
    onboarding: '2024-08-05',
    roi: 88,
    tokenUsage: 560000,
    costSaved: 560,
    sandbox: 'Docker',
    memory: 'Redis',
    sop: 'Data-v1.2',
    supervisor: tt('earlymeow.home.employees.supervisor_data'),
    activeTask: tt('earlymeow.home.employees.task_retention'),
    toolsCount: 18
  },
  {
    id: 'EMP-004',
    name: 'MeowManager',
    role: tt('earlymeow.home.employees.role_manager'),
    department: tt('earlymeow.home.employees.dept_ops'),
    kpi: 99,
    efficiency: 94,
    autonomy: 99,
    taskCount: 3200,
    successRate: 99.9,
    status: tt('earlymeow.home.employees.status_online'),
    skills: ['Orchestration', 'Risk Control', 'Scheduling'],
    onboarding: '2025-01-01',
    roi: 99,
    tokenUsage: 2100000,
    costSaved: 2100,
    sandbox: 'Docker',
    memory: tt('earlymeow.home.employees.memory_active'),
    sop: 'Ops-v3.0',
    supervisor: tt('earlymeow.home.employees.supervisor_admin'),
    activeTask: tt('earlymeow.home.employees.task_resource'),
    toolsCount: 42
  }
]);
</script>

<template>
  <div class="relative z-10">
    <!-- Hero Section -->
    <header class="pt-40 pb-32 px-4 relative overflow-hidden">
      <!-- Matrix Code Rain Placeholder / Background -->
      <div class="absolute inset-0 opacity-10 pointer-events-none overflow-hidden">
        <div class="absolute inset-0 bg-[grid-slate-800] [mask-image:radial-gradient(white,transparent)]"></div>
        <div class="matrix-lines absolute inset-0"></div>
      </div>
      
      <div class="absolute top-0 left-1/2 -translate-x-1/2 w-full h-full bg-[radial-gradient(circle_at_center,rgba(var(--matrix-color-rgb),0.2)_0%,transparent_70%)] -z-10 opacity-20"></div>
      <div class="max-w-7xl mx-auto text-center relative reveal">
        <div class="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-[var(--matrix-color)]/10 border border-[var(--matrix-color)]/20 text-[var(--matrix-color)] text-xs font-bold mb-8 animate-pulse shadow-[0_0_15px_rgba(var(--matrix-color-rgb),0.3)]">
          <Sparkles class="w-4 h-4" />
          {{ tt('earlymeow.home.badge') }}
        </div>
        <h1 class="text-6xl md:text-9xl font-black mb-12 tracking-tighter leading-none text-[var(--text-main)]">
          Early<span class="bg-clip-text text-transparent bg-gradient-to-r from-[var(--matrix-color)] to-[rgba(var(--matrix-color-rgb),0.5)] drop-shadow-[0_10px_10px_rgba(var(--matrix-color-rgb),0.3)]">Meow</span>
        </h1>
        <p class="text-xl md:text-3xl text-[var(--text-muted)] mb-16 max-w-4xl mx-auto font-light leading-relaxed">
          {{ tt('earlymeow.home.subtitle') }}
        </p>
        <div class="flex flex-col sm:flex-row gap-6 justify-center items-center">
          <button 
            @click="router.push('/console')"
            class="px-10 py-5 bg-[var(--matrix-color)] hover:bg-[var(--matrix-color)]/80 text-white rounded-2xl text-xl font-bold transition-all shadow-[0_0_40px_rgba(var(--matrix-color-rgb),0.4)] flex items-center gap-3 group"
          >
            {{ tt('earlymeow.home.cta_primary') }} <ArrowRight class="w-6 h-6 group-hover:translate-x-2 transition-transform" />
          </button>
          <button 
            @click="scrollTo('features')"
            class="px-10 py-5 bg-[var(--bg-body)]/50 hover:bg-[var(--bg-body)]/80 text-[var(--text-main)] rounded-2xl text-xl font-bold transition-all border border-[var(--border-color)] backdrop-blur-sm"
          >
            {{ tt('earlymeow.home.cta_secondary') }}
          </button>
        </div>
      </div>
    </header>

    <!-- Stats Section -->
    <section class="py-20 bg-[var(--bg-body)]/30 border-y border-[var(--border-color)]">
      <div class="max-w-7xl mx-auto px-4">
        <div class="grid grid-cols-1 md:grid-cols-3 gap-12">
          <div v-for="stat in stats" :key="stat.label" class="text-center group">
            <div class="text-5xl md:text-7xl font-black text-[var(--text-main)] mb-2 group-hover:text-[var(--matrix-color)] transition-colors">{{ stat.value }}</div>
            <div class="text-xs text-[var(--text-muted)] font-bold uppercase tracking-[0.3em]">{{ stat.label }}</div>
          </div>
        </div>
      </div>
    </section>

    <!-- Features Grid -->
    <section id="features" class="py-32 relative overflow-hidden">
      <div class="max-w-7xl mx-auto px-4">
        <div class="text-center mb-24">
          <h2 class="text-4xl md:text-6xl font-black mb-6 text-[var(--text-main)]">{{ tt('earlymeow.home.features.title') }}</h2>
          <p class="text-xl text-[var(--text-muted)] max-w-2xl mx-auto">
            {{ tt('earlymeow.home.features.desc') }}
          </p>
        </div>
        
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
          <div v-for="f in features" :key="f.title" class="p-8 rounded-[2rem] bg-[var(--bg-body)]/40 border border-[var(--border-color)] hover:border-[var(--matrix-color)]/50 transition-all duration-500 group relative overflow-hidden h-full">
            <div class="absolute inset-0 bg-gradient-to-br from-[var(--matrix-color)]/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity"></div>
            <div class="relative z-10">
              <div class="w-16 h-16 rounded-2xl bg-[var(--matrix-color)]/10 flex items-center justify-center mb-8 group-hover:scale-110 transition-transform shadow-[0_0_20px_rgba(var(--matrix-color-rgb),0.2)]">
                <component :is="f.icon" class="w-8 h-8 text-[var(--matrix-color)]" />
              </div>
              <h3 class="text-2xl font-black mb-4 text-[var(--text-main)]">{{ f.title }}</h3>
              <p class="text-[var(--text-muted)] leading-relaxed text-lg">{{ f.desc }}</p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Core Brain Section -->
    <section class="py-32 relative overflow-hidden">
      <div class="max-w-7xl mx-auto px-4">
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-24 items-center">
          <div class="relative reveal">
            <div class="absolute -top-24 -left-24 w-64 h-64 bg-[var(--matrix-color)]/20 blur-[100px] -z-10"></div>
            <div class="bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-[3rem] p-12 aspect-square flex items-center justify-center relative overflow-hidden">
              <div class="absolute inset-0 opacity-20">
                <div class="absolute top-0 left-0 w-full h-full bg-[grid-slate-800] [mask-image:radial-gradient(white,transparent)]"></div>
              </div>
              <!-- Central AI Identity -->
              <div class="relative z-10 flex flex-col items-center">
                <div class="w-40 h-40 bg-gradient-to-br from-[var(--matrix-color)] to-[rgba(var(--matrix-color-rgb),0.4)] rounded-full flex items-center justify-center shadow-2xl animate-pulse">
                  <Bot class="w-20 h-20 text-white" />
                </div>
                <div class="mt-8 space-y-4 text-center">
                  <div class="px-6 py-2 rounded-full bg-[var(--matrix-color)] text-white text-sm font-black tracking-widest uppercase">
                    {{ tt('earlymeow.home.core_status_active') }}
                  </div>
                  <div class="flex gap-2 justify-center">
                    <div v-for="i in 3" :key="i" class="w-2 h-2 rounded-full bg-[var(--matrix-color)] animate-bounce" :style="{ animationDelay: `${i * 0.2}s` }"></div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div class="reveal">
            <h2 class="text-5xl md:text-7xl font-black mb-12 tracking-tighter leading-none text-[var(--text-main)]">
              {{ tt('earlymeow.home.core_title') }}<br/>
              <span class="text-[var(--matrix-color)]">{{ tt('earlymeow.home.core_subtitle') }}</span>
            </h2>
            <div class="space-y-12">
              <div v-for="item in [
                { title: tt('earlymeow.home.core_feature1_title'), desc: tt('earlymeow.home.core_feature1_desc') },
                { title: tt('earlymeow.home.core_feature2_title'), desc: tt('earlymeow.home.core_feature2_desc') },
                { title: tt('earlymeow.home.core_feature3_title'), desc: tt('earlymeow.home.core_feature3_desc') }
              ]" :key="item.title" class="flex gap-8 group">
                <div class="w-2 h-auto bg-[var(--border-color)] rounded-full overflow-hidden">
                  <div class="w-full h-0 group-hover:h-full bg-[var(--matrix-color)] transition-all duration-700"></div>
                </div>
                <div>
                  <h4 class="text-2xl font-black mb-4 text-[var(--text-main)] group-hover:text-[var(--matrix-color)] transition-colors">{{ item.title }}</h4>
                  <p class="text-xl text-[var(--text-muted)] leading-relaxed font-light">{{ item.desc }}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Case Studies / Digital Employees -->
    <section class="py-32 bg-[var(--bg-body)]/30">
      <div class="max-w-7xl mx-auto px-4">
        <div class="flex flex-col lg:flex-row justify-between items-end gap-12 mb-24">
          <div class="max-w-2xl">
            <h2 class="text-4xl md:text-6xl font-black mb-8 text-[var(--text-main)]">{{ tt('earlymeow.home.employees.title') }}</h2>
            <p class="text-xl text-[var(--text-muted)]">
              {{ tt('earlymeow.home.employees.desc') }}
            </p>
          </div>
          <button 
            @click="router.push('/meow/pricing')"
            class="px-8 py-4 rounded-2xl bg-[var(--matrix-color)]/10 border border-[var(--border-color)]/20 text-[var(--matrix-color)] font-bold hover:bg-[var(--matrix-color)] hover:text-white transition-all"
          >
            {{ tt('earlymeow.home.employees.view_all') }}
          </button>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
          <div v-for="emp in digitalEmployees" :key="emp.id" class="p-12 rounded-[3.5rem] bg-[var(--bg-body)] border border-[var(--border-color)] hover:border-[var(--matrix-color)]/30 transition-all group relative overflow-hidden">
            <div class="absolute top-0 right-0 w-64 h-64 bg-[var(--matrix-color)]/5 blur-[100px] -z-10 group-hover:bg-[var(--matrix-color)]/10 transition-all"></div>
            
            <div class="flex justify-between items-start mb-12">
              <div class="flex items-center gap-6">
                <div class="w-20 h-20 rounded-3xl bg-[var(--bg-body)]/50 border border-[var(--border-color)] flex items-center justify-center text-[var(--matrix-color)] group-hover:scale-110 transition-transform shadow-xl">
                  <UserCircle class="w-10 h-10" />
                </div>
                <div>
                  <h3 class="text-3xl font-black text-[var(--text-main)]">{{ emp.name }}</h3>
                  <div class="text-[var(--matrix-color)] font-bold text-sm uppercase tracking-widest">{{ emp.role }}</div>
                </div>
              </div>
              <div class="px-4 py-2 bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] text-xs font-black rounded-full border border-[var(--matrix-color)]/20 uppercase tracking-widest">
                {{ emp.status }}
              </div>
            </div>

            <div class="grid grid-cols-3 gap-8 mb-12">
              <div class="space-y-2">
                <div class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ tt('earlymeow.home.employees.kpi') }}</div>
                <div class="text-2xl font-black text-[var(--text-main)]">{{ emp.kpi }}%</div>
                <div class="w-full h-1 bg-[var(--border-color)] rounded-full overflow-hidden">
                  <div class="h-full bg-[var(--matrix-color)]" :style="{ width: emp.kpi + '%' }"></div>
                </div>
              </div>
              <div class="space-y-2">
                <div class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ tt('earlymeow.home.employees.roi') }}</div>
                <div class="text-2xl font-black text-[var(--matrix-color)]">{{ emp.roi }}x</div>
              </div>
              <div class="space-y-2">
                <div class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ tt('earlymeow.home.employees.tasks') }}</div>
                <div class="text-2xl font-black text-[var(--text-main)]">{{ emp.taskCount }}</div>
              </div>
            </div>

            <div class="flex flex-wrap gap-2">
              <span v-for="skill in emp.skills" :key="skill" class="px-4 py-2 bg-[var(--bg-body)]/80 border border-[var(--border-color)] rounded-xl text-xs font-bold text-[var(--text-muted)]">{{ skill }}</span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Bottom CTA -->
    <section class="py-40 relative overflow-hidden border-t border-[var(--border-color)]">
      <div class="absolute inset-0 bg-[radial-gradient(circle_at_center,var(--matrix-color,rgba(168,85,247,0.1))_0%,transparent_70%)] opacity-20"></div>
      <div class="max-w-4xl mx-auto px-4 text-center relative z-10">
        <h2 class="text-5xl md:text-8xl font-black mb-12 tracking-tighter leading-tight text-[var(--text-main)]">
          {{ tt('earlymeow.home.cta_final.title') }}
        </h2>
        <p class="text-2xl text-[var(--text-muted)] mb-16 font-light">
          {{ tt('earlymeow.home.cta_final.desc') }}
        </p>
        <div class="flex flex-col sm:flex-row gap-8 justify-center items-center">
          <button 
            @click="router.push('/console')"
            class="px-12 py-6 bg-[var(--matrix-color)] hover:bg-[var(--matrix-color)]/80 text-white rounded-[2rem] text-2xl font-black transition-all shadow-[0_0_50px_rgba(var(--matrix-color-rgb),0.3)] hover:scale-105 active:scale-95"
          >
            {{ tt('earlymeow.home.cta_final.primary') }}
          </button>
          <button 
            @click="router.push('/about')"
            class="px-12 py-6 bg-[var(--bg-body)]/50 hover:bg-[var(--bg-body)]/80 text-[var(--text-main)] rounded-[2rem] text-2xl font-black transition-all border border-[var(--border-color)]"
          >
            {{ tt('earlymeow.home.cta_final.secondary') }}
          </button>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.animate-bounce-slow {
  animation: bounce 4s infinite ease-in-out;
}

.animate-spin-slow {
  animation: spin 8s linear infinite;
}

.animate-pulse-slow {
  animation: pulse 6s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-20px); }
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>

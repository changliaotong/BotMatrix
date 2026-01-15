<script setup lang="ts">
import { ref, computed } from 'vue';
import { useSystemStore } from '@/stores/system';
import { useI18n } from '@/utils/i18n';
import { 
  Cpu, Zap, Shield, Database, Network, Code2, 
  Workflow, Layers, MessageSquare, Bot, Sparkles,
  BrainCircuit, Gauge, Terminal, Globe, Share2,
  Users, Activity
} from 'lucide-vue-next';

const systemStore = useSystemStore();
const { t: tt } = useI18n();

const techStack = computed(() => [
  {
    name: tt('earlymeow.tech.stack.nexus.name'),
    role: tt('earlymeow.tech.stack.nexus.role'),
    desc: tt('earlymeow.tech.stack.nexus.desc'),
    icon: Network,
    color: 'purple'
  },
  {
    name: tt('earlymeow.tech.stack.llm.name'),
    role: tt('earlymeow.tech.stack.llm.role'),
    desc: tt('earlymeow.tech.stack.llm.desc'),
    icon: BrainCircuit,
    color: 'purple'
  },
  {
    name: tt('earlymeow.tech.stack.memory.name'),
    role: tt('earlymeow.tech.stack.memory.role'),
    desc: tt('earlymeow.tech.stack.memory.desc'),
    icon: Database,
    color: 'purple'
  }
]);

const performanceMetrics = computed(() => [
  { label: tt('earlymeow.tech.metrics.rtt'), value: '< 15ms', icon: Gauge, color: 'text-[var(--matrix-color)]' },
  { label: tt('earlymeow.tech.metrics.concurrency'), value: '50k+', icon: Users, color: 'text-blue-500' },
  { label: tt('earlymeow.tech.metrics.uptime'), value: '99.99%', icon: Activity, color: 'text-purple-500' },
  { label: tt('earlymeow.tech.metrics.throughput'), value: '2.4 GB/s', icon: Zap, color: 'text-yellow-500' }
]);
</script>

<template>
  <div class="pt-40 pb-20 px-6 max-w-7xl mx-auto relative z-10">
    <!-- Header -->
    <div class="mb-32 space-y-6 text-center flex flex-col items-center">
      <div class="inline-flex items-center gap-2 text-[var(--matrix-color)] font-black text-xs uppercase tracking-widest px-3 py-1 rounded-full bg-[var(--matrix-color)]/10 border border-[var(--matrix-color)]/20">
        <Code2 class="w-4 h-4" />
        {{ tt('earlymeow.tech.header.tag') }}
      </div>
      <h1 class="text-6xl md:text-8xl font-black tracking-tighter leading-none text-[var(--text-main)]">
        {{ tt('earlymeow.tech.header.title_prefix') }}<br/>
        <span class="text-transparent bg-clip-text bg-gradient-to-r from-[var(--matrix-color)] to-[var(--matrix-color)]/80">
          {{ tt('earlymeow.tech.header.title_suffix') }}
        </span>
      </h1>
      <p class="text-xl text-[var(--text-muted)] font-medium max-w-2xl leading-relaxed mx-auto">
        {{ tt('earlymeow.tech.header.desc') }}
      </p>
    </div>

    <!-- Tech Stack Grid -->
    <div class="grid lg:grid-cols-3 gap-10 mb-40">
      <div 
        v-for="tech in techStack" 
        :key="tech.name"
        class="group p-10 rounded-[40px] border border-[var(--border-color)] bg-[var(--bg-body)]/40 backdrop-blur-xl hover:border-[var(--matrix-color)]/30 transition-all duration-500"
      >
        <div 
          class="w-20 h-20 rounded-2xl mb-10 flex items-center justify-center transition-transform group-hover:scale-110 group-hover:rotate-6 bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] border border-[var(--matrix-color)]/20"
        >
          <component :is="tech.icon || Cpu" class="w-10 h-10" />
        </div>
        <div class="space-y-4">
          <div class="text-[10px] uppercase font-black tracking-[0.2em] text-[var(--matrix-color)]/60">{{ tech.role }}</div>
          <h3 class="text-3xl font-black text-[var(--text-main)]">{{ tech.name }}</h3>
          <p class="text-sm text-[var(--text-muted)] font-medium leading-relaxed">
            {{ tech.desc }}
          </p>
        </div>
      </div>
    </div>

    <!-- Performance Metrics -->
    <div class="py-32">
      <div class="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
        <div 
          v-for="stat in performanceMetrics" 
          :key="stat.label"
          class="p-10 rounded-[32px] bg-[var(--bg-body)]/40 border border-[var(--border-color)] flex flex-col items-center text-center group hover:border-[var(--matrix-color)]/20 transition-all"
        >
          <div class="w-16 h-16 rounded-2xl bg-[var(--bg-body)] flex items-center justify-center mb-6 group-hover:scale-110 transition-transform">
            <component :is="stat.icon" class="w-8 h-8" :class="stat.color" />
          </div>
          <div class="text-4xl font-black text-[var(--text-main)] mb-2">{{ stat.value }}</div>
          <div class="text-[10px] uppercase font-black tracking-[0.2em] text-[var(--text-muted)]/60">{{ stat.label }}</div>
        </div>
      </div>
    </div>

    <!-- Deep Dive -->
    <div class="grid lg:grid-cols-2 gap-20 items-center py-20 border-t border-[var(--border-color)]">
      <div class="space-y-10">
        <h2 class="text-4xl font-black tracking-tighter text-[var(--text-main)]">
          {{ tt('earlymeow.tech.why.title_prefix') }} 
          <span class="text-[var(--matrix-color)]">EarlyMeow</span> 
          {{ tt('earlymeow.tech.why.title_suffix') }}
        </h2>
        
        <div class="space-y-8">
          <div class="flex gap-6 group">
            <div class="w-12 h-12 rounded-xl bg-[var(--bg-body)]/50 border border-[var(--border-color)] flex-shrink-0 flex items-center justify-center group-hover:border-[var(--matrix-color)]/30 transition-colors">
              <Layers class="w-6 h-6 text-[var(--matrix-color)]" />
            </div>
            <div class="space-y-2">
              <h4 class="text-xl font-bold text-[var(--text-main)]">{{ tt('earlymeow.tech.features.plugin.title') }}</h4>
              <p class="text-sm text-[var(--text-muted)] font-medium leading-relaxed">{{ tt('earlymeow.tech.features.plugin.desc') }}</p>
            </div>
          </div>
          
          <div class="flex gap-6 group">
            <div class="w-12 h-12 rounded-xl bg-[var(--bg-body)]/50 border border-[var(--border-color)] flex-shrink-0 flex items-center justify-center group-hover:border-[var(--matrix-color)]/30 transition-colors">
              <Workflow class="w-6 h-6 text-[var(--matrix-color)]" />
            </div>
            <div class="space-y-2">
              <h4 class="text-xl font-bold text-[var(--text-main)]">{{ tt('earlymeow.tech.features.workflow.title') }}</h4>
              <p class="text-sm text-[var(--text-muted)] font-medium leading-relaxed">{{ tt('earlymeow.tech.features.workflow.desc') }}</p>
            </div>
          </div>

          <div class="flex gap-6 group">
            <div class="w-12 h-12 rounded-xl bg-[var(--bg-body)]/50 border border-[var(--border-color)] flex-shrink-0 flex items-center justify-center group-hover:border-[var(--matrix-color)]/30 transition-colors">
              <Shield class="w-6 h-6 text-[var(--matrix-color)]" />
            </div>
            <div class="space-y-2">
              <h4 class="text-xl font-bold text-[var(--text-main)]">{{ tt('earlymeow.tech.features.security.title') }}</h4>
              <p class="text-sm text-[var(--text-muted)] font-medium leading-relaxed">{{ tt('earlymeow.tech.features.security.desc') }}</p>
            </div>
          </div>
        </div>
      </div>

      <div class="relative aspect-video rounded-[40px] bg-[var(--bg-body)]/60 border border-[var(--border-color)] overflow-hidden group backdrop-blur-sm">
        <div class="absolute inset-0 flex items-center justify-center p-8">
          <div class="w-full max-w-md p-6 bg-[var(--bg-body)]/80 border border-[var(--border-color)] rounded-2xl backdrop-blur-md shadow-2xl space-y-4 font-mono text-[10px]">
            <div class="flex items-center gap-2 mb-4">
              <div class="w-2 h-2 rounded-full bg-red-500/80"></div>
              <div class="w-2 h-2 rounded-full bg-yellow-500/80"></div>
              <div class="w-2 h-2 rounded-full bg-green-500/80"></div>
              <span class="ml-2 text-[var(--text-muted)]/60">meow-kernel-v2.log</span>
            </div>
            <div class="space-y-1">
              <div class="text-[var(--matrix-color)]">{{ tt('earlymeow.tech.log.init') }}</div>
              <div class="text-[var(--text-muted)]">{{ tt('earlymeow.tech.log.connect') }}</div>
              <div class="text-[var(--matrix-color)]/80">{{ tt('earlymeow.tech.log.thinking') }}</div>
              <div class="text-[var(--matrix-color)]">{{ tt('earlymeow.tech.log.executing') }}</div>
              <div class="text-[var(--text-muted)]/40">...</div>
              <div class="text-[var(--matrix-color)]/80">{{ tt('earlymeow.tech.log.success') }}</div>
            </div>
          </div>
        </div>
        <!-- Decorative background elements -->
        <div class="absolute -top-20 -right-20 w-40 h-40 bg-[var(--matrix-color)]/10 rounded-full blur-[80px]"></div>
        <div class="absolute -bottom-20 -left-20 w-40 h-40 bg-[var(--matrix-color)]/20 rounded-full blur-[80px]"></div>
        <!-- Border glow -->
        <div class="absolute inset-0 pointer-events-none border border-[var(--matrix-color)]/0 group-hover:border-[var(--matrix-color)]/20 transition-all duration-700"></div>
      </div>
    </div>

    <!-- Protocols & Ecosystem -->
    <div id="ecosystem" class="py-40 border-t border-[var(--border-color)]">
      <div class="grid lg:grid-cols-2 gap-20">
        <div class="space-y-12">
          <div class="space-y-6">
            <h2 class="text-4xl font-black text-[var(--text-main)] tracking-tighter">{{ tt('earlymeow.tech.protocols.title') }}</h2>
            <p class="text-[var(--text-muted)] font-medium">{{ tt('earlymeow.tech.protocols.desc') }}</p>
          </div>
          
          <div class="grid grid-cols-2 gap-6">
            <div v-for="p in ['OneBot v11', 'OneBot v12', 'Matrix Protocol', 'Custom HTTP/WS']" :key="p" class="flex items-center gap-3 p-4 rounded-xl bg-[var(--bg-body)]/50 border border-[var(--border-color)]">
              <div class="w-2 h-2 rounded-full bg-[var(--matrix-color)]"></div>
              <span class="text-sm font-bold text-[var(--text-muted)]">{{ p }}</span>
            </div>
          </div>
        </div>

        <div class="space-y-12">
          <div class="space-y-6">
            <h2 class="text-4xl font-black text-[var(--text-main)] tracking-tighter">{{ tt('earlymeow.tech.ecosystem.title') }}</h2>
            <p class="text-[var(--text-muted)] font-medium">{{ tt('earlymeow.tech.ecosystem.desc') }}</p>
          </div>
          
          <div class="flex flex-wrap gap-4">
            <div v-for="t in ['TypeScript SDK', 'Python Bridge', 'CLI Tool', 'Debug Panel']" :key="t" class="px-6 py-3 rounded-full bg-[var(--matrix-color)]/10 border border-[var(--matrix-color)]/20 text-[var(--matrix-color)] text-xs font-bold">
              {{ t }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Simplified styles, using Tailwind for most parts */
</style>

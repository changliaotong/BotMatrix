<template>
  <div class="min-h-screen bg-[var(--bg-body)] text-[var(--text-main)] selection:bg-[var(--matrix-color)]/30 overflow-x-hidden" :class="[systemStore.style]">
    <PortalHeader />

    <!-- Hero -->
    <header class="pt-48 pb-32 px-6 relative overflow-hidden">
      <!-- Background Kinetic Effects -->
      <div class="absolute inset-0 pointer-events-none">
        <div class="absolute top-0 left-1/2 -translate-x-1/2 w-full h-full bg-[radial-gradient(circle_at_center,var(--matrix-color,rgba(168,85,247,0.1))_0%,transparent_70%)] opacity-20"></div>
        <div class="absolute inset-0 bg-[grid-slate-800] [mask-image:radial-gradient(white,transparent)] opacity-20"></div>
      </div>

      <div class="max-w-5xl mx-auto text-center relative">
        <div class="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-[var(--matrix-color)]/10 border border-[var(--matrix-color)]/20 text-[var(--matrix-color)] text-xs font-bold mb-10 animate-pulse">
          {{ tt('portal.about.badge') }}
        </div>
        <h1 class="text-6xl md:text-8xl font-black mb-12 tracking-tighter leading-none uppercase text-[var(--text-main)]">
          {{ tt('portal.about.title_prefix') }} <br/>
          <span class="text-transparent bg-clip-text bg-gradient-to-r from-[var(--matrix-color)] via-blue-500 to-cyan-400">{{ tt('portal.about.title_suffix') }}</span>
        </h1>
        <p class="text-xl text-[var(--text-muted)] leading-relaxed max-w-3xl mx-auto font-light">
          {{ tt('portal.about.subtitle') }}
        </p>
      </div>
    </header>

    <!-- Project Vision -->
    <section class="py-32 relative">
      <div class="max-w-7xl mx-auto px-6">
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-24 items-center">
          <div class="space-y-16">
            <h2 class="text-4xl font-black tracking-tighter uppercase italic text-[var(--text-main)]" v-html="tt('portal.about.visions_title')"></h2>
            
            <div class="grid grid-cols-1 gap-12">
              <div v-for="vision in visions" :key="vision.title" class="group relative p-8 rounded-[2.5rem] border border-[var(--border-color)] bg-[var(--bg-body)]/50 hover:border-[var(--matrix-color)]/30 transition-all">
                <div class="flex gap-6 items-start">
                  <div class="flex-shrink-0 w-14 h-14 bg-[var(--matrix-color)]/10 rounded-2xl flex items-center justify-center text-[var(--matrix-color)] group-hover:bg-[var(--matrix-color)] group-hover:text-white transition-all duration-500">
                    <component :is="vision.icon" class="w-7 h-7" />
                  </div>
                  <div>
                    <h4 class="text-xl font-bold mb-3 tracking-tight group-hover:text-[var(--matrix-color)] transition-colors uppercase">{{ vision.title }}</h4>
                    <p class="text-[var(--text-muted)] leading-relaxed font-light group-hover:text-[var(--text-main)] transition-colors">{{ vision.desc }}</p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="relative">
            <div class="bg-[var(--bg-body)]/50 p-12 rounded-[3rem] border border-[var(--border-color)] relative overflow-hidden group">
              <div class="absolute top-0 right-0 w-full h-full bg-[radial-gradient(circle_at_top_right,var(--matrix-color,rgba(168,85,247,0.05)),transparent_50%)] opacity-20"></div>
              <h3 class="text-2xl font-black mb-12 flex items-center gap-4 uppercase tracking-tighter italic text-[var(--text-main)]" v-html="tt('portal.about.stack_title')"></h3>
              <div class="space-y-6">
                <div v-for="tech in techStack" :key="tech.name" class="flex items-center justify-between p-6 bg-[var(--bg-body)]/50 rounded-2xl border border-[var(--border-color)] hover:border-[var(--matrix-color)]/30 transition-all group">
                  <div class="flex items-center gap-4">
                    <div class="w-3 h-3 rounded-full bg-[var(--border-color)] group-hover:bg-[var(--matrix-color)] transition-all"></div>
                    <span class="font-bold text-[var(--text-muted)] group-hover:text-[var(--text-main)] transition-colors uppercase tracking-widest text-sm">{{ tech.name }}</span>
                  </div>
                  <span class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest group-hover:text-[var(--matrix-color)] transition-colors">{{ tech.role }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Development Roadmap -->
    <section class="py-32 bg-[var(--bg-body)]/30 relative border-y border-[var(--border-color)]">
      <div class="max-w-7xl mx-auto px-6">
        <div class="text-center mb-24">
          <h2 class="text-4xl md:text-6xl font-black mb-8 tracking-tighter uppercase italic text-[var(--text-main)]" v-html="tt('portal.about.roadmap_title')"></h2>
          <p class="text-[var(--text-muted)] font-light max-w-2xl mx-auto">{{ tt('portal.about.roadmap_subtitle') }}</p>
        </div>
        
        <div class="relative">
          <!-- Timeline line -->
          <div class="absolute left-1/2 -translate-x-1/2 top-0 bottom-0 w-px bg-[var(--border-color)] hidden md:block"></div>
          
          <div class="space-y-24">
            <div v-for="(milestone, index) in roadmap" :key="milestone.title" 
                 class="relative flex flex-col md:flex-row gap-12 items-center"
                 :class="index % 2 === 0 ? 'md:flex-row-reverse' : ''">
              <!-- Point -->
              <div class="absolute left-1/2 -translate-x-1/2 w-6 h-6 rounded-full bg-[var(--bg-body)] border-4 border-[var(--matrix-color)] shadow-[0_0_20px_rgba(var(--matrix-color-rgb),0.3)] z-10 hidden md:block">
                <div class="absolute inset-0 bg-[var(--matrix-color)] rounded-full animate-ping opacity-20"></div>
              </div>
              
              <div class="flex-1 w-full md:w-1/2">
                <div class="group p-10 rounded-[3rem] border border-[var(--border-color)] bg-[var(--bg-body)] hover:border-[var(--matrix-color)]/30 transition-all duration-500"
                     :class="index % 2 === 0 ? 'text-left' : 'text-left md:text-right'">
                  <div class="inline-block px-4 py-1 rounded-full bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] text-[10px] font-bold uppercase tracking-widest mb-6">
                    {{ milestone.date }}
                  </div>
                  <h4 class="text-2xl font-bold mb-4 tracking-tight uppercase text-[var(--text-main)] group-hover:text-[var(--matrix-color)] transition-colors">{{ milestone.title }}</h4>
                  <p class="text-[var(--text-muted)] leading-relaxed font-light group-hover:text-[var(--text-main)] transition-colors">{{ milestone.desc }}</p>
                </div>
              </div>
              <div class="flex-1 hidden md:block"></div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Join Us CTA -->
    <section class="py-48 text-center relative overflow-hidden">
      <div class="absolute inset-0 bg-[radial-gradient(circle_at_center,rgba(var(--matrix-color-rgb),0.05)_0%,transparent_70%)]"></div>
      <div class="max-w-4xl mx-auto px-6 relative z-10">
        <div class="inline-block p-4 rounded-full bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] mb-8">
          <Github class="w-8 h-8" />
        </div>
        <h2 class="text-5xl md:text-8xl font-black mb-12 tracking-tighter uppercase italic leading-none text-[var(--text-main)]" v-html="tt('portal.about.cta_title')"></h2>
        <p class="text-xl text-[var(--text-muted)] mb-16 font-light leading-relaxed">
          {{ tt('portal.about.cta_subtitle') }}
        </p>
        <div class="flex flex-col md:flex-row gap-6 justify-center items-center">
          <a href="https://github.com/changliaotong/BotMatrix" target="_blank" class="px-12 py-5 bg-[var(--matrix-color)] hover:bg-[var(--matrix-color)]/80 text-white rounded-2xl font-bold uppercase tracking-widest transition-all shadow-xl shadow-[var(--matrix-color)]/20 flex items-center gap-3">
            <Github class="w-5 h-5" />
            {{ tt('portal.about.cta_github') }}
          </a>
          <button class="px-12 py-5 bg-[var(--bg-body)]/50 hover:bg-[var(--bg-body)] text-[var(--text-main)] border border-[var(--border-color)] rounded-2xl font-bold uppercase tracking-widest transition-all">
            {{ tt('portal.about.cta_contact') }}
          </button>
        </div>
      </div>
    </section>

    <PortalFooter />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import PortalHeader from '@/components/layout/PortalHeader.vue';
import PortalFooter from '@/components/layout/PortalFooter.vue';
import { useSystemStore } from '@/stores/system';
import { Layers, Shield, Code, Globe, Github, Settings } from 'lucide-vue-next';
import { useI18n } from '@/utils/i18n';

const { t: tt } = useI18n();
const systemStore = useSystemStore();

const visions = computed(() => [
  {
    icon: Globe,
    title: tt('portal.about.vision1_title'),
    desc: tt('portal.about.vision1_desc')
  },
  {
    icon: Layers,
    title: tt('portal.about.vision2_title'),
    desc: tt('portal.about.vision2_desc')
  },
  {
    icon: Shield,
    title: tt('portal.about.vision3_title'),
    desc: tt('portal.about.vision3_desc')
  }
]);

const techStack = computed(() => [
  { name: 'Golang', role: tt('portal.about.tech_golang') },
  { name: 'Vue 3 + TS', role: tt('portal.about.tech_vue') },
  { name: 'PostgreSQL + pgvector', role: tt('portal.about.tech_pg') },
  { name: 'Docker + K8s', role: tt('portal.about.tech_docker') },
  { name: 'Model Context Protocol', role: tt('portal.about.tech_mcp') }
]);

const roadmap = computed(() => [
  {
    date: tt('portal.about.roadmap1_date'),
    title: tt('portal.about.roadmap1_title'),
    desc: tt('portal.about.roadmap1_desc')
  },
  {
    date: tt('portal.about.roadmap2_date'),
    title: tt('portal.about.roadmap2_title'),
    desc: tt('portal.about.roadmap2_desc')
  },
  {
    date: tt('portal.about.roadmap3_date'),
    title: tt('portal.about.roadmap3_title'),
    desc: tt('portal.about.roadmap3_desc')
  },
  {
    date: tt('portal.about.roadmap4_date'),
    title: tt('portal.about.roadmap4_title'),
    desc: tt('portal.about.roadmap4_desc')
  }
]);
</script>

<style scoped>
.animate-spin-slow {
  animation: spin 8s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>

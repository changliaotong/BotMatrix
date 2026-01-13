<script setup lang="ts">
import { ref, computed } from 'vue';
import { useI18n } from '@/utils/i18n';
import { 
  Puzzle, Globe, Zap, Code2, Rocket, Heart, 
  MessageSquare, Image, Gamepad2, ShieldCheck, 
  ArrowUpRight, Download, Users, Star
} from 'lucide-vue-next';

const { t: tt } = useI18n();

const categories = computed(() => [
  { id: 'all', name: tt('earlymeow.ecosystem.cats.all') },
  { id: 'ai', name: tt('earlymeow.ecosystem.cats.ai') },
  { id: 'tools', name: tt('earlymeow.ecosystem.cats.tools') },
  { id: 'entertainment', name: tt('earlymeow.ecosystem.cats.fun') },
  { id: 'dev', name: tt('earlymeow.ecosystem.cats.dev') }
]);

const activeCategory = ref('all');

const plugins = computed(() => [
  {
    name: tt('earlymeow.ecosystem.plugins.nlp.name'),
    desc: tt('earlymeow.ecosystem.plugins.nlp.desc'),
    category: 'ai',
    author: tt('earlymeow.ecosystem.plugins.author_official'),
    downloads: '12k',
    rating: 4.9,
    icon: MessageSquare,
    color: 'purple'
  },
  {
    name: tt('earlymeow.ecosystem.plugins.draw.name'),
    desc: tt('earlymeow.ecosystem.plugins.draw.desc'),
    category: 'ai',
    author: tt('earlymeow.ecosystem.plugins.author_matrix'),
    downloads: '8.5k',
    rating: 4.8,
    icon: Image,
    color: 'blue'
  },
  {
    name: tt('earlymeow.ecosystem.plugins.rpg.name'),
    desc: tt('earlymeow.ecosystem.plugins.rpg.desc'),
    category: 'entertainment',
    author: tt('earlymeow.ecosystem.plugins.author_community'),
    downloads: '25k',
    rating: 4.7,
    icon: Gamepad2,
    color: 'pink'
  },
  {
    name: tt('earlymeow.ecosystem.plugins.admin.name'),
    desc: tt('earlymeow.ecosystem.plugins.admin.desc'),
    category: 'tools',
    author: tt('earlymeow.ecosystem.plugins.author_official'),
    downloads: '15k',
    rating: 4.9,
    icon: ShieldCheck,
    color: 'emerald'
  }
]);

const filteredPlugins = computed(() => {
  if (activeCategory.value === 'all') return plugins.value;
  return plugins.value.filter(p => p.category === activeCategory.value);
});
</script>

<template>
  <div class="py-20 px-6 max-w-7xl mx-auto relative z-10">
    <!-- Header -->
    <div class="mb-20 space-y-6">
      <div class="inline-flex items-center gap-2 text-[var(--matrix-color)] font-black text-xs uppercase tracking-widest px-3 py-1 rounded-full bg-[var(--matrix-color)]/10 border border-[var(--matrix-color)]/20">
        <Puzzle class="w-4 h-4" />
        {{ tt('earlymeow.ecosystem.header.tag') }}
      </div>
      <h1 class="text-6xl md:text-8xl font-black tracking-tighter leading-none text-[var(--text-main)]">
        {{ tt('earlymeow.ecosystem.header.title_prefix') }}<br/>
        <span class="bg-clip-text text-transparent bg-gradient-to-r from-[var(--matrix-color)] to-[rgba(var(--matrix-color-rgb),0.5)]">
          {{ tt('earlymeow.ecosystem.header.title_suffix') }}
        </span>
      </h1>
      <p class="text-xl text-[var(--text-muted)] font-medium max-w-2xl leading-relaxed">
        {{ tt('earlymeow.ecosystem.header.desc') }}
      </p>
    </div>

    <!-- Marketplace Sections -->
    <div class="space-y-32">
      <!-- Category Tabs -->
      <div class="flex flex-wrap gap-4 border-b border-[var(--border-color)] pb-8">
        <button 
          v-for="cat in categories" 
          :key="cat.id"
          @click="activeCategory = cat.id"
          class="px-6 py-2 rounded-full text-sm font-bold transition-all border"
          :class="activeCategory === cat.id ? 'bg-[var(--matrix-color)] border-[var(--matrix-color)] text-white shadow-lg shadow-[var(--matrix-color)]/30' : 'bg-[var(--bg-body)]/50 border-[var(--border-color)] text-[var(--text-muted)] hover:text-[var(--text-main)] hover:border-[var(--text-main)]/20'"
        >
          {{ cat.name }}
        </button>
      </div>

      <!-- Plugin Grid -->
      <div class="grid md:grid-cols-2 gap-8">
        <div 
          v-for="plugin in filteredPlugins" 
          :key="plugin.name"
          class="group p-8 rounded-[32px] bg-[var(--bg-body)] border border-[var(--border-color)] hover:border-[var(--matrix-color)]/30 transition-all flex gap-8 relative overflow-hidden"
        >
          <div class="absolute top-0 right-0 w-32 h-32 bg-[var(--matrix-color)]/5 blur-[50px] -z-10 group-hover:bg-[var(--matrix-color)]/10 transition-all"></div>
          
          <div class="w-24 h-24 rounded-2xl flex items-center justify-center shrink-0 shadow-xl bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] group-hover:scale-110 transition-transform">
            <component :is="plugin.icon" class="w-10 h-10" />
          </div>
          
          <div class="space-y-4">
            <div class="flex justify-between items-start">
              <div>
                <h3 class="text-2xl font-black text-[var(--text-main)] group-hover:text-[var(--matrix-color)] transition-colors">{{ plugin.name }}</h3>
                <div class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ plugin.author }}</div>
              </div>
              <div class="flex items-center gap-1.5 px-3 py-1 rounded-lg bg-[var(--bg-body)] border border-[var(--border-color)] text-yellow-400 font-black text-xs">
                <Star class="w-3 h-3 fill-current" /> {{ plugin.rating }}
              </div>
            </div>
            
            <p class="text-sm text-[var(--text-muted)] leading-relaxed line-clamp-2">
              {{ plugin.desc }}
            </p>
            
            <div class="flex items-center justify-between pt-4">
              <div class="flex items-center gap-4 text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">
                <span class="flex items-center gap-1.5"><Download class="w-3 h-3" /> {{ plugin.downloads }} {{ tt('earlymeow.ecosystem.plugins.installs') }}</span>
                <span class="flex items-center gap-1.5"><Users class="w-3 h-3" /> {{ (Math.random() * 50 + 10).toFixed(0) }} {{ tt('earlymeow.ecosystem.plugins.devs') }}</span>
              </div>
              <button class="w-10 h-10 rounded-xl bg-[var(--matrix-color)] text-white flex items-center justify-center hover:scale-110 active:scale-95 transition-all shadow-lg shadow-[var(--matrix-color)]/30">
                <ArrowUpRight class="w-5 h-5" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Developer Section -->
    <div class="mt-40 p-12 md:p-24 rounded-[4rem] bg-gradient-to-br from-[var(--matrix-color)]/10 to-[rgba(var(--matrix-color-rgb),0.1)] border border-[var(--matrix-color)]/20 relative overflow-hidden">
      <div class="absolute top-0 right-0 p-12 opacity-5 rotate-12">
        <Code2 class="w-64 h-64 text-[var(--matrix-color)]" />
      </div>
      
      <div class="relative z-10 max-w-3xl space-y-10">
        <h2 class="text-4xl md:text-6xl font-black text-[var(--text-main)] tracking-tighter">{{ tt('earlymeow.ecosystem.dev.title') }}</h2>
        <p class="text-xl text-[var(--text-muted)] font-medium leading-relaxed">
          {{ tt('earlymeow.ecosystem.dev.desc') }}
        </p>
        <div class="flex flex-wrap gap-6">
          <button class="px-10 py-5 rounded-2xl bg-[var(--matrix-color)] text-white font-black text-lg hover:bg-[var(--matrix-color)]/80 transition-all shadow-xl shadow-[var(--matrix-color)]/30">
            {{ tt('earlymeow.ecosystem.dev.cta_docs') }}
          </button>
          <button class="px-10 py-5 rounded-2xl bg-[var(--bg-body)]/50 border border-[var(--border-color)] text-[var(--text-main)] font-black text-lg hover:bg-[var(--bg-body)] transition-all">
            {{ tt('earlymeow.ecosystem.dev.cta_portal') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

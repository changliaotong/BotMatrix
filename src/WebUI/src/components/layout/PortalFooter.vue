<template>
  <footer class="py-24 border-t border-[var(--border-color)] bg-[var(--bg-body)] relative overflow-hidden">
    <!-- Background Decor -->
    <div class="absolute top-0 left-1/4 w-96 h-96 bg-[var(--matrix-color)]/5 blur-[120px] rounded-full"></div>
    <div class="absolute bottom-0 right-1/4 w-96 h-96 bg-indigo-500/5 blur-[120px] rounded-full"></div>

    <div class="max-w-7xl mx-auto px-6 relative z-10">
      <div class="grid grid-cols-2 md:grid-cols-5 gap-16 mb-20">
        <div class="col-span-2">
          <div class="flex items-center gap-3 mb-8">
            <!-- Orbital Logo Effect -->
            <div class="flex items-center gap-1.5 h-10">
              <div v-for="i in 3" :key="i" 
                   class="w-1.5 h-6 bg-[var(--matrix-color)] rounded-full animate-bounce shadow-[0_0_15px_rgba(var(--matrix-color-rgb),0.5)]" 
                   :style="{ animationDelay: i * 0.15 + 's', animationDuration: '1s' }">
              </div>
            </div>
            <div class="flex flex-col">
              <span class="text-xl font-black tracking-tighter text-[var(--text-main)] leading-none uppercase">{{ tt('common.project_name') }}</span>
              <span class="text-[8px] uppercase tracking-[0.4em] font-bold text-[var(--text-muted)]">{{ tt('common.nexus_os') }}</span>
            </div>
          </div>
          <p class="text-[var(--text-muted)] max-w-sm leading-relaxed text-sm font-medium">
            {{ tt('footer.desc') }}
          </p>
        </div>
        <div>
          <h4 class="text-[12px] font-black uppercase tracking-[0.2em] mb-8 text-[var(--text-main)]">{{ tt('footer.zaomiao') }}</h4>
          <ul class="space-y-4 text-[var(--text-muted)] text-[12px] font-black uppercase tracking-widest">
            <li><router-link to="/meow" class="hover:text-[var(--matrix-color)] transition-colors">{{ tt('footer.zaomiao.home') }}</router-link></li>
            <li><router-link to="/meow/tech" class="hover:text-[var(--matrix-color)] transition-colors">{{ tt('footer.zaomiao.tech') }}</router-link></li>
            <li><router-link to="/meow/ecosystem" class="hover:text-[var(--matrix-color)] transition-colors">{{ tt('footer.zaomiao.ecosystem') }}</router-link></li>
            <li><router-link to="/meow/pricing" class="hover:text-[var(--matrix-color)] transition-colors">{{ tt('footer.zaomiao.pricing') }}</router-link></li>
          </ul>
        </div>
        <div>
          <h4 class="text-[12px] font-black uppercase tracking-[0.2em] mb-8 text-[var(--text-main)]">{{ tt('footer.digital_staff') }}</h4>
          <ul class="space-y-4 text-[var(--text-muted)] text-[12px] font-black uppercase tracking-widest">
            <li><router-link to="/bots/digital-employee#vision" class="hover:text-[var(--matrix-color)] transition-colors">{{ tt('footer.digital_staff.vision') }}</router-link></li>
            <li><router-link to="/bots/digital-employee#revolution" class="hover:text-[var(--matrix-color)] transition-colors">{{ tt('footer.digital_staff.revolution') }}</router-link></li>
            <li><router-link to="/bots/digital-employee#kpi" class="hover:text-[var(--matrix-color)] transition-colors">{{ tt('footer.digital_staff.kpi') }}</router-link></li>
            <li><router-link to="/bots/digital-employee#security" class="hover:text-[var(--matrix-color)] transition-colors">{{ tt('footer.digital_staff.security') }}</router-link></li>
          </ul>
        </div>
        <div>
          <h4 class="text-[12px] font-black uppercase tracking-[0.2em] mb-8 text-[var(--text-main)]">{{ tt('footer.more') }}</h4>
          <ul class="space-y-4 text-[var(--text-muted)] text-[12px] font-black uppercase tracking-widest">
            <li><router-link to="/docs" class="hover:text-[var(--matrix-color)] transition-colors">{{ tt('footer.docs') }}</router-link></li>
            <li><router-link to="/pricing" class="hover:text-[var(--matrix-color)] transition-colors">{{ tt('footer.plans') }}</router-link></li>
            <li><router-link to="/about" class="hover:text-[var(--matrix-color)] transition-colors">{{ tt('footer.about') }}</router-link></li>
            <li><a href="#" class="hover:text-[var(--matrix-color)] transition-colors">{{ tt('footer.privacy') }}</a></li>
          </ul>
        </div>
      </div>
      
      <div class="pt-12 border-t border-[var(--border-color)] flex flex-col md:flex-row justify-between items-center gap-8 text-[12px] font-black uppercase tracking-[0.2em] text-[var(--text-muted)]">
        <div class="flex items-center gap-6 flex-wrap justify-center md:justify-start">
          <a href="https://github.com/changliaotong/BotMatrix" target="_blank" class="hover:text-[var(--matrix-color)] transition-colors flex items-center gap-2">
            <Github class="w-4 h-4" />
            <span>GITHUB</span>
          </a>
          <span class="w-1 h-1 rounded-full bg-[var(--border-color)]"></span>
          <span>{{ tt('footer.copyright') }}</span>
          <a href="https://beian.miit.gov.cn/" target="_blank" class="hover:text-[var(--matrix-color)] transition-colors">{{ tt('footer.beian') }}</a>
        </div>
        <div class="flex items-center gap-6">
          <span class="flex items-center gap-2">
            {{ tt('footer.made_with').split('HEART')[0] }} <Heart class="w-3 h-3 text-rose-500 fill-rose-500" /> {{ tt('footer.made_with').split('HEART')[1] }}
          </span>
        </div>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { Github, Heart } from 'lucide-vue-next';
import { useI18n } from '@/utils/i18n';

const { t: tt } = useI18n();
</script>

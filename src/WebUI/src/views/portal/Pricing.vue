<template>
  <div class="min-h-screen bg-[var(--bg-body)] text-[var(--text-main)] selection:bg-[var(--matrix-color)]/30 relative overflow-x-hidden" :class="[systemStore.style]">
    <!-- Unified Background -->
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute top-0 left-1/2 -translate-x-1/2 w-full h-full bg-[radial-gradient(circle_at_center,var(--matrix-color,rgba(168,85,247,0.05))_0%,transparent_70%)] opacity-20"></div>
      <div class="absolute inset-0 bg-[grid-slate-800] [mask-image:radial-gradient(white,transparent)] opacity-20"></div>
    </div>
    
    <PortalHeader />

    <header class="pt-48 pb-24 px-6 text-center relative">
      <div class="inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-[var(--matrix-color)]/10 border border-[var(--matrix-color)]/20 text-[var(--matrix-color)] text-xs font-medium mb-10">
        <Sparkles class="w-4 h-4" /> {{ tt('portal.pricing.badge') }}
      </div>
      <h1 class="text-6xl md:text-7xl font-bold mb-8 tracking-tight text-[var(--text-main)]">
        {{ tt('portal.pricing.title_prefix') }} <br/>
        <span class="text-transparent bg-clip-text bg-gradient-to-r from-[var(--matrix-color)] to-[var(--matrix-glow)]">{{ tt('portal.pricing.title_suffix') }}</span>
      </h1>
      <p class="text-[var(--text-muted)] max-w-3xl mx-auto text-xl leading-relaxed font-light">
        {{ tt('portal.pricing.subtitle') }}
      </p>
    </header>

    <section class="py-24 px-6 relative z-10">
      <div class="max-w-7xl mx-auto grid grid-cols-1 lg:grid-cols-3 gap-8 items-stretch">
        <!-- Community Edition -->
        <div class="p-10 rounded-[2.5rem] bg-[var(--bg-body)]/50 border border-[var(--border-color)] flex flex-col h-full hover:border-[var(--matrix-color)]/30 transition-all duration-500 group">
          <div class="mb-10">
            <h3 class="text-2xl font-bold mb-2 text-[var(--text-main)] group-hover:text-[var(--matrix-color)] transition-colors uppercase tracking-tight">
              {{ tt('portal.pricing.community.name') }}
            </h3>
            <p class="text-[var(--text-muted)] text-sm font-medium uppercase tracking-wider">{{ tt('portal.pricing.community.desc') }}</p>
          </div>
          <div class="text-4xl font-bold mb-10 text-[var(--text-main)] tracking-tight">
            {{ tt('portal.pricing.community.price_free') }} <span class="text-sm font-medium text-[var(--text-muted)] uppercase tracking-widest ml-2">/ {{ tt('portal.pricing.community.price_tag') }}</span>
          </div>
          <ul class="space-y-5 mb-12 flex-1">
            <li v-for="item in communityFeatures" :key="item" class="flex items-start gap-4 text-[var(--text-muted)] group-hover:text-[var(--text-main)] transition-colors">
              <CheckCircle2 class="w-5 h-5 text-[var(--matrix-color)]/50 group-hover:text-[var(--matrix-color)] flex-shrink-0" />
              <span class="text-sm leading-snug">{{ item }}</span>
            </li>
          </ul>
          <button class="w-full py-4 bg-[var(--bg-body)] hover:bg-[var(--matrix-color)] text-white rounded-2xl font-bold transition-all border border-[var(--border-color)] hover:border-[var(--matrix-color)]">
            {{ tt('portal.pricing.community.btn') }}
          </button>
        </div>

        <!-- Business Edition -->
        <div class="p-10 rounded-[2.5rem] bg-[var(--bg-body)] border-2 border-[var(--matrix-color)]/50 flex flex-col h-full relative lg:scale-105 shadow-[0_0_50px_var(--matrix-glow)] group">
          <div class="absolute -top-4 left-1/2 -translate-x-1/2 px-4 py-1 bg-[var(--matrix-color)] rounded-full text-[10px] font-bold text-white uppercase tracking-widest whitespace-nowrap shadow-lg shadow-[var(--matrix-glow)]">{{ tt('portal.pricing.business.most_popular') }}</div>
          <div class="mb-10">
            <h3 class="text-2xl font-bold mb-2 text-[var(--matrix-color)] uppercase tracking-tight">
              {{ tt('portal.pricing.business.name') }}
            </h3>
            <p class="text-[var(--text-muted)] text-sm font-medium uppercase tracking-wider">{{ tt('portal.pricing.business.desc') }}</p>
          </div>
          <div class="text-4xl font-bold mb-10 text-[var(--text-main)] tracking-tight">
            {{ tt('portal.pricing.business.stay_tuned') }} <span class="text-sm font-medium text-[var(--text-muted)] uppercase tracking-widest ml-2">/ {{ tt('portal.pricing.business.price_tag') }}</span>
          </div>
          <ul class="space-y-5 mb-12 flex-1">
            <li v-for="item in proFeatures" :key="item" class="flex items-start gap-4 text-[var(--text-main)]">
              <Sparkles class="w-5 h-5 text-[var(--matrix-color)] flex-shrink-0" />
              <span class="text-sm leading-snug">{{ item }}</span>
            </li>
          </ul>
          <button @click="openAction('contact', 'business')" class="w-full py-4 bg-[var(--matrix-color)] hover:bg-[var(--matrix-color)]/80 text-white rounded-2xl font-bold transition-all shadow-lg shadow-[var(--matrix-glow)]">
            {{ tt('portal.pricing.business.btn') }}
          </button>
        </div>

        <!-- Matrix Edition -->
        <div class="p-10 rounded-[2.5rem] bg-[var(--bg-body)]/50 border border-[var(--border-color)] flex flex-col h-full hover:border-[var(--matrix-color)]/30 transition-all duration-500 group">
          <div class="mb-10">
            <h3 class="text-2xl font-bold mb-2 text-[var(--text-main)] group-hover:text-[var(--matrix-color)] transition-colors uppercase tracking-tight">
              {{ tt('portal.pricing.matrix.name') }}
            </h3>
            <p class="text-[var(--text-muted)] text-sm font-medium uppercase tracking-wider">{{ tt('portal.pricing.matrix.desc') }}</p>
          </div>
          <div class="text-4xl font-bold mb-10 text-[var(--text-main)] tracking-tight group-hover:text-[var(--matrix-color)] transition-colors">
            {{ tt('portal.pricing.matrix.price_custom') }} <span class="text-sm font-medium text-[var(--text-muted)] uppercase tracking-widest ml-2">/ {{ tt('portal.pricing.matrix.price_tag') }}</span>
          </div>
          <ul class="space-y-5 mb-12 flex-1">
            <li v-for="item in matrixFeatures" :key="item" class="flex items-start gap-4 text-[var(--text-muted)] group-hover:text-[var(--text-main)] transition-colors">
              <Zap class="w-5 h-5 text-[var(--border-color)] group-hover:text-[var(--matrix-color)] flex-shrink-0" />
              <span class="text-sm leading-snug">{{ item }}</span>
            </li>
          </ul>
          <button @click="openAction('contact', 'matrix')" class="w-full py-4 bg-[var(--bg-body)] hover:bg-[var(--matrix-color)] text-white rounded-2xl font-bold transition-all border border-[var(--border-color)] hover:border-[var(--matrix-color)]">
            {{ tt('portal.pricing.matrix.btn') }}
          </button>
        </div>
      </div>

      <!-- Comparison Table -->
      <div class="max-w-7xl mx-auto mt-48 px-6">
        <h2 class="text-4xl font-bold mb-16 text-center text-[var(--text-main)] tracking-tight">
          {{ tt('portal.pricing.compare.title') }} <span class="text-[var(--text-muted)] ml-4 font-light text-2xl">/ SPECIFICATIONS</span>
        </h2>
        <div class="overflow-x-auto rounded-[2.5rem] border border-[var(--border-color)] bg-[var(--bg-body)]/30 backdrop-blur-sm">
          <table class="w-full border-collapse">
            <thead>
              <tr class="border-b border-[var(--border-color)]">
                <th class="py-8 px-8 text-left text-xs font-bold uppercase tracking-widest text-[var(--text-muted)]">{{ tt('pricing.features') }}</th>
                <th class="py-8 px-8 text-center text-xs font-bold uppercase tracking-widest text-[var(--text-main)]">{{ tt('pricing.community_edition') }}</th>
                <th class="py-8 px-8 text-center text-xs font-bold uppercase tracking-widest text-[var(--matrix-color)]">{{ tt('pricing.business_edition') }}</th>
                <th class="py-8 px-8 text-center text-xs font-bold uppercase tracking-widest text-[var(--matrix-color)] opacity-80">{{ tt('pricing.matrix_edition') }}</th>
              </tr>
            </thead>
            <tbody class="text-sm">
              <tr v-for="row in comparisonRows" :key="row.feature" class="border-b border-[var(--border-color)]/50 hover:bg-[var(--matrix-color)]/5 transition-colors group">
                <td class="py-6 px-8 text-[var(--text-muted)] group-hover:text-[var(--text-main)] transition-colors font-medium">{{ row.feature }}</td>
                <td class="py-6 px-8 text-center">
                  <Check v-if="row.community" class="w-5 h-5 text-[var(--text-muted)] mx-auto" />
                  <span v-else class="text-[var(--text-muted)] opacity-20">-</span>
                </td>
                <td class="py-6 px-8 text-center">
                  <Check v-if="row.pro" class="w-5 h-5 text-[var(--matrix-color)] mx-auto" />
                  <span v-else class="text-[var(--text-muted)] opacity-20">-</span>
                </td>
                <td class="py-6 px-8 text-center">
                  <Check v-if="row.matrix" class="w-5 h-5 text-[var(--matrix-color)] mx-auto opacity-80" />
                  <span v-else class="text-[var(--text-muted)] opacity-20">-</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Pricing FAQ -->
      <div class="max-w-4xl mx-auto mt-48 px-6 pb-24">
        <h2 class="text-4xl font-bold mb-16 text-center text-[var(--text-main)] tracking-tight">
          {{ tt('portal.pricing.faq.title') }} <span class="text-[var(--matrix-color)]/50 ml-4 font-light text-2xl">/ FAQ</span>
        </h2>
        <div class="space-y-4">
          <div v-for="(item, index) in faqs" :key="index" 
            class="bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-2xl overflow-hidden transition-all hover:border-[var(--matrix-color)]/30">
            <button @click="activeFaq = activeFaq === index ? -1 : index" 
              class="w-full p-6 text-left flex justify-between items-center transition-colors">
              <span class="text-lg font-bold text-[var(--text-main)]">{{ item.q }}</span>
              <ChevronDown class="w-5 h-5 transition-transform duration-300 text-[var(--text-muted)]" :class="{ 'rotate-180': activeFaq === index }" />
            </button>
            <transition name="fade">
              <div v-show="activeFaq === index" class="px-6 pb-6 text-[var(--text-muted)] leading-relaxed font-light border-t border-[var(--border-color)]/50 pt-4">
                {{ item.a }}
              </div>
            </transition>
          </div>
        </div>
      </div>
    </section>

    <!-- Contact Modal -->
    <div v-if="activeModal" class="fixed inset-0 z-[100] flex items-center justify-center p-6">
      <div class="absolute inset-0 bg-[var(--bg-body)]/80 backdrop-blur-md" @click="activeModal = null"></div>
      <div class="relative w-full max-w-lg bg-[var(--bg-body)] border border-[var(--border-color)] rounded-[2.5rem] p-10 shadow-2xl overflow-hidden">
        <div class="absolute top-0 left-0 w-full h-1 bg-[var(--matrix-color)]"></div>
        <button @click="activeModal = null" class="absolute top-6 right-6 p-2 hover:bg-[var(--bg-body)]/80 rounded-full transition-all text-[var(--text-muted)] hover:text-[var(--text-main)]">
          <X class="w-5 h-5" />
        </button>
        
        <div class="text-center">
          <div class="w-16 h-16 rounded-2xl bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] flex items-center justify-center mx-auto mb-8">
            <Sparkles class="w-8 h-8" />
          </div>
          <h3 class="text-2xl font-bold mb-4 text-[var(--text-main)] uppercase tracking-tight">
            {{ modalType === 'matrix' ? tt('portal.pricing.modal.title_matrix') : tt('portal.pricing.modal.title_business') }}
          </h3>
          <p class="text-[var(--text-muted)] mb-8 leading-relaxed font-light">
            {{ tt('portal.pricing.modal.subtitle') }}
          </p>
          
          <div class="bg-[var(--bg-body)]/50 rounded-3xl p-8 mb-8 flex flex-col items-center gap-4 border border-[var(--border-color)]">
            <div class="w-40 h-40 bg-white rounded-2xl p-4 flex items-center justify-center shadow-inner">
              <div class="w-full h-full bg-slate-100 rounded-lg flex items-center justify-center text-slate-400 text-[10px] font-bold tracking-widest text-center px-4 uppercase">
                QR CODE<br/>{{ modalType === 'matrix' ? 'MATRIX' : 'BUSINESS' }}
              </div>
            </div>
            <div class="text-xs font-bold tracking-widest text-[var(--matrix-color)] uppercase">WeChat：BotMatrix_Global</div>
          </div>

          <button @click="activeModal = null" class="w-full py-4 bg-[var(--matrix-color)] hover:bg-[var(--matrix-color)]/80 text-white rounded-xl font-bold transition-all">
            {{ tt('portal.pricing.modal.btn') }}
          </button>
        </div>
      </div>
    </div>

    <PortalFooter />
  </div>
</template>

<script setup lang="ts">
import PortalHeader from '@/components/layout/PortalHeader.vue';
import PortalFooter from '@/components/layout/PortalFooter.vue';
import { 
  CheckCircle2, 
  Sparkles, 
  Check, 
  ChevronDown, 
  X, 
  Zap, 
  ArrowRight,
  Shield,
  Bot,
  MessageSquare,
  Globe,
  Lock,
  Cpu
} from 'lucide-vue-next';
import { ref, computed } from 'vue';
import { useSystemStore } from '@/stores/system';
import { useI18n } from '@/utils/i18n';

const { t: tt } = useI18n();
const systemStore = useSystemStore();
const activeFaq = ref(-1);
const activeModal = ref(false);
const modalType = ref('business');

const openAction = (type: string, subType: string = 'business') => {
  if (type === 'contact') {
    modalType.value = subType;
    activeModal.value = true;
  }
};

const communityFeatures = computed(() => [
  tt('portal.pricing.features.core'),
  tt('portal.pricing.features.onebot'),
  tt('portal.pricing.features.mcp'),
  tt('portal.pricing.features.rag'),
  tt('portal.pricing.features.ota')
]);

const proFeatures = computed(() => [
  tt('portal.pricing.features.digital_employee'),
  tt('portal.pricing.features.kpi'),
  tt('portal.pricing.features.bastion'),
  tt('portal.pricing.features.audit'),
  tt('portal.pricing.features.support')
]);

const matrixFeatures = computed(() => [
  tt('portal.pricing.features.mesh'),
  tt('portal.pricing.features.swarm'),
  tt('portal.pricing.features.vision'),
  tt('portal.pricing.features.hub'),
  tt('portal.pricing.features.custom')
]);

const comparisonRows = computed(() => [
  { feature: tt('portal.pricing.features.core'), community: true, pro: true, matrix: true },
  { feature: tt('portal.pricing.features.onebot'), community: true, pro: true, matrix: true },
  { feature: tt('portal.pricing.features.mcp'), community: true, pro: true, matrix: true },
  { feature: tt('portal.pricing.features.rag'), community: true, pro: true, matrix: true },
  { feature: tt('portal.pricing.features.digital_employee'), community: false, pro: true, matrix: true },
  { feature: tt('portal.pricing.features.kpi'), community: false, pro: true, matrix: true },
  { feature: tt('portal.pricing.features.bastion'), community: false, pro: true, matrix: true },
  { feature: tt('portal.pricing.features.mesh'), community: false, pro: false, matrix: true },
  { feature: tt('portal.pricing.features.swarm'), community: false, pro: false, matrix: true },
  { feature: tt('portal.pricing.features.vision'), community: false, pro: false, matrix: true },
  { feature: tt('portal.pricing.features.hub'), community: false, pro: false, matrix: true }
]);

const faqs = computed(() => [
  {
    q: tt('portal.pricing.faq.q1'),
    a: tt('portal.pricing.faq.a1')
  },
  {
    q: tt('portal.pricing.faq.q2'),
    a: tt('portal.pricing.faq.a2')
  },
  {
    q: tt('portal.pricing.faq.q3'),
    a: tt('portal.pricing.faq.a3')
  }
]);
</script>

<style scoped>
.fade-enter-active, .fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
</style>

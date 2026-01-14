<script setup lang="ts">
import { ref, computed } from 'vue';
import { 
  CheckCircle2, Star, Zap, Crown, HelpCircle, 
  ArrowRight, ShieldCheck, Cpu, X
} from 'lucide-vue-next';
import { useI18n } from '@/utils/i18n';

const { t: tt } = useI18n();

const plans = computed(() => [
  {
    name: tt('earlymeow.pricing.plans.basic.name'),
    price: '0',
    desc: tt('earlymeow.pricing.plans.basic.desc'),
    features: [
      tt('earlymeow.pricing.plans.basic.feat1'),
      tt('earlymeow.pricing.plans.basic.feat2'),
      tt('earlymeow.pricing.plans.basic.feat3'),
      tt('earlymeow.pricing.plans.basic.feat4')
    ],
    cta: tt('earlymeow.pricing.plans.basic.cta'),
    featured: false,
    color: 'slate'
  },
  {
    name: tt('earlymeow.pricing.plans.pro.name'),
    price: '99',
    period: tt('earlymeow.pricing.plans.period.month'),
    desc: tt('earlymeow.pricing.plans.pro.desc'),
    features: [
      tt('earlymeow.pricing.plans.pro.feat1'),
      tt('earlymeow.pricing.plans.pro.feat2'),
      tt('earlymeow.pricing.plans.pro.feat3'),
      tt('earlymeow.pricing.plans.pro.feat4'),
      tt('earlymeow.pricing.plans.pro.feat5')
    ],
    cta: tt('earlymeow.pricing.plans.pro.cta'),
    featured: true,
    color: 'purple'
  },
  {
    name: tt('earlymeow.pricing.plans.enterprise.name'),
    price: tt('earlymeow.pricing.plans.enterprise.price'),
    desc: tt('earlymeow.pricing.plans.enterprise.desc'),
    features: [
      tt('earlymeow.pricing.plans.enterprise.feat1'),
      tt('earlymeow.pricing.plans.enterprise.feat2'),
      tt('earlymeow.pricing.plans.enterprise.feat3'),
      tt('earlymeow.pricing.plans.enterprise.feat4'),
      tt('earlymeow.pricing.plans.enterprise.feat5')
    ],
    cta: tt('earlymeow.pricing.plans.enterprise.cta'),
    featured: false,
    color: 'slate'
  }
]);

const faqs = computed(() => [
  {
    q: tt('earlymeow.pricing.faq.q1'),
    a: tt('earlymeow.pricing.faq.a1')
  },
  {
    q: tt('earlymeow.pricing.faq.q2'),
    a: tt('earlymeow.pricing.faq.a2')
  },
  {
    q: tt('earlymeow.pricing.faq.q3'),
    a: tt('earlymeow.pricing.faq.a3')
  },
  {
    q: tt('earlymeow.pricing.faq.q4'),
    a: tt('earlymeow.pricing.faq.a4')
  }
]);

const trustFeatures = computed(() => [
  { icon: ShieldCheck, title: tt('earlymeow.pricing.trust.security'), desc: tt('earlymeow.pricing.trust.security_desc') },
  { icon: Cpu, title: tt('earlymeow.pricing.trust.uptime'), desc: tt('earlymeow.pricing.trust.uptime_desc') },
  { icon: Zap, title: tt('earlymeow.pricing.trust.speed'), desc: tt('earlymeow.pricing.trust.speed_desc') }
]);
</script>

<template>
  <div class="pt-40 pb-20 px-6 max-w-7xl mx-auto relative z-10">
    <!-- Header -->
    <div class="text-center mb-32 space-y-6">
      <div class="inline-flex items-center gap-2 text-[var(--matrix-color)] font-black text-xs uppercase tracking-widest px-3 py-1 rounded-full bg-[var(--matrix-color)]/10 border border-[var(--matrix-color)]/20 mx-auto">
        <Crown class="w-4 h-4" />
        {{ tt('earlymeow.pricing.header.tag') }}
      </div>
      <h1 class="text-6xl md:text-8xl font-black tracking-tighter leading-none text-[var(--text-main)]">
        {{ tt('earlymeow.pricing.header.title_prefix') }}<br/>
        <span class="bg-clip-text text-transparent bg-gradient-to-r from-[var(--matrix-color)] to-[rgba(var(--matrix-color-rgb),0.5)]">
          {{ tt('earlymeow.pricing.header.title_suffix') }}
        </span>
      </h1>
      <p class="text-xl text-[var(--text-muted)] font-medium max-w-2xl mx-auto leading-relaxed">
        {{ tt('earlymeow.pricing.header.desc') }}
      </p>
    </div>

    <!-- Pricing Grid -->
    <div class="grid lg:grid-cols-3 gap-8 mb-40">
      <div 
        v-for="plan in plans" 
        :key="plan.name"
        class="relative group p-10 rounded-[48px] border transition-all duration-500 flex flex-col backdrop-blur-xl"
        :class="[
          plan.featured 
            ? 'border-[var(--matrix-color)]/30 bg-[var(--matrix-color)]/10 -translate-y-4 shadow-2xl shadow-[var(--matrix-color)]/30' 
            : 'border-[var(--border-color)] bg-[var(--bg-body)]/40 hover:border-[var(--matrix-color)]/20'
        ]"
      >
        <div v-if="plan.featured" class="absolute -top-4 left-1/2 -translate-x-1/2 px-4 py-1 rounded-full bg-[var(--matrix-color)] text-white text-[10px] font-black uppercase tracking-widest">
          {{ tt('earlymeow.pricing.recommended') }}
        </div>

        <div class="mb-10">
          <h3 class="text-2xl font-black mb-2 text-[var(--text-main)]">{{ plan.name }}</h3>
          <p class="text-sm text-[var(--text-muted)] font-medium">{{ plan.desc }}</p>
        </div>

        <div class="mb-10">
          <div class="flex items-baseline gap-1">
            <span class="text-sm font-black text-[var(--text-muted)]">￥</span>
            <span class="text-6xl font-black text-[var(--text-main)]">{{ plan.price }}</span>
            <span v-if="plan.period" class="text-sm font-bold text-[var(--text-muted)]">{{ plan.period }}</span>
          </div>
        </div>

        <div class="flex-1 space-y-4 mb-12">
          <div v-for="feat in plan.features" :key="feat" class="flex gap-3 items-center">
            <CheckCircle2 class="w-5 h-5 flex-shrink-0" :class="plan.featured ? 'text-[var(--matrix-color)]' : 'text-[var(--text-muted)]'" />
            <span class="text-sm font-medium text-[var(--text-main)]/80">{{ feat }}</span>
          </div>
        </div>

        <button 
          class="w-full py-4 rounded-2xl font-black text-lg transition-all flex items-center justify-center gap-2 group"
          :class="[
            plan.featured 
              ? 'bg-[var(--matrix-color)] text-white hover:bg-[var(--matrix-color)]/80 hover:shadow-xl hover:shadow-[var(--matrix-color)]/30' 
              : 'bg-[var(--bg-body)]/80 text-[var(--text-main)] hover:bg-[var(--bg-body)] border border-[var(--border-color)]'
          ]"
        >
          {{ plan.cta }}
          <ArrowRight class="w-5 h-5 group-hover:translate-x-1 transition-transform" />
        </button>
      </div>
    </div>

    <!-- Trust Badges -->
    <div class="grid md:grid-cols-3 gap-8 mb-40">
      <div 
        v-for="feature in trustFeatures" 
        :key="feature.title"
        class="flex items-start gap-6 p-8 rounded-[32px] bg-[var(--bg-body)]/40 border border-[var(--border-color)]"
      >
        <div class="w-12 h-12 rounded-2xl bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] flex items-center justify-center shrink-0 border border-[var(--matrix-color)]/20">
          <component :is="feature.icon" class="w-6 h-6" />
        </div>
        <div>
          <h4 class="text-[var(--text-main)] font-bold mb-1">{{ feature.title }}</h4>
          <p class="text-sm text-[var(--text-muted)] leading-relaxed">{{ feature.desc }}</p>
        </div>
      </div>
    </div>

    <!-- Feature Comparison -->
    <div class="mb-40 overflow-hidden rounded-[40px] border border-[var(--border-color)] bg-[var(--bg-body)]/40 backdrop-blur-xl">
      <div class="p-10 border-b border-[var(--border-color)] bg-[var(--bg-body)]/50">
        <h2 class="text-3xl font-black text-[var(--text-main)]">{{ tt('earlymeow.pricing.compare.title') }}</h2>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-left">
          <thead>
            <tr class="border-b border-[var(--border-color)] text-[10px] font-black uppercase tracking-widest text-[var(--text-muted)]">
              <th class="px-10 py-6">{{ tt('earlymeow.pricing.compare.feature') }}</th>
              <th class="px-10 py-6 text-center">{{ tt('earlymeow.pricing.plans.basic.name') }}</th>
              <th class="px-10 py-6 text-center text-[var(--matrix-color)]">{{ tt('earlymeow.pricing.plans.pro.name') }}</th>
              <th class="px-10 py-6 text-center">{{ tt('earlymeow.pricing.plans.enterprise.name') }}</th>
            </tr>
          </thead>
          <tbody class="text-sm text-[var(--text-muted)] font-medium">
            <tr class="border-b border-[var(--border-color)] hover:bg-[var(--text-main)]/5 transition-colors">
              <td class="px-10 py-6 text-[var(--text-main)]">{{ tt('earlymeow.pricing.compare.nlp_engine') }}</td>
              <td class="px-10 py-6 text-center">{{ tt('earlymeow.pricing.compare.engine_std') }}</td>
              <td class="px-10 py-6 text-center">{{ tt('earlymeow.pricing.compare.engine_turbo') }}</td>
              <td class="px-10 py-6 text-center text-[var(--text-main)] font-bold">{{ tt('earlymeow.pricing.compare.engine_ultra') }}</td>
            </tr>
            <tr class="border-b border-[var(--border-color)] hover:bg-[var(--text-main)]/5 transition-colors">
              <td class="px-10 py-6 text-[var(--text-main)]">{{ tt('earlymeow.pricing.compare.memory_depth') }}</td>
              <td class="px-10 py-6 text-center">{{ tt('earlymeow.pricing.compare.memory_10') }}</td>
              <td class="px-10 py-6 text-center">{{ tt('earlymeow.pricing.compare.memory_50') }}</td>
              <td class="px-10 py-6 text-center text-[var(--text-main)] font-bold">{{ tt('earlymeow.pricing.compare.unlimited') }}</td>
            </tr>
            <tr class="border-b border-[var(--border-color)] hover:bg-[var(--text-main)]/5 transition-colors">
              <td class="px-10 py-6 text-[var(--text-main)]">{{ tt('earlymeow.pricing.compare.onebot_support') }}</td>
              <td class="px-10 py-6 text-center">{{ tt('earlymeow.pricing.compare.onebot_v11') }}</td>
              <td class="px-10 py-6 text-center">{{ tt('earlymeow.pricing.compare.onebot_v11_12') }}</td>
              <td class="px-10 py-6 text-center text-[var(--text-main)] font-bold">{{ tt('earlymeow.pricing.compare.full_protocol') }}</td>
            </tr>
            <tr class="border-b border-[var(--border-color)] hover:bg-[var(--text-main)]/5 transition-colors">
              <td class="px-10 py-6 text-[var(--text-main)]">{{ tt('earlymeow.pricing.compare.multi_agent') }}</td>
              <td class="px-10 py-6 text-center">
                <X class="w-4 h-4 mx-auto text-[var(--text-muted)]" />
              </td>
              <td class="px-10 py-6 text-center">
                <CheckCircle2 class="w-4 h-4 mx-auto text-[var(--matrix-color)]" />
              </td>
              <td class="px-10 py-6 text-center">
                <CheckCircle2 class="w-4 h-4 mx-auto text-[var(--text-main)]" />
              </td>
            </tr>
            <tr class="border-b border-[var(--border-color)] hover:bg-[var(--text-main)]/5 transition-colors">
              <td class="px-10 py-6 text-[var(--text-main)]">{{ tt('earlymeow.pricing.compare.api_access') }}</td>
              <td class="px-10 py-6 text-center">
                <X class="w-4 h-4 mx-auto text-[var(--text-muted)]" />
              </td>
              <td class="px-10 py-6 text-center">
                <CheckCircle2 class="w-4 h-4 mx-auto text-[var(--matrix-color)]" />
              </td>
              <td class="px-10 py-6 text-center">
                <CheckCircle2 class="w-4 h-4 mx-auto text-[var(--text-main)]" />
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- FAQ -->
    <div class="max-w-3xl mx-auto py-20 border-t border-[var(--border-color)]">
      <h2 class="text-4xl font-black tracking-tighter text-center mb-16 text-[var(--text-main)]">{{ tt('earlymeow.pricing.faq.title') }}</h2>
      <div class="space-y-12">
        <div v-for="faq in faqs" :key="faq.q" class="space-y-4 group">
          <h4 class="text-xl font-bold flex items-center gap-3 text-[var(--text-main)] group-hover:text-[var(--matrix-color)] transition-colors">
            <HelpCircle class="w-5 h-5 text-[var(--matrix-color)]" />
            {{ faq.q }}
          </h4>
          <p class="text-sm text-[var(--text-muted)] font-medium leading-relaxed pl-8">
            {{ faq.a }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Simplified styles */
</style>

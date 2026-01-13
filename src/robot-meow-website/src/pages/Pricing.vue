<template>
  <section class="pricing p-4">
    <h2>{{ $t('pricing.title') }}</h2>
    <div class="cards">
      <div class="card" v-for="p in plans" :key="p.name">
        <h3>{{ p.name }}</h3>
        <div class="price">{{ p.price }}</div>
        <ul>
          <li v-for="f in p.features" :key="f">{{ f }}</li>
        </ul>
        <router-link to="/contact" class="cta">{{ $t('pricing.contact') }}</router-link>
      </div>
    </div>
  </section>
</template>

<script setup>
import { ref } from 'vue'
const plans = ref([
  { name: 'Basic', price: '$19', features: ['Feature 1','Feature 2','Email support'] },
  { name: 'Pro', price: '$49', features: ['Feature 1','Feature 2','Priority support'] },
  { name: 'Enterprise', price: '$99', features: ['All features','Dedicated manager','SLA'] }
])
</script>

<style scoped>
.pricing{ padding:20px; }
.cards{ display:flex; gap:20px; flex-wrap:wrap; }
.card{ border:1px solid #eee; padding:16px; border-radius:8px; min-width:260px; }
.price{ font-size:28px; font-weight:700; margin:6px 0; }
.cta{ display:inline-block; margin-top:8px; padding:10px 14px; background:#2C8F7C; color:white; border-radius:6px; text-decoration:none; }
</style>

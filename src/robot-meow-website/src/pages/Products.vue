<template>
  <section class="products p-4">
    <h2>{{ $t('products.title') }}</h2>
    <div class="grid">
      <div class="card" v-for="p in products" :key="p.slug">
        <div class="thumb" />
        <h3>{{ p.title }}</h3>
        <p>{{ p.description }}</p>
        <router-link :to="`/products/${p.slug}`">{{ $t('products.view') }}</router-link>
      </div>
    </div>
  </section>
</template>

<script setup>
import { ref } from 'vue'
const products = ref([
  { slug: 'alpha', title: 'Alpha Robot', description: '智能协作助手' },
  { slug: 'beta', title: 'Beta Robot', description: '协同办公伙伴' }
])
</script>

<style scoped>
.products{ padding: 20px; }
.grid{ display:grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap:20px; }
.card{ border:1px solid #eee; padding:16px; border-radius:8px; text-align:left; }
.thumb{ height:120px; background:#f5f5f5; border-radius:6px; margin-bottom:8px; }
</style>

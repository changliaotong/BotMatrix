<template>
  <section class="product-detail p-4">
    <h2>{{ productTitle }}</h2>
    <div class="media">
      <div class="image" />
      <div class="desc">
        <p>{{ description }}</p>
        <div class="features">
          <span class="tag" v-for="f in features" :key="f">{{ f }}</span>
        </div>
      </div>
    </div>
    <router-link to="/contact" class="cta">{{ $t('product.contact') || '联系经销商' }}</router-link>
  </section>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
const route = useRoute()
// simple dummy data based on slug
const slug = route.params.slug
const dataMap = {
  alpha: { productTitle: 'Alpha Robot', description: '智能协作助手，提升工作效率', features: ['特征A','特征B','高可靠性'] },
  beta: { productTitle: 'Beta Robot', description: '协同办公伙伴，提升团队协同', features: ['特征X','特征Y','低成本'] }
}
const item = dataMap[slug] || dataMap.alpha
const productTitle = item.productTitle
const description = item.description
const features = item.features
</script>

<style scoped>
.product-detail{ padding:20px; }
.media{ display:flex; gap:20px; align-items:flex-start; }
.image{ width:60%; height:300px; background:#eee; border-radius:6px; }
.desc{ flex:1; }
.tag{ display:inline-block; padding:6px 10px; border:1px solid #ddd; border-radius:999px; margin:4px; font-size:12px; }
.cta{ display:inline-block; margin-top:12px; padding:12px 20px; background:#2C8F7C; color:white; border-radius:6px; text-decoration:none; }
</style>

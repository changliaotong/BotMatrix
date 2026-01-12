<template>
  <nav class="navbar">
    <div class="brand">机器人早喵</div>
    <ul class="nav-links" aria-label="Main">
      <li><router-link to="/">{{ $t('nav.home') }}</router-link></li>
      <li><router-link to="/products">{{ $t('nav.products') }}</router-link></li>
      <li><router-link to="/pricing">{{ $t('nav.pricing') }}</router-link></li>
      <li><router-link to="/about">{{ $t('nav.about') }}</router-link></li>
      <li><router-link to="/contact">{{ $t('nav.contact') }}</router-link></li>
      <li><button class="lang-toggle" @click="toggleLang">{{ currentLang === 'zh' ? 'EN' : '中文' }}</button></li>
    </ul>
  </nav>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { locale } = useI18n()
const currentLang = computed(() => locale.value)
function toggleLang(){
  locale.value = (locale.value === 'zh') ? 'en' : 'zh'
}
</script>

<style scoped>
.navbar{ display:flex; align-items:center; justify-content:space-between; padding:16px 24px; background:#fff; border-bottom:1px solid #eee; }
.brand{ font-weight:700; font-size:1.25rem; }
.nav-links{ list-style:none; display:flex; gap:16px; padding:0; margin:0; align-items:center; }
.nav-links a{ text-decoration:none; color:#333; }
.lang-toggle{ background:transparent; border:1px solid #ccc; padding:6px 10px; border-radius:6px; cursor:pointer; }
</style>

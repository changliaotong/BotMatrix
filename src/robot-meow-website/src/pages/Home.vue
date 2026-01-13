<template>
  <section class="home p-4">
    <div class="hero">
      <h1>{{ $t('home.title') }}</h1>
      <p class="subtitle">{{ $t('home.subtitle') }}</p>
      <router-link to="/products" class="cta">{{ $t('home.cta') }}</router-link>
    </div>
  </section>
</template>

<script setup>
</script>

<style scoped>
.home{ padding:40px 20px; }
.hero{ max-width: 1100px; margin:0 auto; text-align:left; }
.hero h1{ font-size: 2.5rem; margin-bottom: 0.5rem; }
.subtitle{ color:#555; margin-bottom:20px; }
.cta{ display:inline-block; padding:12px 20px; background:#2C8F7C; color:white; border-radius:6px; text-decoration:none; }
</style>

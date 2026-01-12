<template>
  <div id="app" class="app">
    <NavBar />
    <router-view />
  </div>
</template>

<script setup>
import NavBar from './components/NavBar.vue'
</script>

<style>
/* Global styles can be extended in src/style.css */
</style>

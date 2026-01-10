<template>
  <nav class="fixed top-0 w-full z-50 bg-slate-900/80 backdrop-blur-md border-b border-slate-800">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex justify-between h-16 items-center">
        <router-link to="/" class="flex items-center gap-2 group cursor-pointer">
          <div class="w-8 h-8 bg-pink-500 rounded-lg flex items-center justify-center font-bold text-white group-hover:rotate-12 transition-transform">
            <Cat class="w-5 h-5" />
          </div>
          <span class="text-xl font-bold tracking-tight text-white">早喵机器人</span>
        </router-link>
        
        <!-- Desktop Nav -->
        <div class="hidden md:flex items-center gap-8 text-sm font-medium text-slate-400">
          <router-link to="/" class="hover:text-pink-400 transition-colors" :class="{ 'text-pink-400': route.path === '/' }">首页</router-link>
          
          <!-- Bots Dropdown -->
          <div class="relative group">
            <button class="flex items-center gap-1 hover:text-pink-400 transition-colors py-4" :class="{ 'text-pink-400': route.path.startsWith('/bots') }">
              产品矩阵
              <ChevronDown class="w-4 h-4" />
            </button>
            <div class="absolute top-full left-0 w-48 py-2 bg-slate-800 border border-slate-700 rounded-xl shadow-xl opacity-0 translate-y-2 pointer-events-none group-hover:opacity-100 group-hover:translate-y-0 group-hover:pointer-events-auto transition-all">
              <router-link to="/botmatrix" class="block px-4 py-2 hover:bg-slate-700 hover:text-cyan-400 font-bold text-cyan-500">BotMatrix 核心</router-link>
              <router-link to="/bots/nexus-guard" class="block px-4 py-2 hover:bg-slate-700 hover:text-cyan-400">Nexus Guard</router-link>
              <router-link to="/bots/digital-employee" class="block px-4 py-2 hover:bg-slate-700 hover:text-purple-400 font-medium border-t border-slate-700/50 mt-1">数字员工 (AI Worker)</router-link>
              <div class="px-4 py-2 text-xs text-slate-500 border-t border-slate-700 mt-1">更多机器人敬请期待...</div>
            </div>
          </div>

          <router-link to="/docs" class="hover:text-pink-400 transition-colors" :class="{ 'text-pink-400': route.path === '/docs' }">文档中心</router-link>
          <router-link to="/news" class="hover:text-pink-400 transition-colors" :class="{ 'text-pink-400': route.path === '/news' }">动态更新</router-link>
          <router-link to="/pricing" class="hover:text-pink-400 transition-colors" :class="{ 'text-pink-400': route.path === '/pricing' }">版本计划</router-link>
          <router-link to="/about" class="hover:text-pink-400 transition-colors" :class="{ 'text-pink-400': route.path === '/about' }">关于我们</router-link>
          
          <router-link :to="authStore.isAuthenticated ? '/console' : '/login'" class="px-5 py-2 bg-pink-500 hover:bg-pink-400 text-white rounded-full transition-all font-bold">
            {{ authStore.isAuthenticated ? '进入控制台' : '开始使用' }}
          </router-link>
        </div>

        <!-- Mobile Menu Toggle -->
        <div class="md:hidden">
          <button @click="isMobileMenuOpen = !isMobileMenuOpen" class="p-2 text-slate-400">
            <Menu v-if="!isMobileMenuOpen" class="w-6 h-6" />
            <X v-else class="w-6 h-6" />
          </button>
        </div>
      </div>
    </div>

    <!-- Mobile Menu -->
    <transition name="fade">
      <div v-if="isMobileMenuOpen" class="md:hidden bg-slate-900 border-b border-slate-800 py-4 px-4 space-y-4">
        <router-link to="/" class="block text-slate-400 hover:text-white" @click="isMobileMenuOpen = false">首页</router-link>
        <div class="space-y-2">
          <div class="text-xs font-bold text-slate-500 uppercase tracking-widest px-2">官方机器人</div>
          <router-link to="/bots/early-meow" class="block pl-4 text-slate-400 hover:text-white" @click="isMobileMenuOpen = false">早喵机器人</router-link>
          <router-link to="/bots/nexus-guard" class="block pl-4 text-slate-400 hover:text-white" @click="isMobileMenuOpen = false">Nexus Guard</router-link>
          <router-link to="/bots/digital-employee" class="block pl-4 text-purple-400 font-medium" @click="isMobileMenuOpen = false">数字员工 (AI Worker)</router-link>
        </div>
        <router-link to="/docs" class="block text-slate-400 hover:text-white" @click="isMobileMenuOpen = false">文档中心</router-link>
        <router-link to="/news" class="block text-slate-400 hover:text-white" @click="isMobileMenuOpen = false">动态更新</router-link>
        <router-link to="/pricing" class="block text-slate-400 hover:text-white" @click="isMobileMenuOpen = false">版本计划</router-link>
        <router-link to="/about" class="block text-slate-400 hover:text-white" @click="isMobileMenuOpen = false">关于我们</router-link>
        <router-link :to="authStore.isAuthenticated ? '/console' : '/login'" class="block w-full py-3 bg-cyan-500 text-slate-900 text-center rounded-xl font-bold" @click="isMobileMenuOpen = false">
          {{ authStore.isAuthenticated ? '进入控制台' : '开始使用' }}
        </router-link>
      </div>
    </transition>
  </nav>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRoute } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { ChevronDown, Menu, X } from 'lucide-vue-next';

const route = useRoute();
const authStore = useAuthStore();
const isMobileMenuOpen = ref(false);
</script>

<style scoped>
.fade-enter-active, .fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
</style>

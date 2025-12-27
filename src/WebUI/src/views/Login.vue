<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { Bot, Lock, User, ArrowRight, Loader2 } from 'lucide-vue-next';

const router = useRouter();
const authStore = useAuthStore();

const username = ref('');
const password = ref('');
const loading = ref(false);
const error = ref('');

const handleLogin = async () => {
  if (!username.value || !password.value) return;
  
  loading.value = true;
  error.value = '';
  
  try {
    const success = await authStore.login(username.value, password.value);
    if (success) {
      router.push('/');
    } else {
      error.value = '用户名或密码错误';
    }
  } catch (err: any) {
    error.value = err.message || '登录失败，请稍后再试';
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-black flex items-center justify-center p-4 relative overflow-hidden">
    <!-- Matrix Rain Background Placeholder -->
    <div class="absolute inset-0 opacity-10 pointer-events-none">
      <div class="absolute inset-0 bg-gradient-to-b from-transparent via-matrix/20 to-transparent animate-pulse"></div>
    </div>

    <div class="w-full max-w-md relative">
      <div class="p-8 sm:p-12 rounded-[2.5rem] bg-white dark:bg-zinc-900 border border-black/5 dark:border-white/5 shadow-2xl space-y-8">
        <!-- Logo -->
        <div class="text-center space-y-4">
          <div class="inline-flex p-5 rounded-3xl bg-matrix/10 text-matrix animate-bounce-slow">
            <Bot class="w-10 h-10" />
          </div>
          <div class="space-y-1">
            <h1 class="text-3xl font-black dark:text-white tracking-tighter uppercase italic">Bot<span class="text-matrix">Matrix</span></h1>
            <p class="text-xs font-bold text-gray-400 uppercase tracking-[0.2em]">智能机器人管理系统</p>
          </div>
        </div>

        <!-- Form -->
        <form @submit.prevent="handleLogin" class="space-y-6">
          <div class="space-y-4">
            <div class="relative group">
              <div class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-matrix transition-colors">
                <User class="w-5 h-5" />
              </div>
              <input 
                v-model="username"
                type="text" 
                placeholder="用户名" 
                class="w-full bg-gray-50 dark:bg-black border border-black/5 dark:border-white/10 rounded-2xl pl-12 pr-4 py-4 focus:outline-none focus:border-matrix transition-all dark:text-white font-bold placeholder:text-gray-400"
              />
            </div>
            <div class="relative group">
              <div class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-matrix transition-colors">
                <Lock class="w-5 h-5" />
              </div>
              <input 
                v-model="password"
                type="password" 
                placeholder="密码" 
                class="w-full bg-gray-50 dark:bg-black border border-black/5 dark:border-white/10 rounded-2xl pl-12 pr-4 py-4 focus:outline-none focus:border-matrix transition-all dark:text-white font-bold placeholder:text-gray-400"
              />
            </div>
          </div>

          <div v-if="error" class="p-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-500 text-xs font-bold text-center">
            {{ error }}
          </div>

          <button 
            type="submit" 
            :disabled="loading"
            class="w-full bg-matrix hover:bg-matrix/90 disabled:opacity-50 text-black font-black py-4 rounded-2xl flex items-center justify-center gap-2 transition-all group active:scale-95 shadow-lg shadow-matrix/20 uppercase tracking-widest"
          >
            <template v-if="loading">
              <Loader2 class="w-5 h-5 animate-spin" /> 验证中...
            </template>
            <template v-else>
              进入矩阵 <ArrowRight class="w-5 h-5 group-hover:translate-x-1 transition-transform" />
            </template>
          </button>
        </form>

        <div class="text-center">
          <p class="text-[10px] font-bold text-gray-500 uppercase tracking-widest">
            &copy; 2025 BotMatrix. Industry Best Practice.
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.text-matrix {
  color: var(--matrix-color);
}
.bg-matrix {
  background-color: var(--matrix-color);
}
.bg-matrix\/10 {
  background-color: rgba(0, 255, 65, 0.1);
}
.animate-bounce-slow {
  animation: bounce 3s infinite;
}
@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}
</style>

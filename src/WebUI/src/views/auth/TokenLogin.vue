<template>
  <div class="min-h-screen flex items-center justify-center bg-slate-900 text-white p-4">
    <div class="max-w-md w-full bg-slate-800 rounded-2xl p-8 border border-slate-700 shadow-2xl text-center">
      <div v-if="loading">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-cyan-500 mx-auto mb-4"></div>
        <h2 class="text-xl font-bold mb-2">正在验证登录...</h2>
        <p class="text-slate-400">请稍候，我们正在通过安全令牌为您登录。</p>
      </div>
      
      <div v-else-if="error">
        <div class="w-16 h-16 bg-red-500/10 text-red-500 rounded-full flex items-center justify-center mx-auto mb-4">
          <i class="pi pi-exclamation-triangle text-2xl"></i>
        </div>
        <h2 class="text-xl font-bold mb-2">登录失败</h2>
        <p class="text-red-400 mb-6">{{ error }}</p>
        <router-link to="/login" class="inline-block px-6 py-2 bg-slate-700 hover:bg-slate-600 rounded-lg transition-colors">
          返回普通登录
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();

const loading = ref(true);
const error = ref('');

onMounted(async () => {
  const token = route.query.token as string;
  const platform = route.query.platform as string;
  const platformID = route.query.platform_id as string;

  if (!token || !platform || !platformID) {
    error.value = '无效的登录链接或 Token 已过期。';
    loading.value = false;
    return;
  }

  try {
    const success = await authStore.loginWithToken(platform, platformID, token);
    if (success) {
      // 登录成功，跳转到控制台首页
      router.push({ name: 'console-dashboard' });
    } else {
      error.value = 'Token 验证失败，可能已过期或已被使用。';
    }
  } catch (err: any) {
    error.value = err.response?.data?.error || '登录过程中发生错误，请稍后重试。';
  } finally {
    loading.value = false;
  }
});
</script>

<style scoped>
@import "primeicons/primeicons.css";
</style>

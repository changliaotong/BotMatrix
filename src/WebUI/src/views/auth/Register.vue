<template>
  <div class="min-h-screen flex items-center justify-center bg-slate-900 text-white p-4">
    <div class="max-w-md w-full bg-slate-800 rounded-2xl p-8 border border-slate-700 shadow-2xl">
      <div class="text-center mb-8">
        <h2 class="text-3xl font-bold bg-gradient-to-r from-cyan-400 to-blue-500 bg-clip-text text-transparent">用户注册</h2>
        <p class="text-slate-400 mt-2">创建您的 BotMatrix 账号</p>
      </div>

      <form @submit.prevent="handleRegister" class="space-y-6">
        <div>
          <label class="block text-sm font-medium text-slate-300 mb-2">用户名</label>
          <input 
            v-model="form.username" 
            type="text" 
            required
            class="w-full bg-slate-900 border border-slate-700 rounded-xl px-4 py-3 focus:ring-2 focus:ring-cyan-500 focus:border-transparent transition-all outline-none"
            placeholder="请输入用户名"
          >
        </div>
        <div>
          <label class="block text-sm font-medium text-slate-300 mb-2">密码</label>
          <input 
            v-model="form.password" 
            type="password" 
            required
            class="w-full bg-slate-900 border border-slate-700 rounded-xl px-4 py-3 focus:ring-2 focus:ring-cyan-500 focus:border-transparent transition-all outline-none"
            placeholder="请输入密码"
          >
        </div>
        <div>
          <label class="block text-sm font-medium text-slate-300 mb-2">确认密码</label>
          <input 
            v-model="form.confirmPassword" 
            type="password" 
            required
            class="w-full bg-slate-900 border border-slate-700 rounded-xl px-4 py-3 focus:ring-2 focus:ring-cyan-500 focus:border-transparent transition-all outline-none"
            placeholder="请再次输入密码"
          >
        </div>
        
        <div class="pt-2">
          <button 
            type="submit" 
            :disabled="loading"
            class="w-full bg-cyan-500 hover:bg-cyan-400 text-slate-900 font-bold py-3 rounded-xl transition-all shadow-lg shadow-cyan-500/20 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {{ loading ? '注册中...' : '立即注册' }}
          </button>
        </div>
      </form>

      <div class="mt-8 text-center text-slate-400 text-sm">
        已有账号？ 
        <router-link to="/login" class="text-cyan-400 hover:text-cyan-300 font-medium transition-colors">立即登录</router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';

const router = useRouter();
const authStore = useAuthStore();

const loading = ref(false);
const form = reactive({
  username: '',
  password: '',
  confirmPassword: ''
});

const handleRegister = async () => {
  if (form.password !== form.confirmPassword) {
    alert('两次输入的密码不一致');
    return;
  }

  loading.value = true;
  try {
    const success = await authStore.register(form.username, form.password);
    if (success) {
      alert('注册成功，请登录');
      router.push({ name: 'login' });
    }
  } catch (err: any) {
    alert(err.response?.data?.error || '注册失败，请重试');
  } finally {
    loading.value = false;
  }
};
</script>

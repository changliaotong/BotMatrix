<template>
  <div class="min-h-screen flex items-center justify-center bg-slate-900 text-white p-4">
    <div class="max-w-md w-full bg-slate-800 rounded-2xl p-8 border border-slate-700 shadow-2xl text-center">
      <div v-if="loading">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-cyan-500 mx-auto mb-4"></div>
        <h2 class="text-xl font-bold mb-2">{{ tt('verifying_login') }}</h2>
        <p class="text-slate-400">{{ tt('verifying_login_desc') }}</p>
      </div>
      
      <div v-else-if="error">
        <div class="w-16 h-16 bg-red-500/10 text-red-500 rounded-full flex items-center justify-center mx-auto mb-4">
          <i class="pi pi-exclamation-triangle text-2xl"></i>
        </div>
        <h2 class="text-xl font-bold mb-2">{{ tt('login_failed') }}</h2>
        <p class="text-red-400 mb-6">{{ error }}</p>
        <router-link to="/login" class="inline-block px-6 py-2 bg-slate-700 hover:bg-slate-600 rounded-lg transition-colors">
          {{ tt('back_to_login') }}
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { useI18n } from '@/utils/i18n';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const { tt } = useI18n();

const loading = ref(true);
const error = ref('');

onMounted(async () => {
  const token = route.query.token as string;
  const platform = route.query.platform as string;
  const platformID = route.query.platform_id as string;

  if (!token || !platform || !platformID) {
    error.value = tt('invalid_token');
    loading.value = false;
    return;
  }

  try {
    const success = await authStore.loginWithToken(platform, platformID, token);
    if (success) {
      // Get redirect path from query or determine default based on role
      const redirect = route.query.redirect as string;
      if (redirect) {
        router.push(redirect);
      } else {
        // Default landing pages
        if (authStore.isAdmin) {
          router.push('/console'); // Control Center (Dashboard)
        } else {
          router.push('/setup/bot'); // Bot Settings
        }
      }
    } else {
      error.value = tt('token_verify_failed');
    }
  } catch (err: any) {
    error.value = err.response?.data?.error || tt('login_error');
  } finally {
    loading.value = false;
  }
});
</script>

<style scoped>
@import "primeicons/primeicons.css";
</style>

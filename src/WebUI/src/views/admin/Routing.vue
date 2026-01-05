<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useSystemStore } from '@/stores/system';
import { useBotStore } from '@/stores/bot';
import { useAuthStore } from '@/stores/auth';
import { 
  Route, 
  Plus, 
  Trash2, 
  ArrowRight, 
  Cpu, 
  Search,
  AlertCircle,
  CheckCircle2,
  X,
  Shield
} from 'lucide-vue-next';

const systemStore = useSystemStore();
const botStore = useBotStore();
const authStore = useAuthStore();
const t = (key: string) => systemStore.t(key);

const rules = ref<any[]>([]);
const workers = ref<any[]>([]);
const loading = ref(false);
const showAddModal = ref(false);
const searchQuery = ref('');
const isEditing = ref(false);
const error = ref<string | null>(null);

const newRule = ref({
  key: '',
  worker_id: ''
});

const filteredRules = computed(() => {
  if (!searchQuery.value) return rules.value;
  const query = searchQuery.value.toLowerCase();
  return rules.value.filter(rule => 
    rule.key.toLowerCase().includes(query) || 
    rule.worker_id.toLowerCase().includes(query)
  );
});

const fetchRules = async () => {
  if (!authStore.isAdmin) {
    loading.value = false;
    return;
  }
  loading.value = true;
  error.value = null;
  try {
    const [rulesData, workersData] = await Promise.all([
      botStore.fetchRoutingRules(),
      botStore.fetchWorkers()
    ]);
    
    if (rulesData && rulesData.success && rulesData.data) {
      // Convert map to array if necessary
      const rulesObj = rulesData.data.rules;
      if (rulesObj && typeof rulesObj === 'object' && !Array.isArray(rulesObj)) {
        rules.value = Object.entries(rulesObj).map(([key, worker_id]) => ({
          key,
          worker_id
        }));
      } else {
        rules.value = rulesObj || [];
      }
    } else {
      error.value = rulesData?.message || 'Failed to fetch rules';
    }

    if (workersData && workersData.success && workersData.data) {
      workers.value = workersData.data.workers || [];
    }
  } catch (err: any) {
    console.error('Failed to fetch routing data:', err);
    error.value = err.message || 'Network error';
  } finally {
    loading.value = false;
  }
};

const handleAddRule = async () => {
  if (!newRule.value.key || !newRule.value.worker_id) return;
  
  try {
    const data = await botStore.setRoutingRule(newRule.value);
    if (data.success) {
      await fetchRules();
      showAddModal.value = false;
      resetForm();
    }
  } catch (err) {
    console.error('Failed to add/update rule:', err);
  }
};

const handleEditRule = (rule: any) => {
  newRule.value = { ...rule };
  isEditing.value = true;
  showAddModal.value = true;
};

const resetForm = () => {
  newRule.value = { key: '', worker_id: '' };
  isEditing.value = false;
};

const handleDeleteRule = async (key: string) => {
  if (!confirm(t('confirm_delete_routing_rule').replace('{key}', key))) return;
  
  try {
    const data = await botStore.deleteRoutingRule(key);
    if (data.success) {
      await fetchRules();
    }
  } catch (err) {
    console.error('Failed to delete rule:', err);
  }
};

const getWorkerName = (workerId: string) => {
  const worker = workers.value.find(w => w.id === workerId);
  if (worker && worker.name) return worker.name;
  return t(workerId);
};

onMounted(fetchRules);
</script>

<template>
  <div class="p-6 space-y-6">
    <div v-if="!authStore.isAdmin" class="min-h-[60vh] flex flex-col items-center justify-center space-y-4 opacity-30">
      <Shield class="w-16 h-16 text-[var(--text-main)]" />
      <p class="text-xs font-black uppercase tracking-[0.2em] text-[var(--text-main)]">{{ t('admin_required') }}</p>
    </div>
    <template v-else>
      <!-- Header -->
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-black text-[var(--text-main)] tracking-tight">{{ t('routing') }}</h1>
        <p class="text-sm font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('routing_desc') }}</p>
      </div>
      <div class="flex items-center gap-3">
        <div class="relative flex-1 md:w-64">
          <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
          <input 
            v-model="searchQuery"
            type="text" 
            :placeholder="t('search_rules_placeholder')"
            class="w-full pl-10 pr-4 py-2 rounded-2xl bg-[var(--bg-card)] border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold transition-all"
          />
        </div>
        <button 
          @click="resetForm(); showAddModal = true"
          class="flex items-center gap-2 px-4 py-2 rounded-2xl bg-[var(--matrix-color)] text-black font-black text-xs uppercase tracking-widest hover:opacity-90 transition-opacity whitespace-nowrap"
        >
          <Plus class="w-4 h-4" />
          {{ t('add_rule') }}
        </button>
      </div>
    </div>

    <!-- Error Alert -->
    <div v-if="error" class="p-4 rounded-3xl bg-red-500/10 border border-red-500/20 flex items-center gap-4 text-red-500">
      <div class="p-3 rounded-2xl bg-red-500/10">
        <AlertCircle class="w-6 h-6" />
      </div>
      <div class="flex-1">
        <h4 class="font-black text-xs uppercase tracking-widest">{{ t('error') }}</h4>
        <p class="text-sm font-medium">{{ error }}</p>
      </div>
      <button @click="fetchRules" class="p-3 rounded-2xl hover:bg-red-500/10 transition-colors">
        <Plus class="w-5 h-5 rotate-45" />
      </button>
    </div>

    <!-- Rules List -->
    <div v-if="loading" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div v-for="i in 6" :key="i" class="h-48 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] animate-pulse"></div>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div 
        v-for="rule in filteredRules" 
        :key="rule.key"
        class="group p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] hover:border-[var(--matrix-color)]/30 transition-all duration-500 relative overflow-hidden flex flex-col justify-between h-full"
      >
        <div>
          <div class="flex items-start justify-between mb-6">
            <div class="p-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
              <Route class="w-6 h-6 text-[var(--matrix-color)]" />
            </div>
            <div class="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
              <button 
                @click="handleEditRule(rule)"
                class="p-2 rounded-xl bg-[var(--matrix-color)]/10 border border-[var(--matrix-color)]/20 text-[var(--matrix-color)] hover:bg-[var(--matrix-color)] hover:text-[var(--sidebar-text-active)] transition-all"
              >
                <Plus class="w-4 h-4 rotate-45 group-hover:rotate-0 transition-transform" />
              </button>
              <button 
                @click="handleDeleteRule(rule.key)"
                class="p-2 rounded-xl bg-red-500/10 border border-red-500/20 text-red-500 hover:bg-red-500 hover:text-[var(--sidebar-text)] transition-all"
              >
                <Trash2 class="w-4 h-4" />
              </button>
            </div>
          </div>

          <div class="space-y-4">
            <div>
              <span class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('routing_key') }}</span>
              <div class="flex items-center gap-2 mt-1">
                <h3 class="font-black text-[var(--text-main)] break-all">{{ rule.key }}</h3>
                <div v-if="rule.key.includes('*')" class="px-2 py-0.5 rounded text-[8px] font-black uppercase tracking-tighter bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] border border-[var(--matrix-color)]/20">
                  {{ t('pattern') }}
                </div>
              </div>
            </div>
            
            <div class="flex items-center gap-3">
              <div class="flex-1 h-px bg-[var(--border-color)]"></div>
              <ArrowRight class="w-4 h-4 text-[var(--matrix-color)]" />
              <div class="flex-1 h-px bg-[var(--border-color)]"></div>
            </div>

            <div>
              <span class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('target_worker') }}</span>
              <div class="flex items-center gap-2 mt-1">
                <Cpu class="w-4 h-4 text-[var(--matrix-color)]" />
                <span class="font-bold text-[var(--text-main)] truncate">{{ getWorkerName(rule.worker_id) }}</span>
              </div>
            </div>
          </div>
        </div>

        <div class="mt-6 pt-6 border-t border-[var(--border-color)] flex items-center justify-between">
          <div class="flex items-center gap-2">
            <div :class="[
              'w-1.5 h-1.5 rounded-full',
              workers.find(w => w.id === rule.worker_id)?.status === 'online' ? 'bg-green-500' : 'bg-red-500 animate-pulse'
            ]"></div>
            <span class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">
              {{ t(workers.find(w => w.id === rule.worker_id)?.status || 'unknown') }}
            </span>
          </div>
          <span class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">
            {{ t('type') }}: {{ t(workers.find(w => w.id === rule.worker_id)?.type || 'unknown') }}
          </span>
        </div>
      </div>

      <!-- Empty State -->
      <div v-if="filteredRules.length === 0" class="col-span-full flex flex-col items-center justify-center py-20 bg-[var(--bg-card)] border border-[var(--border-color)] rounded-3xl">
        <Route class="w-16 h-16 text-[var(--text-muted)] mb-4 opacity-20" />
        <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">
          {{ searchQuery ? t('no_matching_rules') : t('no_routing_rules') }}
        </h2>
        <p class="text-[var(--text-muted)] text-sm font-bold uppercase tracking-widest mt-2 text-center max-w-xs">
          {{ searchQuery ? t('no_matching_rules_desc') : t('no_routing_rules_desc') }}
        </p>
        <button 
          v-if="searchQuery"
          @click="searchQuery = ''"
          class="mt-6 text-xs font-black text-[var(--matrix-color)] uppercase tracking-widest hover:underline"
        >
          {{ t('clear_search') }}
        </button>
      </div>
    </div>

    <!-- Add/Edit Rule Modal -->
    <div v-if="showAddModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
      <div class="w-full max-w-md bg-[var(--bg-card)] border border-[var(--border-color)] rounded-[2rem] sm:rounded-[2.5rem] shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
        <div class="p-6 sm:p-8">
          <div class="flex items-center justify-between mb-8">
            <div>
              <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">
                {{ isEditing ? t('edit_routing_rule') : t('add_routing_rule') }}
              </h2>
              <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest mt-1">{{ t('configure_request_routing') }}</p>
            </div>
            <button @click="showAddModal = false" class="p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
              <X class="w-6 h-6 text-[var(--text-muted)]" />
            </button>
          </div>

          <form @submit.prevent="handleAddRule" class="space-y-6">
            <div class="space-y-2">
              <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ t('routing_key') }}</label>
              <input 
                v-model="newRule.key"
                type="text"
                :placeholder="t('routing_key_placeholder')"
                class="w-full px-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold text-sm"
                required
                :disabled="isEditing"
              />
              <p v-if="!isEditing" class="text-[8px] font-bold text-[var(--text-muted)] uppercase tracking-tighter ml-1">
                {{ t('routing_key_desc') }}
              </p>
            </div>

            <div class="space-y-2">
              <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ t('select_worker') }}</label>
              <div class="relative">
                <select 
                  v-model="newRule.worker_id"
                  class="w-full px-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold appearance-none text-sm"
                  required
                >
                  <option value="" disabled>{{ t('select_worker_placeholder') }}</option>
                  <option v-for="worker in workers" :key="worker.id" :value="worker.id">
                    {{ getWorkerName(worker.id) }} ({{ t(worker.status || 'unknown') }}) - {{ t(worker.type || 'unknown') }}
                  </option>
                </select>
                <div class="absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none">
                  <ArrowRight class="w-4 h-4 text-[var(--text-muted)] rotate-90" />
                </div>
              </div>
            </div>

            <button 
              type="submit"
              class="w-full py-4 rounded-2xl bg-[var(--matrix-color)] text-black font-black uppercase tracking-widest hover:opacity-90 transition-opacity mt-4 flex items-center justify-center gap-2"
            >
              <CheckCircle2 v-if="isEditing" class="w-4 h-4" />
              <Plus v-else class="w-4 h-4" />
              {{ isEditing ? t('update_rule') : t('create_rule') }}
            </button>
          </form>
        </div>
      </div>
    </div>
    </template>
  </div>
</template>

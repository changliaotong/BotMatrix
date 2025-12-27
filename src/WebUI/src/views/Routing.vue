<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useSystemStore } from '@/stores/system';
import { useBotStore } from '@/stores/bot';
import { 
  Route, 
  Plus, 
  Trash2, 
  ArrowRight, 
  Cpu, 
  Search,
  AlertCircle,
  CheckCircle2,
  X
} from 'lucide-vue-next';

const systemStore = useSystemStore();
const botStore = useBotStore();
const t = (key: string) => systemStore.t(key);

const rules = ref<any[]>([]);
const workers = ref<any[]>([]);
const loading = ref(false);
const showAddModal = ref(false);
const searchQuery = ref('');
const isEditing = ref(false);

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
  loading.value = true;
  try {
    const [rulesData, workersData] = await Promise.all([
      botStore.fetchRoutingRules(),
      botStore.fetchWorkers()
    ]);
    
    if (rulesData.success) {
      // Convert map to array if necessary
      if (rulesData.rules && typeof rulesData.rules === 'object' && !Array.isArray(rulesData.rules)) {
        rules.value = Object.entries(rulesData.rules).map(([key, worker_id]) => ({
          key,
          worker_id
        }));
      } else {
        rules.value = rulesData.rules || [];
      }
    }
    if (workersData.success) {
      workers.value = workersData.workers;
    }
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
  if (!confirm(`Are you sure you want to delete the routing rule for "${key}"?`)) return;
  
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
  return worker ? worker.id : workerId;
};

onMounted(fetchRules);
</script>

<template>
  <div class="p-6 space-y-6">
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
            placeholder="Search rules..."
            class="w-full pl-10 pr-4 py-2 rounded-2xl bg-[var(--bg-card)] border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold transition-all"
          />
        </div>
        <button 
          @click="resetForm(); showAddModal = true"
          class="flex items-center gap-2 px-4 py-2 rounded-2xl bg-[var(--matrix-color)] text-black font-black text-xs uppercase tracking-widest hover:opacity-90 transition-opacity whitespace-nowrap"
        >
          <Plus class="w-4 h-4" />
          Add Rule
        </button>
      </div>
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
                class="p-2 rounded-xl bg-[var(--matrix-color)]/10 border border-[var(--matrix-color)]/20 text-[var(--matrix-color)] hover:bg-[var(--matrix-color)] hover:text-black transition-all"
              >
                <Plus class="w-4 h-4 rotate-45 group-hover:rotate-0 transition-transform" />
              </button>
              <button 
                @click="handleDeleteRule(rule.key)"
                class="p-2 rounded-xl bg-red-500/10 border border-red-500/20 text-red-500 hover:bg-red-500 hover:text-white transition-all"
              >
                <Trash2 class="w-4 h-4" />
              </button>
            </div>
          </div>

          <div class="space-y-4">
            <div>
              <span class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">Routing Key</span>
              <div class="flex items-center gap-2 mt-1">
                <h3 class="font-black text-[var(--text-main)] break-all">{{ rule.key }}</h3>
                <div v-if="rule.key.includes('*')" class="px-2 py-0.5 rounded text-[8px] font-black uppercase tracking-tighter bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] border border-[var(--matrix-color)]/20">
                  Pattern
                </div>
              </div>
            </div>
            
            <div class="flex items-center gap-3">
              <div class="flex-1 h-px bg-[var(--border-color)]"></div>
              <ArrowRight class="w-4 h-4 text-[var(--matrix-color)]" />
              <div class="flex-1 h-px bg-[var(--border-color)]"></div>
            </div>

            <div>
              <span class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">Target Worker</span>
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
            <span class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">
              {{ workers.find(w => w.id === rule.worker_id)?.status || 'Unknown' }}
            </span>
          </div>
          <span class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">
            Type: {{ workers.find(w => w.id === rule.worker_id)?.type || 'N/A' }}
          </span>
        </div>
      </div>

      <!-- Empty State -->
      <div v-if="filteredRules.length === 0" class="col-span-full flex flex-col items-center justify-center py-20 bg-[var(--bg-card)] border border-[var(--border-color)] rounded-3xl">
        <Route class="w-16 h-16 text-[var(--text-muted)] mb-4 opacity-20" />
        <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">
          {{ searchQuery ? 'No Matching Rules' : 'No Routing Rules' }}
        </h2>
        <p class="text-[var(--text-muted)] text-sm font-bold uppercase tracking-widest mt-2 text-center max-w-xs">
          {{ searchQuery ? `No rules match your search for "${searchQuery}"` : 'Create a rule to route requests to specific workers' }}
        </p>
        <button 
          v-if="searchQuery"
          @click="searchQuery = ''"
          class="mt-6 text-xs font-black text-[var(--matrix-color)] uppercase tracking-widest hover:underline"
        >
          Clear Search
        </button>
      </div>
    </div>

    <!-- Add/Edit Rule Modal -->
    <div v-if="showAddModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showAddModal = false"></div>
      <div class="relative w-full max-w-md bg-[var(--bg-main)] border border-[var(--border-color)] rounded-3xl p-8 shadow-2xl">
        <div class="flex items-center justify-between mb-8">
          <div>
            <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">
              {{ isEditing ? 'Edit Routing Rule' : 'Add Routing Rule' }}
            </h2>
            <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest mt-1">Configure request routing</p>
          </div>
          <button @click="showAddModal = false" class="p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
            <X class="w-6 h-6 text-[var(--text-muted)]" />
          </button>
        </div>

        <form @submit.prevent="handleAddRule" class="space-y-6">
          <div class="space-y-2">
            <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">Routing Key</label>
            <input 
              v-model="newRule.key"
              type="text"
              placeholder="e.g. user_123 or * (for default)"
              class="w-full px-4 py-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold"
              required
              :disabled="isEditing"
            />
            <p v-if="!isEditing" class="text-[8px] font-bold text-[var(--text-muted)] uppercase tracking-tighter ml-1">
              Supports exact IDs or * wildcards (e.g. group_*, user_123)
            </p>
          </div>

          <div class="space-y-2">
            <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">Select Worker</label>
            <div class="relative">
              <select 
                v-model="newRule.worker_id"
                class="w-full px-4 py-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold appearance-none"
                required
              >
                <option value="" disabled>Select a worker</option>
                <option v-for="worker in workers" :key="worker.id" :value="worker.id">
                  {{ worker.id }} ({{ worker.status }}) - {{ worker.type }}
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
            {{ isEditing ? 'Update Rule' : 'Create Rule' }}
          </button>
        </form>
      </div>
    </div>
  </div>
</template>

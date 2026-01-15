<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { useSystemStore } from '@/stores/system';
import api from '@/api';
import { 
  Bot, 
  RefreshCw, 
  Plus, 
  Network, 
  Terminal, 
  GitBranch, 
  Hammer, 
  TestTube, 
  Sparkles,
  Activity,
  CheckCircle2,
  Circle,
  AlertCircle,
  Loader2
} from 'lucide-vue-next';

const systemStore = useSystemStore();
const t = (key: string) => systemStore.t(key);

interface Task {
  id: number;
  title: string;
  description: string;
  status: string;
  progress: number;
  createdAt: string;
}

interface TaskStep {
  id: number;
  name: string;
  status: string;
  outputData: string;
}

interface SubTask extends Task {
  jobKey: string;
  steps?: TaskStep[];
}

const tasks = ref<Task[]>([]);
const currentTaskId = ref<number | null>(null);
const currentTask = ref<Task | null>(null);
const subTasks = ref<SubTask[]>([]);
const globalPrompt = ref('');
const isSubmitting = ref(false);
const isLoading = ref(false);
const isLoadingDetail = ref(false);

let refreshInterval: any = null;

const fetchTasks = async () => {
  try {
    isLoading.value = true;
    const response = await api.get('/api/Task/list?limit=20');
    tasks.value = response.data;
  } catch (err) {
    console.error('Failed to fetch tasks', err);
  } finally {
    isLoading.value = false;
  }
};

const loadTaskDetail = async (id: number) => {
  currentTaskId.value = id;
  isLoadingDetail.value = true;
  try {
    const [taskRes, subTasksRes] = await Promise.all([
      api.get(`/api/Task/${id}`),
      api.get(`/api/Task/${id}/subtasks`)
    ]);
    
    currentTask.value = taskRes.data;
    const rawSubTasks = subTasksRes.data;
    
    // Fetch steps for each subtask
    const subTasksWithSteps = await Promise.all(rawSubTasks.map(async (sub: SubTask) => {
      try {
        const stepsRes = await api.get(`/api/Task/${sub.id}/steps`);
        return { ...sub, steps: stepsRes.data };
      } catch (err) {
        return { ...sub, steps: [] };
      }
    }));
    
    subTasks.value = subTasksWithSteps;
  } catch (err) {
    console.error('Failed to load task detail', err);
  } finally {
    isLoadingDetail.value = false;
  }
};

const submitGlobalTask = async () => {
  if (!globalPrompt.value.trim()) return;
  
  isSubmitting.value = true;
  try {
    const response = await api.post('/api/Task/submit', { prompt: globalPrompt.value });
    globalPrompt.value = '';
    await fetchTasks();
    if (response.data?.taskId) {
      loadTaskDetail(response.data.taskId);
    }
  } catch (err) {
    console.error('Failed to submit task', err);
  } finally {
    isSubmitting.value = false;
  }
};

const getStatusClass = (status: string) => {
  switch (status.toLowerCase()) {
    case 'executing': return 'bg-blue-500/20 text-blue-400 border-blue-500/30';
    case 'completed': return 'bg-green-500/20 text-green-400 border-green-500/30';
    case 'failed': return 'bg-red-500/20 text-red-400 border-red-500/30';
    default: return 'bg-slate-500/20 text-slate-400 border-slate-500/30';
  }
};

const getStatusIcon = (status: string) => {
  switch (status.toLowerCase()) {
    case 'executing': return Loader2;
    case 'completed': return CheckCircle2;
    case 'failed': return AlertCircle;
    default: return Circle;
  }
};

onMounted(() => {
  fetchTasks();
  refreshInterval = setInterval(() => {
    fetchTasks();
    if (currentTaskId.value) {
      loadTaskDetail(currentTaskId.value);
    }
  }, 5000);
});

onUnmounted(() => {
  if (refreshInterval) clearInterval(refreshInterval);
});
</script>

<template>
  <div class="h-full flex flex-col gap-6 p-6 overflow-hidden">
    <!-- Top Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 bg-[var(--matrix-color)] rounded-xl flex items-center justify-center shadow-lg shadow-[var(--matrix-color)]/20">
          <Bot class="text-white w-6 h-6" />
        </div>
        <div>
          <h1 class="text-xl font-bold tracking-tight text-[var(--text-main)]">
            {{ t('global_dashboard') }}
          </h1>
          <p class="text-xs text-[var(--text-muted)] uppercase tracking-widest font-bold">
            Project Orchestration Hub
          </p>
        </div>
      </div>
      <div class="flex items-center gap-4">
        <div class="flex items-center gap-2 text-xs font-bold text-green-500 bg-green-500/10 px-3 py-1.5 rounded-full border border-green-500/20">
          <span class="w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
          {{ t('system_online') }}
        </div>
        <button @click="fetchTasks" class="p-2 hover:bg-white/5 rounded-full transition-colors text-[var(--text-muted)] hover:text-[var(--text-main)]">
          <RefreshCw class="w-5 h-5" :class="{ 'animate-spin': isLoading }" />
        </button>
      </div>
    </div>

    <div class="flex-1 grid grid-cols-12 gap-6 overflow-hidden">
      <!-- Left: Task List -->
      <div class="col-span-12 lg:col-span-3 flex flex-col gap-4 overflow-hidden">
        <div class="flex items-center justify-between">
          <h2 class="text-[10px] font-black uppercase tracking-[0.2em] text-[var(--text-muted)] opacity-60">
            {{ t('recent_tasks') }}
          </h2>
          <button class="p-1.5 bg-[var(--matrix-color)]/10 hover:bg-[var(--matrix-color)]/20 text-[var(--matrix-color)] rounded-lg transition-all">
            <Plus class="w-4 h-4" />
          </button>
        </div>
        <div class="flex-1 overflow-y-auto space-y-3 pr-2 custom-scrollbar">
          <div v-if="isLoading && tasks.length === 0" class="space-y-3">
            <div v-for="i in 5" :key="i" class="h-24 bg-white/5 rounded-2xl animate-pulse border border-white/5"></div>
          </div>
          <div v-else v-for="task in tasks" :key="task.id" 
               @click="loadTaskDetail(task.id)"
               :class="[
                 'p-4 rounded-2xl cursor-pointer transition-all border group relative overflow-hidden',
                 currentTaskId === task.id 
                   ? 'bg-[var(--matrix-color)]/10 border-[var(--matrix-color)]/30 shadow-lg shadow-[var(--matrix-color)]/5' 
                   : 'bg-white/5 border-white/5 hover:border-white/10 hover:bg-white/10'
               ]">
            <div class="flex justify-between items-start mb-2 relative z-10">
              <h3 class="text-sm font-bold truncate pr-2 text-[var(--text-main)]">{{ task.title }}</h3>
              <span :class="['status-badge text-[10px] px-2 py-0.5 rounded-full border font-bold uppercase', getStatusClass(task.status)]">
                {{ task.status }}
              </span>
            </div>
            <div class="flex items-center justify-between text-[10px] text-[var(--text-muted)] mb-3 relative z-10">
              <span>{{ new Date(task.createdAt).toLocaleString() }}</span>
              <span class="font-black text-[var(--matrix-color)]">{{ task.progress }}%</span>
            </div>
            <div class="w-full bg-white/5 h-1 rounded-full overflow-hidden relative z-10">
              <div class="bg-[var(--matrix-color)] h-full transition-all duration-500 shadow-[0_0_8px_var(--matrix-color)]" :style="{ width: task.progress + '%' }"></div>
            </div>
            <!-- Glow Effect on Active -->
            <div v-if="currentTaskId === task.id" class="absolute inset-0 bg-gradient-to-r from-[var(--matrix-color)]/5 to-transparent"></div>
          </div>
        </div>
      </div>

      <!-- Center: Task Detail -->
      <div class="col-span-12 lg:col-span-6 bg-white/5 border border-white/5 rounded-3xl p-6 flex flex-col overflow-hidden relative">
        <template v-if="currentTask">
          <div class="flex items-start justify-between mb-8">
            <div class="flex-1">
              <div class="flex items-center gap-3 mb-2">
                <h2 class="text-2xl font-black text-[var(--text-main)] tracking-tight">{{ currentTask.title }}</h2>
                <span :class="['px-3 py-1 rounded-full text-[10px] font-black uppercase border', getStatusClass(currentTask.status)]">
                  {{ currentTask.status }}
                </span>
              </div>
              <p class="text-sm text-[var(--text-muted)] leading-relaxed max-w-2xl">{{ currentTask.description }}</p>
            </div>
            <div class="text-right">
              <div class="text-4xl font-black text-[var(--matrix-color)] drop-shadow-[0_0_10px_var(--matrix-color)]">{{ currentTask.progress }}%</div>
              <div class="text-[10px] uppercase tracking-[0.2em] font-black text-[var(--text-muted)] mt-1">{{ t('completion_rate') }}</div>
            </div>
          </div>

          <!-- Subtasks Timeline -->
          <div class="flex-1 overflow-y-auto pr-4 custom-scrollbar relative">
            <div v-if="isLoadingDetail" class="flex items-center justify-center h-full">
              <Loader2 class="w-8 h-8 text-[var(--matrix-color)] animate-spin" />
            </div>
            <div v-else-if="subTasks.length === 0" class="flex flex-col items-center justify-center h-full text-[var(--text-muted)] opacity-50">
            <Network class="w-12 h-12 mb-4" />
          </div>
          <div v-else class="space-y-10 pl-8 relative">
            <!-- Vertical Line -->
            <div class="absolute left-3.5 top-2 bottom-2 w-px bg-white/10"></div>
            
            <div v-for="sub in subTasks" :key="sub.id" class="relative">
              <!-- Dot -->
              <div :class="[
                'absolute -left-8 w-7 h-7 rounded-full border-4 border-[var(--bg-body)] flex items-center justify-center z-10 transition-all duration-300',
                sub.status === 'executing' ? 'bg-[var(--matrix-color)] shadow-[0_0_15px_var(--matrix-color)] scale-110' : 'bg-white/10'
              ]">
                <component :is="getStatusIcon(sub.status)" class="w-3 h-3 text-white" :class="{ 'animate-spin': sub.status === 'executing' }" />
              </div>

              <div class="flex items-center justify-between mb-3">
                <div class="flex items-center gap-3">
                  <h4 class="font-black text-[var(--text-main)] text-sm tracking-wide">{{ sub.title }}</h4>
                  <span class="text-[9px] font-black px-2 py-0.5 rounded bg-white/5 text-[var(--text-muted)] border border-white/5 uppercase tracking-tighter">{{ sub.jobKey }}</span>
                </div>
                <span :class="['text-[9px] font-black px-2 py-0.5 rounded-full border uppercase', getStatusClass(sub.status)]">
                  {{ sub.status }}
                </span>
              </div>
              <p class="text-xs text-[var(--text-muted)] mb-4 leading-relaxed bg-white/5 p-3 rounded-xl border border-white/5">{{ sub.description }}</p>
              
              <!-- Steps -->
              <div v-if="sub.steps && sub.steps.length > 0" class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div v-for="step in sub.steps" :key="step.id" class="bg-black/20 rounded-2xl p-4 border border-white/5 hover:border-[var(--matrix-color)]/30 transition-all group">
                  <div class="flex items-center justify-between mb-2">
                    <span class="text-[10px] font-black text-[var(--text-muted)] group-hover:text-[var(--matrix-color)] transition-colors uppercase tracking-widest">{{ step.name }}</span>
                    <span class="text-[9px] font-bold px-1.5 py-0.5 rounded-md bg-white/5 text-[var(--text-muted)] border border-white/5">{{ step.status }}</span>
                  </div>
                  <div class="text-[11px] text-[var(--text-muted)] font-mono line-clamp-2 bg-black/40 p-2 rounded-lg border border-white/5">
                    {{ step.outputData || 'Waiting for output...' }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </template>
        <div v-else class="flex-1 flex flex-col items-center justify-center text-[var(--text-muted)] text-center p-12">
          <div class="w-24 h-24 rounded-full bg-[var(--matrix-color)]/5 flex items-center justify-center mb-6 border border-[var(--matrix-color)]/10">
            <Network class="w-10 h-10 text-[var(--matrix-color)] opacity-50" />
          </div>
          <h3 class="text-xl font-black text-[var(--text-main)] mb-2">{{ t('select_task_to_view') }}</h3>
          <p class="text-sm opacity-60 max-w-xs">{{ t('monitor_agent_actions_realtime') }}</p>
        </div>
      </div>

      <!-- Right: Controls -->
      <div class="col-span-12 lg:col-span-3 flex flex-col gap-6 overflow-y-auto pr-2 custom-scrollbar">
        <!-- Global Command -->
        <div class="bg-white/5 border border-white/5 rounded-3xl p-6 relative overflow-hidden group">
          <div class="absolute top-0 right-0 w-32 h-32 bg-[var(--matrix-color)]/10 blur-3xl rounded-full -mr-16 -mt-16 transition-all group-hover:bg-[var(--matrix-color)]/20"></div>
          
          <h3 class="text-xs font-black mb-4 flex items-center gap-2 text-[var(--text-main)] uppercase tracking-[0.2em]">
            <Terminal class="w-4 h-4 text-[var(--matrix-color)]" />
            {{ t('global_orchestration') }}
          </h3>
          <textarea 
            v-model="globalPrompt"
            class="w-full bg-black/30 border border-white/5 rounded-2xl p-4 text-sm text-[var(--text-main)] placeholder:text-[var(--text-muted)]/30 focus:outline-none focus:border-[var(--matrix-color)]/50 transition-all mb-4 resize-none custom-scrollbar" 
            rows="5" 
            :placeholder="t('enter_global_goal_placeholder')"
          ></textarea>
          <button 
            @click="submitGlobalTask"
            :disabled="isSubmitting || !globalPrompt.trim()"
            class="w-full bg-[var(--matrix-color)] hover:scale-[1.02] active:scale-[0.98] disabled:opacity-50 disabled:scale-100 text-white font-black py-4 rounded-2xl shadow-xl shadow-[var(--matrix-color)]/20 transition-all flex items-center justify-center gap-2"
          >
            <template v-if="isSubmitting">
              <Loader2 class="w-5 h-5 animate-spin" />
              {{ t('submitting') }}
            </template>
            <template v-else>
              <Sparkles class="w-5 h-5" />
              {{ t('orchestrate_now') }}
            </template>
          </button>
        </div>

        <!-- Quick Actions -->
        <div class="bg-white/5 border border-white/5 rounded-3xl p-6">
          <h3 class="text-xs font-black mb-4 text-[var(--text-main)] uppercase tracking-[0.2em]">{{ t('quick_controls') }}</h3>
          <div class="grid grid-cols-2 gap-3">
            <button class="flex flex-col items-center justify-center p-4 rounded-2xl bg-white/5 hover:bg-[var(--matrix-color)]/10 border border-white/5 hover:border-[var(--matrix-color)]/30 transition-all group">
              <GitBranch class="w-5 h-5 text-purple-400 mb-2 group-hover:scale-110 transition-transform" />
              <span class="text-[10px] font-black text-[var(--text-muted)] group-hover:text-[var(--text-main)] transition-colors uppercase tracking-tighter">Sync Git</span>
            </button>
            <button class="flex flex-col items-center justify-center p-4 rounded-2xl bg-white/5 hover:bg-[var(--matrix-color)]/10 border border-white/5 hover:border-[var(--matrix-color)]/30 transition-all group">
              <Hammer class="w-5 h-5 text-orange-400 mb-2 group-hover:scale-110 transition-transform" />
              <span class="text-[10px] font-black text-[var(--text-muted)] group-hover:text-[var(--text-main)] transition-colors uppercase tracking-tighter">Full Build</span>
            </button>
            <button class="flex flex-col items-center justify-center p-4 rounded-2xl bg-white/5 hover:bg-[var(--matrix-color)]/10 border border-white/5 hover:border-[var(--matrix-color)]/30 transition-all group">
              <TestTube class="w-5 h-5 text-green-400 mb-2 group-hover:scale-110 transition-transform" />
              <span class="text-[10px] font-black text-[var(--text-muted)] group-hover:text-[var(--text-main)] transition-colors uppercase tracking-tighter">Run Tests</span>
            </button>
            <button class="flex flex-col items-center justify-center p-4 rounded-2xl bg-white/5 hover:bg-[var(--matrix-color)]/10 border border-white/5 hover:border-[var(--matrix-color)]/30 transition-all group">
              <Sparkles class="w-5 h-5 text-pink-400 mb-2 group-hover:scale-110 transition-transform" />
              <span class="text-[10px] font-black text-[var(--text-muted)] group-hover:text-[var(--text-main)] transition-colors uppercase tracking-tighter">Auto Evolve</span>
            </button>
          </div>
        </div>

        <!-- Agent Status -->
        <div class="bg-white/5 border border-white/5 rounded-3xl p-6">
          <h3 class="text-xs font-black mb-4 text-[var(--text-main)] uppercase tracking-[0.2em]">{{ t('agent_status_board') }}</h3>
          <div class="space-y-4">
            <div class="flex items-center justify-between p-3 rounded-2xl bg-black/20 border border-white/5">
              <div class="flex items-center gap-3">
                <div class="w-2 h-2 bg-blue-500 rounded-full shadow-[0_0_8px_rgba(59,130,246,0.5)]"></div>
                <span class="text-xs font-bold text-[var(--text-main)]">Dev Orchestrator</span>
              </div>
              <span class="text-[9px] font-black bg-blue-500/20 text-blue-400 px-2 py-0.5 rounded-full border border-blue-500/30 uppercase tracking-tighter">BUSY</span>
            </div>
            <div class="flex items-center justify-between p-3 rounded-2xl bg-black/20 border border-white/5">
              <div class="flex items-center gap-3">
                <div class="w-2 h-2 bg-slate-500 rounded-full"></div>
                <span class="text-xs font-bold text-[var(--text-main)]">Code Reviewer</span>
              </div>
              <span class="text-[9px] font-black bg-white/5 text-[var(--text-muted)] px-2 py-0.5 rounded-full border border-white/5 uppercase tracking-tighter">IDLE</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 10px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.2);
}

.status-badge {
  @apply transition-all duration-300;
}
</style>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useSystemStore } from '@/stores/system';
import { useBotStore } from '@/stores/bot';
import { 
  ListTodo, 
  Plus, 
  Play, 
  Pause, 
  Trash2, 
  Clock, 
  Calendar,
  Settings2,
  AlertCircle,
  CheckCircle2,
  X,
  Type,
  Hash,
  Sparkles,
  Brain,
  Wand2,
  Send,
  Loader2
} from 'lucide-vue-next';

const systemStore = useSystemStore();
const botStore = useBotStore();
const t = (key: string) => systemStore.t(key);

const tasks = ref<any[]>([]);
const loading = ref(true);
const showCreateModal = ref(false);
const isEditing = ref(false);
const editingTaskId = ref<string | null>(null);
const capabilities = ref({ actions: [], interceptors: [] });

// AI Parsing State
const aiInput = ref('');
const aiParsing = ref(false);
const aiResult = ref<any>(null);
const showAIConfirm = ref(false);

const newTask = ref({
  name: '',
  schedule: '',
  type: 'once',
  action_type: 'send_message',
  action_params: '{}',
  priority: 1,
  max_retries: 3
});

const fetchTasks = async () => {
  loading.value = true;
  try {
    const [tasksData, capsData] = await Promise.all([
      botStore.fetchTasks(),
      botStore.fetchTaskCapabilities()
    ]);
    
    if (tasksData.success) {
      tasks.value = tasksData.tasks;
    }
    if (capsData.success) {
      capabilities.value = capsData.data;
    }
  } finally {
    loading.value = false;
  }
};

const handleAIParse = async () => {
  if (!aiInput.value.trim()) return;
  
  aiParsing.value = true;
  aiResult.value = null;
  try {
    const data = await botStore.aiParse(aiInput.value);
    if (data.success) {
      aiResult.value = data.data;
      showAIConfirm.value = true;
    }
  } catch (err) {
    console.error('AI parse failed:', err);
  } finally {
    aiParsing.value = false;
  }
};

const handleAIConfirm = async () => {
  if (!aiResult.value?.draft_id) return;
  
  try {
    const data = await botStore.aiConfirm(aiResult.value.draft_id);
    if (data.success) {
      await fetchTasks();
      showAIConfirm.value = false;
      aiInput.value = '';
      aiResult.value = null;
    }
  } catch (err) {
    console.error('AI confirm failed:', err);
  }
};

const handleCreateTask = async () => {
  try {
    let data;
    const payload = {
      ...newTask.value,
      priority: Number(newTask.value.priority),
      max_retries: Number(newTask.value.max_retries)
    };

    if (isEditing.value && editingTaskId.value) {
      data = await botStore.updateTask(editingTaskId.value, payload);
    } else {
      data = await botStore.createTask(payload);
    }
    
    if (data.success) {
      await fetchTasks();
      showCreateModal.value = false;
      resetForm();
    }
  } catch (err) {
    console.error('Failed to save task:', err);
  }
};

const handleEditTask = (task: any) => {
  isEditing.value = true;
  editingTaskId.value = task.ID || task.id;
  newTask.value = {
    name: task.Name || task.name,
    schedule: task.Schedule || task.schedule,
    type: task.Type || task.type,
    action_type: task.ActionType || task.action_type || 'send_message',
    action_params: task.ActionParams || task.action_params || '{}',
    priority: task.Priority || task.priority || 1,
    max_retries: task.MaxRetries || task.max_retries || 3
  };
  showCreateModal.value = true;
};

const resetForm = () => {
  isEditing.value = false;
  editingTaskId.value = null;
  newTask.value = { 
    name: '', 
    schedule: '', 
    type: 'once', 
    action_type: 'send_message',
    action_params: '{}',
    priority: 1,
    max_retries: 3
  };
};

const handleToggleTask = async (task: any) => {
  const newStatus = task.status === 'running' ? 'paused' : 'running';
  try {
    const data = await botStore.toggleTask(task.id, newStatus);
    if (data.success) {
      await fetchTasks();
    }
  } catch (err) {
    console.error('Failed to toggle task:', err);
  }
};

const handleDeleteTask = async (taskId: string) => {
  if (!confirm(t('confirm_delete_task'))) return;
  try {
    const data = await botStore.deleteTask(taskId);
    if (data.success) {
      await fetchTasks();
    }
  } catch (err) {
    console.error('Failed to delete task:', err);
  }
};

onMounted(fetchTasks);

const getStatusStyle = (status: string) => {
  switch (status.toLowerCase()) {
    case 'running': return 'text-green-500 bg-green-500/10 border-green-500/20';
    case 'paused': return 'text-yellow-500 bg-yellow-500/10 border-yellow-500/20';
    case 'failed': return 'text-red-500 bg-red-500/10 border-red-500/20';
    default: return 'text-[var(--text-muted)] bg-black/5 dark:bg-white/5 border-[var(--border-color)]';
  }
};

</script>

<template>
  <div class="p-6 space-y-6">
    <!-- AI Intent Parser -->
    <div class="p-8 rounded-[32px] bg-gradient-to-br from-[var(--matrix-color)]/10 via-purple-600/10 to-pink-600/10 border border-[var(--matrix-color)]/20">
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-6">
        <div class="flex items-center gap-4">
          <div class="p-4 rounded-2xl bg-[var(--matrix-color)]/20 border border-[var(--matrix-color)]/30">
            <Brain class="w-8 h-8 text-[var(--matrix-color)]" />
          </div>
          <div>
            <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">{{ t('ai_task_orchestrator') }}</h2>
            <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest mt-1">{{ t('ai_task_desc') }}</p>
          </div>
        </div>
        
        <div class="flex-1 max-w-2xl">
          <div class="relative group">
            <input 
              v-model="aiInput"
              @keyup.enter="handleAIParse"
              type="text"
              :placeholder="t('ai_task_placeholder')"
              class="w-full pl-6 pr-32 py-5 rounded-3xl bg-black/20 dark:bg-white/5 border border-white/10 focus:border-[var(--matrix-color)]/50 outline-none text-[var(--text-main)] transition-all font-bold placeholder:text-[var(--text-muted)]/50"
            />
            <button 
              @click="handleAIParse"
              :disabled="aiParsing || !aiInput"
              class="absolute right-2 top-2 bottom-2 px-6 rounded-2xl bg-[var(--matrix-color)] text-black font-black text-xs uppercase tracking-widest hover:opacity-90 disabled:opacity-50 transition-all flex items-center gap-2"
            >
              <Loader2 v-if="aiParsing" class="w-4 h-4 animate-spin" />
              <Sparkles v-else class="w-4 h-4" />
              {{ aiParsing ? t('analyzing') : t('generate') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Task List Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-black text-[var(--text-main)] tracking-tight">{{ t('tasks') }}</h1>
        <p class="text-sm font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('tasks_desc') }}</p>
      </div>
      <button 
        @click="resetForm(); showCreateModal = true"
        class="flex items-center gap-2 px-4 py-2 rounded-2xl bg-[var(--matrix-color)] text-black font-black text-xs uppercase tracking-widest hover:opacity-90 transition-opacity"
      >
        <Plus class="w-4 h-4" />
        {{ t('create_task') }}
      </button>
    </div>

    <!-- AI Result Confirmation Modal -->
    <div v-if="showAIConfirm" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
      <div class="w-full max-w-lg bg-[var(--bg-card)] rounded-[2rem] sm:rounded-[2.5rem] border border-[var(--border-color)] shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
        <div class="p-6 sm:p-8 space-y-4 sm:space-y-6 max-h-[90vh] overflow-y-auto no-scrollbar relative">
          <!-- Decoration -->
          <div class="absolute -top-24 -right-24 w-48 h-48 bg-[var(--matrix-color)]/10 blur-[100px] rounded-full pointer-events-none"></div>
          <div class="absolute -bottom-24 -left-24 w-48 h-48 bg-purple-500/10 blur-[100px] rounded-full pointer-events-none"></div>

          <div class="relative">
            <div class="flex items-center justify-between mb-6">
              <div class="flex items-center gap-4">
                <div class="p-3 rounded-2xl bg-[var(--matrix-color)]/10 border border-[var(--matrix-color)]/20">
                  <Wand2 class="w-6 h-6 text-[var(--matrix-color)]" />
                </div>
                <div>
                  <h3 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">{{ t('ai_proposal') }}</h3>
                  <p class="text-[10px] font-bold text-[var(--matrix-color)] uppercase tracking-widest">{{ t('confidence_score') }}: 98%</p>
                </div>
              </div>
              <button @click="showAIConfirm = false" class="p-2 hover:bg-black/5 dark:hover:bg-white/5 rounded-xl transition-colors">
                <X class="w-5 h-5 text-[var(--text-muted)]" />
              </button>
            </div>

            <div class="space-y-6">
              <div class="p-6 rounded-3xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] space-y-4">
                <div class="flex items-start gap-4">
                  <div class="p-2 rounded-xl bg-[var(--matrix-color)]/10">
                    <CheckCircle2 class="w-5 h-5 text-[var(--matrix-color)]" />
                  </div>
                  <div>
                    <p class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest mb-1">{{ t('intent_analysis') }}</p>
                    <p class="text-sm text-[var(--text-main)] font-bold">{{ aiResult?.summary }}</p>
                  </div>
                </div>
                <div class="flex items-start gap-4">
                  <div class="p-2 rounded-xl bg-[var(--matrix-color)]/10">
                    <Brain class="w-5 h-5 text-[var(--matrix-color)]" />
                  </div>
                  <div>
                    <p class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest mb-1">{{ t('ai_reasoning') }}</p>
                    <p class="text-xs text-[var(--text-main)] font-medium leading-relaxed opacity-80">{{ aiResult?.analysis }}</p>
                  </div>
                </div>
              </div>

              <div class="grid grid-cols-2 gap-4">
                <div class="p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
                  <p class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest mb-1">{{ t('action_type') }}</p>
                  <p class="text-xs font-black text-[var(--matrix-color)] uppercase">{{ aiResult?.data?.action_type || aiResult?.intent }}</p>
                </div>
                <div class="p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
                  <p class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest mb-1">{{ t('schedule') }}</p>
                  <p class="text-xs font-black text-[var(--text-main)]">{{ aiResult?.data?.trigger_config ? JSON.parse(aiResult.data.trigger_config).cron || t('once') : t('once') }}</p>
                </div>
              </div>

              <div class="flex gap-4 pt-4">
                <button 
                  @click="showAIConfirm = false"
                  class="flex-1 py-4 rounded-2xl border border-[var(--border-color)] text-[10px] font-black uppercase tracking-widest hover:bg-black/5 dark:hover:bg-white/5 transition-all text-[var(--text-muted)]"
                >
                  {{ t('cancel') }}
                </button>
                <button 
                  @click="handleAIConfirm"
                  class="flex-1 py-4 rounded-2xl bg-[var(--matrix-color)] text-black text-[10px] font-black uppercase tracking-widest hover:opacity-90 transition-all shadow-lg shadow-[var(--matrix-color)]/20 flex items-center justify-center gap-2"
                >
                  <Send class="w-4 h-4" />
                  {{ t('execute_plan') }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Task List -->
    <div v-if="loading" class="space-y-4 animate-pulse">
      <div v-for="i in 4" :key="i" class="h-24 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)]"></div>
    </div>

    <div v-else class="space-y-4">
      <div 
        v-for="task in tasks" 
        :key="task.id"
        class="group p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] hover:border-[var(--matrix-color)]/30 transition-all duration-500"
      >
        <div class="flex flex-col md:flex-row md:items-center justify-between gap-6">
          <div class="flex items-center gap-4">
            <div class="p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
              <ListTodo class="w-6 h-6 text-[var(--matrix-color)]" />
            </div>
            <div>
              <h3 class="font-black text-[var(--text-main)] uppercase tracking-tight">{{ task.name }}</h3>
              <div class="flex items-center gap-4 mt-1">
                <div class="flex items-center gap-1.5">
                  <Clock class="w-3 h-3 text-[var(--text-muted)]" />
                  <span class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ task.schedule }}</span>
                </div>
                <div class="flex items-center gap-1.5">
                  <Calendar class="w-3 h-3 text-[var(--text-muted)]" />
                  <span class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('last_run') }}: {{ task.last_run || t('never') }}</span>
                </div>
              </div>
            </div>
          </div>

          <div class="flex items-center gap-3">
            <span 
              :class="getStatusStyle(task.status)"
              class="px-2 py-0.5 rounded-lg border text-[8px] font-black uppercase tracking-widest"
            >
              {{ t('status_' + task.status.toLowerCase()) }}
            </span>
            <span class="text-[10px] font-black text-[var(--matrix-color)] uppercase tracking-widest">{{ t(task.type) }}</span>
          </div>
            
            <div class="flex items-center gap-2">
              <button 
                @click="handleEditTask(task)"
                class="p-2 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] hover:border-[var(--matrix-color)]/30 text-[var(--text-muted)] hover:text-[var(--matrix-color)] transition-all"
                :title="t('edit_task')"
              >
                <Settings2 class="w-4 h-4" />
              </button>
              <button 
                @click="handleToggleTask(task)"
                :class="task.status === 'paused' ? 'bg-green-500/10 border-green-500/20 text-green-500 hover:bg-green-500 hover:text-[var(--sidebar-text)]' : 'bg-yellow-500/10 border-yellow-500/20 text-yellow-500 hover:bg-yellow-500 hover:text-[var(--sidebar-text)]'"
                class="p-2 rounded-xl border transition-all"
              >
                <Play v-if="task.status === 'paused'" class="w-4 h-4" />
                <Pause v-else class="w-4 h-4" />
              </button>
              <button 
                @click="handleDeleteTask(task.id)"
                class="p-2 rounded-xl bg-red-500/10 border border-red-500/20 text-red-500 hover:bg-red-500 hover:text-[var(--sidebar-text)] transition-all"
              >
                <Trash2 class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-if="tasks.length === 0" class="flex flex-col items-center justify-center py-20 bg-[var(--bg-card)] border border-[var(--border-color)] rounded-3xl">
        <ListTodo class="w-16 h-16 text-[var(--text-muted)] mb-4 opacity-20" />
        <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">{{ t('no_tasks_configured') }}</h2>
        <p class="text-[var(--text-muted)] text-sm font-bold uppercase tracking-widest mt-2">{{ t('create_first_task_desc') }}</p>
      </div>

    <!-- Create Task Modal -->
    <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
      <div class="w-full max-w-lg bg-[var(--bg-card)] rounded-[2rem] sm:rounded-[2.5rem] border border-[var(--border-color)] shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
        <div class="p-6 sm:p-8 space-y-4 sm:space-y-6 max-h-[90vh] overflow-y-auto no-scrollbar">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">{{ isEditing ? t('edit_automation_task') : t('create_automation_task') }}</h2>
              <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest mt-1">{{ isEditing ? t('modify_existing_automation') : t('configure_scheduled_automation') }}</p>
            </div>
            <button @click="showCreateModal = false" class="p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
              <X class="w-6 h-6 text-[var(--text-muted)]" />
            </button>
          </div>

          <form @submit.prevent="handleCreateTask" class="space-y-6">
            <div class="space-y-2">
              <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ t('task_name') }}</label>
              <div class="relative">
                <Type class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
                <input 
                  v-model="newTask.name"
                  type="text"
                  :placeholder="t('task_name_placeholder')"
                  class="w-full pl-12 pr-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold text-sm"
                  required
                />
              </div>
            </div>

            <div class="space-y-2">
              <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ t('schedule_cron') }}</label>
              <div class="relative">
                <Clock class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
                <input 
                  v-model="newTask.schedule"
                  type="text"
                  :placeholder="t('cron_placeholder')"
                  class="w-full pl-12 pr-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold text-sm"
                  required
                />
              </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-2">
                <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ t('task_type') }}</label>
                <div class="relative">
                  <select 
                    v-model="newTask.type"
                    class="w-full px-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold appearance-none text-sm"
                  >
                    <option value="once">{{ t('once') }}</option>
                    <option value="cron">{{ t('cron') }}</option>
                    <option value="condition">{{ t('condition') }}</option>
                  </select>
                </div>
              </div>
              <div class="space-y-2">
                <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ t('action_type') }}</label>
                <div class="relative">
                  <select 
                    v-model="newTask.action_type"
                    class="w-full px-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold appearance-none text-sm"
                  >
                    <option v-for="action in capabilities.actions" :key="action" :value="action">{{ t(action) }}</option>
                    <option v-if="capabilities.actions.length === 0" value="send_message">{{ t('send_message') }}</option>
                  </select>
                </div>
              </div>
            </div>

            <div class="space-y-2">
              <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ t('action_params') }}</label>
              <div class="relative">
                <Hash class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
                <input 
                  v-model="newTask.action_params"
                  type="text"
                  :placeholder='t("action_params_placeholder")'
                  class="w-full pl-12 pr-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold text-sm"
                />
              </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-2">
                <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ t('priority') }}</label>
                <input 
                  v-model="newTask.priority"
                  type="number"
                  class="w-full px-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold text-sm"
                />
              </div>
              <div class="space-y-2">
                <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ t('max_retries') }}</label>
                <input 
                  v-model="newTask.max_retries"
                  type="number"
                  class="w-full px-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold text-sm"
                />
              </div>
            </div>

            <div class="flex gap-4 pt-2">
              <button 
                type="button"
                @click="showCreateModal = false"
                class="flex-1 py-4 rounded-2xl border border-[var(--border-color)] text-[10px] font-black uppercase tracking-widest hover:bg-black/5 dark:hover:bg-white/5 transition-all text-[var(--text-muted)]"
              >
                {{ t('cancel') }}
              </button>
              <button 
                type="submit"
                class="flex-1 py-4 rounded-2xl bg-[var(--matrix-color)] text-black text-[10px] font-black uppercase tracking-widest hover:opacity-90 transition-all shadow-lg shadow-[var(--matrix-color)]/20"
              >
                {{ isEditing ? t('update_task') : t('create_task') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>
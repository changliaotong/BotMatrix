<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useSystemStore } from '@/stores/system';
import { useBotStore } from '@/stores/bot';
import { 
  Share2, 
  Plus, 
  Settings, 
  TrendingUp, 
  Users, 
  MessageSquare,
  ChevronRight,
  Save,
  Activity,
  CheckCircle2,
  Clock,
  X,
  Type,
  Calendar,
  Trash2
} from 'lucide-vue-next';

const systemStore = useSystemStore();
const botStore = useBotStore();
const t = (key: string) => systemStore.t(key);

const config = ref<any>({
  invite_keyword: 'invite',
  welcome_msg: 'Welcome to our group!',
  invite_msg: 'Invite your friends to join!',
  max_invites_per_day: 10,
  auto_approve: true
});

const tasks = ref<any[]>([]);
const stats = ref<any>(null);
const leaderboard = ref<any[]>([]);
const invitations = ref<any[]>([]);
const loading = ref(true);
const activeTab = ref('tasks'); // 'tasks', 'leaderboard', 'invitations', 'config'
const showCreateModal = ref(false);

const newTask = ref({
  name: '',
  keyword: '',
  target_group: '',
  end_date: '',
  status: 'active'
});

const fetchData = async () => {
  loading.value = true;
  try {
    const [configData, tasksData, statsData, leaderboardData, invitationsData] = await Promise.all([
      botStore.fetchFissionConfig(),
      botStore.fetchFissionTasks(),
      botStore.fetchFissionStats(),
      botStore.fetchFissionLeaderboard(),
      botStore.fetchFissionInvitations()
    ]);
    
    if (configData.success) {
      config.value = { ...config.value, ...configData.config };
    }
    if (tasksData.success) {
      tasks.value = tasksData.tasks;
    }
    if (statsData.success) {
      stats.value = statsData.stats;
    }
    if (leaderboardData.success) {
      leaderboard.value = leaderboardData.leaderboard;
    }
    if (invitationsData.success) {
      invitations.value = invitationsData.invitations;
    }
  } finally {
    loading.value = false;
  }
};

const handleSaveConfig = async () => {
  try {
    const data = await botStore.updateFissionConfig(config.value);
    if (data.success) {
      // Success logic could go here
    }
  } catch (err) {
    console.error('Failed to save config:', err);
  }
};

const handleCreateTask = async () => {
  try {
    const data = await botStore.saveFissionTask(newTask.value);
    if (data.success) {
      await fetchData();
      showCreateModal.value = false;
      newTask.value = { name: '', keyword: '', target_group: '', end_date: '', status: 'active' };
    }
  } catch (err) {
    console.error('Failed to create fission task:', err);
  }
};

const handleDeleteTask = async (taskId: string) => {
  if (!confirm(t('confirm_delete_campaign'))) return;
  try {
    const data = await botStore.deleteFissionTask(taskId);
    if (data.success) {
      await fetchData();
    }
  } catch (err) {
    console.error('Failed to delete fission task:', err);
  }
};

onMounted(fetchData);

const getTaskStatusStyle = (status: string) => {
  switch (status.toLowerCase()) {
    case 'active': return 'text-green-500 bg-green-500/10 border-green-500/20';
    case 'paused': return 'text-yellow-500 bg-yellow-500/10 border-yellow-500/20';
    default: return 'text-[var(--text-muted)] bg-black/5 dark:bg-white/5 border-[var(--border-color)]';
  }
};
</script>

<template>
  <div class="p-6 space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-black text-[var(--text-main)] tracking-tight">{{ t('fission') }}</h1>
        <p class="text-sm font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('fission_desc') }}</p>
      </div>
      <div class="flex bg-[var(--bg-card)] border border-[var(--border-color)] p-1 rounded-2xl overflow-x-auto no-scrollbar">
        <button 
          @click="activeTab = 'tasks'"
          :class="activeTab === 'tasks' ? 'bg-[var(--matrix-color)] text-black' : 'text-[var(--text-muted)] hover:text-[var(--text-main)]'"
          class="px-4 py-2 rounded-xl text-xs font-black uppercase tracking-widest transition-all whitespace-nowrap"
        >
          {{ t('tasks') }}
        </button>
        <button 
          @click="activeTab = 'leaderboard'"
          :class="activeTab === 'leaderboard' ? 'bg-[var(--matrix-color)] text-black' : 'text-[var(--text-muted)] hover:text-[var(--text-main)]'"
          class="px-4 py-2 rounded-xl text-xs font-black uppercase tracking-widest transition-all whitespace-nowrap"
        >
          {{ t('leaderboard') }}
        </button>
        <button 
          @click="activeTab = 'invitations'"
          :class="activeTab === 'invitations' ? 'bg-[var(--matrix-color)] text-black' : 'text-[var(--text-muted)] hover:text-[var(--text-main)]'"
          class="px-4 py-2 rounded-xl text-xs font-black uppercase tracking-widest transition-all whitespace-nowrap"
        >
          {{ t('invitations') }}
        </button>
        <button 
          @click="activeTab = 'config'"
          :class="activeTab === 'config' ? 'bg-[var(--matrix-color)] text-black' : 'text-[var(--text-muted)] hover:text-[var(--text-main)]'"
          class="px-4 py-2 rounded-xl text-xs font-black uppercase tracking-widest transition-all whitespace-nowrap"
        >
          {{ t('config') }}
        </button>
      </div>
    </div>

    <!-- Stats Overview -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] relative overflow-hidden group">
        <div class="absolute top-0 right-0 p-8 opacity-5 group-hover:scale-110 transition-transform">
          <TrendingUp class="w-24 h-24 text-[var(--matrix-color)]" />
        </div>
        <div class="relative z-10">
          <div class="p-3 rounded-2xl bg-blue-500/10 text-blue-500 w-fit mb-4">
            <TrendingUp class="w-6 h-6" />
          </div>
          <div class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('growth_rate') }}</div>
          <div class="text-3xl font-black text-[var(--text-main)] mt-1">{{ stats?.growth_rate || '+0.0%' }}</div>
          <div class="text-[10px] font-bold text-green-500 uppercase tracking-widest mt-2">↑ {{ stats?.growth_increase || '0.0%' }} {{ t('from_last_week') }}</div>
        </div>
      </div>
      <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] relative overflow-hidden group">
        <div class="absolute top-0 right-0 p-8 opacity-5 group-hover:scale-110 transition-transform">
          <Users class="w-24 h-24 text-[var(--matrix-color)]" />
        </div>
        <div class="relative z-10">
          <div class="p-3 rounded-2xl bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] w-fit mb-4">
            <Users class="w-6 h-6" />
          </div>
          <div class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('total_invites') }}</div>
          <div class="text-3xl font-black text-[var(--text-main)] mt-1">{{ stats?.total_invites || 0 }}</div>
          <div class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest mt-2">{{ t('across_all_campaigns') }}</div>
        </div>
      </div>
      <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] relative overflow-hidden group">
        <div class="absolute top-0 right-0 p-8 opacity-5 group-hover:scale-110 transition-transform">
          <Activity class="w-24 h-24 text-[var(--matrix-color)]" />
        </div>
        <div class="relative z-10">
          <div class="p-3 rounded-2xl bg-purple-500/10 text-purple-500 w-fit mb-4">
            <Activity class="w-6 h-6" />
          </div>
          <div class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('active_tasks') }}</div>
          <div class="text-3xl font-black text-[var(--text-main)] mt-1">{{ tasks.filter(t => t.status === 'active').length }}</div>
          <div class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest mt-2">{{ t('currently_running') }}</div>
        </div>
      </div>
    </div>

    <!-- Content Sections -->
    <div v-if="activeTab === 'tasks'" class="space-y-6">
      <div class="flex items-center justify-between">
        <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">{{ t('active_campaigns') }}</h2>
        <button 
          @click="showCreateModal = true"
          class="flex items-center gap-2 px-4 py-2 rounded-2xl bg-[var(--matrix-color)] text-black font-black text-xs uppercase tracking-widest hover:opacity-90 transition-opacity"
        >
          <Plus class="w-4 h-4" />
          {{ t('new_campaign') }}
        </button>
      </div>

      <div v-if="loading" class="space-y-4 animate-pulse">
        <div v-for="i in 3" :key="i" class="h-24 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)]"></div>
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
                <Share2 class="w-6 h-6 text-[var(--matrix-color)]" />
              </div>
              <div>
                <h3 class="font-black text-[var(--text-main)] uppercase tracking-tight">{{ task.name }}</h3>
                <div class="flex items-center gap-4 mt-1">
                  <div class="flex items-center gap-1.5">
                    <Users class="w-3 h-3 text-[var(--text-muted)]" />
                    <span class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ task.participants || 0 }} {{ t('participants') }}</span>
                  </div>
                  <div class="flex items-center gap-1.5">
                    <Clock class="w-3 h-3 text-[var(--text-muted)]" />
                    <span class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('ends') }}: {{ task.end_date || t('ongoing') }}</span>
                  </div>
                </div>
              </div>
            </div>

            <div class="flex items-center gap-6">
              <div :class="getTaskStatusStyle(task.status)" class="px-3 py-1 rounded-full border text-[10px] font-black uppercase tracking-widest">
                {{ t(task.status) || task.status }}
              </div>
              <div class="flex items-center gap-2">
                <button class="p-2 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] text-[var(--text-muted)] hover:text-[var(--matrix-color)] transition-all">
                  <ChevronRight class="w-5 h-5" />
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

        <div v-if="tasks.length === 0" class="flex flex-col items-center justify-center py-20 bg-[var(--bg-card)] border border-[var(--border-color)] rounded-3xl">
          <Share2 class="w-16 h-16 text-[var(--text-muted)] mb-4 opacity-20" />
          <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">{{ t('no_active_campaigns') }}</h2>
          <p class="text-[var(--text-muted)] text-sm font-bold uppercase tracking-widest mt-2">{{ t('launch_fission_desc') }}</p>
        </div>
      </div>
    </div>

    <div v-else-if="activeTab === 'leaderboard'" class="space-y-6">
      <div class="flex items-center justify-between">
        <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">{{ t('top_inviters') }}</h2>
      </div>
      
      <div class="bg-[var(--bg-card)] border border-[var(--border-color)] rounded-3xl overflow-hidden">
        <table class="w-full">
          <thead>
            <tr class="border-b border-[var(--border-color)] bg-black/5 dark:bg-white/5">
              <th class="px-6 py-4 text-left text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('rank') }}</th>
              <th class="px-6 py-4 text-left text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('user') }}</th>
              <th class="px-6 py-4 text-right text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('invites') }}</th>
              <th class="px-6 py-4 text-right text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('points') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[var(--border-color)]">
            <tr v-for="(user, index) in leaderboard" :key="user.id" class="hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
              <td class="px-6 py-4">
                <div class="flex items-center justify-center w-8 h-8 rounded-full font-black text-xs" 
                     :class="index === 0 ? 'bg-yellow-500 text-black' : index === 1 ? 'bg-slate-300 text-black' : index === 2 ? 'bg-amber-600 text-black' : 'text-[var(--text-muted)]'">
                  {{ index + 1 }}
                </div>
              </td>
              <td class="px-6 py-4">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-full bg-[var(--matrix-color)]/20 flex items-center justify-center">
                    <Users class="w-4 h-4 text-[var(--matrix-color)]" />
                  </div>
                  <span class="font-bold text-[var(--text-main)]">{{ user.name }}</span>
                </div>
              </td>
              <td class="px-6 py-4 text-right font-black text-[var(--text-main)]">{{ user.invites }}</td>
              <td class="px-6 py-4 text-right text-[var(--matrix-color)] font-black">{{ user.points }}</td>
            </tr>
            <tr v-if="leaderboard.length === 0">
              <td colspan="4" class="px-6 py-20 text-center text-[var(--text-muted)] font-bold uppercase tracking-widest">{{ t('no_leaderboard_data') }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div v-else-if="activeTab === 'invitations'" class="space-y-6">
      <div class="flex items-center justify-between">
        <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">{{ t('recent_invitations') }}</h2>
      </div>
      
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div v-for="invite in invitations" :key="invite.id" class="p-4 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] flex items-center gap-4">
          <div class="p-3 rounded-2xl bg-green-500/10 text-green-500">
            <CheckCircle2 class="w-5 h-5" />
          </div>
          <div>
            <div class="font-bold text-[var(--text-main)]">{{ invite.inviter }} {{ t('invited') }} {{ invite.invitee }}</div>
            <div class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest mt-1">{{ invite.time }} • {{ invite.group }}</div>
          </div>
        </div>
        <div v-if="invitations.length === 0" class="col-span-full py-20 text-center text-[var(--text-muted)] font-bold uppercase tracking-widest bg-[var(--bg-card)] border border-[var(--border-color)] rounded-3xl">
          {{ t('no_invitations_recorded') }}
        </div>
      </div>
    </div>

    <div v-else-if="activeTab === 'config'" class="max-w-3xl mx-auto">
      <div class="bg-[var(--bg-card)] border border-[var(--border-color)] rounded-3xl p-8 space-y-8">
        <div class="flex items-center justify-between">
          <div>
            <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">{{ t('global_fission_config') }}</h2>
            <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest mt-1">{{ t('fission_config_desc') }}</p>
          </div>
          <button 
            @click="handleSaveConfig"
            class="flex items-center gap-2 px-6 py-3 rounded-2xl bg-[var(--matrix-color)] text-black font-black text-xs uppercase tracking-widest hover:opacity-90 transition-opacity shadow-lg shadow-[var(--matrix-color)]/20"
          >
            <Save class="w-4 h-4" />
            {{ t('save_changes') }}
          </button>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div class="space-y-2">
            <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ t('invite_keyword') }}</label>
            <input v-model="config.invite_keyword" type="text" class="w-full px-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] font-bold" />
          </div>
          <div class="space-y-2">
            <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ t('max_invites_per_day') }}</label>
            <input v-model.number="config.max_invites_per_day" type="number" class="w-full px-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] font-bold" />
          </div>
          <div class="md:col-span-2 space-y-2">
            <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ t('welcome_message') }}</label>
            <textarea v-model="config.welcome_msg" rows="3" class="w-full px-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] font-bold resize-none"></textarea>
          </div>
          <div class="md:col-span-2 space-y-2">
            <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ t('invite_template') }}</label>
            <textarea v-model="config.invite_msg" rows="3" class="w-full px-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] font-bold resize-none"></textarea>
          </div>
          <div class="flex items-center gap-3 ml-1">
            <div 
              @click="config.auto_approve = !config.auto_approve"
              :class="config.auto_approve ? 'bg-[var(--matrix-color)]' : 'bg-black/20 dark:bg-white/10'"
              class="w-12 h-6 rounded-full relative cursor-pointer transition-colors"
            >
              <div 
                :class="config.auto_approve ? 'translate-x-7' : 'translate-x-1'"
                class="absolute top-1 w-4 h-4 rounded-full bg-white shadow-sm transition-transform"
              ></div>
            </div>
            <span class="text-[10px] font-black text-[var(--text-main)] uppercase tracking-widest">{{ t('auto_approve_requests') }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Create Campaign Modal -->
    <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
      <div class="w-full max-w-md bg-[var(--bg-card)] border border-[var(--border-color)] rounded-[2rem] sm:rounded-[2.5rem] shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
        <div class="p-6 sm:p-8">
          <div class="flex items-center justify-between mb-8">
            <div>
              <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">{{ t('new_fission_campaign') }}</h2>
              <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest mt-1">{{ t('campaign_config_desc') }}</p>
            </div>
            <button @click="showCreateModal = false" class="p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
              <X class="w-6 h-6 text-[var(--text-muted)]" />
            </button>
          </div>

          <form @submit.prevent="handleCreateTask" class="space-y-6">
            <div class="space-y-2">
              <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ t('campaign_name') }}</label>
              <div class="relative">
                <Type class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
                <input 
                  v-model="newTask.name"
                  type="text"
                  :placeholder="t('campaign_name')"
                  class="w-full pl-12 pr-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold text-sm"
                  required
                />
              </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-2">
                <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ t('trigger_keyword') }}</label>
                <input 
                  v-model="newTask.keyword"
                  type="text"
                  placeholder="#invite"
                  class="w-full px-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold text-sm"
                  required
                />
              </div>
              <div class="space-y-2">
                <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ t('target_group') }}</label>
                <input 
                  v-model="newTask.target_group"
                  type="text"
                  :placeholder="t('target_group_id')"
                  class="w-full px-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold text-sm"
                />
              </div>
            </div>

            <div class="space-y-2">
              <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ t('end_date') }}</label>
              <div class="relative">
                <Calendar class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
                <input 
                  v-model="newTask.end_date"
                  type="date"
                  class="w-full pl-12 pr-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold text-sm"
                />
              </div>
            </div>

            <button type="submit" class="w-full py-4 rounded-2xl bg-[var(--matrix-color)] text-black font-black uppercase tracking-widest hover:opacity-90 transition-opacity mt-4 flex items-center justify-center gap-2">
              <Plus class="w-4 h-4" />
              {{ t('launch_campaign') }}
            </button>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.no-scrollbar::-webkit-scrollbar {
  display: none;
}
.no-scrollbar {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
</style>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useSystemStore } from '@/stores/system';
import { useBotStore } from '@/stores/bot';
import { useAuthStore } from '@/stores/auth';
import { 
  Users, 
  Plus, 
  Shield, 
  Mail, 
  Lock, 
  Trash2, 
  Edit2,
  X,
  CheckCircle2,
  XCircle,
  Search,
  Key,
  ArrowRight,
  User
} from 'lucide-vue-next';
import { 
  getPlatformIcon, 
  getPlatformColor, 
  isPlatformAvatar, 
  getPlatformFromAvatar 
} from '@/utils/avatar';

const systemStore = useSystemStore();
const botStore = useBotStore();
const authStore = useAuthStore();
const t = (key: string) => systemStore.t(key);

const users = ref<any[]>([]);
const identities = ref<any[]>([]);
const loading = ref(true);
const showModal = ref(false);
const showIdentityModal = ref(false);
const search = ref('');
const isEditing = ref(false);

const newUser = ref({
  id: '',
  username: '',
  password: '',
  qq: '',
  role: 'user',
  status: 'active'
});

const newIdentity = ref({
  NexusUID: '',
  Platform: 'qq',
  PlatformUID: '',
  Nickname: '',
  Metadata: '{}'
});

const fetchUsers = async () => {
  if (!authStore.isAdmin) {
    loading.value = false;
    return;
  }
  loading.value = true;
  try {
    const [usersData, identityData] = await Promise.all([
      botStore.fetchUsers(),
      botStore.fetchIdentities()
    ]);
    
    if (usersData.success) {
      users.value = usersData.data?.users || [];
    }
    if (identityData.success) {
      identities.value = identityData.data;
    }
  } finally {
    loading.value = false;
  }
};

const handleManageUser = async () => {
  try {
    const data = await botStore.manageUser(newUser.value);
    if (data.success) {
      await fetchUsers();
      showModal.value = false;
      resetForm();
    }
  } catch (err) {
    console.error('Failed to manage user:', err);
  }
};

const handleManageIdentity = async () => {
  try {
    const data = await botStore.saveIdentity(newIdentity.value);
    if (data.success) {
      await fetchUsers();
      showIdentityModal.value = false;
      resetIdentityForm();
    }
  } catch (err) {
    console.error('Failed to manage identity:', err);
  }
};

const handleEditIdentity = (identity: any) => {
  isEditing.value = true;
  newIdentity.value = { ...identity };
  showIdentityModal.value = true;
};

const handleDeleteIdentity = async (id: number) => {
  if (!confirm(t('confirm_delete_identity'))) return;
  try {
    const data = await botStore.deleteIdentity(id);
    if (data.success) {
      await fetchUsers();
    }
  } catch (err) {
    console.error('Failed to delete identity:', err);
  }
};

const resetIdentityForm = () => {
  isEditing.value = false;
  newIdentity.value = { NexusUID: '', Platform: 'qq', PlatformUID: '', Nickname: '', Metadata: '{}' };
};

const handleEditUser = (user: any) => {
  isEditing.value = true;
  newUser.value = { 
    ...user, 
    password: '',
    qq: user.qq || ''
  }; // Don't prefill password
  showModal.value = true;
};

const handleDeleteUser = async (userId: string) => {
  if (!confirm(t('confirm_delete_user'))) return;
  try {
    const data = await botStore.deleteUser(userId);
    if (data.success) {
      await fetchUsers();
    }
  } catch (err) {
    console.error('Failed to delete user:', err);
  }
};

const resetForm = () => {
  isEditing.value = false;
  newUser.value = { id: '', username: '', password: '', qq: '', role: 'user', status: 'active' };
};

const filteredUsers = () => {
  return users.value.filter(u => 
    u.username.toLowerCase().includes(search.value.toLowerCase()) ||
    u.role.toLowerCase().includes(search.value.toLowerCase())
  );
};

onMounted(fetchUsers);
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
        <h1 class="text-2xl font-black text-[var(--text-main)] tracking-tight">{{ t('users') }}</h1>
        <p class="text-sm font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('users_desc') }}</p>
      </div>
      <div class="flex items-center gap-3">
        <div class="relative">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
          <input 
            v-model="search"
            type="text"
            :placeholder="t('search_users_placeholder')"
            class="pl-10 pr-4 py-2 rounded-2xl bg-[var(--bg-card)] border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] w-64 transition-all"
          />
        </div>
        <button 
          @click="fetchUsers"
          :disabled="loading"
          class="p-2 rounded-2xl bg-[var(--bg-card)] border border-[var(--border-color)] hover:border-[var(--matrix-color)]/30 text-[var(--text-muted)] hover:text-[var(--matrix-color)] transition-all disabled:opacity-50"
        >
          <Key class="w-5 h-5" :class="{ 'animate-spin': loading }" />
        </button>
        <button 
          @click="resetForm(); showModal = true"
          class="flex items-center gap-2 px-4 py-2 rounded-2xl bg-[var(--matrix-color)] text-black font-black text-xs uppercase tracking-widest hover:opacity-90 transition-opacity"
        >
          <Plus class="w-4 h-4" />
          {{ t('add_new_user') }}
        </button>
      </div>
    </div>

    <!-- Stats Row -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)]">
        <div class="flex items-center gap-4">
          <div class="p-3 rounded-2xl bg-[var(--matrix-color)]/10 text-[var(--matrix-color)]">
            <Users class="w-6 h-6" />
          </div>
          <div>
            <div class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('total_users') }}</div>
            <div class="text-2xl font-black text-[var(--text-main)]">{{ users.length }}</div>
          </div>
        </div>
      </div>
      <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)]">
        <div class="flex items-center gap-4">
          <div class="p-3 rounded-2xl bg-[var(--matrix-color)]/10 text-[var(--matrix-color)]">
            <Shield class="w-6 h-6" />
          </div>
          <div>
            <div class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('admins') }}</div>
            <div class="text-2xl font-black text-[var(--text-main)]">{{ users.filter(u => u.role === 'admin').length }}</div>
          </div>
        </div>
      </div>
      <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)]">
        <div class="flex items-center gap-4">
          <div class="p-3 rounded-2xl bg-[var(--status-online)]/10 text-[var(--status-online)]">
            <CheckCircle2 class="w-6 h-6" />
          </div>
          <div>
            <div class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('active') }}</div>
            <div class="text-2xl font-black text-[var(--text-main)]">{{ users.filter(u => u.status === 'active').length }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- User List -->
    <div v-if="loading" class="space-y-4 animate-pulse">
      <div v-for="i in 4" :key="i" class="h-20 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)]"></div>
    </div>

    <div v-else class="space-y-4">
      <div 
        v-for="user in filteredUsers()" 
        :key="user.id"
        class="group p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] hover:border-[var(--matrix-color)]/30 transition-all duration-500"
      >
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-6">
            <div class="w-12 h-12 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] flex items-center justify-center">
              <Users class="w-6 h-6 text-[var(--matrix-color)]" />
            </div>
            <div>
              <div class="flex items-center gap-3">
                <h3 class="font-black text-[var(--text-main)] tracking-tight">{{ user.username }}</h3>
                <span 
                  :class="user.role === 'admin' ? 'bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] border-[var(--matrix-color)]/20' : 'bg-[var(--text-muted)]/10 text-[var(--text-muted)] border-[var(--text-muted)]/20'"
                  class="px-2 py-0.5 rounded-lg border text-[8px] font-black uppercase tracking-widest"
                >
                  {{ t(user.role) }}
                </span>
              </div>
              <div class="flex items-center gap-4 mt-1">
                <div class="flex items-center gap-1.5">
                  <span class="w-1.5 h-1.5 rounded-full" :class="user.status === 'active' ? 'bg-[var(--status-online)]' : 'bg-[var(--status-offline)]'"></span>
                  <span class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t(user.status) }}</span>
                </div>
                <div v-if="user.qq" class="flex items-center gap-1.5">
                  <span class="text-[10px] font-bold text-[var(--matrix-color)] uppercase tracking-widest">QQ: {{ user.qq }}</span>
                </div>
                <div class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('last_login') }}: {{ user.last_login || t('never') }}</div>
              </div>
            </div>
          </div>

          <div class="flex items-center gap-2">
            <button 
              @click="handleEditUser(user)"
              class="p-2 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] text-[var(--text-muted)] hover:border-[var(--matrix-color)]/30 hover:text-[var(--matrix-color)] transition-all"
            >
              <Edit2 class="w-4 h-4" />
            </button>
            <button 
              @click="handleDeleteUser(user.id)"
              class="p-2 rounded-xl bg-red-500/10 border border-red-500/20 text-red-500 hover:bg-red-500 hover:text-[var(--sidebar-text)] transition-all"
            >
              <Trash2 class="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-if="filteredUsers().length === 0" class="flex flex-col items-center justify-center py-20 bg-[var(--bg-card)] border border-[var(--border-color)] rounded-3xl">
        <Users class="w-16 h-16 text-[var(--text-muted)] mb-4 opacity-20" />
        <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">{{ t('no_users_found') }}</h2>
        <p class="text-[var(--text-muted)] text-sm font-bold uppercase tracking-widest mt-2">{{ t('search_adjust_desc') }}</p>
      </div>
    </div>

    <!-- Identity Mapping Section -->
    <div class="space-y-6 pt-12">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">{{ t('identity_mapping') }}</h2>
          <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest mt-1">{{ t('identity_mapping_desc') }}</p>
        </div>
        <button 
          @click="resetIdentityForm(); showIdentityModal = true"
          class="flex items-center gap-2 px-4 py-2 rounded-2xl bg-[var(--matrix-color)] text-[var(--sidebar-text-active)] font-black text-xs uppercase tracking-widest hover:opacity-90 transition-opacity"
        >
          <Plus class="w-4 h-4" />
          {{ t('map_identity') }}
        </button>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div 
          v-for="id in identities" 
          :key="id.ID"
          class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] flex items-center justify-between group hover:border-[var(--matrix-color)]/30 transition-all"
        >
          <div class="flex items-center gap-4">
            <div class="w-12 h-12 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] flex items-center justify-center overflow-hidden">
              <template v-if="id.Avatar && !isPlatformAvatar(id.Avatar)">
                <img :src="id.Avatar" class="w-full h-full object-cover" />
              </template>
              <template v-else>
                <component 
                  :is="isPlatformAvatar(id.Avatar) ? getPlatformIcon(getPlatformFromAvatar(id.Avatar)) : getPlatformIcon(id.Platform)" 
                  :class="['w-6 h-6', isPlatformAvatar(id.Avatar) ? getPlatformColor(getPlatformFromAvatar(id.Avatar)) : getPlatformColor(id.Platform)]" 
                />
              </template>
            </div>
            <div>
              <p class="font-black text-sm text-[var(--text-main)] uppercase tracking-tight">{{ id.Nickname || t('unknown') }}</p>
              <div class="flex items-center gap-3 mt-1">
                <span class="text-[10px] font-black px-2 py-0.5 rounded bg-black/5 dark:bg-white/5 text-[var(--text-muted)] uppercase tracking-widest">{{ t('platform_' + id.Platform.toLowerCase()) || id.Platform }}</span>
                <span class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('nexus_uid') }}: {{ id.NexusUID }}</span>
              </div>
            </div>
          </div>
          <div class="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
            <button @click="handleEditIdentity(id)" class="p-2 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] text-[var(--text-muted)] hover:text-[var(--matrix-color)] transition-all">
              <Edit2 class="w-4 h-4" />
            </button>
            <button @click="handleDeleteIdentity(id.ID)" class="p-2 rounded-xl bg-red-500/10 border border-red-500/20 text-red-500 hover:bg-red-500 hover:text-[var(--sidebar-text)] transition-all">
              <Trash2 class="w-4 h-4" />
            </button>
          </div>
        </div>
        <div v-if="identities.length === 0" class="lg:col-span-2 text-center py-12 bg-[var(--bg-card)] border border-[var(--border-color)] border-dashed rounded-3xl text-[var(--text-muted)] text-[10px] font-black uppercase tracking-widest">
          {{ t('no_identities_defined') }}
        </div>
      </div>
    </div>

    <!-- Add/Edit Identity Modal -->
    <div v-if="showIdentityModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
      <div class="w-full max-w-md bg-[var(--bg-card)] border border-[var(--border-color)] rounded-[2rem] sm:rounded-[2.5rem] shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
        <div class="p-6 sm:p-8">
          <div class="flex items-center justify-between mb-8">
            <div>
              <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">{{ isEditing ? t('edit_mapping') : t('map_new_identity') }}</h2>
              <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest mt-1">{{ t('connect_platform_id_desc') }}</p>
            </div>
            <button @click="showIdentityModal = false; resetIdentityForm()" class="p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
              <X class="w-6 h-6 text-[var(--text-muted)]" />
            </button>
          </div>

          <form @submit.prevent="handleManageIdentity" class="space-y-6">
            <div class="space-y-2">
              <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ t('nexus_uid') }}</label>
              <input v-model="newIdentity.NexusUID" type="text" class="w-full px-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold text-sm" required />
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-2">
                <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ t('platform') }}</label>
                <div class="relative">
                  <select v-model="newIdentity.Platform" class="w-full px-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold appearance-none text-sm">
                    <option value="qq">{{ t('platform_qq') }}</option>
                    <option value="wechat">{{ t('platform_wechat') }}</option>
                    <option value="telegram">{{ t('platform_telegram') }}</option>
                  </select>
                  <div class="absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none">
                    <ArrowRight class="w-4 h-4 text-[var(--text-muted)] rotate-90" />
                  </div>
                </div>
              </div>
              <div class="space-y-2">
                <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ t('platform_uid') }}</label>
                <input v-model="newIdentity.PlatformUID" type="text" class="w-full px-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold text-sm" required />
              </div>
            </div>
            <div class="space-y-2">
              <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ t('nickname') }}</label>
              <input v-model="newIdentity.Nickname" type="text" class="w-full px-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold text-sm" />
            </div>
            <button type="submit" class="w-full py-4 rounded-2xl bg-[var(--matrix-color)] text-black font-black uppercase tracking-widest hover:opacity-90 transition-opacity mt-4">
              {{ isEditing ? t('update_mapping') : t('create_mapping') }}
            </button>
          </form>
        </div>
      </div>
    </div>

    <!-- Add/Edit User Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
      <div class="w-full max-w-md bg-[var(--bg-card)] border border-[var(--border-color)] rounded-[2rem] sm:rounded-[2.5rem] shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
        <div class="p-6 sm:p-8">
          <div class="flex items-center justify-between mb-8">
            <div>
              <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">{{ isEditing ? t('edit_user') : t('add_new_user') }}</h2>
              <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest mt-1">{{ t('configure_system_access') }}</p>
            </div>
            <button @click="showModal = false; resetForm()" class="p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
              <X class="w-6 h-6 text-[var(--text-muted)]" />
            </button>
          </div>

          <form @submit.prevent="handleManageUser" class="space-y-6">
            <div class="space-y-2">
              <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ t('username') }}</label>
              <div class="relative">
                <Mail class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
                <input 
                  v-model="newUser.username"
                  type="text"
                  :placeholder="t('enter_username')"
                  class="w-full pl-12 pr-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold text-sm"
                  required
                  :disabled="isEditing"
                />
              </div>
            </div>

            <div class="space-y-2">
              <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ isEditing ? t('new_password_desc') : t('password') }}</label>
              <div class="relative">
                <Lock class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
                <input 
                  v-model="newUser.password"
                  type="password"
                  :placeholder="t('enter_password')"
                  class="w-full pl-12 pr-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold text-sm"
                  :required="!isEditing"
                />
              </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-2">
              <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">QQ</label>
              <div class="relative">
                <Users class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
                <input 
                  v-model="newUser.qq"
                  type="text"
                  placeholder="QQ 号码"
                  class="w-full pl-12 pr-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold text-sm"
                />
              </div>
            </div>

            <div class="space-y-2">
              <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ t('role') }}</label>
                <div class="relative">
                  <select 
                    v-model="newUser.role"
                    class="w-full px-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold appearance-none text-sm"
                  >
                    <option value="user">{{ t('user') }}</option>
                    <option value="admin">{{ t('admin') }}</option>
                  </select>
                  <div class="absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none">
                    <ArrowRight class="w-4 h-4 text-[var(--text-muted)] rotate-90" />
                  </div>
                </div>
              </div>
              <div class="space-y-2">
                <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ t('status') }}</label>
                <div class="relative">
                  <select 
                    v-model="newUser.status"
                    class="w-full px-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold appearance-none text-sm"
                  >
                    <option value="active">{{ t('active') }}</option>
                    <option value="inactive">{{ t('inactive') }}</option>
                  </select>
                  <div class="absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none">
                    <ArrowRight class="w-4 h-4 text-[var(--text-muted)] rotate-90" />
                  </div>
                </div>
              </div>
            </div>

            <button type="submit" class="w-full py-4 rounded-2xl bg-[var(--matrix-color)] text-black font-black uppercase tracking-widest hover:opacity-90 transition-opacity mt-4 flex items-center justify-center gap-2">
              <CheckCircle2 v-if="isEditing" class="w-4 h-4" />
              <Plus v-else class="w-4 h-4" />
              {{ isEditing ? t('update_user') : t('create_user') }}
            </button>
          </form>
        </div>
      </div>
    </div>
    </template>
  </div>
</template>

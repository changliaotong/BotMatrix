<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useSystemStore } from '@/stores/system';
import { useBotStore } from '@/stores/bot';
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
  Key
} from 'lucide-vue-next';

const systemStore = useSystemStore();
const botStore = useBotStore();
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
  loading.value = true;
  try {
    const [usersData, identityData] = await Promise.all([
      botStore.fetchUsers(),
      botStore.fetchIdentities()
    ]);
    
    if (usersData.success) {
      users.value = usersData.users;
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
  if (!confirm('Delete this identity mapping?')) return;
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
  newUser.value = { ...user, password: '' }; // Don't prefill password
  showModal.value = true;
};

const handleDeleteUser = async (userId: string) => {
  if (!confirm('Are you sure you want to delete this user?')) return;
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
  newUser.value = { id: '', username: '', password: '', role: 'user', status: 'active' };
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
            placeholder="Search users..."
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
          Add User
        </button>
      </div>
    </div>

    <!-- Stats Row -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)]">
        <div class="flex items-center gap-4">
          <div class="p-3 rounded-2xl bg-blue-500/10 text-blue-500">
            <Users class="w-6 h-6" />
          </div>
          <div>
            <div class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">Total Users</div>
            <div class="text-2xl font-black text-[var(--text-main)]">{{ users.length }}</div>
          </div>
        </div>
      </div>
      <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)]">
        <div class="flex items-center gap-4">
          <div class="p-3 rounded-2xl bg-purple-500/10 text-purple-500">
            <Shield class="w-6 h-6" />
          </div>
          <div>
            <div class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">Admins</div>
            <div class="text-2xl font-black text-[var(--text-main)]">{{ users.filter(u => u.role === 'admin').length }}</div>
          </div>
        </div>
      </div>
      <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)]">
        <div class="flex items-center gap-4">
          <div class="p-3 rounded-2xl bg-green-500/10 text-green-500">
            <CheckCircle2 class="w-6 h-6" />
          </div>
          <div>
            <div class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">Active</div>
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
                  :class="user.role === 'admin' ? 'bg-purple-500/10 text-purple-500 border-purple-500/20' : 'bg-blue-500/10 text-blue-500 border-blue-500/20'"
                  class="px-2 py-0.5 rounded-lg border text-[8px] font-black uppercase tracking-widest"
                >
                  {{ user.role }}
                </span>
              </div>
              <div class="flex items-center gap-4 mt-1">
                <div class="flex items-center gap-1.5">
                  <span class="w-1.5 h-1.5 rounded-full" :class="user.status === 'active' ? 'bg-green-500' : 'bg-red-500'"></span>
                  <span class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ user.status }}</span>
                </div>
                <div class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">Last Login: {{ user.last_login || 'Never' }}</div>
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
              class="p-2 rounded-xl bg-red-500/10 border border-red-500/20 text-red-500 hover:bg-red-500 hover:text-white transition-all"
            >
              <Trash2 class="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-if="filteredUsers().length === 0" class="flex flex-col items-center justify-center py-20 bg-[var(--bg-card)] border border-[var(--border-color)] rounded-3xl">
        <Users class="w-16 h-16 text-[var(--text-muted)] mb-4 opacity-20" />
        <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">No Users Found</h2>
        <p class="text-[var(--text-muted)] text-sm font-bold uppercase tracking-widest mt-2">Try adjusting your search criteria</p>
      </div>
    </div>

    <!-- Identity Mapping Section -->
    <div class="space-y-6 pt-12">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">Identity Mapping</h2>
          <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest mt-1">Unified Nexus IDs for multi-platform users</p>
        </div>
        <button 
          @click="resetIdentityForm(); showIdentityModal = true"
          class="flex items-center gap-2 px-4 py-2 rounded-2xl bg-blue-500 text-white font-black text-xs uppercase tracking-widest hover:opacity-90 transition-opacity"
        >
          <Plus class="w-4 h-4" />
          Map Identity
        </button>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div 
          v-for="id in identities" 
          :key="id.ID"
          class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] flex items-center justify-between group hover:border-blue-500/30 transition-all"
        >
          <div class="flex items-center gap-4">
            <div class="p-3 rounded-2xl bg-blue-500/10 text-blue-500">
              <Shield class="w-5 h-5" />
            </div>
            <div>
              <p class="font-black text-sm text-[var(--text-main)] uppercase tracking-tight">{{ id.Nickname || 'Unknown User' }}</p>
              <div class="flex items-center gap-3 mt-1">
                <span class="text-[10px] font-black px-2 py-0.5 rounded bg-black/5 dark:bg-white/5 text-[var(--text-muted)] uppercase tracking-widest">{{ id.Platform }}</span>
                <span class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">NexusUID: {{ id.NexusUID }}</span>
              </div>
            </div>
          </div>
          <div class="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
            <button @click="handleEditIdentity(id)" class="p-2 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] text-[var(--text-muted)] hover:text-blue-500 transition-all">
              <Edit2 class="w-4 h-4" />
            </button>
            <button @click="handleDeleteIdentity(id.ID)" class="p-2 rounded-xl bg-red-500/10 border border-red-500/20 text-red-500 hover:bg-red-500 hover:text-white transition-all">
              <Trash2 class="w-4 h-4" />
            </button>
          </div>
        </div>
        <div v-if="identities.length === 0" class="lg:col-span-2 text-center py-12 bg-[var(--bg-card)] border border-[var(--border-color)] border-dashed rounded-3xl text-[var(--text-muted)] text-[10px] font-black uppercase tracking-widest">
          No identity mappings defined
        </div>
      </div>
    </div>

    <!-- Add/Edit Identity Modal -->
    <div v-if="showIdentityModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showIdentityModal = false; resetIdentityForm()"></div>
      <div class="relative w-full max-w-md bg-[var(--bg-main)] border border-[var(--border-color)] rounded-3xl p-8 shadow-2xl">
        <div class="flex items-center justify-between mb-8">
          <div>
            <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">{{ isEditing ? 'Edit Mapping' : 'Map New Identity' }}</h2>
            <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest mt-1">Connect platform ID to Nexus ID</p>
          </div>
          <button @click="showIdentityModal = false; resetIdentityForm()" class="p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
            <X class="w-6 h-6 text-[var(--text-muted)]" />
          </button>
        </div>

        <form @submit.prevent="handleManageIdentity" class="space-y-6">
          <div class="space-y-2">
            <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">Nexus UID</label>
            <input v-model="newIdentity.NexusUID" type="text" class="w-full px-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-blue-500 outline-none text-[var(--text-main)] transition-all font-bold" required />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">Platform</label>
              <select v-model="newIdentity.Platform" class="w-full px-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-blue-500 outline-none text-[var(--text-main)] transition-all font-bold appearance-none">
                <option value="qq">QQ</option>
                <option value="wechat">WeChat</option>
                <option value="telegram">Telegram</option>
              </select>
            </div>
            <div class="space-y-2">
              <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">Platform UID</label>
              <input v-model="newIdentity.PlatformUID" type="text" class="w-full px-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-blue-500 outline-none text-[var(--text-main)] transition-all font-bold" required />
            </div>
          </div>
          <div class="space-y-2">
            <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">Nickname</label>
            <input v-model="newIdentity.Nickname" type="text" class="w-full px-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-blue-500 outline-none text-[var(--text-main)] transition-all font-bold" />
          </div>
          <button type="submit" class="w-full py-4 rounded-2xl bg-blue-500 text-white font-black uppercase tracking-widest hover:opacity-90 transition-opacity mt-4">
            {{ isEditing ? 'Update Mapping' : 'Create Mapping' }}
          </button>
        </form>
      </div>
    </div>

    <!-- Add/Edit User Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showModal = false; resetForm()"></div>
      <div class="relative w-full max-w-md bg-[var(--bg-main)] border border-[var(--border-color)] rounded-3xl p-8 shadow-2xl">
        <div class="flex items-center justify-between mb-8">
          <div>
            <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">{{ isEditing ? 'Edit User' : 'Add New User' }}</h2>
            <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest mt-1">Configure system access</p>
          </div>
          <button @click="showModal = false; resetForm()" class="p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
            <X class="w-6 h-6 text-[var(--text-muted)]" />
          </button>
        </div>

        <form @submit.prevent="handleManageUser" class="space-y-6">
          <div class="space-y-2">
            <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">Username</label>
            <div class="relative">
              <Mail class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
              <input 
                v-model="newUser.username"
                type="text"
                placeholder="Enter username"
                class="w-full pl-12 pr-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold"
                required
                :disabled="isEditing"
              />
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">{{ isEditing ? 'New Password (leave blank to keep current)' : 'Password' }}</label>
            <div class="relative">
              <Lock class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
              <input 
                v-model="newUser.password"
                type="password"
                placeholder="Enter password"
                class="w-full pl-12 pr-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold"
                :required="!isEditing"
              />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">Role</label>
              <select 
                v-model="newUser.role"
                class="w-full px-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold appearance-none"
              >
                <option value="user">User</option>
                <option value="admin">Admin</option>
              </select>
            </div>
            <div class="space-y-2">
              <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest ml-1">Status</label>
              <select 
                v-model="newUser.status"
                class="w-full px-4 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-[var(--text-main)] transition-all font-bold appearance-none"
              >
                <option value="active">Active</option>
                <option value="disabled">Disabled</option>
              </select>
            </div>
          </div>

          <button 
            type="submit"
            class="w-full py-4 rounded-2xl bg-[var(--matrix-color)] text-black font-black uppercase tracking-widest hover:opacity-90 transition-opacity mt-4 shadow-lg shadow-[var(--matrix-color)]/20"
          >
            {{ isEditing ? 'Save Changes' : 'Create User' }}
          </button>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { useSystemStore } from '@/stores/system';
import { useBotStore } from '@/stores/bot';
import { 
  Box, 
  Activity, 
  Play, 
  Square, 
  RotateCcw, 
  RefreshCw, 
  ShieldCheck, 
  AlertCircle,
  Search,
  Server,
  Cpu,
  User,
  Info,
  Upload,
  Trash2
} from 'lucide-vue-next';

const systemStore = useSystemStore();
const botStore = useBotStore();
const t = (key: string) => systemStore.t(key);

const plugins = ref<any[]>([]);
const workers = ref<any[]>([]);
const loading = ref(true);
const batchLoading = ref(false);
const actionLoading = ref<string | null>(null);
const error = ref<string | null>(null);
const fileInput = ref<HTMLInputElement | null>(null);
const installing = ref(false);
const installTarget = ref('nexus');

const triggerUpload = () => {
  fileInput.value?.click();
};

const handleUpload = async (event: Event) => {
  const target = (event.target as HTMLInputElement);
  if (!target.files?.length) return;

  const file = target.files[0];
  if (!file.name.endsWith('.bmpk')) {
    alert(t('upload_bmpk_only'));
    return;
  }

  installing.value = true;
  try {
    const res = await botStore.installPlugin(file, installTarget.value);
    if (res.success) {
      alert(t('plugin_install_success'));
      fetchPlugins();
    } else {
      alert(res.message || t('plugin_install_failed'));
    }
  } catch (err: any) {
    alert(err.message || t('plugin_install_failed'));
  } finally {
    installing.value = false;
    target.value = ''; // Reset input
  }
};

const handleDelete = async (plugin: any) => {
  if (!confirm(t('confirm_delete_plugin'))) {
    return;
  }

  const actionKey = `${plugin.source}-${plugin.id}-delete`;
  actionLoading.value = actionKey;
  try {
    const res = await botStore.deletePlugin(plugin.id, plugin.version, plugin.source);
    if (res.success) {
      // Refresh list
      setTimeout(fetchPlugins, 1000);
    } else {
      alert(res.message || t('plugin_delete_failed'));
    }
  } catch (err: any) {
    alert(err.message || t('plugin_delete_failed'));
  } finally {
    actionLoading.value = null;
  }
};

const handleBatchAction = async (action: 'start' | 'stop') => {
  if (batchLoading.value) return;
  
  const targets = filteredPlugins.value.filter(p => 
    !p.is_internal && p.online && (action === 'start' ? p.state !== 'running' : p.state === 'running')
  );

  if (targets.length === 0) {
    alert(`没有可${action === 'start' ? '启动' : '停止'}的插件`);
    return;
  }

  if (!confirm(`确定要${action === 'start' ? '批量启动' : '批量停止'} ${targets.length} 个插件吗？`)) {
    return;
  }

  batchLoading.value = true;
  let successCount = 0;
  let failCount = 0;

  try {
    // 串行或并行发送请求，这里选择并行
    const results = await Promise.all(
      targets.map(p => botStore.pluginAction(p.id, action, p.type, p.source))
    );

    results.forEach(res => {
      if (res.success) successCount++;
      else failCount++;
    });

    if (failCount > 0) {
      alert(`批量操作完成: ${successCount} 成功, ${failCount} 失败`);
    }
    
    // 刷新列表
    setTimeout(fetchPlugins, 1000);
  } catch (err) {
    console.error('Batch action failed:', err);
    alert('批量操作过程中发生错误');
  } finally {
    batchLoading.value = false;
  }
};
const searchQuery = ref('');
const filterType = ref<'all' | 'central' | 'worker'>('all');

let refreshTimer: number | null = null;

const filteredPlugins = computed(() => {
  let result = [...plugins.value];
  
  // Filter by search query
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase();
    result = result.filter(p => 
      (p.id && p.id.toLowerCase().includes(q)) || 
      (p.name && p.name.toLowerCase().includes(q)) ||
      (p.description && p.description.toLowerCase().includes(q)) ||
      (p.author && p.author.toLowerCase().includes(q))
    );
  }

  // Filter by type
  if (filterType.value !== 'all') {
    result = result.filter(p => p.type === filterType.value);
  }
  
  return result;
});

const fetchPlugins = async () => {
  error.value = null;
  loading.value = true;
  try {
    const [pluginsRes, workersRes] = await Promise.all([
      botStore.fetchPlugins(),
      botStore.fetchWorkers()
    ]);

    if (pluginsRes.success) {
      plugins.value = pluginsRes.data?.plugins || [];
    } else {
      error.value = pluginsRes.message || 'Failed to fetch plugins';
    }

    if (workersRes.success) {
      workers.value = workersRes.workers || [];
    }
  } catch (err: any) {
    console.error('Failed to fetch plugins:', err);
    error.value = err.message || 'An error occurred while fetching plugins';
  } finally {
    loading.value = false;
  }
};

const handleAction = async (plugin: any, action: string) => {
  const actionKey = `${plugin.source}-${plugin.id}-${action}`;
  actionLoading.value = actionKey;
  try {
    const res = await botStore.pluginAction(plugin.id, action, plugin.type, plugin.source);
    if (res.success) {
      // Refresh after short delay
      setTimeout(fetchPlugins, 1000);
    } else {
      alert(res.message || 'Operation failed');
    }
  } catch (err: any) {
    alert(err.message || 'Operation failed');
  } finally {
    actionLoading.value = null;
  }
};

onMounted(() => {
  fetchPlugins();
  refreshTimer = window.setInterval(fetchPlugins, 10000);
});

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer);
});

const getStatusColor = (status: string) => {
  switch (status?.toLowerCase()) {
    case 'running': return 'text-green-500 bg-green-500/10 border-green-500/20';
    case 'stopped': return 'text-gray-500 bg-gray-500/10 border-gray-500/20';
    case 'error': return 'text-red-500 bg-red-500/10 border-red-500/20';
    default: return 'text-yellow-500 bg-yellow-500/10 border-yellow-500/20';
  }
};

const getTypeLabel = (type: string) => {
  return type === 'central' ? 'Nexus' : 'Worker';
};

const getTypeColor = (type: string) => {
  return type === 'central' ? 'text-purple-500 bg-purple-500/10 border-purple-500/20' : 'text-blue-500 bg-blue-500/10 border-blue-500/20';
};

</script>

<template>
  <div class="p-6 space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-black text-[var(--text-main)] tracking-tight">插件管理</h1>
        <p class="text-sm font-bold text-[var(--text-muted)] uppercase tracking-widest">管理 Nexus 中心端和所有 Worker 节点的插件</p>
      </div>
      
      <div class="flex flex-col sm:flex-row items-center gap-4">
        <div class="relative w-full sm:w-64 group">
          <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)] group-focus-within:text-[var(--matrix-color)] transition-colors" />
          <input 
            v-model="searchQuery"
            type="text" 
            placeholder="搜索插件..."
            class="w-full pl-11 pr-4 py-3 bg-[var(--bg-card)] border border-[var(--border-color)] rounded-2xl text-xs font-bold text-[var(--text-main)] focus:outline-none focus:border-[var(--matrix-color)]/50 transition-all"
          >
        </div>

        <div class="flex items-center gap-2 bg-[var(--bg-card)] border border-[var(--border-color)] p-1 rounded-2xl">
          <button 
            v-for="type in ['all', 'central', 'worker']" 
            :key="type"
            @click="filterType = type as any"
            :class="['px-3 py-2 rounded-xl text-[10px] font-black uppercase tracking-widest transition-all', 
              filterType === type ? 'bg-[var(--matrix-color)] text-black' : 'text-[var(--text-muted)] hover:bg-black/5 dark:hover:bg-white/5']"
          >
            {{ type === 'all' ? '全部' : (type === 'central' ? 'Nexus' : 'Worker') }}
          </button>
        </div>

        <button 
          @click="fetchPlugins"
          class="p-3 rounded-2xl bg-[var(--bg-card)] border border-[var(--border-color)] hover:border-[var(--matrix-color)]/30 transition-all group"
          title="刷新列表"
        >
          <RefreshCw class="w-5 h-5 text-[var(--text-muted)] group-hover:text-[var(--matrix-color)]" :class="{ 'animate-spin': loading }" />
        </button>

        <div class="h-8 w-[1px] bg-[var(--border-color)] mx-2 hidden sm:block"></div>

        <input 
          ref="fileInput"
          type="file"
          accept=".bmpk"
          class="hidden"
          @change="handleUpload"
        >

        <div class="flex items-center gap-2">
          <select 
            v-model="installTarget"
            class="px-4 py-3 bg-[var(--bg-card)] border border-[var(--border-color)] rounded-2xl text-xs font-bold text-[var(--text-main)] focus:outline-none focus:border-[var(--matrix-color)]/50 transition-all"
          >
            <option value="nexus">Nexus (Central)</option>
            <option v-for="worker in workers" :key="worker.id" :value="worker.id">
              Worker: {{ worker.name || worker.id }}
            </option>
          </select>

          <button 
            @click="triggerUpload"
            :disabled="installing"
            class="px-4 py-3 rounded-2xl bg-[var(--matrix-color)] text-black text-xs font-black uppercase tracking-widest transition-all hover:shadow-lg hover:shadow-[var(--matrix-color)]/20 disabled:opacity-50 flex items-center gap-2"
          >
            <Upload class="w-4 h-4" :class="{ 'animate-bounce': installing }" />
            {{ t('install_plugin') }}
          </button>
        </div>

        <div class="h-8 w-[1px] bg-[var(--border-color)] mx-2 hidden sm:block"></div>

        <div class="flex items-center gap-2">
          <button 
            @click="handleBatchAction('start')"
            :disabled="batchLoading"
            class="px-4 py-3 rounded-2xl bg-green-500/10 hover:bg-green-500/20 text-green-500 text-xs font-black uppercase tracking-widest transition-all disabled:opacity-50 flex items-center gap-2"
          >
            <Play class="w-4 h-4" :class="{ 'animate-pulse': batchLoading }" />
            全部启动
          </button>
          <button 
            @click="handleBatchAction('stop')"
            :disabled="batchLoading"
            class="px-4 py-3 rounded-2xl bg-red-500/10 hover:bg-red-500/20 text-red-500 text-xs font-black uppercase tracking-widest transition-all disabled:opacity-50 flex items-center gap-2"
          >
            <Square class="w-4 h-4" :class="{ 'animate-pulse': batchLoading }" />
            全部停止
          </button>
        </div>
      </div>
    </div>

    <!-- Error State -->
    <div v-if="error" class="p-4 rounded-2xl bg-red-500/10 border border-red-500/20 flex items-center gap-3 text-red-500">
      <AlertCircle class="w-5 h-5" />
      <div class="flex-1">
        <p class="text-xs font-bold uppercase tracking-widest">错误</p>
        <p class="text-sm font-black">{{ error }}</p>
      </div>
      <button @click="fetchPlugins" class="px-3 py-1 rounded-lg bg-red-500 text-white text-[10px] font-black uppercase tracking-widest">
        重试
      </button>
    </div>

    <!-- Plugins Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div 
        v-for="plugin in filteredPlugins" 
        :key="plugin.source + plugin.id"
        class="group p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] hover:border-[var(--matrix-color)]/30 transition-all duration-500 relative overflow-hidden flex flex-col"
      >
        <!-- Top Info -->
        <div class="flex items-start justify-between mb-4">
          <div class="flex items-center gap-4">
            <div class="p-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
              <Box class="w-6 h-6 text-[var(--matrix-color)]" />
            </div>
            <div>
              <h3 class="font-black text-[var(--text-main)]">{{ plugin.name || plugin.id }}</h3>
              <div class="flex items-center gap-2 mt-1">
                <span class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">v{{ plugin.version }}</span>
                <span :class="getTypeColor(plugin.type)" class="px-2 py-0.5 rounded-md border text-[8px] font-black uppercase tracking-widest">
                  {{ getTypeLabel(plugin.type) }}
                </span>
              </div>
            </div>
          </div>
          <div class="flex flex-col items-end gap-1">
            <div :class="getStatusColor(plugin.state)" class="px-3 py-1 rounded-full border text-[10px] font-black uppercase tracking-widest">
              {{ plugin.state }}
            </div>
            <div v-if="!plugin.online" class="px-2 py-0.5 rounded-md bg-red-500/10 border border-red-500/20 text-red-500 text-[8px] font-black uppercase tracking-widest">
              节点离线
            </div>
          </div>
        </div>

        <!-- Description -->
        <p class="text-xs text-[var(--text-muted)] mb-6 line-clamp-2 flex-1">
          {{ plugin.description || '暂无描述' }}
        </p>

        <!-- Metadata -->
        <div class="grid grid-cols-2 gap-3 mb-6">
          <div class="flex items-center gap-2 text-[var(--text-muted)]">
            <User class="w-3 h-3" />
            <span class="text-[10px] font-bold truncate">{{ plugin.author || 'system' }}</span>
          </div>
          <div class="flex items-center gap-2 text-[var(--text-muted)]">
            <Server class="w-3 h-3" />
            <span class="text-[10px] font-bold truncate">{{ plugin.source === 'nexus' ? 'Nexus' : plugin.source }}</span>
          </div>
        </div>

        <!-- Actions -->
        <div v-if="!plugin.is_internal" class="flex items-center gap-2 mt-auto">
          <button 
            v-if="plugin.state !== 'running'"
            @click="handleAction(plugin, 'start')"
            :disabled="!plugin.online || actionLoading === `${plugin.source}-${plugin.id}-start`"
            class="flex-1 flex items-center justify-center gap-2 py-2 rounded-xl bg-green-500/10 hover:bg-green-500/20 text-green-500 text-[10px] font-black uppercase tracking-widest transition-all disabled:opacity-30 disabled:cursor-not-allowed"
          >
            <Play class="w-3 h-3" :class="{ 'animate-pulse': actionLoading === `${plugin.source}-${plugin.id}-start` }" />
            启动
          </button>
          <button 
            v-if="plugin.state === 'running'"
            @click="handleAction(plugin, 'stop')"
            :disabled="!plugin.online || actionLoading === `${plugin.source}-${plugin.id}-stop`"
            class="flex-1 flex items-center justify-center gap-2 py-2 rounded-xl bg-red-500/10 hover:bg-red-500/20 text-red-500 text-[10px] font-black uppercase tracking-widest transition-all disabled:opacity-30 disabled:cursor-not-allowed"
          >
            <Square class="w-3 h-3" :class="{ 'animate-pulse': actionLoading === `${plugin.source}-${plugin.id}-stop` }" />
            停止
          </button>
          <button 
            @click="handleAction(plugin, 'restart')"
            :disabled="!plugin.online || actionLoading === `${plugin.source}-${plugin.id}-restart`"
            class="p-2 rounded-xl bg-blue-500/10 hover:bg-blue-500/20 text-blue-500 transition-all disabled:opacity-30 disabled:cursor-not-allowed"
            :title="plugin.online ? '重启' : '节点离线无法重启'"
          >
            <RotateCcw class="w-4 h-4" :class="{ 'animate-spin': actionLoading === `${plugin.source}-${plugin.id}-restart` }" />
          </button>
          <button 
            @click="handleAction(plugin, 'reload')"
            :disabled="!plugin.online || actionLoading === `${plugin.source}-${plugin.id}-reload`"
            class="p-2 rounded-xl bg-yellow-500/10 hover:bg-yellow-500/20 text-yellow-500 transition-all disabled:opacity-30 disabled:cursor-not-allowed"
            :title="plugin.online ? '重载配置' : '节点离线无法重载'"
          >
            <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': actionLoading === `${plugin.source}-${plugin.id}-reload` }" />
          </button>
          <button 
            @click="handleDelete(plugin)"
            :disabled="!plugin.online || actionLoading === `${plugin.source}-${plugin.id}-delete`"
            class="p-2 rounded-xl bg-red-500/10 hover:bg-red-500/20 text-red-500 transition-all disabled:opacity-30 disabled:cursor-not-allowed"
            :title="t('delete_plugin')"
          >
            <Trash2 class="w-4 h-4" :class="{ 'animate-pulse': actionLoading === `${plugin.source}-${plugin.id}-delete` }" />
          </button>
        </div>

        <div v-else class="flex items-center justify-center gap-2 mt-auto py-2 rounded-xl bg-gray-500/5 border border-dashed border-gray-500/20">
          <ShieldCheck class="w-3 h-3 text-gray-400" />
          <span class="text-[10px] font-black text-gray-400 uppercase tracking-widest">核心组件 (只读)</span>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-if="!loading && filteredPlugins.length === 0" class="flex flex-col items-center justify-center py-20 text-[var(--text-muted)]">
      <Box class="w-16 h-16 mb-4 opacity-20" />
      <p class="text-sm font-bold uppercase tracking-widest">未找到插件</p>
    </div>
  </div>
</template>

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;  
  overflow: hidden;
}
</style>

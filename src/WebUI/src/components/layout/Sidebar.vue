<script setup lang="ts">
import { useSystemStore } from '@/stores/system';
import { useAuthStore } from '@/stores/auth';
import { useRoute, useRouter } from 'vue-router';
import { 
  Cpu, 
  PanelLeftClose, 
  LayoutDashboard, 
  Bot, 
  Network, 
  Terminal, 
  Settings,
  User,
  Users,
  ListTodo,
  Share2,
  Box,
  Route,
  UserCog,
  BookOpen
} from 'lucide-vue-next';

const systemStore = useSystemStore();
const authStore = useAuthStore();
const route = useRoute();
const router = useRouter();

const checkMobile = () => window.innerWidth < 1024;

const iconMap: Record<string, any> = {
  LayoutDashboard,
  Bot,
  Cpu,
  Network,
  Terminal,
  Settings,
  Users,
  ListTodo,
  Share2,
  Box,
  Route,
  UserCog,
  BookOpen
};

// Map item IDs to routes
const routeMap: Record<string, string> = {
  'dashboard': '/',
  'bots': '/bots',
  'workers': '/workers',
  'contacts': '/contacts',
  'nexus': '/nexus',
  'tasks': '/tasks',
  'fission': '/fission',
  'docker': '/docker',
  'routing': '/routing',
  'users': '/users',
  'logs': '/logs',
  'settings': '/settings',
  'manual': '/manual'
};

const navigateTo = (itemId: string) => {
  const path = routeMap[itemId];
  if (path) {
    router.push(path);
    if (checkMobile()) systemStore.showMobileMenu = false;
  }
};

const isItemActive = (itemId: string) => {
  return routeMap[itemId] === route.path;
};

// Translation function
const t = (key: string) => systemStore.t(key);
</script>

<template>
  <!-- Sidebar Overlay (Mobile) -->
  <div v-if="systemStore.showMobileMenu" 
       @click="systemStore.toggleMobileMenu()" 
       class="fixed inset-0 bg-black/50 backdrop-blur-sm z-40 lg:hidden"></div>

  <!-- Sidebar -->
  <aside class="fixed lg:static inset-y-0 left-0 flex-shrink-0 border-r border-[var(--border-color)] bg-[var(--bg-sidebar)] flex flex-col z-50 transition-all duration-300 overflow-hidden"
     :class="[
         systemStore.isSidebarCollapsed ? 'w-64 lg:w-20' : 'w-64 lg:w-64',
         systemStore.showMobileMenu ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'
     ]">
  
    <!-- Logo Area -->
    <div class="h-16 flex items-center border-b border-[var(--border-color)] overflow-hidden transition-all duration-300"
         :class="systemStore.isSidebarCollapsed ? 'justify-center px-0' : 'justify-between px-6'">
        <div class="flex items-center gap-3">
            <div class="w-8 h-8 rounded-lg bg-[var(--matrix-color)] flex items-center justify-center shadow-lg shadow-[var(--matrix-color)]/20 flex-shrink-0">
                <Cpu class="w-5 h-5 text-black" />
            </div>
            <span v-show="!systemStore.isSidebarCollapsed" class="font-bold tracking-tight text-[var(--text-main)] whitespace-nowrap">BotMatrix</span>
        </div>
        <!-- Collapse Toggle (Desktop) -->
        <button v-show="!systemStore.isSidebarCollapsed"
                @click="systemStore.toggleSidebar()" 
                class="hidden lg:flex p-1.5 rounded-lg hover:bg-black/5 dark:hover:bg-white/5 text-[var(--text-muted)] transition-colors flex-shrink-0">
            <PanelLeftClose class="w-4 h-4" />
        </button>
        <!-- Close Toggle (Mobile) -->
        <button @click="systemStore.showMobileMenu = false" 
                class="lg:hidden p-2 text-[var(--text-muted)] hover:text-[var(--matrix-color)]">
            <PanelLeftClose class="w-5 h-5" />
        </button>
    </div>

    <!-- Navigation -->
    <nav class="flex-1 overflow-y-auto py-6 px-3 space-y-1 custom-scrollbar">
        <template v-for="(group, idx) in systemStore.menuGroups" :key="group?.id || idx">
            <div v-if="group">
                <div v-show="!systemStore.isSidebarCollapsed" class="px-3 mb-2 text-xs font-bold uppercase tracking-[0.2em] text-[var(--text-muted)]">
                    {{ t(group.titleKey) }}
                </div>
                <div class="space-y-1">
                    <template v-for="(item, itemIdx) in (group?.items || [])" :key="item?.id || itemIdx">
                        <button v-if="item" 
                                @click="navigateTo(item.id)"
                                :class="[
                                    'w-full flex items-center transition-all duration-200 group relative border border-transparent',
                                    systemStore.isSidebarCollapsed ? 'justify-center px-0 py-2.5 rounded-lg' : 'gap-3 px-3 py-2.5 rounded-xl',
                                    isItemActive(item.id) 
                                        ? 'bg-[var(--matrix-color)]/10 border-[var(--matrix-color)]/20 text-[var(--matrix-color)] shadow-sm' 
                                        : 'text-[var(--text-muted)] hover:bg-[var(--matrix-color)]/10 hover:text-[var(--matrix-color)]'
                                ]">
                            <component :is="iconMap[item.icon]" 
                               :class="[
                                   'w-5 h-5 flex-shrink-0 transition-colors',
                                   isItemActive(item.id) ? 'text-[var(--matrix-color)]' : 'text-[var(--text-muted)] group-hover:text-[var(--matrix-color)]'
                               ]" />
                            <span v-show="!systemStore.isSidebarCollapsed" 
                                  :class="[
                                      'text-sm font-medium transition-colors',
                                      isItemActive(item.id) ? 'text-[var(--matrix-color)]' : 'text-[var(--text-muted)] group-hover:text-[var(--matrix-color)]'
                                  ]">{{ t(item.titleKey) }}</span>
                            
                            <!-- Active Indicator -->
                            <div v-if="isItemActive(item.id)" 
                                 class="absolute left-0 w-1 h-6 bg-[var(--matrix-color)] rounded-r-full"></div>
                        </button>
                    </template>
                </div>
                <div class="h-6"></div>
            </div>
        </template>
    </nav>

    <div class="p-4 border-t border-[var(--border-color)]">
        <div class="flex items-center p-2 rounded-xl bg-black/5 dark:bg-white/5 transition-all duration-300"
             :class="systemStore.isSidebarCollapsed ? 'justify-center' : 'gap-3'">
            <div class="w-8 h-8 rounded-full bg-[var(--matrix-color)]/20 flex items-center justify-center text-[var(--matrix-color)] overflow-hidden flex-shrink-0">
                <img v-if="authStore.user?.avatar" :src="authStore.user.avatar" class="w-full h-full object-cover">
                <User v-else class="w-4 h-4" />
            </div>
            <div v-show="!systemStore.isSidebarCollapsed" class="flex-1 min-w-0">
                <div class="text-sm font-semibold truncate text-[var(--text-main)]">{{ authStore.user?.username || 'ADMIN' }}</div>
                <div class="text-xs text-[var(--text-muted)] truncate uppercase tracking-wider">{{ t(authStore.user?.role === 'admin' ? 'admin' : 'user') }}</div>
            </div>
        </div>
    </div>
  </aside>
</template>

<style scoped>
.bg-matrix {
  background-color: var(--matrix-color);
}
.text-matrix {
  color: var(--matrix-color);
}
.bg-matrix\/10 {
  background-color: rgba(0, 255, 65, 0.1);
}
.bg-matrix\/20 {
  background-color: rgba(0, 255, 65, 0.2);
}
.border-matrix\/20 {
  border-color: rgba(0, 255, 65, 0.2);
}
.shadow-matrix\/20 {
  shadow-color: rgba(0, 255, 65, 0.2);
}
</style>

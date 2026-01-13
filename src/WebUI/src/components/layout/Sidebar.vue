<script setup lang="ts">
import { ref, computed } from 'vue';
import { useSystemStore } from '@/stores/system';
import { useAuthStore } from '@/stores/auth';
import { useRoute, useRouter } from 'vue-router';
import { 
  Cpu, 
  PanelLeftClose, 
  PanelLeft,
  LayoutDashboard, 
  Bot, 
  MessageSquare,
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
  BookOpen,
  Activity,
  Sparkles,
  Wrench,
  Settings2
} from 'lucide-vue-next';

const systemStore = useSystemStore();
const authStore = useAuthStore();
const route = useRoute();
const router = useRouter();

const checkMobile = () => window.innerWidth < 1024;

const iconMap: Record<string, any> = {
  LayoutDashboard,
  Bot,
  MessageSquare,
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
  BookOpen,
  Activity,
  Sparkles,
  Wrench,
  Settings2
};

// Map item IDs to routes
const routeMap: Record<string, string> = {
  'dashboard': '/console',
  'bots': '/console/bots',
  'bot-setup': '/setup/bot',
  'group-setup': '/setup/group',
  'contacts': '/console/contacts',
  'messages': '/console/messages',
  'tasks': '/console/tasks',
  'fission': '/console/fission',
  'settings': '/console/settings',
  'manual': '/console/manual',
  'workers': '/admin/workers',
  'users': '/admin/users',
  'logs': '/admin/logs',
  'monitor': '/admin/monitor',
  'nexus': '/admin/nexus',
  'ai': '/admin/ai',
  'routing': '/admin/routing',
  'docker': '/admin/docker',
  'plugins': '/admin/plugins',
};

const isNavigating = ref(false);
let navTimeout: any = null;

const navigateTo = async (itemId: string) => {
  if (isNavigating.value) return;
  
  const path = routeMap[itemId];
  if (path) {
    if (path === route.path) return;
    
    isNavigating.value = true;
    
    // Safety timeout to prevent deadlock
    if (navTimeout) clearTimeout(navTimeout);
    navTimeout = setTimeout(() => {
      isNavigating.value = false;
    }, 2000);

    try {
      await router.push(path);
      if (checkMobile()) systemStore.showMobileMenu = false;
    } catch (err: any) {
      console.error('Navigation failed:', err);
      // Fallback for critical failures
      if (err.name !== 'NavigationDuplicated') {
        window.location.href = path;
      }
    } finally {
      setTimeout(() => {
        isNavigating.value = false;
        if (navTimeout) {
          clearTimeout(navTimeout);
          navTimeout = null;
        }
      }, 300);
    }
  }
};

const isItemActive = (itemId: string) => {
  return routeMap[itemId] === route.path;
};

// Translation function
const t = (key: string) => systemStore.t(key);

// Filtered menu groups based on user role
const filteredMenuGroups = computed(() => {
    return systemStore.rawMenuGroups.filter(group => {
        // If group is admin only and user is not admin, hide group
        if (group.adminOnly && !authStore.isAdmin) return false;
        
        // Filter items within group
        const visibleItems = group.items.filter(item => {
            if (item.adminOnly && !authStore.isAdmin) return false;
            return true;
        });
        
        // If no items left in group, hide group (unless it's a special group)
        if (visibleItems.length === 0) return false;
        
        // Return group with only visible items
        return {
            ...group,
            items: visibleItems
        };
    });
});
</script>

<template>
  <!-- Sidebar Overlay (Mobile) -->
  <div v-if="systemStore.showMobileMenu" 
       @click="systemStore.toggleMobileMenu()" 
       class="fixed inset-0 bg-black/50 backdrop-blur-sm z-40 lg:hidden"></div>

  <!-- Sidebar -->
  <aside class="fixed md:static inset-y-0 left-0 flex-shrink-0 border-r border-[var(--border-color)] bg-[var(--bg-sidebar)] flex flex-col z-50 transition-all duration-300 overflow-hidden"
     :class="[
         systemStore.isSidebarCollapsed ? 'w-64 md:w-20' : 'w-64 md:w-64',
         systemStore.showMobileMenu ? 'translate-x-0' : '-translate-x-full md:translate-x-0'
     ]">
  
    <!-- Logo Area -->
    <div class="h-16 flex items-center border-b border-[var(--border-color)] overflow-hidden transition-all duration-300"
         :class="systemStore.isSidebarCollapsed ? 'justify-center px-0' : 'justify-between px-6'">
        <div class="flex items-center gap-3">
            <div class="w-8 h-8 rounded-lg bg-[var(--matrix-color)] flex items-center justify-center shadow-lg shadow-[var(--matrix-color)]/20 flex-shrink-0">
                <Cpu class="w-5 h-5 text-black" />
            </div>
            <span v-show="!systemStore.isSidebarCollapsed" class="font-bold tracking-tight text-[var(--sidebar-text)] whitespace-nowrap">{{ t('botmatrix') }}</span>
        </div>
        <!-- Collapse Toggle (Desktop) -->
        <button @click="systemStore.toggleSidebar()" 
                class="hidden md:flex p-1.5 rounded-lg hover:bg-black/5 dark:hover:bg-white/5 text-[var(--sidebar-text-muted)] transition-colors flex-shrink-0">
            <PanelLeftClose v-if="!systemStore.isSidebarCollapsed" class="w-4 h-4" />
            <PanelLeft v-else class="w-4 h-4" />
        </button>
        <!-- Close Toggle (Mobile) -->
        <button @click="systemStore.showMobileMenu = false" 
                class="md:hidden p-2 text-[var(--sidebar-text-muted)] hover:text-[var(--sidebar-text)]">
            <PanelLeftClose class="w-5 h-5" />
        </button>
    </div>

    <!-- Navigation -->
    <nav class="flex-1 overflow-y-auto py-2 px-3 space-y-2 custom-scrollbar">
        <template v-for="(group, idx) in filteredMenuGroups" :key="group?.id || idx">
            <div v-if="group">
                <div v-show="!systemStore.isSidebarCollapsed" class="px-3 mt-4 mb-1 text-[9px] font-black uppercase tracking-[0.2em] text-[var(--sidebar-text-muted)] opacity-80">
                    {{ t(group.titleKey) }}
                </div>
                <div class="space-y-1">
                    <template v-for="(item, itemIdx) in (group?.items || [])" :key="item?.id || itemIdx">
                        <button v-if="item" 
        @click="navigateTo(item.id)"
        :class="[
            'w-full flex items-center transition-all duration-200 group relative border border-transparent',
            systemStore.isSidebarCollapsed ? 'justify-center px-0 py-2.5 rounded-lg' : 'gap-3 px-3 py-2 rounded-lg',
            isItemActive(item.id) 
                ? 'bg-[var(--matrix-color)] border-[var(--matrix-color)] text-[var(--sidebar-text-active)] shadow-lg active scale-[1.02]' 
                : 'text-[var(--sidebar-text-muted)] hover:bg-[var(--matrix-color)]/10 hover:text-[var(--sidebar-text)]'
        ]">
                            <component :is="iconMap[item.icon]" 
                               :class="[
                                   'w-5 h-5 flex-shrink-0 transition-colors',
                                   isItemActive(item.id) ? 'text-[var(--sidebar-text-active)]' : 'text-[var(--sidebar-text-muted)] group-hover:text-[var(--sidebar-text)]'
                               ]" />
                            <span v-show="!systemStore.isSidebarCollapsed" 
                                  :class="[
                                      'text-sm font-medium transition-colors',
                                      isItemActive(item.id) ? 'text-[var(--sidebar-text-active)]' : 'text-[var(--sidebar-text-muted)] group-hover:text-[var(--sidebar-text)]'
                                  ]">{{ t(item.titleKey) }}</span>
                            
                            <!-- Active Indicator -->
                            <div v-if="isItemActive(item.id)" 
                                 class="absolute left-0 w-1 h-5 bg-[var(--matrix-color)] rounded-r-full"></div>
                        </button>
                    </template>
                </div>
                <div class="h-0.5"></div>
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
                <div class="text-sm font-semibold truncate text-[var(--sidebar-text)]">{{ authStore.user?.username || 'ADMIN' }}</div>
                <div class="text-xs text-[var(--sidebar-text-muted)] truncate uppercase tracking-wider">{{ t(authStore.user?.role === 'admin' ? 'admin' : 'user') }}</div>
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
  background-color: color-mix(in srgb, var(--matrix-color) 10%, transparent);
}
.bg-matrix\/20 {
  background-color: color-mix(in srgb, var(--matrix-color) 20%, transparent);
}
.border-matrix\/20 {
  border-color: color-mix(in srgb, var(--matrix-color) 20%, transparent);
}
.shadow-matrix\/20 {
  shadow-color: color-mix(in srgb, var(--matrix-color) 20%, transparent);
}
</style>

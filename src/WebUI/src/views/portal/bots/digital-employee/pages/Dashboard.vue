<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { 
  Users, 
  Activity, 
  Zap, 
  Target, 
  Cpu, 
  Bot, 
  ShieldCheck, 
  Terminal,
  TrendingUp,
  Clock,
  Briefcase,
  AlertCircle,
  Radio,
  Settings2
} from 'lucide-vue-next'
import { useI18n } from '@/utils/i18n'

const { t: tt } = useI18n()

// Real-time telemetry simulation
const stats = ref({
  activeAgents: 128,
  tasksPerHour: 450,
  efficiency: 98.4,
  systemUptime: '99.99%'
})

const agents = ref([
  { id: 'DE-101', name: 'Market Analyst Alpha', status: 'WORKING', load: 85, task: 'Sentiment Analysis' },
  { id: 'DE-204', name: 'Customer Support Beta', status: 'IDLE', load: 0, task: 'Waiting for Queue' },
  { id: 'DE-052', name: 'Lead Gen Gamma', status: 'WORKING', load: 92, task: 'Social Outreach' },
  { id: 'DE-318', name: 'Data Sync Delta', status: 'SYNCING', load: 45, task: 'Postgres Replication' }
])

const activityLog = ref([
  { time: '10:24:05', type: 'success', msg: 'DE-101 completed quarterly report generation.' },
  { time: '10:23:42', type: 'info', msg: 'New Digital Employee DE-405 initialized in HR cluster.' },
  { time: '10:22:15', type: 'warning', msg: 'System load spike detected in Tokyo node (88%).' },
  { time: '10:21:03', type: 'success', msg: 'Global sync completed across 14 distributed meshes.' }
])

// Simulate dynamic data
let timer: number
onMounted(() => {
  timer = window.setInterval(() => {
    stats.value.activeAgents = 120 + Math.floor(Math.random() * 20)
    stats.value.tasksPerHour = 440 + Math.floor(Math.random() * 30)
    
    // Rotate logs
    if (Math.random() > 0.7) {
      const now = new Date()
      const timeStr = `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}:${now.getSeconds().toString().padStart(2, '0')}`
      const msgs = [
        'Quantum link handshake verified.',
        'Mesh node DE-202 recalibrated.',
        'High-priority task delegated to Gamma cluster.',
        'Neural memory sync completed.'
      ]
      activityLog.value.unshift({
        time: timeStr,
        type: Math.random() > 0.8 ? 'warning' : 'info',
        msg: msgs[Math.floor(Math.random() * msgs.length)]
      })
      if (activityLog.value.length > 5) activityLog.value.pop()
    }

    agents.value.forEach(a => {
      if (a.status === 'WORKING') {
        a.load = Math.min(100, Math.max(70, a.load + (Math.random() * 10 - 5)))
      }
    })
  }, 3000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <div class="digital-employee-dashboard min-h-screen bg-[var(--bg-body)] text-[var(--text-main)] font-sans relative overflow-hidden selection:bg-[var(--matrix-color)] selection:text-white industrial pt-20">
    <!-- Sci-Fi Background Elements -->
    <div class="fixed inset-0 pointer-events-none overflow-hidden">
      <!-- Scanline Overlay -->
      <div class="absolute inset-0 z-50 opacity-[0.03] pointer-events-none overflow-hidden">
        <div class="absolute inset-0 bg-gradient-to-b from-transparent via-[var(--matrix-color)] to-transparent h-[200%] animate-scanline"></div>
      </div>

      <!-- Animated Starfield -->
      <div class="absolute inset-0 opacity-30">
        <div v-for="i in 80" :key="i" 
             class="absolute rounded-full animate-pulse"
             :style="{
               top: Math.random() * 100 + '%',
               left: Math.random() * 100 + '%',
               width: Math.random() * 2 + 1 + 'px',
               height: Math.random() * 2 + 1 + 'px',
               backgroundColor: ['var(--matrix-color)', 'var(--accent-cyan)', 'var(--accent-purple)', 'var(--accent-orange)'][Math.floor(Math.random() * 4)],
               animationDelay: Math.random() * 5 + 's',
               animationDuration: Math.random() * 3 + 2 + 's'
             }">
        </div>
      </div>
      
      <!-- Glowing Orbs for "Bright" feel -->
      <div class="absolute top-[-10%] left-[-10%] w-[50%] h-[50%] bg-[var(--matrix-color)]/5 blur-[120px] rounded-full"></div>
      <div class="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] bg-[var(--accent-purple)]/5 blur-[120px] rounded-full"></div>
      <div class="absolute top-[20%] right-[10%] w-[30%] h-[30%] bg-[var(--accent-cyan)]/5 blur-[100px] rounded-full"></div>
    </div>

    <!-- Dashboard Content -->
    <main class="relative z-10 max-w-[1600px] mx-auto px-6 lg:px-12 py-12 pt-40">
      <!-- Main Header: Holographic Station Status -->
      <header class="mb-20 flex flex-col lg:flex-row justify-between items-start lg:items-end gap-12 relative">
        <div class="space-y-6 relative">
          <!-- Hologram Glow -->
          <div class="absolute -inset-8 bg-[var(--matrix-color)]/5 blur-3xl rounded-full pointer-events-none"></div>
          
          <div class="flex items-center gap-3">
            <div class="relative flex items-center justify-center">
              <div class="absolute inset-0 bg-[var(--matrix-color)] rounded-full animate-ping opacity-20"></div>
              <Radio class="w-4 h-4 text-[var(--matrix-color)] relative z-10" />
            </div>
            <span class="text-[10px] font-black uppercase tracking-[0.6em] text-[var(--matrix-color)] filter drop-shadow-[0_0_12px_rgba(var(--matrix-color-rgb),0.7)]">
              {{ tt('portal.digital_employee.dashboard_title_small', 'Global Command Center') }}
            </span>
          </div>
          
          <h1 class="text-8xl font-black uppercase tracking-tighter leading-none relative">
            <span class="relative z-10 text-transparent bg-clip-text bg-[var(--gradient-highlight)] filter drop-shadow-[0_0_20px_rgba(var(--matrix-color-rgb),0.2)]">
              {{ tt('portal.digital_employee.dashboard_title', 'Agent Matrix') }}
            </span>
            <span class="absolute inset-0 text-[var(--matrix-color)]/10 blur-[3px] translate-x-1.5 translate-y-1.5 select-none">Agent Matrix</span>
          </h1>
          
          <div class="flex items-center gap-8 text-[11px] font-black tracking-[0.4em] uppercase opacity-50">
            <span class="flex items-center gap-3"><Activity class="w-4 h-4 text-[var(--matrix-color)]" /> {{ tt('portal.digital_employee.dashboard_subtitle', 'Operations Control') }}</span>
            <span class="flex items-center gap-3"><Cpu class="w-4 h-4" /> SEC-LINK ACTIVE</span>
            <span class="text-[var(--matrix-color)] animate-pulse flex items-center gap-2"><span class="w-2 h-2 rounded-full bg-[var(--matrix-color)] shadow-[0_0_8px_var(--matrix-color)]"></span> System Live</span>
          </div>
        </div>

        <div class="flex gap-8">
          <!-- Glassmorphism Stats -->
          <div class="px-12 py-8 bg-[var(--bg-card)] backdrop-blur-3xl border border-[var(--border-color)] rounded-[3rem] flex items-center gap-10 shadow-[0_30px_60px_rgba(0,0,0,0.12)] hover:border-[var(--matrix-color)]/50 transition-all group relative overflow-hidden">
            <div class="absolute inset-0 bg-gradient-to-br from-[var(--matrix-color)]/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity"></div>
            <div class="text-right relative z-10">
              <div class="text-[11px] font-black uppercase opacity-40 tracking-[0.3em] mb-2">{{ tt('portal.digital_employee.control_status', 'Control Status') }}</div>
              <div class="text-4xl font-light tracking-tighter tabular-nums flex items-baseline gap-3 text-[var(--matrix-color)]">
                {{ tt('portal.digital_employee.operational', 'ACTIVE') }}
              </div>
            </div>
            <div class="w-16 h-16 rounded-2xl bg-[var(--matrix-color)]/15 flex items-center justify-center text-[var(--matrix-color)] shadow-inner group-hover:scale-110 group-hover:shadow-[0_0_20px_rgba(var(--matrix-color-rgb),0.3)] transition-all relative z-10">
              <ShieldCheck class="w-8 h-8" />
            </div>
          </div>
        </div>
      </header>

      <!-- Stats Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8 mb-20">
        <div v-for="(val, key) in stats" :key="key" 
             class="relative p-12 bg-[var(--bg-card)] backdrop-blur-2xl border border-[var(--border-color)] rounded-[3.5rem] hover:bg-[var(--bg-card)] hover:border-[var(--matrix-color)]/60 transition-all group overflow-hidden shadow-2xl">
          <!-- Decorative Background Icon -->
          <component :is="key === 'activeAgents' ? Users : key === 'tasksPerHour' ? Activity : key === 'efficiency' ? Zap : Clock" 
                     class="absolute -right-6 -bottom-6 w-40 h-40 opacity-[0.04] group-hover:opacity-[0.08] transition-all group-hover:scale-110 group-hover:-rotate-6" />
          
          <div class="flex justify-between items-start mb-12 relative z-10">
            <div class="w-14 h-14 rounded-2xl bg-[var(--matrix-color)]/10 flex items-center justify-center text-[var(--matrix-color)] group-hover:bg-[var(--matrix-color)] group-hover:text-white group-hover:shadow-[0_0_25px_rgba(var(--matrix-color-rgb),0.5)] transition-all">
              <component :is="key === 'activeAgents' ? Users : key === 'tasksPerHour' ? Activity : key === 'efficiency' ? Zap : Clock" class="w-7 h-7" />
            </div>
            <div class="text-[11px] font-black uppercase tracking-[0.4em] opacity-40">{{ key.replace(/([A-Z])/g, ' $1') }}</div>
          </div>
          
          <div class="relative z-10">
            <div class="text-6xl font-black tracking-tighter mb-6 tabular-nums flex items-baseline gap-3">
              {{ val }}
              <span class="text-sm opacity-40 font-black uppercase tracking-widest">{{ key === 'efficiency' ? '%' : '' }}</span>
            </div>
            
            <div class="h-2.5 bg-[var(--matrix-color)]/10 rounded-full overflow-hidden p-[1.5px] border border-white/5">
              <div class="h-full bg-[var(--gradient-highlight)] rounded-full transition-all duration-1000 shadow-[0_0_15px_rgba(var(--matrix-color-rgb),0.4)]"
                   :style="{ width: (key === 'efficiency' ? val : (parseInt(val.toString()) / (key === 'activeAgents' ? 200 : 600) * 100)) + '%' }"></div>
            </div>
          </div>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-12 gap-12">
        <!-- Main Fleet Table -->
        <section class="lg:col-span-8 bg-[var(--bg-card)] backdrop-blur-3xl border border-[var(--border-color)] rounded-[4rem] overflow-hidden shadow-2xl flex flex-col">
          <div class="px-12 py-12 border-b border-[var(--border-color)] flex justify-between items-center bg-[var(--matrix-color)]/[0.03]">
            <div class="flex items-center gap-8">
              <div class="w-14 h-14 rounded-2xl bg-[var(--matrix-color)]/15 flex items-center justify-center shadow-lg">
                <Bot class="w-7 h-7 text-[var(--matrix-color)] animate-pulse" />
              </div>
              <div>
                <h2 class="text-3xl font-black tracking-tight uppercase">{{ tt('portal.digital_employee.fleet_deployment', 'Active Fleet Deployment') }}</h2>
                <p class="text-[11px] font-bold opacity-40 tracking-[0.3em] uppercase mt-1.5">{{ tt('portal.digital_employee.fleet_monitoring', 'Real-time Agent Monitoring') }}</p>
              </div>
            </div>
            <button class="px-8 py-4 bg-[var(--matrix-color)]/10 hover:bg-[var(--matrix-color)] hover:text-white border border-[var(--border-color)] rounded-2xl text-[11px] font-black tracking-[0.3em] uppercase transition-all active:scale-95 shadow-lg">
              {{ tt('portal.digital_employee.optimize_allocation', 'Optimize Allocation') }}
            </button>
          </div>

          <div class="p-6 overflow-x-auto">
            <table class="w-full text-left">
              <thead>
                <tr class="text-[11px] font-black uppercase tracking-[0.4em] opacity-30">
                  <th class="px-10 py-8">{{ tt('portal.digital_employee.agent_id', 'Agent ID') }}</th>
                  <th class="px-10 py-8">{{ tt('portal.digital_employee.agent_name', 'Agent Name') }}</th>
                  <th class="px-10 py-8">{{ tt('portal.digital_employee.status', 'Status') }}</th>
                  <th class="px-10 py-8">{{ tt('portal.digital_employee.load', 'Load Factor') }}</th>
                  <th class="px-10 py-8 text-right">{{ tt('portal.digital_employee.objective', 'Objective') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-white/5">
                <tr v-for="agent in agents" :key="agent.id" class="group hover:bg-[var(--matrix-color)]/[0.03] transition-all">
                  <td class="px-10 py-10">
                    <span class="font-mono text-xs text-[var(--matrix-color)] opacity-70 bg-[var(--matrix-color)]/10 px-4 py-1.5 rounded-lg border border-[var(--matrix-color)]/20 shadow-sm">
                      {{ agent.id }}
                    </span>
                  </td>
                  <td class="px-10 py-10">
                    <div class="flex items-center gap-4">
                      <div class="w-2.5 h-2.5 rounded-full bg-[var(--matrix-color)] shadow-[0_0_12px_var(--matrix-color)]"></div>
                      <span class="font-black text-base uppercase tracking-tight text-[var(--text-main)] group-hover:translate-x-1 transition-transform">{{ agent.name }}</span>
                    </div>
                  </td>
                  <td class="px-10 py-10">
                    <div class="flex items-center gap-3">
                      <span class="relative flex h-2.5 w-2.5">
                        <span :class="agent.status === 'WORKING' ? 'bg-emerald-400' : 'bg-yellow-400'" class="animate-ping absolute inline-flex h-full w-full rounded-full opacity-75"></span>
                        <span :class="agent.status === 'WORKING' ? 'bg-emerald-500' : 'bg-yellow-500'" class="relative inline-flex rounded-full h-2.5 w-2.5"></span>
                      </span>
                      <span class="text-[11px] font-black uppercase tracking-[0.2em]" :class="agent.status === 'WORKING' ? 'text-emerald-400' : 'text-yellow-400'">
                        {{ agent.status }}
                      </span>
                    </div>
                  </td>
                  <td class="px-10 py-10">
                    <div class="flex items-center gap-6">
                      <div class="flex-1 h-2 bg-[var(--matrix-color)]/10 rounded-full overflow-hidden max-w-[140px] border border-white/5 p-[1px]">
                        <div class="h-full bg-gradient-to-r from-[var(--matrix-color)] to-[var(--accent-cyan)] transition-all duration-1000 shadow-[0_0_8px_rgba(var(--matrix-color-rgb),0.3)]" :style="{ width: agent.load + '%' }"></div>
                      </div>
                      <span class="text-sm font-black tabular-nums opacity-60">{{ agent.load }}%</span>
                    </div>
                  </td>
                  <td class="px-10 py-10 text-right">
                    <span class="text-sm font-medium text-[var(--text-muted)] italic opacity-60 group-hover:opacity-100 transition-opacity">"{{ agent.task }}"</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <!-- Right: Activity Log -->
        <aside class="lg:col-span-4 space-y-12 flex flex-col">
          <section class="bg-[var(--bg-card)] backdrop-blur-3xl border border-[var(--border-color)] rounded-[4rem] overflow-hidden shadow-2xl flex-1 flex flex-col">
            <div class="px-12 py-12 border-b border-[var(--border-color)] flex items-center gap-8 bg-[var(--matrix-color)]/[0.03]">
              <div class="w-14 h-14 rounded-2xl bg-[var(--matrix-color)]/15 flex items-center justify-center shadow-lg">
                <Terminal class="w-7 h-7 text-[var(--matrix-color)]" />
              </div>
              <div>
                <h2 class="text-3xl font-black tracking-tight uppercase">{{ tt('portal.digital_employee.activity_log', 'Activity Log') }}</h2>
                <p class="text-[11px] font-bold opacity-40 tracking-[0.3em] uppercase mt-1.5">{{ tt('portal.digital_employee.live_feed', 'Live Transmission') }}</p>
              </div>
            </div>
            
            <div class="p-10 space-y-8 overflow-y-auto max-h-[600px] custom-scrollbar">
              <div v-for="(log, idx) in activityLog" :key="idx" 
                   class="flex gap-6 p-8 rounded-[2.5rem] bg-white/[0.02] border border-white/5 hover:border-[var(--matrix-color)]/30 transition-all group relative overflow-hidden">
                <div class="absolute inset-0 bg-gradient-to-r from-[var(--matrix-color)]/[0.02] to-transparent opacity-0 group-hover:opacity-100 transition-opacity"></div>
                <div class="w-2.5 h-2.5 rounded-full mt-2.5 shrink-0 relative z-10" 
                     :class="log.type === 'success' ? 'bg-emerald-500 shadow-[0_0_12px_rgba(16,185,129,0.6)]' : log.type === 'warning' ? 'bg-yellow-500 shadow-[0_0_12px_rgba(245,158,11,0.6)]' : 'bg-[var(--matrix-color)] shadow-[0_0_12px_rgba(var(--matrix-color-rgb),0.6)]'"></div>
                <div class="space-y-3 relative z-10 w-full">
                  <div class="flex justify-between items-center">
                    <span class="text-[11px] font-black opacity-30 tabular-nums tracking-widest">{{ log.time }}</span>
                    <span class="text-[9px] font-black uppercase tracking-[0.2em] px-2.5 py-1 rounded-md bg-white/5 opacity-40 group-hover:opacity-60 transition-opacity border border-white/5">{{ log.type }}</span>
                  </div>
                  <p class="text-sm font-medium leading-relaxed text-[var(--text-main)] opacity-70 group-hover:opacity-100 transition-opacity">
                    {{ log.msg }}
                  </p>
                </div>
              </div>
            </div>
          </section>

          <!-- System Health Card -->
          <section class="bg-gradient-to-br from-[var(--matrix-color)] to-blue-600 p-12 rounded-[3.5rem] text-white shadow-[0_40px_80px_-20px_rgba(var(--matrix-color-rgb),0.5)] relative overflow-hidden group">
            <div class="absolute top-[-20%] right-[-20%] w-[100%] h-[100%] bg-white/10 blur-[100px] rounded-full pointer-events-none group-hover:scale-125 transition-transform duration-1000"></div>
            <Cpu class="absolute -bottom-12 -right-12 w-56 h-56 opacity-[0.08] group-hover:scale-110 group-hover:-rotate-12 transition-all duration-1000" />
            
            <div class="relative z-10">
              <h3 class="text-3xl font-black uppercase tracking-tight mb-6 leading-tight">Neural Mesh Health</h3>
              <p class="text-white/80 text-sm mb-12 leading-relaxed font-medium">Global synchronization with all digital employee nodes is currently optimal. No latency detected in main processing clusters.</p>
              
              <div class="flex items-center justify-between p-6 bg-white/15 backdrop-blur-xl border border-white/25 rounded-3xl shadow-inner group-hover:bg-white/20 transition-all">
                <div class="space-y-1">
                  <div class="text-[10px] font-black uppercase opacity-60 tracking-widest mb-1">Mesh Latency</div>
                  <div class="text-3xl font-black tabular-nums">12.4ms</div>
                </div>
                <div class="w-16 h-16 rounded-2xl bg-white/25 flex items-center justify-center shadow-lg group-hover:rotate-12 transition-transform">
                  <Activity class="w-8 h-8" />
                </div>
              </div>
            </div>
          </section>
        </aside>
      </div>
    </main>
  </div>
</template>

<style scoped>
.digital-employee-dashboard {
  background-size: 100px 100px;
  background-image: radial-gradient(circle at 2px 2px, var(--border-color) 1px, transparent 0);
}

.animate-scanline {
  animation: scanline 8s linear infinite;
}

@keyframes scanline {
  0% { transform: translateY(-100%); }
  100% { transform: translateY(100%); }
}

.animate-pulse-slow {
  animation: pulse 4s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 0.1; }
  50% { opacity: 0.3; }
}

.animate-bounce-slow {
  animation: bounce-slow 3s ease-in-out infinite;
}

@keyframes bounce-slow {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 10px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: var(--matrix-color);
}
</style>

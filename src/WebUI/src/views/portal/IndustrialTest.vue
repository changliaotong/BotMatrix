<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import PortalHeader from '@/components/layout/PortalHeader.vue'
import PortalFooter from '@/components/layout/PortalFooter.vue'
import { Activity, ShieldCheck, Cpu, Settings2, Database, Zap, Thermometer, Gauge, Rocket, Radio, Navigation2, Orbit } from 'lucide-vue-next'

const telemetry = ref({
  energy: 94.2,
  oxygen: 98.5,
  gravity: 1.0,
  shields: 100.0,
  velocity: 28400,
  warp: 0.0,
  thermal: 42.1,
  cpu_load: 34.5
})

const missionLog = ref([
  { time: '14:20:05', msg: 'Quantum Link established with Sector-7', type: 'info' },
  { time: '14:21:12', msg: 'Orbital correction completed (+0.04°)', type: 'success' },
  { time: '14:22:45', msg: 'AI Core synchronization: 100%', type: 'info' },
  { time: '14:23:10', msg: 'Warp Drive pre-heating initiated', type: 'warning' }
])

const interval = ref<any>(null)

onMounted(() => {
  interval.value = setInterval(() => {
    telemetry.value.energy = +(94 + Math.random() * 0.5).toFixed(1)
    telemetry.value.velocity = Math.floor(28400 + Math.random() * 50)
    telemetry.value.oxygen = +(98.4 + Math.random() * 0.2).toFixed(1)
    telemetry.value.cpu_load = +(30 + Math.random() * 10).toFixed(1)
    telemetry.value.thermal = +(42 + Math.random() * 0.8).toFixed(1)
  }, 2000)
})

onUnmounted(() => {
  if (interval.value) clearInterval(interval.value)
})

const sectors = [
  { id: 'SEC-01', name: 'Navigation Array', status: 'ACTIVE', load: 12 },
  { id: 'SEC-02', name: 'Quantum Processor', status: 'ACTIVE', load: 64 },
  { id: 'SEC-03', name: 'Life Support', status: 'ACTIVE', load: 28 },
  { id: 'SEC-04', name: 'Shield Generator', status: 'STANDBY', load: 0 }
]
</script>

<template>
  <div class="industrial-test min-h-screen font-sans bg-[var(--bg-body)] text-[var(--text-main)] transition-colors duration-500 industrial overflow-hidden">
    <!-- Bright Space Sci-Fi Background Elements -->
    <div class="fixed inset-0 pointer-events-none">
      <!-- Animated Starfield -->
      <div class="absolute inset-0 opacity-30">
        <div v-for="i in 60" :key="i" 
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

    <PortalHeader />
    
    <main class="max-w-[1600px] mx-auto px-6 lg:px-12 py-12 pt-32 relative z-10">
      <!-- Main Header: Holographic Station Status -->
      <header class="mb-16 flex flex-col lg:flex-row justify-between items-start lg:items-end gap-8 relative">
        <div class="space-y-4 relative">
          <!-- Hologram Glow -->
          <div class="absolute -inset-4 bg-[var(--matrix-color)]/5 blur-2xl rounded-full pointer-events-none"></div>
          
          <div class="flex items-center gap-4">
            <div class="flex gap-1">
              <div v-for="i in 3" :key="i" class="w-1.5 h-6 bg-[var(--matrix-color)] rounded-full animate-bounce" :style="{ animationDelay: i * 0.1 + 's' }"></div>
            </div>
            <span class="text-xs font-black uppercase tracking-[0.5em] text-[var(--matrix-color)] filter drop-shadow-[0_0_8px_rgba(var(--matrix-color-rgb),0.6)]">
              Orbital Command Center
            </span>
          </div>
          
          <h1 class="text-7xl font-black uppercase tracking-tighter leading-none relative">
            <span class="relative z-10 text-transparent bg-clip-text bg-[var(--gradient-highlight)]">
              Future Space
            </span>
            <span class="absolute inset-0 text-[var(--matrix-color)]/10 blur-[2px] translate-x-1 translate-y-1 select-none">Future Space</span>
          </h1>
          
          <div class="flex items-center gap-6 text-[10px] font-bold tracking-[0.3em] uppercase opacity-40">
            <span class="flex items-center gap-2"><Orbit class="w-3 h-3" /> Sector 7-G</span>
            <span class="flex items-center gap-2"><Database class="w-3 h-3" /> Syncing...</span>
            <span class="text-[var(--matrix-color)] animate-pulse">● System Live</span>
          </div>
        </div>

        <div class="flex gap-6">
          <!-- Glassmorphism Stats -->
          <div class="px-10 py-6 bg-[var(--bg-card)] backdrop-blur-2xl border border-[var(--border-color)] rounded-[2.5rem] flex items-center gap-8 shadow-[0_20px_50px_rgba(0,0,0,0.1)] hover:border-[var(--matrix-color)]/40 transition-all group">
            <div class="text-right">
              <div class="text-[10px] font-black uppercase opacity-40 tracking-[0.2em] mb-1">Local Time</div>
              <div class="text-3xl font-light tracking-tighter tabular-nums flex items-baseline gap-2">
                14:24:58 <span class="text-xs opacity-40 font-black">UTC</span>
              </div>
            </div>
            <div class="w-14 h-14 rounded-2xl bg-[var(--matrix-color)]/20 flex items-center justify-center text-[var(--matrix-color)] shadow-inner group-hover:scale-110 transition-transform">
              <Navigation2 class="w-7 h-7" />
            </div>
          </div>
          
          <div class="px-10 py-6 bg-[var(--gradient-highlight)] text-white rounded-[2.5rem] flex items-center gap-8 shadow-[0_30px_60px_-12px_rgba(var(--matrix-color-rgb),0.5)] hover:translate-y-[-4px] transition-all cursor-pointer group active:scale-95">
            <div class="text-right">
              <div class="text-[10px] font-black uppercase opacity-70 tracking-[0.2em] mb-1">System Load</div>
              <div class="text-3xl font-black tracking-tight">NOMINAL</div>
            </div>
            <div class="w-14 h-14 rounded-2xl bg-white/20 flex items-center justify-center backdrop-blur-md">
              <ShieldCheck class="w-8 h-8 group-hover:rotate-12 transition-transform" />
            </div>
          </div>
        </div>
      </header>

      <!-- Dashboard Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8 mb-16">
        <!-- Telemetry Cards with Glassmorphism and Glow -->
        <div v-for="(val, key) in telemetry" :key="key" 
             class="relative p-10 bg-[var(--bg-card)] backdrop-blur-xl border border-[var(--border-color)] rounded-[3rem] hover:bg-[var(--bg-card)] hover:border-[var(--matrix-color)]/50 transition-all group overflow-hidden shadow-xl">
          <!-- Decorative Background Icon -->
          <component :is="key === 'energy' ? Zap : key === 'oxygen' ? Activity : key === 'velocity' ? Rocket : Orbit" 
                     class="absolute -right-4 -bottom-4 w-32 h-32 opacity-[0.05] group-hover:opacity-[0.1] transition-all group-hover:scale-110 rotate-12" />
          
          <div class="flex justify-between items-start mb-10 relative z-10">
            <div class="w-12 h-12 rounded-2xl bg-[var(--matrix-color)]/10 flex items-center justify-center text-[var(--matrix-color)] group-hover:bg-[var(--matrix-color)] group-hover:text-white group-hover:shadow-[0_0_20px_rgba(var(--matrix-color-rgb),0.4)] transition-all">
              <component :is="key === 'energy' ? Zap : key === 'oxygen' ? Activity : key === 'velocity' ? Rocket : Orbit" class="w-6 h-6" />
            </div>
            <div class="text-[10px] font-black uppercase tracking-[0.3em] opacity-50">{{ key }}</div>
          </div>
          
          <div class="relative z-10">
            <div class="text-5xl font-black tracking-tighter mb-4 tabular-nums flex items-baseline gap-2">
              {{ key === 'velocity' ? val.toLocaleString() : val }}
              <span class="text-sm opacity-40 font-black uppercase tracking-widest">{{ key === 'velocity' ? 'km/h' : '%' }}</span>
            </div>
            
            <div class="h-2 bg-[var(--matrix-color)]/5 rounded-full overflow-hidden p-[1px]">
              <div class="h-full bg-[var(--gradient-highlight)] rounded-full transition-all duration-1000 shadow-[0_0_10px_rgba(var(--matrix-color-rgb),0.3)]"
                   :style="{ width: (key === 'velocity' ? (val/30000)*100 : val) + '%' }"></div>
            </div>
          </div>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-12 gap-12">
        <!-- Left: High-Tech Table -->
        <div class="lg:col-span-8 space-y-12">
          <section class="bg-[var(--bg-card)] backdrop-blur-3xl border border-[var(--border-color)] rounded-[3.5rem] overflow-hidden shadow-2xl">
            <div class="px-10 py-10 border-b border-[var(--border-color)] flex justify-between items-center bg-[var(--matrix-color)]/[0.02]">
              <div class="flex items-center gap-6">
                <div class="w-12 h-12 rounded-2xl bg-[var(--matrix-color)]/10 flex items-center justify-center">
                  <Cpu class="w-6 h-6 text-[var(--matrix-color)] animate-pulse" />
                </div>
                <div>
                  <h2 class="text-2xl font-black tracking-tight uppercase">Sector Diagnostics</h2>
                  <p class="text-[10px] font-bold opacity-50 tracking-[0.2em] uppercase mt-1">Real-time Node Monitoring</p>
                </div>
              </div>
              <button class="px-6 py-3 bg-[var(--matrix-color)]/5 hover:bg-[var(--matrix-color)]/10 border border-[var(--border-color)] rounded-2xl text-[10px] font-black tracking-[0.2em] uppercase transition-all active:scale-95">
                Refresh Matrix
              </button>
            </div>
            
            <div class="p-4 overflow-x-auto">
              <table class="w-full text-left">
                <thead>
                  <tr class="text-[10px] font-black uppercase tracking-[0.3em] opacity-30">
                    <th class="px-8 py-6">Node ID</th>
                    <th class="px-8 py-6">Designation</th>
                    <th class="px-8 py-6">Status</th>
                    <th class="px-8 py-6">Load Factor</th>
                    <th class="px-8 py-6 text-right">Action</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-white/5">
                  <tr v-for="sec in sectors" :key="sec.id" class="group hover:bg-white/[0.03] transition-all">
                    <td class="px-8 py-8">
                      <span class="font-mono text-xs text-[var(--matrix-color)] opacity-60 bg-[var(--matrix-color)]/5 px-3 py-1 rounded-md border border-[var(--matrix-color)]/10">
                        {{ sec.id }}
                      </span>
                    </td>
                    <td class="px-8 py-8 font-black tracking-tight text-lg">{{ sec.name }}</td>
                    <td class="px-8 py-8">
                      <div class="flex items-center gap-2">
                        <span class="relative flex h-2 w-2">
                          <span :class="sec.status === 'ACTIVE' ? 'bg-cyan-400' : 'bg-white/20'" class="animate-ping absolute inline-flex h-full w-full rounded-full opacity-75"></span>
                          <span :class="sec.status === 'ACTIVE' ? 'bg-cyan-500' : 'bg-white/40'" class="relative inline-flex rounded-full h-2 w-2"></span>
                        </span>
                        <span :class="sec.status === 'ACTIVE' ? 'text-cyan-400' : 'text-white/30'" class="text-[10px] font-black uppercase tracking-[0.2em]">
                          {{ sec.status }}
                        </span>
                      </div>
                    </td>
                    <td class="px-8 py-8">
                      <div class="flex items-center gap-6">
                        <div class="flex-1 h-1.5 bg-[var(--matrix-color)]/10 rounded-full overflow-hidden max-w-[120px]">
                          <div class="h-full bg-gradient-to-r from-[var(--matrix-color)] to-[var(--accent-cyan)] transition-all duration-1000" :style="{ width: sec.load + '%' }"></div>
                        </div>
                        <span class="text-xs font-black tabular-nums opacity-40">{{ sec.load }}%</span>
                      </div>
                    </td>
                    <td class="px-8 py-8 text-right">
                      <button class="w-10 h-10 flex items-center justify-center bg-[var(--matrix-color)]/5 hover:bg-[var(--matrix-color)] hover:text-white rounded-xl transition-all shadow-lg group-hover:scale-110">
                        <Settings2 class="w-4 h-4" />
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>

          <!-- Terminal Style Mission Log -->
          <section class="bg-[var(--bg-card)] backdrop-blur-2xl border border-[var(--border-color)] rounded-[3.5rem] p-10 shadow-inner">
            <div class="flex items-center justify-between mb-10">
              <div class="flex items-center gap-6">
                <div class="w-12 h-12 rounded-2xl bg-[var(--matrix-color)]/10 flex items-center justify-center">
                  <Radio class="w-6 h-6 text-[var(--matrix-color)]" />
                </div>
                <h2 class="text-2xl font-black tracking-tight uppercase">Event Stream</h2>
              </div>
              <div class="flex gap-2">
                <div v-for="i in 3" :key="i" class="w-2 h-2 rounded-full bg-[var(--matrix-color)]/10"></div>
              </div>
            </div>
            
            <div class="space-y-4 font-mono">
              <div v-for="(log, i) in missionLog" :key="i" 
                   class="flex gap-8 items-center p-5 bg-[var(--matrix-color)]/[0.02] rounded-2xl border border-[var(--border-color)]/30 hover:border-[var(--matrix-color)]/50 transition-all group">
                <span class="text-[10px] font-black text-[var(--matrix-color)] opacity-60 whitespace-nowrap">[{{ log.time }}]</span>
                <span class="w-2 h-2 rounded-full" :class="log.type === 'warning' ? 'bg-yellow-500 shadow-[0_0_8px_rgba(234,179,8,0.5)]' : log.type === 'success' ? 'bg-cyan-500 shadow-[0_0_8px_rgba(6,182,212,0.5)]' : 'bg-[var(--matrix-color)] shadow-[0_0_8px_rgba(var(--matrix-color-rgb),0.5)]'"></span>
                <p class="text-sm font-medium tracking-wide group-hover:translate-x-1 transition-transform text-[var(--text-main)]">{{ log.msg }}</p>
              </div>
            </div>
          </section>
        </div>

        <!-- Right: Control Panel & Widgets -->
        <div class="lg:col-span-4 space-y-12">
          <!-- Main Action Card -->
          <section class="bg-[var(--gradient-highlight)] rounded-[4rem] p-12 text-white shadow-[0_50px_100px_-20px_rgba(var(--matrix-color-rgb),0.5)] relative overflow-hidden group">
            <!-- Decorative Elements -->
            <div class="absolute top-[-20%] right-[-20%] w-[80%] h-[80%] bg-white/10 blur-[100px] rounded-full pointer-events-none group-hover:scale-125 transition-transform duration-1000"></div>
            <Rocket class="absolute -bottom-12 -right-12 w-64 h-64 opacity-10 group-hover:scale-110 group-hover:-rotate-12 transition-all duration-1000" />
            
            <div class="relative z-10">
              <h3 class="text-4xl font-black uppercase tracking-tighter mb-6 leading-none">Warp Control</h3>
              <p class="text-white/80 text-sm mb-12 leading-relaxed font-medium">Initiate FTL travel protocols. Coordinate sync with Global Agent Mesh required for jump.</p>
              
              <div class="space-y-6">
                <button class="w-full py-6 bg-white text-[var(--matrix-color)] font-black text-xs uppercase tracking-[0.3em] rounded-[2rem] hover:shadow-2xl hover:scale-[1.02] active:scale-95 transition-all">
                  Execute Warp Jump
                </button>
                <button class="w-full py-6 bg-black/20 backdrop-blur-xl border border-white/30 font-black text-xs uppercase tracking-[0.3em] rounded-[2rem] hover:bg-black/30 transition-all">
                  Deep Space Scan
                </button>
              </div>
            </div>
          </section>

          <!-- Technical Widget -->
          <section class="bg-[var(--bg-card)] backdrop-blur-2xl border border-[var(--border-color)] rounded-[3.5rem] p-12 relative overflow-hidden">
            <div class="absolute -top-10 -left-10 w-40 h-40 bg-[var(--matrix-color)]/5 rounded-full blur-3xl"></div>
            <div class="flex items-center gap-4 mb-10">
              <Gauge class="w-6 h-6 text-[var(--matrix-color)]" />
              <h3 class="text-xl font-black uppercase tracking-tight">Environmental</h3>
            </div>
            
            <div class="space-y-8">
              <div v-for="i in 3" :key="i" class="relative">
                <div class="flex justify-between items-end mb-3">
                  <div class="space-y-1">
                    <div class="text-[9px] font-black uppercase tracking-[0.2em] opacity-40">Sensor Module {{ i }}</div>
                    <div class="text-md font-black tracking-tight uppercase text-[var(--text-main)]">ATMOS-X{{ 400 + i }}</div>
                  </div>
                  <div class="text-right">
                    <div class="text-lg font-mono font-black text-[var(--matrix-color)] filter drop-shadow-[0_0_8px_rgba(var(--matrix-color-rgb),0.3)]">102.4 <span class="text-[10px] opacity-40">hPa</span></div>
                    <div class="text-[9px] font-black text-cyan-500 uppercase tracking-widest mt-1 animate-pulse">Stable</div>
                  </div>
                </div>
                <div class="h-1 bg-[var(--matrix-color)]/10 rounded-full overflow-hidden">
                  <div class="h-full bg-[var(--matrix-color)] opacity-40" :style="{ width: (70 + i * 5) + '%' }"></div>
                </div>
              </div>
            </div>
          </section>
        </div>
      </div>
    </main>

    <PortalFooter />
  </div>
</template>

<style scoped>
.industrial-test {
  /* Removed grid background */
  background-size: 100px 100px;
}

@keyframes scanline {
  0% { transform: translateY(-100%); }
  100% { transform: translateY(100%); }
}

/* Removed scanline overlay */
</style>

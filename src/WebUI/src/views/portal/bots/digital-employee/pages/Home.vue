<template>
  <div class="relative z-10">
    <!-- Hero Section -->
    <header id="hero" class="pt-40 pb-32 px-4 relative overflow-hidden">
      <!-- Matrix Code Rain Placeholder / Background -->
      <div class="absolute inset-0 opacity-10 pointer-events-none overflow-hidden">
        <div class="absolute inset-0 bg-[grid-slate-800] [mask-image:radial-gradient(white,transparent)]"></div>
        <div class="matrix-lines absolute inset-0"></div>
      </div>
      
      <div class="absolute top-0 left-1/2 -translate-x-1/2 w-full h-full bg-[radial-gradient(circle_at_center,var(--matrix-color,rgba(168,85,247,0.2))_0%,transparent_70%)] -z-10 opacity-20"></div>
      <div class="max-w-7xl mx-auto text-center relative reveal">
        <div class="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-[var(--matrix-color)]/10 border border-[var(--matrix-color)]/20 text-[var(--matrix-color)] text-xs font-bold mb-8 animate-pulse shadow-[0_0_15px_rgba(var(--matrix-color-rgb),0.3)]">
          <Sparkles class="w-4 h-4" />
          {{ tt('portal.digital_employee.badge') }}
        </div>
        <h1 class="text-6xl md:text-9xl font-black mb-12 tracking-tighter leading-none text-[var(--text-main)]">
          {{ tt('portal.digital_employee.title_prefix') }} <br/>
          <span class="bg-clip-text text-transparent bg-gradient-to-r from-[var(--matrix-color)] to-[rgba(var(--matrix-color-rgb),0.5)] drop-shadow-[0_10px_10px_rgba(var(--matrix-color-rgb),0.3)]">
            {{ tt('portal.digital_employee.title_suffix') }}
          </span>
        </h1>
        <p class="text-xl md:text-3xl text-[var(--text-muted)] mb-16 max-w-4xl mx-auto font-light leading-relaxed">
          {{ tt('portal.digital_employee.subtitle') }}
        </p>
        <div class="flex flex-col sm:flex-row gap-6 justify-center items-center">
          <router-link to="/bots/digital-employee/dashboard" class="px-10 py-5 bg-[var(--matrix-color)] hover:bg-[var(--matrix-color)]/80 text-white rounded-2xl text-xl font-bold transition-all shadow-[0_0_40px_rgba(var(--matrix-color-rgb),0.3)] flex items-center gap-3 group">
            {{ tt('portal.digital_employee.cta_dashboard', '查看看板') }} <LayoutDashboard class="w-6 h-6 group-hover:rotate-12 transition-transform" />
          </router-link>
          <button @click="openAction('join')" class="px-10 py-5 bg-[var(--bg-body)]/50 hover:bg-[var(--bg-body)]/80 text-[var(--text-main)] rounded-2xl text-xl font-bold transition-all border border-[var(--border-color)] backdrop-blur-sm flex items-center gap-3 group">
            {{ tt('portal.digital_employee.cta_primary') }} <ArrowRight class="w-6 h-6 group-hover:translate-x-2 transition-transform" />
          </button>
        </div>
      </div>
    </header>

    <!-- Vision Section: Production Relations Revolution -->
    <section id="vision" class="py-32 relative overflow-hidden">
      <div class="max-w-7xl mx-auto px-4">
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-24 items-center">
          <div>
            <h2 class="text-4xl md:text-6xl font-black mb-8 leading-tight text-[var(--text-main)]">
              {{ tt('portal.digital_employee.vision.title_line1') }} <br/>
              <span class="text-[var(--matrix-color)]">{{ tt('portal.digital_employee.vision.title_line2') }}</span>
            </h2>
            <p class="text-xl text-[var(--text-muted)] mb-12 leading-relaxed">
              {{ tt('portal.digital_employee.vision.desc') }}
            </p>
            <div class="space-y-8">
              <div v-for="v in visionPoints" :key="v.title" class="flex gap-6 group">
                <div class="w-14 h-14 shrink-0 bg-[var(--bg-body)] border border-[var(--border-color)] rounded-2xl flex items-center justify-center text-[var(--matrix-color)] group-hover:bg-[var(--matrix-color)] group-hover:text-white transition-all">
                  <component :is="v.icon" class="w-7 h-7" />
                </div>
                <div>
                  <h4 class="text-2xl font-bold mb-2 group-hover:text-[var(--matrix-color)] transition-colors text-[var(--text-main)]">{{ v.title }}</h4>
                  <p class="text-[var(--text-muted)] leading-relaxed">{{ v.desc }}</p>
                </div>
              </div>
            </div>
          </div>
          <div class="relative">
            <div class="absolute inset-0 bg-[var(--matrix-color)]/20 blur-[100px] -z-10 opacity-20"></div>
            <div class="bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-[3rem] p-12 aspect-square flex items-center justify-center relative overflow-hidden">
              <div class="absolute top-0 left-0 w-full h-full opacity-20">
                <div class="absolute top-0 left-0 w-full h-full bg-[grid-slate-800] [mask-image:radial-gradient(white,transparent)]"></div>
              </div>
              <!-- Central AI Identity -->
              <div class="relative z-10 flex flex-col items-center">
                <div class="w-40 h-40 bg-gradient-to-br from-[var(--matrix-color)] to-[rgba(var(--matrix-color-rgb),0.4)] rounded-full flex items-center justify-center shadow-2xl animate-pulse">
                  <UserCheck class="w-20 h-20 text-white" />
                </div>
                <div class="mt-8 text-center">
                  <div class="text-3xl font-black mb-2 tracking-widest text-[var(--text-main)]">{{ tt('portal.digital_employee.vision.id_label') }}: AX-709</div>
                  <div class="text-[var(--matrix-color)] font-bold uppercase tracking-[0.3em] text-sm">{{ tt('portal.digital_employee.vision.role_senior_analyst') }}</div>
                </div>
              </div>
              <!-- Orbiting Stats -->
              <div class="absolute top-12 right-12 bg-[var(--bg-body)]/80 backdrop-blur border border-[var(--border-color)] p-4 rounded-2xl shadow-xl animate-bounce-slow">
                <div class="text-xs text-[var(--text-muted)] uppercase font-bold mb-1">{{ tt('portal.digital_employee.vision.stats_perf') }}</div>
                <div class="text-xl font-black text-[var(--matrix-color)]">99.8%</div>
              </div>
              <div class="absolute bottom-20 left-12 bg-[var(--bg-body)]/80 backdrop-blur border border-[var(--border-color)] p-4 rounded-2xl shadow-xl animate-bounce-slow" style="animation-delay: 1s">
                <div class="text-xs text-[var(--text-muted)] uppercase font-bold mb-1">{{ tt('portal.digital_employee.vision.stats_token') }}</div>
                <div class="text-xl font-black text-[var(--matrix-color)]">12.5M</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Matrix Revolution: The Global Agent Mesh -->
    <section id="revolution" class="py-32 bg-[var(--bg-body)]/30">
      <div class="max-w-7xl mx-auto px-4 text-center">
        <h2 class="text-4xl md:text-7xl font-black mb-12 text-[var(--text-main)]">{{ tt('portal.digital_employee.revolution.title') }}</h2>
        <p class="text-xl text-[var(--text-muted)] max-w-3xl mx-auto mb-20 leading-relaxed">
          {{ tt('portal.digital_employee.revolution.desc') }}
        </p>
        
        <div class="grid grid-cols-1 md:grid-cols-3 gap-12">
          <div v-for="r in revolutionSteps" :key="r.title" class="p-12 rounded-[3rem] bg-[var(--bg-body)] border border-[var(--border-color)] hover:border-[var(--matrix-color)]/50 transition-all group">
            <div class="w-20 h-20 bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] rounded-3xl flex items-center justify-center mx-auto mb-8 group-hover:scale-110 group-hover:bg-[var(--matrix-color)] group-hover:text-white transition-all">
              <component :is="r.icon" class="w-10 h-10" />
            </div>
            <h3 class="text-2xl font-bold mb-6 text-[var(--text-main)]">{{ r.title }}</h3>
            <p class="text-[var(--text-muted)] leading-relaxed">{{ r.desc }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- KPI & HR Section -->
    <section id="kpi" class="py-32 bg-[var(--bg-body)]">
      <div class="max-w-7xl mx-auto px-4">
        <div class="bg-gradient-to-br from-[var(--bg-body)] to-[var(--matrix-color)]/10 rounded-[4rem] p-12 md:p-24 border border-[var(--border-color)] relative overflow-hidden">
          <div class="absolute top-0 right-0 w-full h-full bg-[radial-gradient(circle_at_top_right,var(--matrix-color,rgba(168,85,247,0.1)),transparent_50%)]"></div>
          <div class="flex flex-col lg:flex-row gap-20 items-center relative z-10">
            <div class="flex-1">
              <h2 class="text-4xl md:text-5xl font-black mb-8 leading-tight text-[var(--text-main)]">{{ tt('portal.digital_employee.hr.title') }}</h2>
              <p class="text-xl text-[var(--text-muted)] mb-10 leading-relaxed">
                {{ tt('portal.digital_employee.hr.desc') }}
              </p>
              <div class="grid grid-cols-2 gap-6">
                <div v-for="stat in hrStats" :key="stat.label" class="p-6 bg-[var(--bg-body)]/50 rounded-2xl border border-[var(--border-color)]">
                  <div class="text-3xl font-black text-[var(--matrix-color)] mb-2">{{ stat.value }}</div>
                  <div class="text-xs text-[var(--text-muted)] font-bold uppercase tracking-widest">{{ stat.label }}</div>
                </div>
              </div>
            </div>
            <div class="flex-1 w-full space-y-6">
              <div v-for="role in roles" :key="role.name" class="p-6 bg-[var(--bg-body)]/80 backdrop-blur rounded-3xl border border-[var(--border-color)] hover:border-[var(--matrix-color)]/50 transition-all">
                <div class="flex justify-between items-center mb-4">
                  <div class="flex items-center gap-4">
                    <div class="w-12 h-12 bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-xl flex items-center justify-center">
                      <component :is="role.icon" class="w-6 h-6 text-[var(--matrix-color)]" />
                    </div>
                    <div>
                      <div class="font-bold text-lg text-[var(--text-main)]">{{ role.name }}</div>
                      <div class="text-xs text-[var(--text-muted)]">{{ role.dept }}</div>
                    </div>
                  </div>
                  <div class="px-3 py-1 bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] text-[10px] font-bold rounded-full uppercase">{{ tt('portal.digital_employee.hr.status_active') }}</div>
                </div>
                <div class="mb-2 text-xs text-[var(--text-muted)] line-clamp-1">{{ tt('portal.digital_employee.hr.role_skills') }}: {{ role.skills }}</div>
                <div class="w-full bg-[var(--border-color)] h-1.5 rounded-full overflow-hidden">
                  <div class="bg-[var(--matrix-color)] h-full" :style="{ width: role.progress + '%' }"></div>
                </div>
                <div class="flex justify-between mt-2 text-[10px] font-bold text-[var(--text-muted)]">
                  <span>{{ tt('portal.digital_employee.hr.kpi_label') }}</span>
                  <span>{{ role.progress }}%</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Industry Roles Showcase -->
    <section id="jobs" class="py-32 bg-[var(--bg-body)]">
      <div class="max-w-7xl mx-auto px-4">
        <div class="text-center mb-20">
          <h2 class="text-4xl md:text-6xl font-black mb-6 text-[var(--text-main)]">{{ tt('portal.digital_employee.jobs.title') }}</h2>
          <p class="text-xl text-[var(--text-muted)] max-w-3xl mx-auto">
            {{ tt('portal.digital_employee.jobs.desc') }}
          </p>
        </div>
        
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          <div v-for="job in industryJobs" :key="job.title" class="group bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-[2.5rem] p-10 hover:border-[var(--matrix-color)]/30 transition-all hover:-translate-y-2">
            <div class="w-16 h-16 rounded-2xl bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] flex items-center justify-center mb-8 group-hover:bg-[var(--matrix-color)] group-hover:text-white transition-all">
              <component :is="job.icon" class="w-8 h-8" />
            </div>
            <h3 class="text-2xl font-bold mb-4 text-[var(--text-main)]">{{ job.title }}</h3>
            <p class="text-[var(--text-muted)] mb-8 leading-relaxed text-sm">{{ job.desc }}</p>
            <div class="space-y-4">
              <div class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest border-b border-[var(--border-color)] pb-2">{{ tt('portal.digital_employee.jobs.core_skills') }}</div>
              <div class="flex flex-wrap gap-2">
                <span v-for="skill in job.skills" :key="skill" class="px-3 py-1 bg-[var(--bg-body)]/80 border border-[var(--border-color)] rounded-full text-xs text-[var(--text-muted)]">{{ skill }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Lifecycle & HITL -->
    <section id="security" class="py-32 bg-[var(--bg-body)]/30">
      <div class="max-w-7xl mx-auto px-4">
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-24 items-center">
          <div>
            <h2 class="text-4xl md:text-5xl font-black mb-8 leading-tight text-[var(--text-main)]">
              {{ tt('portal.digital_employee.security.title') }}
            </h2>
            <p class="text-xl text-[var(--text-muted)] mb-12 leading-relaxed">
              {{ tt('portal.digital_employee.security.desc') }}
            </p>
            <div class="space-y-6">
               <div v-for="f in securityFeatures" :key="f.title" class="flex items-start gap-4 p-6 bg-[var(--bg-body)]/50 rounded-2xl border border-[var(--border-color)]">
                 <div class="p-3 bg-[var(--matrix-color)]/10 rounded-xl text-[var(--matrix-color)]">
                   <component :is="f.icon" class="w-6 h-6" />
                 </div>
                 <div>
                   <h4 class="font-bold mb-1 text-[var(--text-main)]">{{ f.title }}</h4>
                   <p class="text-sm text-[var(--text-muted)]">{{ f.desc }}</p>
                 </div>
               </div>
            </div>
          </div>
          <div class="relative group">
            <div class="absolute inset-0 bg-[var(--matrix-color)]/10 blur-[120px] -z-10 group-hover:bg-[var(--matrix-color)]/20 transition-all"></div>
            <div class="bg-[var(--bg-body)] border border-[var(--border-color)] rounded-[3rem] p-8 md:p-12 shadow-2xl">
               <!-- Trace Visualization Placeholder -->
               <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-6 mb-12">
                <div>
                  <h3 class="text-2xl font-bold mb-2 flex items-center gap-3 text-[var(--text-main)]">
                    <ListTree class="w-6 h-6 text-[var(--matrix-color)]" />
                    {{ tt('portal.digital_employee.collaboration.trace_title') }}
                    <span class="flex h-2 w-2 relative ml-2">
                      <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-[var(--matrix-color)]/40 opacity-75"></span>
                      <span class="relative inline-flex rounded-full h-2 w-2 bg-[var(--matrix-color)]"></span>
                    </span>
                  </h3>
                  <p class="text-[var(--text-muted)] text-sm">{{ tt('portal.digital_employee.collaboration.trace_subtitle') }}</p>
                </div>
              </div>

              <div class="space-y-4">
                <div v-for="trace in collaborationTraces" :key="trace.id" 
                     class="grid grid-cols-1 md:grid-cols-[120px_1fr_200px] items-center gap-4 lg:gap-8 p-6 bg-[var(--bg-body)]/80 rounded-2xl border border-[var(--border-color)] hover:border-[var(--matrix-color)]/30 transition-all group">
                  
                  <!-- ID Column -->
                  <div class="flex items-center gap-4">
                    <div class="w-2 h-2 rounded-full bg-[var(--matrix-color)] shadow-[0_0_10px_rgba(var(--matrix-color-rgb),0.3)]"></div>
                    <span class="font-mono text-xs text-[var(--text-muted)]">{{ trace.id }}</span>
                  </div>

                  <!-- Content Column: From -> To | Action -->
                  <div class="flex flex-col xl:flex-row xl:items-center gap-4 xl:gap-8 overflow-hidden">
                    <div class="flex items-center gap-3 shrink-0">
                      <div class="px-3 py-1 bg-[var(--bg-body)]/50 rounded-lg border border-[var(--border-color)] text-[10px] font-bold text-[var(--text-main)] truncate max-w-[120px]" :title="trace.from">{{ trace.from }}</div>
                      <ArrowRight class="w-4 h-4 text-[var(--text-muted)] shrink-0" />
                      <div class="px-3 py-1 bg-[var(--bg-body)]/50 rounded-lg border border-[var(--border-color)] text-[10px] font-bold text-[var(--text-main)] truncate max-w-[120px]" :title="trace.to">{{ trace.to }}</div>
                    </div>
                    <div class="text-sm font-medium text-[var(--text-main)] group-hover:text-[var(--matrix-color)] transition-colors truncate" :title="trace.action">
                      {{ trace.action }}
                    </div>
                  </div>

                  <!-- Status & Time Column -->
                  <div class="flex items-center justify-between md:justify-end gap-6">
                    <div class="flex items-center gap-2 shrink-0" :class="trace.color">
                      <component :is="trace.icon" class="w-4 h-4" />
                      <span class="text-[10px] font-black uppercase tracking-widest">{{ trace.status }}</span>
                    </div>
                    <div class="text-[10px] font-bold text-[var(--text-muted)] uppercase shrink-0 whitespace-nowrap">{{ trace.time }}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Futuristic Vision (Computer Use & Swarm) -->
    <section id="future" class="py-32 bg-gradient-to-b from-[var(--bg-body)] to-[var(--matrix-color)]/20">
      <div class="max-w-7xl mx-auto px-4">
        <div class="flex flex-col lg:flex-row items-center gap-24">
          <div class="flex-1">
            <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] text-xs font-bold mb-6 border border-[var(--matrix-color)]/20">
              <Sparkles class="w-4 h-4" /> {{ tt('portal.digital_employee.future_lab') }}
            </div>
            <h2 class="text-4xl md:text-5xl font-black mb-8 leading-tight text-[var(--text-main)]">
              {{ tt('portal.digital_employee.future_evolution_prefix') }} <br/>
              <span class="text-[var(--matrix-color)]">{{ tt('portal.digital_employee.future_evolution_suffix') }}</span>
            </h2>
            <p class="text-xl text-[var(--text-muted)] mb-12">
              {{ tt('portal.digital_employee.future_desc') }}
            </p>
            <div class="space-y-8">
              <div v-for="feature in futuristicFeatures" :key="feature.title" class="group relative p-8 bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-3xl hover:border-[var(--matrix-color)]/50 transition-all">
                <div class="absolute top-4 right-8 px-2 py-1 bg-[var(--matrix-color)]/20 text-[var(--matrix-color)] text-[10px] font-bold rounded uppercase tracking-tighter">{{ feature.tag }}</div>
                <div class="flex gap-6">
                  <div class="shrink-0 w-12 h-12 bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] rounded-xl flex items-center justify-center">
                    <component :is="feature.icon" class="w-6 h-6" />
                  </div>
                  <div>
                    <h4 class="text-xl font-bold mb-2 text-[var(--text-main)]">{{ feature.title }}</h4>
                    <p class="text-[var(--text-muted)] text-sm leading-relaxed">{{ feature.desc }}</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div class="flex-1 relative">
            <div class="aspect-video bg-[var(--bg-body)]/50 rounded-[2.5rem] border border-[var(--border-color)] p-2 overflow-hidden shadow-2xl">
              <div class="w-full h-full bg-[var(--bg-body)] rounded-[2rem] relative flex items-center justify-center">
                <!-- Decorative elements -->
                <div class="absolute -top-12 -right-12 w-64 h-64 bg-[var(--matrix-color)]/10 blur-[100px] -z-10"></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- The Future Vision -->
    <section class="py-40 text-center px-4 relative">
      <div class="absolute inset-0 bg-[radial-gradient(circle_at_center,rgba(168,85,247,0.05)_0%,transparent_70%)]"></div>
      <h2 class="text-5xl md:text-8xl font-black mb-12 tracking-tighter">
        {{ tt('portal.digital_employee.future.motto_line1') }} <br/>
        {{ tt('portal.digital_employee.future.motto_line2') }}
      </h2>
      <p class="text-xl md:text-2xl text-slate-400 mb-16 max-w-3xl mx-auto font-light leading-relaxed">
        {{ tt('portal.digital_employee.future.motto_desc') }}
      </p>
      <button @click="openAction('join')" class="px-12 py-6 bg-[var(--matrix-color)] hover:bg-[var(--matrix-color)]/80 text-white rounded-3xl text-2xl font-black transition-all shadow-2xl shadow-[var(--matrix-color)]/40">
        {{ tt('portal.digital_employee.cta_primary') }}
      </button>
    </section>

    <!-- Action Modal -->
    <div v-if="showModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-[var(--bg-body)]/90 backdrop-blur-md" @click="showModal = false"></div>
      <div class="relative w-full max-w-lg bg-[var(--bg-body)] border border-[var(--border-color)] rounded-[3rem] p-10 shadow-2xl overflow-hidden">
        <div class="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-[var(--matrix-color)] to-[rgba(var(--matrix-color-rgb),0.5)]"></div>
        <button @click="showModal = false" class="absolute top-8 right-8 p-2 hover:bg-[var(--bg-body)]/80 rounded-full transition-colors">
          <X class="w-6 h-6 text-[var(--text-muted)]" />
        </button>
        
        <div class="text-center">
          <div class="w-20 h-20 bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] rounded-[2rem] flex items-center justify-center mx-auto mb-8 animate-pulse">
            <Zap class="w-10 h-10" />
          </div>
          <h3 class="text-3xl font-bold mb-6 text-[var(--text-main)]">{{ tt('portal.digital_employee.modal.title') }}</h3>
          <p class="text-[var(--text-muted)] mb-10 leading-relaxed text-lg">
            {{ tt('portal.digital_employee.modal.desc') }}
          </p>
          <div class="bg-[var(--bg-body)]/50 rounded-2xl p-6 mb-10 text-left border border-[var(--border-color)]">
            <div class="text-xs font-bold text-[var(--text-muted)] uppercase tracking-widest mb-4">{{ tt('portal.digital_employee.modal.priority_access') }}</div>
            <div class="flex items-center gap-4 text-[var(--matrix-color)] font-bold">
              <MessageCircle class="w-6 h-6" />
              <span>{{ tt('portal.digital_employee.modal.wechat') }}</span>
            </div>
          </div>
          <button @click="showModal = false" class="w-full py-5 bg-[var(--matrix-color)] hover:bg-[var(--matrix-color)]/80 text-white rounded-2xl font-black transition-all text-xl">
            {{ tt('portal.digital_employee.modal.accept') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { 
  BrainCircuit, Sparkles, ArrowRight, UserCheck, Zap, MessageCircle, X, Target, 
  ShieldCheck, Cpu, TrendingUp, Briefcase, Activity, Headphones, Calculator, 
  Calendar, ListTree, Eye, Lock, Network, Workflow, Monitor, Users, LayoutDashboard
} from 'lucide-vue-next';
import { ref, computed } from 'vue';
import { useI18n } from '@/utils/i18n';

const { t: tt } = useI18n();
const showModal = ref(false);

const openAction = (type: string) => {
  showModal.value = true;
};

const scrollTo = (id: string) => {
  const el = document.getElementById(id);
  if (el) {
    const headerOffset = 120; // Adjusted for sub-nav
    const elementPosition = el.getBoundingClientRect().top;
    const offsetPosition = elementPosition + window.pageYOffset - headerOffset;

    window.scrollTo({
      top: offsetPosition,
      behavior: 'smooth'
    });
  }
};

const visionPoints = computed(() => [
  {
    icon: Target,
    title: tt('portal.digital_employee.vision.point1_title'),
    desc: tt('portal.digital_employee.vision.point1_desc')
  },
  {
    icon: Briefcase,
    title: tt('portal.digital_employee.vision.point2_title'),
    desc: tt('portal.digital_employee.vision.point2_desc')
  },
  {
    icon: TrendingUp,
    title: tt('portal.digital_employee.vision.point3_title'),
    desc: tt('portal.digital_employee.vision.point3_desc')
  }
]);

const revolutionSteps = computed(() => [
  {
    icon: Network,
    title: tt('portal.digital_employee.revolution.step1_title'),
    desc: tt('portal.digital_employee.revolution.step1_desc')
  },
  {
    icon: Workflow,
    title: tt('portal.digital_employee.revolution.step2_title'),
    desc: tt('portal.digital_employee.revolution.step2_desc')
  },
  {
    icon: Cpu,
    title: tt('portal.digital_employee.revolution.step3_title'),
    desc: tt('portal.digital_employee.revolution.step3_desc')
  }
]);

const hrStats = computed(() => [
  { label: tt('portal.digital_employee.hr.stat1'), value: '420%' },
  { label: tt('portal.digital_employee.hr.stat2'), value: '1.2B' },
  { label: tt('portal.digital_employee.hr.stat3'), value: '50M+' },
  { label: tt('portal.digital_employee.hr.stat4'), value: '99.99%' }
]);

const roles = computed(() => [
  { name: 'AX-709', dept: tt('portal.digital_employee.hr.dept_support'), icon: Headphones, progress: 98, skills: tt('portal.digital_employee.jobs.job1_desc') },
  { name: 'BX-102', dept: tt('portal.digital_employee.hr.dept_finance'), icon: Calculator, progress: 85, skills: tt('portal.digital_employee.jobs.job2_desc') },
  { name: 'CX-443', dept: tt('portal.digital_employee.hr.dept_admin'), icon: Calendar, progress: 92, skills: tt('portal.digital_employee.jobs.job3_desc') }
]);

const industryJobs = computed(() => [
  {
    icon: Headphones,
    title: tt('portal.digital_employee.jobs.job1_title'),
    desc: tt('portal.digital_employee.jobs.job1_desc'),
    skills: [tt('portal.digital_employee.jobs.skill_search'), tt('portal.digital_employee.jobs.skill_routing'), tt('portal.digital_employee.jobs.skill_multilingual')]
  },
  {
    icon: Calculator,
    title: tt('portal.digital_employee.jobs.job2_title'),
    desc: tt('portal.digital_employee.jobs.job2_desc'),
    skills: [tt('portal.digital_employee.jobs.skill_audit'), tt('portal.digital_employee.jobs.skill_tax'), tt('portal.digital_employee.jobs.skill_erp')]
  },
  {
    icon: Monitor,
    title: tt('portal.digital_employee.jobs.job3_title'),
    desc: tt('portal.digital_employee.jobs.job3_desc'),
    skills: [tt('portal.digital_employee.jobs.skill_ai_review'), tt('portal.digital_employee.jobs.skill_cicd')]
  }
]);

const securityFeatures = computed(() => [
  {
    icon: Lock,
    title: tt('portal.digital_employee.security.f1_title'),
    desc: tt('portal.digital_employee.security.f1_desc')
  },
  {
    icon: Eye,
    title: tt('portal.digital_employee.security.f2_title'),
    desc: tt('portal.digital_employee.security.f2_desc')
  }
]);

const collaborationTraces = computed(() => [
  { id: 'TR-8821', from: 'Finance-Bot', to: 'CEO-Assistant', action: tt('portal.digital_employee.collaboration.trace_action1'), status: tt('portal.digital_employee.collaboration.trace_status_pending'), time: tt('portal.digital_employee.collaboration.trace_time_2m'), icon: Activity, color: 'text-yellow-500' },
  { id: 'TR-8820', from: 'Dev-Agent', to: 'QA-Bot', action: tt('portal.digital_employee.collaboration.trace_action2'), status: tt('portal.digital_employee.collaboration.trace_status_success'), time: tt('portal.digital_employee.collaboration.trace_time_15m'), icon: ShieldCheck, color: 'text-green-500' },
  { id: 'TR-8819', from: 'Support-01', to: 'Knowledge-Base', action: tt('portal.digital_employee.collaboration.trace_action3'), status: tt('portal.digital_employee.collaboration.trace_status_processing'), time: tt('portal.digital_employee.collaboration.trace_time_1h'), icon: Cpu, color: 'text-[var(--matrix-color)]' }
]);

const futuristicFeatures = computed(() => [
  {
    tag: tt('portal.digital_employee.future.f1_tag'),
    icon: Monitor,
    title: tt('portal.digital_employee.future.f1_title'),
    desc: tt('portal.digital_employee.future.f1_desc')
  },
  {
    tag: tt('portal.digital_employee.future.f2_tag'),
    icon: Users,
    title: tt('portal.digital_employee.future.f2_title'),
    desc: tt('portal.digital_employee.future.f2_desc')
  }
]);
</script>

<style scoped>
.matrix-lines {
  background: linear-gradient(0deg, rgba(var(--matrix-color-rgb), 0.2) 0%, transparent 100%);
  background-size: 100% 20px;
  animation: matrix-scroll 20s linear infinite;
}

@keyframes matrix-scroll {
  0% { background-position: 0 0; }
  100% { background-position: 0 1000px; }
}

.animate-bounce-slow {
  animation: bounce 4s infinite ease-in-out;
}

.animate-spin-slow {
  animation: spin 12s linear infinite;
}

@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-20px); }
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.reveal {
  animation: reveal 1s cubic-bezier(0.23, 1, 0.32, 1) forwards;
}

@keyframes reveal {
  from { opacity: 0; transform: translateY(40px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>

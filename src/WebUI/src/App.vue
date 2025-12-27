<script setup lang="ts">
import { onMounted, computed, watch, ref } from 'vue';
import { useRoute } from 'vue-router';
import { useSystemStore } from '@/stores/system';
import Sidebar from '@/components/layout/Sidebar.vue';
import Header from '@/components/layout/Header.vue';
import MatrixRain from '@/components/common/MatrixRain.vue';
import KawaiiSparkles from '@/components/common/KawaiiSparkles.vue';

const systemStore = useSystemStore();
const route = useRoute();

const isBlankLayout = computed(() => route.meta.layout === 'blank');

// Audio states
const isMuted = ref(localStorage.getItem('matrix_muted') === 'true');
const showAudioSettings = ref(false);
const availableVoices = ref<SpeechSynthesisVoice[]>([]);
const ttsSettings = ref({
  voiceURI: localStorage.getItem('matrix_tts_voice') || '',
  pitch: parseFloat(localStorage.getItem('matrix_tts_pitch') || '0.8'),
  rate: parseFloat(localStorage.getItem('matrix_tts_rate') || '1.2'),
  volume: parseFloat(localStorage.getItem('matrix_tts_volume') || '1.0')
});

let ambientAudio: HTMLAudioElement | null = null;
let audioInitialized = false;
let sharedAudioCtx: AudioContext | null = null;

const getAudioCtx = () => {
  if (typeof window === 'undefined') return null;
  if (!sharedAudioCtx) {
    try {
      const AudioContextClass = (window.AudioContext || (window as any).webkitAudioContext);
      sharedAudioCtx = new AudioContextClass();
    } catch (e) {
      console.warn('AudioContext not supported');
    }
  }
  return sharedAudioCtx;
};

const toggleMute = () => {
  isMuted.value = !isMuted.value;
  localStorage.setItem('matrix_muted', isMuted.value.toString());
  
  if (isMuted.value) {
    if (ambientAudio) ambientAudio.pause();
  } else {
    audioInitialized = false;
    initAndPlayAmbient();
  }
};

// Audio assets (using more reliable public URLs)
const MATRIX_AMBIENT_URL = 'https://www.soundjay.com/nature/sounds/rain-01.mp3';
const MATRIX_CLICK_URL = 'https://actions.google.com/sounds/v1/foley/button_click_on.ogg';
const MATRIX_HOVER_URL = 'https://actions.google.com/sounds/v1/foley/draw_knife_from_sheath.ogg';

// Synthesized fallback beep
const playSynthBeep = (freq = 880, duration = 0.1, type: OscillatorType = 'square') => {
  if (isMuted.value) return;
  try {
    const ctx = getAudioCtx();
    if (!ctx) return;
    
    // Auto-resume if suspended (only if already allowed)
    if (ctx.state === 'suspended') {
      ctx.resume().catch(() => {});
    }

    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    osc.type = type;
    osc.frequency.setValueAtTime(freq, ctx.currentTime);
    gain.gain.setValueAtTime(0.1, ctx.currentTime);
    gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + duration);
    osc.connect(gain);
    gain.connect(ctx.destination);
    osc.start();
    osc.stop(ctx.currentTime + duration);
  } catch (e) {
    // Silently fail to avoid console clutter if still blocked
  }
};

const playClick = () => {
  if (systemStore.style === 'matrix' && !isMuted.value) {
    const audio = new Audio(MATRIX_CLICK_URL);
    audio.volume = 0.4;
    audio.play().catch(() => playSynthBeep(880, 0.1, 'square'));
  }
};

const playHover = (e: MouseEvent) => {
  if (systemStore.style === 'matrix' && !isMuted.value) {
    const target = e.target as HTMLElement;
    const clickable = target.closest('a, button, .cursor-pointer, li');
    if (clickable) {
      const audio = new Audio(MATRIX_HOVER_URL);
      audio.volume = 0.15;
      audio.play().catch(() => playSynthBeep(440, 0.05, 'sine'));
      
      // Debounce TTS to avoid rapid-fire errors
      if ((window as any).ttsTimer) clearTimeout((window as any).ttsTimer);
      (window as any).ttsTimer = setTimeout(() => {
        speakText(clickable.textContent || '');
      }, 50);
    }
  }
};

const loadVoices = () => {
  const getVoices = () => {
    let voices = window.speechSynthesis.getVoices();
    if (voices.length > 0) {
      availableVoices.value = voices;
      if (!ttsSettings.value.voiceURI) {
        const defaultVoice = voices.find(v => v.lang.startsWith('en') && (v.name.includes('David') || v.name.includes('Google US'))) 
                          || voices.find(v => v.lang.startsWith('zh'))
                          || voices[0];
        if (defaultVoice) ttsSettings.value.voiceURI = defaultVoice.voiceURI;
      }
      return true;
    }
    return false;
  };

  if (!getVoices()) {
    // Retry with exponential backoff or simple interval
    let attempts = 0;
    const interval = setInterval(() => {
      attempts++;
      if (getVoices() || attempts > 10) {
        clearInterval(interval);
      }
    }, 100);
  }
};

// Web Speech API Integration
const speakText = (text: string, force = false) => {
  if (systemStore.style !== 'matrix' || !text.trim()) return;
  if (isMuted.value && !force) return;

  try {
    // 1. Force clear and resume synthesis state
    window.speechSynthesis.cancel();
    
    // Resume AudioContext if it exists (for browsers that link synthesis to audio context)
    try {
      const ctx = getAudioCtx();
      if (ctx && ctx.state === 'suspended') ctx.resume().catch(() => {});
    } catch (e) {}

    if (window.speechSynthesis.paused) window.speechSynthesis.resume();
    
    const cleanText = text.replace(/[^\w\s\u4e00-\u9fa5]/gi, '').trim();
    if (!cleanText) return;

    // 2. Initial attempt with full settings
    const attemptSpeech = (useSettings = true) => {
      const utterance = new SpeechSynthesisUtterance(cleanText);
      
      if (useSettings) {
        if (availableVoices.value.length === 0) {
          availableVoices.value = window.speechSynthesis.getVoices();
        }

        let voice = availableVoices.value.find(v => v.voiceURI === ttsSettings.value.voiceURI);
        if (!voice) {
          voice = availableVoices.value.find(v => v.lang.includes('zh')) || 
                  availableVoices.value.find(v => v.lang.includes('en')) || 
                  availableVoices.value[0];
        }
        
        if (voice) {
          utterance.voice = voice;
          utterance.lang = voice.lang;
        }

        utterance.pitch = ttsSettings.value.pitch;
        utterance.rate = ttsSettings.value.rate;
        utterance.volume = ttsSettings.value.volume;
      }

      utterance.onerror = (e) => {
        console.warn('TTS Attempt Failed (Settings:' + useSettings + '):', e.error);
        // 3. Fallback: If full settings failed, try a "naked" utterance
        if (useSettings) {
          console.log('Retrying with safe-mode fallback...');
          setTimeout(() => attemptSpeech(false), 100);
        }
      };

      window.speechSynthesis.speak(utterance);
    };

    // Start the process with a small delay to ensure cancel() took effect
    setTimeout(() => attemptSpeech(true), 50);

  } catch (e) {
    console.error('TTS execution failed:', e);
  }
};

const saveTTSSettings = () => {
  localStorage.setItem('matrix_tts_voice', ttsSettings.value.voiceURI);
  localStorage.setItem('matrix_tts_pitch', ttsSettings.value.pitch.toString());
  localStorage.setItem('matrix_tts_rate', ttsSettings.value.rate.toString());
  localStorage.setItem('matrix_tts_volume', ttsSettings.value.volume.toString());
  speakText('Voice settings synchronization complete', true);
};

const testVoice = () => {
  speakText('System voice diagnostic sequence initiated. Audio transmission operational.', true);
};

// Pre-load voices
if (typeof window !== 'undefined' && window.speechSynthesis) {
  loadVoices();
  if (speechSynthesis.onvoiceschanged !== undefined) {
    speechSynthesis.onvoiceschanged = loadVoices;
  }
}

const initAndPlayAmbient = () => {
  if (systemStore.style === 'matrix' && !audioInitialized && !isMuted.value) {
    // 1. Force Resume AudioContext (User Gesture)
    const ctx = getAudioCtx();
    if (ctx && ctx.state === 'suspended') {
      ctx.resume().catch(e => console.warn('Ctx resume failed', e));
    }

    // 2. Unlock Speech Synthesis
    try {
      const unlockUtterance = new SpeechSynthesisUtterance('');
      unlockUtterance.volume = 0;
      window.speechSynthesis.speak(unlockUtterance);
    } catch (e) {}

    if (!ambientAudio) {
      ambientAudio = new Audio(MATRIX_AMBIENT_URL);
      ambientAudio.loop = true;
      ambientAudio.volume = 0.35;
    }
    ambientAudio.play().then(() => {
      audioInitialized = true;
      sessionStorage.setItem('matrix_audio_authorized', 'true');
    }).catch(() => {
      // Still blocked
    });
  }
};

// Check for existing authorization on mount or refresh
const checkStoredAuth = () => {
  if (systemStore.style === 'matrix' && sessionStorage.getItem('matrix_audio_authorized') === 'true') {
    initAndPlayAmbient();
  }
};

watch(() => systemStore.style, (newStyle) => {
  if (newStyle === 'matrix') {
    audioInitialized = false;
    checkStoredAuth(); // Try to resume if already authorized in this session
    window.addEventListener('click', playClick);
    window.addEventListener('mouseover', playHover);
    window.addEventListener('click', initAndPlayAmbient);
  } else {
    if (ambientAudio) {
      ambientAudio.pause();
    }
    sessionStorage.removeItem('matrix_audio_authorized');
    window.removeEventListener('click', playClick);
    window.removeEventListener('mouseover', playHover);
    window.removeEventListener('click', initAndPlayAmbient);
  }
}, { immediate: true });

onMounted(() => {
  // Check auth again after mount to ensure DOM is ready
  setTimeout(checkStoredAuth, 1000);
  
  // Update time every second
  setInterval(() => {
    systemStore.updateTime();
  }, 1000);

  // Initial theme check
  systemStore.initTheme();
});
</script>

<template>
  <!-- Style Specific Backgrounds -->
  <div v-if="systemStore.style === 'matrix'" class="matrix-scanlines"></div>
  <MatrixRain v-if="systemStore.style === 'matrix'" />
  <KawaiiSparkles v-if="systemStore.style === 'kawaii'" />

  <!-- Matrix Audio Control Panel -->
  <div v-if="systemStore.style === 'matrix'" 
       class="fixed bottom-6 right-6 z-[10001] flex flex-col items-end gap-3">
    
    <!-- TTS Settings Panel -->
    <transition name="fade">
      <div v-if="showAudioSettings" 
           class="matrix-btn p-4 rounded-lg w-72 mb-2 backdrop-blur-xl border-2 border-[#00ff41] shadow-[0_0_20px_rgba(0,255,65,0.3)]">
        <div class="flex justify-between items-center mb-4">
          <h3 class="text-[#00ff41] font-bold tracking-widest text-sm uppercase">Voice Matrix</h3>
          <button @click="showAudioSettings = false" class="text-[#00ff41] hover:scale-110 transition-transform">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- Voice Selection -->
          <div class="space-y-4 text-xs">
            <div class="flex flex-col gap-1">
              <label class="text-[#00ff41]/70 uppercase tracking-tighter">
                Voice Entity 
                <span v-if="availableVoices.length === 0" class="text-red-500 animate-pulse ml-2">(LOADING...)</span>
              </label>
              <select v-model="ttsSettings.voiceURI" 
                      class="bg-black/80 border border-[#00ff41] text-[#00ff41] p-1 rounded outline-none text-[10px]">
                <option v-if="availableVoices.length === 0" value="">No voices detected</option>
                <option v-for="voice in availableVoices" :key="voice.voiceURI" :value="voice.voiceURI">
                  {{ voice.name }} ({{ voice.lang }})
                </option>
              </select>
            </div>

          <!-- Pitch Slider -->
          <div class="flex flex-col gap-1">
            <div class="flex justify-between">
              <label class="text-[#00ff41]/70 uppercase tracking-tighter">Frequency (Pitch)</label>
              <span class="text-[#00ff41]">{{ ttsSettings.pitch }}</span>
            </div>
            <input type="range" v-model.number="ttsSettings.pitch" min="0" max="2" step="0.1" 
                   class="accent-[#00ff41] bg-black/50">
          </div>

          <!-- Rate Slider -->
          <div class="flex flex-col gap-1">
            <div class="flex justify-between">
              <label class="text-[#00ff41]/70 uppercase tracking-tighter">Transmission (Rate)</label>
              <span class="text-[#00ff41]">{{ ttsSettings.rate }}</span>
            </div>
            <input type="range" v-model.number="ttsSettings.rate" min="0.1" max="3" step="0.1" 
                   class="accent-[#00ff41] bg-black/50">
          </div>

          <!-- Volume Slider -->
          <div class="flex flex-col gap-1">
            <div class="flex justify-between">
              <label class="text-[#00ff41]/70 uppercase tracking-tighter">Amplitude (Volume)</label>
              <span class="text-[#00ff41]">{{ Math.round(ttsSettings.volume * 100) }}%</span>
            </div>
            <input type="range" v-model.number="ttsSettings.volume" min="0" max="1" step="0.1" 
                   class="accent-[#00ff41] bg-black/50">
          </div>

          <div class="pt-2 flex gap-2">
            <button @click="testVoice" 
                    class="flex-1 border border-[#00ff41] text-[#00ff41] py-1 rounded hover:bg-[#00ff41]/20 transition-colors uppercase tracking-widest font-bold">
              Test
            </button>
            <button @click="saveTTSSettings" 
                    class="flex-1 bg-[#00ff41] text-black py-1 rounded hover:bg-[#00ff41]/80 transition-colors uppercase tracking-widest font-bold">
              Save
            </button>
          </div>
        </div>
      </div>
    </transition>

    <!-- Main Controls -->
    <div class="flex items-center gap-3">
      <div v-if="!audioInitialized && !isMuted" 
           class="bg-black border border-[#00ff41] px-3 py-1 text-[#00ff41] text-xs animate-pulse">
        CLICK TO INITIALIZE AUDIO
      </div>
      
      <div class="flex gap-2">
        <button @click="showAudioSettings = !showAudioSettings" 
                class="matrix-btn p-3 rounded-full flex items-center justify-center group border-[#00ff41]/50"
                title="Voice Settings">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-[#00ff41]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
          </svg>
        </button>

        <button @click="toggleMute" 
                class="matrix-btn p-3 rounded-full flex items-center justify-center group min-w-[120px]"
                :title="isMuted ? 'Unmute Matrix' : 'Mute Matrix'">
          <div v-if="isMuted" class="flex items-center gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" />
            </svg>
            <span class="text-xs font-bold mr-1">AUDIO OFF</span>
          </div>
          <div v-else class="flex items-center gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.536 8.464a5 5 0 010 7.072m2.828-9.9a9 9 0 010 12.728M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" />
            </svg>
            <span class="text-xs font-bold mr-1">AUDIO ON</span>
          </div>
        </button>
      </div>
    </div>
  </div>

  <div v-if="isBlankLayout" class="min-h-screen relative z-10">
    <router-view />
  </div>
  <div v-else class="flex h-screen bg-[var(--bg-body)] text-[var(--text-main)] overflow-hidden transition-colors duration-300 relative z-10">
    <!-- Sidebar -->
    <Sidebar />

    <!-- Main Content -->
    <div class="flex-1 flex flex-col min-w-0 overflow-hidden bg-[var(--bg-body)] transition-colors duration-300">
      <!-- Header -->
      <Header />

      <!-- Page Content -->
      <main class="flex-1 overflow-y-auto custom-scrollbar">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </main>
    </div>
  </div>
</template>

<style>
:root {
  --matrix-color: #00ff41;
}

.mono {
  font-family: 'JetBrains Mono', 'Fira Code', 'Courier New', monospace;
}

.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(0, 255, 65, 0.1);
  border-radius: 10px;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 255, 65, 0.2);
}

.fade-enter-active,
.fade-leave-active {
  transition: all 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(5px);
}
</style>

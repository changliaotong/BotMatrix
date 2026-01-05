<script setup lang="ts">
import { ref, onMounted, reactive, computed, nextTick, onUnmounted, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useSystemStore } from '@/stores/system';
import { 
  Sparkles, 
  Plus, 
  Settings2, 
  Trash2, 
  Send, 
  MessageSquare, 
  Brain, 
  Server, 
  Layers,
  Search,
  Loader2,
  ChevronRight,
  User,
  Bot as BotIcon,
  X,
  Edit3,
  Terminal,
  RefreshCw,
  Cpu,
  History,
  Share2,
  Volume2,
  VolumeX,
  Pause,
  Play,
  Wand2,
  ThumbsUp,
  ThumbsDown,
  RotateCcw,
  Copy,
  Check,
  Activity,
  FileUp,
  FileText,
  Database
} from 'lucide-vue-next';
import { aiApi, type AIAgent, type AISession, type AIChatMessage } from '@/api/ai';

const systemStore = useSystemStore();
const route = useRoute();
const t = (key: string) => systemStore.t(key);

// --- State ---
const activeTab = ref('sessions'); // sessions, agents, models, providers
const isLoading = ref(false);
const searchQuery = ref('');
const errorMessage = ref('');

// Data
interface AIProvider {
  id: number;
  name: string;
  type: string;
  base_url: string;
  api_key: string;
}

interface AIModel {
  id: number;
  provider_id: number;
  model_id: string;
  model_name: string;
  context_size: number;
}

const providers = ref<AIProvider[]>([]);
const models = ref<AIModel[]>([]);
const agents = ref<AIAgent[]>([]);
const sessions = ref<AISession[]>([]);
const knowledgeDocs = ref<any[]>([]);

// Selection & Editing
const selectedAgent = ref<AIAgent | null>(null);
const selectedSession = ref<AISession | null>(null);
const agentDetails = ref<Record<number, AIAgent>>({});
const showAgentModal = ref(false);
const showModelModal = ref(false);
const showProviderModal = ref(false);
const editingAgent = reactive<Partial<AIAgent>>({});
const editingModel = reactive<Partial<AIModel>>({});
const editingProvider = reactive<Partial<AIProvider>>({});

// Chat Trial
const chatHistories = ref<Record<string, (AIChatMessage | { role: string; content: string })[]>>({});
const currentSessionId = ref<string | null>(null);

const chatMessages = computed({
  get: () => {
    if (currentSessionId.value) {
      return chatHistories.value[currentSessionId.value] || [];
    }
    if (selectedAgent.value) {
      return chatHistories.value[`agent_${selectedAgent.value.id}`] || [];
    }
    return [];
  },
  set: (val) => {
    if (currentSessionId.value) {
      chatHistories.value[currentSessionId.value] = val;
    } else if (selectedAgent.value) {
      chatHistories.value[`agent_${selectedAgent.value.id}`] = val;
    }
  }
});
const userInput = ref('');
const isGenerating = ref(false);
const isLoadingHistory = ref(false);
const contextLength = ref(5);
const contextOptions = [
  { label: '1 条', value: 1 },
  { label: '3 条', value: 3 },
  { label: '5 条', value: 5 },
  { label: '10 条', value: 10 },
  { label: '20 条', value: 20 }
];
const hasMoreHistory = ref<Record<string, boolean>>({});
const scrollContainer = ref<HTMLElement | null>(null);

const scrollToBottom = () => {
  if (scrollContainer.value) {
    scrollContainer.value.scrollTop = scrollContainer.value.scrollHeight;
  }
};

// --- History Fetching ---
const fetchHistory = async (id: string | number, isSession: boolean = false, beforeId?: number) => {
  if (isLoadingHistory.value) return;
  isLoadingHistory.value = true;

  try {
    const historyKey = isSession ? (id as string) : `agent_${id}`;
    let messages: AIChatMessage[] = [];
    
    if (isSession) {
      const res = await aiApi.getChatHistory(id as string, beforeId);
      const data = res.data;
      if (data.success) {
        messages = data.data || [];
      }
    } else {
      const url = `/api/admin/ai/chat/history?agent_id=${id}${beforeId ? `&before_id=${beforeId}` : ''}&limit=20`;
      const res = await fetch(url, { headers: getHeaders() });
      const data = await res.json();
      if (data.success) {
        messages = data.data || [];
      }
    }

    if (!chatHistories.value[historyKey]) {
      chatHistories.value[historyKey] = [];
    }

    if (beforeId) {
      // Prepend history and deduplicate
      const existingIds = new Set(chatHistories.value[historyKey].map((m: any) => m.id));
      const newMessages = messages.filter((m: any) => !existingIds.has(m.id));
      
      if (newMessages.length > 0) {
        chatHistories.value[historyKey] = [...newMessages, ...chatHistories.value[historyKey]];
      } else if (messages.length > 0) {
        // If we got messages but they are all duplicates, we should still try to find older ones
        // but to avoid infinite loop, if we get exactly what we asked for and it's all duplicates,
        // it might mean something is wrong. However, usually the backend's before_id should prevent this.
        // For safety, if we get 0 new messages, we stop trying for this session/agent until refresh.
        hasMoreHistory.value[historyKey] = false;
      }
    } else {
      // Initial load
      chatHistories.value[historyKey] = messages;
    }

    // Update hasMoreHistory based on whether we reached the limit
    if (messages.length < 20) {
      hasMoreHistory.value[historyKey] = false;
    } else if (!beforeId) {
      // If initial load and we got exactly 20, assume there might be more
      hasMoreHistory.value[historyKey] = true;
    }

    // Handle scroll position when prepending
    if (beforeId && scrollContainer.value) {
      const container = scrollContainer.value;
      const oldHeight = container.scrollHeight;
      await nextTick();
      container.scrollTop = container.scrollHeight - oldHeight;
    } else {
      await nextTick(scrollToBottom);
    }
  } catch (error) {
    console.error('Failed to fetch history:', error);
  } finally {
    isLoadingHistory.value = false;
  }
};

const handleScroll = async () => {
  if (!scrollContainer.value || isLoadingHistory.value) return;

  const historyKey = currentSessionId.value || (selectedAgent.value ? `agent_${selectedAgent.value.id}` : null);
  if (!historyKey || !hasMoreHistory.value[historyKey]) return;

  const container = scrollContainer.value;
  // Trigger when near top (threshold 50px)
  if (container.scrollTop <= 50) {
    const currentMessages = chatHistories.value[historyKey] || [];
    if (currentMessages.length > 0) {
      const firstMsg = currentMessages.find((m: any) => m.id);
      const firstMsgId = firstMsg ? (firstMsg as any).id : null;
      
      if (firstMsgId) {
        if (currentSessionId.value) {
          await fetchHistory(currentSessionId.value, true, firstMsgId);
        } else if (selectedAgent.value) {
          await fetchHistory(selectedAgent.value.id, false, firstMsgId);
        }
      }
    }
  }
};

// --- Built-in Agent IDs ---
const BUILTIN_AGENTS = {
  TITLE_GENERATOR: 1, // 会话标题生成
  PROMPT_GENERATOR: 2, // 提示词生成
  DESC_GENERATOR: 3 // 简介生成
};

const chatStyle = ref<'default' | 'wechat'>(systemStore.style === 'classic' ? 'wechat' : 'default');

// Watch for style changes to update chat style default
watch(() => systemStore.style, (newStyle) => {
  if (newStyle === 'classic') {
    chatStyle.value = 'wechat';
  } else {
    chatStyle.value = 'default';
  }
});

// --- TTS Function ---
const isSpeaking = ref(false);
const isAutoVoiceEnabled = ref(localStorage.getItem('auto_voice_enabled') === 'true');
const highlightedSentence = ref('');
const speakingMsgId = ref<string | number | null>(null);
const synth = window.speechSynthesis;
let currentUtterance: SpeechSynthesisUtterance | null = null; // 保持引用防止被垃圾回收销毁导致无声
const voices = ref<SpeechSynthesisVoice[]>([]);

const updateVoices = () => {
  const allVoices = synth.getVoices();
  voices.value = allVoices;
  console.log('[TTS DEBUG] Available Voices List:');
  console.table(allVoices.map(v => ({
    name: v.name,
    lang: v.lang,
    uri: v.voiceURI
  })));
};

if (synth.onvoiceschanged !== undefined) {
  synth.onvoiceschanged = updateVoices;
}
updateVoices();

const sentenceEndChars = ['。', '！', '？', '；', '，', '：', '.', '!', '?', ';', ',', ':', '\n'];
const MAX_SENTENCE_LENGTH = 30; // 超过30个字强制切分，提升播报响应速度和高亮精度
let sentenceQueue: string[] = []; 
let sentenceBuffer = '';
let currentVoiceId: string | undefined = undefined;
let currentVoiceName: string | undefined = undefined;
let currentVoiceLang: string | undefined = undefined;
let currentVoiceRate: number = 1.0;

const toggleAutoVoice = () => {
  isAutoVoiceEnabled.value = !isAutoVoiceEnabled.value;
  localStorage.setItem('auto_voice_enabled', isAutoVoiceEnabled.value.toString());
};

const isPaused = ref(false);

const stopSpeaking = () => {
  console.log('[TTS DEBUG] Stopping speech...');
  sentenceQueue = [];
  if (synth.speaking || synth.pending) {
    synth.cancel();
  }
  isSpeaking.value = false;
  isPaused.value = false;
  highlightedSentence.value = '';
  speakingMsgId.value = null;
};

const pauseSpeaking = () => {
  if (synth.speaking && !synth.paused) {
    console.log('[TTS DEBUG] Pausing speech...');
    synth.pause();
    isPaused.value = true;
  }
};

const resumeSpeaking = () => {
  if (synth.paused) {
    console.log('[TTS DEBUG] Resuming speech...');
    synth.resume();
    isPaused.value = false;
  }
};

// 获取可用语音列表，处理 voices 为空的情况
const getAvailableVoices = (): Promise<SpeechSynthesisVoice[]> => {
  return new Promise((resolve) => {
    let currentVoices = synth.getVoices();
    if (currentVoices.length > 0) {
      console.log(`[TTS DEBUG] getAvailableVoices found ${currentVoices.length} voices`);
      resolve(currentVoices);
    } else {
      console.log('[TTS DEBUG] getAvailableVoices: No voices yet, waiting for voiceschanged...');
      const handler = () => {
        currentVoices = synth.getVoices();
        console.log(`[TTS DEBUG] getAvailableVoices (after event) found ${currentVoices.length} voices`);
        resolve(currentVoices);
        synth.removeEventListener('voiceschanged', handler);
      };
      synth.addEventListener('voiceschanged', handler);
      
      // 备选方案：如果500ms还没触发，直接返回当前列表
      setTimeout(() => {
        currentVoices = synth.getVoices();
        resolve(currentVoices);
      }, 500);
    }
  });
};

const speak = async (text: string, voiceName?: string, voiceLang?: string, rate?: number, msgId?: string | number, highlightText?: string) => {
  if (!text || !text.trim()) {
    console.log('[TTS DEBUG] Empty text, skipping');
    return;
  }

  // 优先级：参数传入 > 智能体配置 > 默认值
  const finalVoiceName = voiceName || selectedAgent.value?.voice_name;
  const finalVoiceLang = voiceLang || selectedAgent.value?.voice_lang;
  const finalRate = (rate !== undefined && rate > 0) ? rate : (selectedAgent.value?.voice_rate || 1.0);

  // 如果提供了 msgId，说明是点击播放或者开始新的消息播报
  if (msgId !== undefined) {
    console.log(`[TTS DEBUG] New session start, msgId: ${msgId}`);
    if (speakingMsgId.value !== msgId) {
      stopSpeaking();
      speakingMsgId.value = msgId;
    }
    
    // 初始化配置
    currentVoiceName = finalVoiceName;
    currentVoiceLang = finalVoiceLang;
    currentVoiceRate = finalRate;

    // 处理分句逻辑
    const hasEndChar = sentenceEndChars.some(char => text.includes(char));
    if (hasEndChar && !isGenerating.value) {
      // 非流式生成时，直接按标点切分并放入队列
      sentenceQueue = text.split(/([。！？；，：\.\!\?\;\,：\n])/g)
        .filter(s => s.trim().length > 0)
        .reduce((acc: string[], curr, i, arr) => {
          if (i % 2 === 0) {
            acc.push(curr + (arr[i+1] || ''));
          }
          return acc;
        }, []);
      
      console.log(`[TTS DEBUG] Text split into ${sentenceQueue.length} sentences`);
      playNextSentence();
      return;
    }
  }

  // 如果不是新的播报请求，而是分句播放，则直接调用底层播放
  const currentVoices = await getAvailableVoices();
  const utterance = new SpeechSynthesisUtterance(text);
  currentUtterance = utterance; // 关键：保存到外部变量，防止被 GC
  
  utterance.rate = finalRate;
  utterance.volume = 1.0; // 显式设置音量为 1.0
  utterance.pitch = 1.0;  // 显式设置音调为 1.0
  
  // 1. 自动选择匹配语音 (增强匹配逻辑：先语言后名称)
  let selectedVoice: SpeechSynthesisVoice | undefined = undefined;
  
  const normalize = (s: string | undefined | null) => {
    if (!s) return '';
    // 关键修复：处理不可见字符（如 \u00A0）和多余空格
    return s.toLowerCase()
      .trim()
      .replace(/[\u00A0\s]+/g, '') // 替换所有空格（包括 &nbsp;）
      .replace(/[_-]/g, '');       // 忽略下划线和连字符
  };

  const getLangCode = (l: string) => {
    const n = normalize(l);
    return n.substring(0, 2);
  };

  if (finalVoiceLang && finalVoiceLang.trim()) {
    const targetLangNorm = normalize(finalVoiceLang);
    const targetLangCode = getLangCode(finalVoiceLang);
    console.log(`[TTS DEBUG] Target Lang: "${finalVoiceLang}" (Norm: "${targetLangNorm}", Code: "${targetLangCode}")`);
    
    // 过滤出匹配该语言的语音 (优先完全匹配，其次前缀匹配，最后语言代码匹配)
    let langMatchedVoices = currentVoices.filter(v => normalize(v.lang) === targetLangNorm);
    if (langMatchedVoices.length === 0) {
      langMatchedVoices = currentVoices.filter(v => normalize(v.lang).startsWith(targetLangNorm) || targetLangNorm.startsWith(normalize(v.lang)));
    }
    if (langMatchedVoices.length === 0) {
      langMatchedVoices = currentVoices.filter(v => getLangCode(v.lang) === targetLangCode);
    }
    
    if (langMatchedVoices.length > 0) {
      console.log(`[TTS DEBUG] Found ${langMatchedVoices.length} voices for language group: ${finalVoiceLang}`);
      
      if (finalVoiceName && finalVoiceName.trim()) {
        const targetNameNorm = normalize(finalVoiceName);
        console.log(`[TTS DEBUG] Target Name (Normalized): "${targetNameNorm}"`);
        // 在语言匹配的基础上，寻找名称匹配的语音
        selectedVoice = langMatchedVoices.find(v => normalize(v.name).includes(targetNameNorm) || normalize(v.voiceURI).includes(targetNameNorm));
        
        if (selectedVoice) {
          console.log(`[TTS DEBUG] Found specific voice in language group: ${selectedVoice.name}`);
        } else {
          console.log(`[TTS DEBUG] No specific voice name "${finalVoiceName}" found in language group, using first available: ${langMatchedVoices[0].name}`);
          selectedVoice = langMatchedVoices[0];
        }
      } else {
        selectedVoice = langMatchedVoices[0];
      }
    }
  }

  // 如果按语言没找到，尝试全局按名称匹配
  if (!selectedVoice && finalVoiceName && finalVoiceName.trim()) {
    const targetNameNorm = normalize(finalVoiceName);
    selectedVoice = currentVoices.find(v => normalize(v.name).includes(targetNameNorm) || normalize(v.voiceURI).includes(targetNameNorm));
    if (selectedVoice) {
      console.log(`[TTS DEBUG] Found voice by global name match: ${selectedVoice.name}`);
    }
  }

  // 2. 语言自动识别与回退逻辑
  const hasChinese = /[\u4e00-\u9fa5]/.test(text);
  const hasJapanese = /[\u3040-\u30FF]/.test(text);
  const hasKorean = /[\uac00-\ud7af]/.test(text);
  const hasEnglish = /[a-zA-Z]/.test(text);

  if (selectedVoice) {
    const lang = selectedVoice.lang.toLowerCase();
    const voiceLangCode = lang.split(/[-_]/)[0];
    
    // 如果选择的语音与文本主要语言不匹配，进行回退
    let needsFallback = false;
    if (hasChinese && voiceLangCode !== 'zh') needsFallback = true;
    else if (hasJapanese && voiceLangCode !== 'ja') needsFallback = true;
    else if (hasKorean && voiceLangCode !== 'ko') needsFallback = true;
    // 如果只有英文且当前语音不是英文，也考虑回退
    else if (hasEnglish && !hasChinese && !hasJapanese && !hasKorean && voiceLangCode !== 'en') needsFallback = true;

    if (needsFallback) {
      console.log(`[TTS DEBUG] Voice language mismatch (${lang}), finding fallback for text...`);
      let targetLang = 'zh';
      if (hasJapanese) targetLang = 'ja';
      else if (hasKorean) targetLang = 'ko';
      else if (hasEnglish) targetLang = 'en';
      
      const fallbackVoice = currentVoices.find(v => v.lang.toLowerCase().startsWith(targetLang));
      if (fallbackVoice) {
        console.log(`[TTS DEBUG] Found fallback voice: ${fallbackVoice.name} (${fallbackVoice.lang})`);
        selectedVoice = fallbackVoice;
      }
    }
  } else {
    // 如果没有指定 VoiceName，则按文本内容匹配最合适的语言
    let targetLang = 'zh';
    if (hasJapanese) targetLang = 'ja';
    else if (hasKorean) targetLang = 'ko';
    else if (hasEnglish && !hasChinese) targetLang = 'en';

    selectedVoice = currentVoices.find(v => v.lang.toLowerCase().startsWith(targetLang)) || currentVoices[0];
  }

  if (selectedVoice) {
    utterance.voice = selectedVoice;
    console.log(`[TTS DEBUG] Final Voice selected: ${selectedVoice.name} (${selectedVoice.lang})`);
  }

  // 兜底检查：确保没有被暂停
  if (synth.paused) {
    console.log('[TTS DEBUG] Synth was paused, resuming...');
    synth.resume();
  }

  utterance.onstart = () => {
    // 确保状态同步
    isSpeaking.value = true;
    isPaused.value = false;
    highlightedSentence.value = highlightText || text;
    console.log('[TTS DEBUG] Utterance started');
  };

  utterance.onend = () => {
    console.log('[TTS DEBUG] Utterance ended');
    isSpeaking.value = false;
    highlightedSentence.value = '';
    currentUtterance = null; // 释放引用
    playNextSentence();
  };

  utterance.onerror = (e: any) => {
    // canceled 是由 synth.cancel() 触发的正常中断，不作为错误打印
    if (e.error === 'canceled' || e.error === 'interrupted') {
      console.log(`[TTS DEBUG] Utterance ${e.error} (normal stop)`);
    } else {
      console.error('[TTS DEBUG] Utterance Error:', e);
    }
    
    isSpeaking.value = false;
    highlightedSentence.value = '';
    currentUtterance = null; // 释放引用
    
    // 如果是 interrupted 或 canceled，通常意味着手动停止或切换，不需要自动播放下一句
    if (e.error !== 'interrupted' && e.error !== 'canceled') {
      // 稍微延迟一点点再播下一句，给引擎喘息时间
      setTimeout(() => {
        playNextSentence();
      }, 50);
    }
  };

  // 关键修复：某些浏览器在 cancel() 之后立即 speak() 会静默失败
  // 增加 50ms 的微延迟，确保引擎状态已重置
  setTimeout(() => {
    console.log('[TTS DEBUG] Calling synth.speak() now...');
    synth.speak(utterance);
    
    // 再次兜底检查：如果 speak() 后仍处于 paused，强制 resume
    if (synth.paused) {
      synth.resume();
    }
  }, 50);
};

const playNextSentence = () => {
  if (sentenceQueue.length === 0) {
    console.log('[TTS DEBUG] Queue empty, stopping.');
    isSpeaking.value = false;
    speakingMsgId.value = null;
    return;
  }

  if (isSpeaking.value) {
    console.log('[TTS DEBUG] Already speaking, skipping playNextSentence call');
    return;
  }

  const nextText = sentenceQueue.shift();
  if (nextText) {
    console.log(`[TTS DEBUG] Playing next sentence: "${nextText.substring(0, 20)}..."`);
    // 立即标记为正在播放，防止并发调用导致 interrupted
    isSpeaking.value = true;
    speak(nextText, currentVoiceName, currentVoiceLang, currentVoiceRate);
  }
};

const processSentenceBuffer = () => {
  // 流式输出时的处理逻辑
  while (sentenceBuffer.trim()) {
    let endIdx = -1;
    for (const char of sentenceEndChars) {
      const idx = sentenceBuffer.indexOf(char);
      if (idx !== -1 && (endIdx === -1 || idx < endIdx)) {
        endIdx = idx;
      }
    }

    if (endIdx === -1) {
      if (sentenceBuffer.length > MAX_SENTENCE_LENGTH) {
        endIdx = MAX_SENTENCE_LENGTH;
      } else if (!isGenerating.value) {
        endIdx = sentenceBuffer.length;
      } else {
        // 还没到切分点，且还在生成中，跳出循环等待更多内容
        break;
      }
    }

    const sentence = sentenceBuffer.slice(0, endIdx + 1).trim();
    sentenceBuffer = sentenceBuffer.slice(endIdx + 1);
    
    if (sentence) {
      // 流式输出时，直接将句子加入队列并尝试播放
      sentenceQueue.push(sentence);
      playNextSentence();
    }
  }
};

// --- Title Generation ---
const isGeneratingTitle = ref(false);

const generateTitle = async (session: AISession) => {
  if (isGeneratingTitle.value) return;
  isGeneratingTitle.value = true;
  
  try {
    // 1. Get recent messages for context
    const historyRes = await aiApi.getChatHistory(session.session_id);
    const history = historyRes.data.data || [];
    if (history.length === 0) return;
    
    const context = history.slice(-5).map(m => `${m.role}: ${m.content}`).join('\n');
    
    // 2. Call the title generator agent
    const res = await fetch('/api/ai/chat', {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({
        agent_id: BUILTIN_AGENTS.TITLE_GENERATOR,
        message: `请根据以下对话内容，生成一个简短的会话标题（不超过10个字）：\n\n${context}`,
        stream: false
      })
    });
    
    const data = await res.json();
    if (data.success && data.content) {
      const newTitle = data.content.trim().replace(/^"|"$/g, '');
      
      // 3. Update session title
      const updateRes = await fetch(`/api/admin/ai/sessions/${session.session_id}/topic`, {
        method: 'PUT',
        headers: getHeaders(),
        body: JSON.stringify({ topic: newTitle })
      });
      
      const updateData = await updateRes.json();
      if (updateData.success) {
        session.topic = newTitle;
      }
    }
  } catch (error) {
    console.error('Failed to generate title:', error);
  } finally {
    isGeneratingTitle.value = false;
  }
};

// --- Helpers ---
const getHeaders = () => {
  const token = localStorage.getItem('wxbot_token');
  return {
    'Content-Type': 'application/json',
    'Authorization': token ? `Bearer ${token}` : ''
  };
};

// --- Fetch Data ---
const fetchData = async () => {
  isLoading.value = true;
  try {
    const headers = getHeaders();
    const [pRes, mRes, aRes, sRes, kRes] = await Promise.all([
      fetch('/api/admin/ai/providers', { headers }),
      fetch('/api/admin/ai/models', { headers }),
      aiApi.getAgents(),
      aiApi.getRecentSessions(),
      aiApi.getKnowledgeList()
    ]);

    const pData = await pRes.json();
    const mData = await mRes.json();
    const aData = aRes.data;
    const sData = sRes.data;
    const kData = kRes.data;

    if (pData.success) providers.value = pData.data || [];
    if (mData.success) models.value = mData.data || [];
    if (aData.success) agents.value = aData.data || [];
    if (sData.success) sessions.value = sData.data || [];
    if (kData.success) knowledgeDocs.value = kData.data || [];
    
    // Select first agent by default if none selected and in agents tab
    // Update agent list to ensure ID 4 is Lu Xun
      agents.value = agents.value.map(a => {
        if (a.id === 4 && (a.name === '早喵' || !a.name)) {
          return { ...a, name: '鲁迅' };
        }
        return a;
      });

      if (activeTab.value === 'agents' && agents.value.length > 0 && !selectedAgent.value) {
      selectedAgent.value = agents.value[0];
    }
  } catch (error) {
    console.error('Failed to fetch AI data:', error);
  } finally {
    isLoading.value = false;
  }
};

const kbFileInput = ref<HTMLInputElement | null>(null);
const isUploadingKB = ref(false);

const handleKBUpload = async (event: Event) => {
  const target = event.target as HTMLInputElement;
  if (!target.files?.length) return;

  const file = target.files[0];
  isUploadingKB.value = true;
  
  try {
    const res = await aiApi.uploadKnowledge(file);
    if (res.data.success) {
      alert(res.data.message || '上传成功，正在后台解析...');
      // 延迟刷新列表，给后端一点时间开始处理
      setTimeout(() => {
        aiApi.getKnowledgeList().then(r => {
          if (r.data.success) knowledgeDocs.value = r.data.data || [];
        });
      }, 2000);
    } else {
      alert(res.data.message || '上传失败');
    }
  } catch (err: any) {
    alert('上传请求失败: ' + (err.message || err));
  } finally {
    isUploadingKB.value = false;
    target.value = '';
  }
};

const deleteKnowledge = async (id: number) => {
  if (!confirm('确定要删除这个知识文档吗？关联的切片也将被清除。')) return;

  try {
    const res = await aiApi.deleteKnowledge(id);
    if (res.data.success) {
      knowledgeDocs.value = knowledgeDocs.value.filter(d => d.id !== id);
    } else {
      alert(res.data.message || '删除失败');
    }
  } catch (err: any) {
    alert('请求失败: ' + (err.message || err));
  }
};

onMounted(async () => {
  await fetchData();
  
  // Handle URL query parameters
  const queryAgentId = route.query.agent_id;
  const querySessionId = route.query.session_id;

  if (querySessionId) {
    const session = sessions.value.find(s => s.session_id === querySessionId);
    if (session) {
      await selectSession(session);
      activeTab.value = 'sessions';
    }
  } else if (queryAgentId) {
    const agentId = Number(queryAgentId);
    const agent = agents.value.find(a => a.id === agentId);
    if (agent) {
      startNewChat(agent);
    } else {
      // If not in list, try to fetch it (for link_only agents)
      try {
        const res = await aiApi.getAgentDetail(agentId);
        if (res.data.success && res.data.data) {
          startNewChat(res.data.data);
        }
      } catch (e) {
        console.error('Failed to fetch agent from query:', e);
      }
    }
  }
});

// --- Selection Logic ---
const selectSession = async (session: AISession) => {
  selectedSession.value = session;
  currentSessionId.value = session.session_id;
  
  // Find and select the corresponding agent
  if (selectedAgent.value?.id !== session.agent_id) {
    let agent = agents.value.find(a => a.id === session.agent_id);
    
    // Ensure ID 4 is Lu Xun
    if (agent && agent.id === 4 && (agent.name === '早喵' || !agent.name)) {
      agent = { ...agent, name: '鲁迅' };
    }

    if (agent) {
      selectedAgent.value = agent;
    } else {
      // If agent not in list, fetch it
      try {
        const res = await aiApi.getAgentDetail(session.agent_id);
        const data = res.data;
        if (data.success && data.data) {
          selectedAgent.value = data.data;
        } else {
          // Fallback if agent cannot be found
          console.warn(`[DATA DEBUG] Agent ID ${session.agent_id} not found in backend, using fallback`);
          const fallbackName = session.agent_id === 4 ? '鲁迅' : `智能体 #${session.agent_id}`;
          selectedAgent.value = {
            id: session.agent_id,
            name: fallbackName,
            description: '系统默认智能体',
            visibility: 'public',
            revenue_rate: 0,
            owner_id: 0,
            model_id: models.value[0]?.id || 0,
            system_prompt: '',
            temperature: 0.7,
            max_tokens: 2048,
            is_voice: false,
            voice_id: '',
            voice_rate: 1.0
          } as AIAgent;
        }
      } catch (e) {
        console.error('Failed to fetch agent for session:', e);
        // Fallback
        const fallbackName = session.agent_id === 4 ? '鲁迅' : `智能体 #${session.agent_id}`;
        selectedAgent.value = {
          id: session.agent_id,
          name: fallbackName,
          description: '系统默认智能体',
          visibility: 'public',
          revenue_rate: 0,
          owner_id: 0,
          model_id: models.value[0]?.id || 0,
          system_prompt: '',
          temperature: 0.7,
          max_tokens: 2048,
          is_voice: false,
          voice_id: '',
          voice_rate: 1.0
        } as AIAgent;
      }
    }
  }

  // History will be loaded by currentSessionId watcher
};

const startNewChat = (agent: AIAgent, force: boolean = false) => {
  if (!force && selectedAgent.value?.id === agent.id && !currentSessionId.value) {
    // 已经在准备新对话了，不需要重置
    return;
  }
  
  if (!force && selectedAgent.value?.id === agent.id && currentSessionId.value) {
    // 已经在跟这个智能体对话且已有会话，如果不是强制新建，则保持现状
    return;
  }

  selectedAgent.value = agent;
  selectedSession.value = null;
  currentSessionId.value = null; // New session
  activeTab.value = 'agents'; // Switch to agents tab if not already there
  
  // 清除该智能体的临时历史，准备开始新会话
  const historyKey = `agent_${agent.id}`;
  chatHistories.value[historyKey] = [];
  
  nextTick(scrollToBottom);
};

// --- Agent Management ---
const openAddAgent = () => {
  errorMessage.value = '';
  Object.assign(editingAgent, {
    name: '',
    description: '',
    model_id: models.value[0]?.id || 0,
    system_prompt: '',
    temperature: 0.7,
    max_tokens: 2048,
    visibility: 'public',
    revenue_rate: 0,
    is_voice: false,
    voice_id: '',
    voice_rate: 1.0
  });
  showAgentModal.value = true;
};

const openEditAgent = async (agent: AIAgent) => {
  errorMessage.value = '';
  // 确保有详情数据
  if (!agentDetails.value[agent.id]) {
    try {
      const res = await aiApi.getAgentDetail(agent.id);
      const data = res.data;
      if (data.success) {
        agentDetails.value[agent.id] = data.data;
      }
    } catch (error) {
      console.error('Failed to fetch agent details for editing:', error);
    }
  }
  const details = agentDetails.value[agent.id] || agent;
  Object.assign(editingAgent, {
    ...details,
    visibility: details.visibility || 'public',
    revenue_rate: details.revenue_rate || 0,
    is_voice: details.is_voice || false,
    voice_id: details.voice_id || '',
    voice_rate: details.voice_rate || 1.0
  });
  showAgentModal.value = true;
};

const saveAgent = async () => {
  if (!editingAgent.name) {
    errorMessage.value = t('ai_name_required');
    return;
  }
  if (!editingAgent.model_id) {
    errorMessage.value = t('ai_model_required');
    return;
  }

  try {
    const res = await fetch('/api/admin/ai/agents', {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify(editingAgent)
    });
    const data = await res.json();
    if (data.success) {
      showAgentModal.value = false;
      // 清除详情缓存以便下次获取最新数据
      if (editingAgent.id) delete agentDetails.value[editingAgent.id];
      await fetchData();
    } else {
      errorMessage.value = data.message || t('ai_save_failed');
    }
  } catch (error) {
    console.error('Save agent failed:', error);
    errorMessage.value = t('ai_save_failed');
  }
};

const deleteAgent = async (id: number) => {
  if (!confirm(t('ai_delete_confirm'))) return;
  try {
    const res = await fetch(`/api/admin/ai/agents/${id}`, { 
      method: 'DELETE',
      headers: getHeaders()
    });
    const data = await res.json();
    if (data.success) {
      if (selectedAgent.value?.id === id) selectedAgent.value = null;
      // Also delete history and details
      delete chatHistories.value[`agent_${id}`];
      delete agentDetails.value[id];
      await fetchData();
    }
  } catch (error) {
    console.error('Delete agent failed:', error);
  }
};

const shareAgent = (agent: AIAgent) => {
  const url = new URL(window.location.href);
  url.searchParams.set('agent_id', agent.id.toString());
  url.searchParams.delete('session_id');
  
  navigator.clipboard.writeText(url.toString()).then(() => {
    alert(t('ai_copy_success') || '链接已复制到剪贴板');
  }).catch(err => {
    console.error('Failed to copy link:', err);
  });
};

// --- Chat History Logic ---
watch(currentSessionId, async (newId, oldId) => {
  if (newId) {
    // If we just got a real session ID from a temporary one, the history is already moved
    // but we might want to refresh it from the server to be sure it's in sync
    if (!chatHistories.value[newId] || chatHistories.value[newId].length === 0) {
      await fetchHistory(newId, true);
    }
  }
  await nextTick(scrollToBottom);
}, { immediate: true });

watch(selectedAgent, async (newAgent) => {
  if (newAgent) {
      // 只有在没有手动设置过（即本地存储为空）时，才遵循智能体的默认语音开关
      if (localStorage.getItem('auto_voice_enabled') === null) {
        isAutoVoiceEnabled.value = newAgent.is_voice;
      }
      
      // 立即尝试从详情中同步一次（如果列表接口已经带了部分数据）
      console.log(`[DATA DEBUG] Switching to Agent: ${newAgent.name} (ID: ${newAgent.id})`);
      console.log(`[DATA DEBUG] Initial Agent Data from List:`, JSON.parse(JSON.stringify(newAgent)));

      if (selectedAgent.value && selectedAgent.value.id === newAgent.id) {
        // 如果 newAgent 里已经有了这些字段，直接用
        if (newAgent.voice_name !== undefined) {
          console.log(`[DATA DEBUG] Setting voice_name from list: ${newAgent.voice_name}`);
          selectedAgent.value.voice_name = newAgent.voice_name;
        }
        if (newAgent.voice_lang !== undefined) {
          console.log(`[DATA DEBUG] Setting voice_lang from list: ${newAgent.voice_lang}`);
          selectedAgent.value.voice_lang = newAgent.voice_lang;
        }
        if (newAgent.voice_rate !== undefined) {
          console.log(`[DATA DEBUG] Setting voice_rate from list: ${newAgent.voice_rate}`);
          selectedAgent.value.voice_rate = newAgent.voice_rate;
        }
      }
      
      if (!currentSessionId.value) {
        // 按需加载详情
        if (!agentDetails.value[newAgent.id]) {
          try {
            console.log(`[DATA DEBUG] Fetching detail for Agent ID: ${newAgent.id}`);
            const res = await aiApi.getAgentDetail(newAgent.id);
            const data = res.data;
            if (data.success) {
              console.log(`[DATA DEBUG] Detail Fetched Successfully:`, JSON.parse(JSON.stringify(data.data)));
              agentDetails.value[newAgent.id] = data.data;
              // 同步详情中的语音设置到 selectedAgent
              if (selectedAgent.value && selectedAgent.value.id === newAgent.id) {
                console.log(`[DATA DEBUG] Syncing detail fields to selectedAgent`);
                selectedAgent.value.voice_name = data.data.voice_name;
                selectedAgent.value.voice_lang = data.data.voice_lang;
                selectedAgent.value.voice_rate = data.data.voice_rate;
                selectedAgent.value.is_voice = data.data.is_voice;
              }
            // 如果详情里有语音设置且用户从未手动设置过
            if (data.data.is_voice !== undefined && localStorage.getItem('auto_voice_enabled') === null) {
              isAutoVoiceEnabled.value = data.data.is_voice;
            }
          }
        } catch (e) {
          console.error('Failed to fetch agent details:', e);
        }
      }
      
      // 加载最近历史
      const historyKey = `agent_${newAgent.id}`;
      if (!chatHistories.value[historyKey] || chatHistories.value[historyKey].length === 0) {
        await fetchHistory(newAgent.id, false);
      } else {
        nextTick(scrollToBottom);
      }
    } else {
      nextTick(scrollToBottom);
    }
  }
});

// --- Model Management ---
const openAddModel = () => {
  errorMessage.value = '';
  Object.assign(editingModel, {
    provider_id: providers.value[0]?.id || 0,
    model_id: '',
    model_name: '',
    context_size: 4096
  });
  showModelModal.value = true;
};

const openEditModel = (model: AIModel) => {
  errorMessage.value = '';
  Object.assign(editingModel, model);
  showModelModal.value = true;
};

const saveModel = async () => {
  if (!editingModel.model_name) {
    errorMessage.value = t('ai_name_required');
    return;
  }
  if (!editingModel.model_id) {
    errorMessage.value = t('ai_id_required');
    return;
  }
  if (!editingModel.provider_id) {
    errorMessage.value = t('ai_provider_required');
    return;
  }

  try {
    const res = await fetch('/api/admin/ai/models', {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify(editingModel)
    });
    const data = await res.json();
    if (data.success) {
      showModelModal.value = false;
      await fetchData();
    } else {
      errorMessage.value = data.message || t('ai_save_failed');
    }
  } catch (error) {
    console.error('Save model failed:', error);
    errorMessage.value = t('ai_save_failed');
  }
};

const deleteModel = async (id: number) => {
  if (!confirm(t('ai_delete_confirm'))) return;
  try {
    const res = await fetch(`/api/admin/ai/models/${id}`, { 
      method: 'DELETE',
      headers: getHeaders()
    });
    const data = await res.json();
    if (data.success) await fetchData();
  } catch (error) {
    console.error('Delete model failed:', error);
  }
};

// --- Provider Management ---
const openAddProvider = () => {
  errorMessage.value = '';
  Object.assign(editingProvider, {
    name: '',
    type: 'openai',
    base_url: '',
    api_key: ''
  });
  showProviderModal.value = true;
};

const openEditProvider = (provider: AIProvider) => {
  errorMessage.value = '';
  Object.assign(editingProvider, provider);
  showProviderModal.value = true;
};

const saveProvider = async () => {
  if (!editingProvider.name) {
    errorMessage.value = t('ai_name_required');
    return;
  }
  if (!editingProvider.type) {
    errorMessage.value = t('ai_type_required');
    return;
  }

  try {
    const res = await fetch('/api/admin/ai/providers', {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify(editingProvider)
    });
    const data = await res.json();
    if (data.success) {
      showProviderModal.value = false;
      await fetchData();
    } else {
      errorMessage.value = data.message || t('ai_save_failed');
    }
  } catch (error) {
    console.error('Save provider failed:', error);
    errorMessage.value = t('ai_save_failed');
  }
};

const deleteProvider = async (id: number) => {
  if (!confirm(t('ai_delete_confirm'))) return;
  try {
    const res = await fetch(`/api/admin/ai/providers/${id}`, { 
      method: 'DELETE',
      headers: getHeaders()
    });
    const data = await res.json();
    if (data.success) await fetchData();
  } catch (error) {
    console.error('Delete provider failed:', error);
  }
};

// --- Chat Trial (Streaming) ---
const sendMessage = async () => {
  if (!userInput.value.trim() || !selectedAgent.value || isGenerating.value) return;

  const content = userInput.value;
  userInput.value = '';
  isGenerating.value = false;
  
  // Reset TTS buffer for new message
  sentenceBuffer = '';
  if (isAutoVoiceEnabled.value) {
    stopSpeaking();
  }

  const historyKey = currentSessionId.value || `agent_${selectedAgent.value.id}`;
  if (!chatHistories.value[historyKey]) {
    chatHistories.value[historyKey] = [];
  }

  // Add user message
  chatHistories.value[historyKey].push({ role: 'user', content });
  await nextTick(scrollToBottom);

  // Add assistant placeholder
  const assistantMsg = reactive({ role: 'assistant', content: '_', tempId: Date.now().toString() });
  chatHistories.value[historyKey].push(assistantMsg);

  isGenerating.value = true;
  if (isAutoVoiceEnabled.value) {
    speakingMsgId.value = assistantMsg.tempId;
    // 优先从详情中获取最新的语音配置
    const detail = agentDetails.value[selectedAgent.value?.id || 0];
    currentVoiceId = detail?.voice_id || selectedAgent.value?.voice_id;
    currentVoiceName = detail?.voice_name || selectedAgent.value?.voice_name;
    currentVoiceLang = detail?.voice_lang || selectedAgent.value?.voice_lang;
    currentVoiceRate = detail?.voice_rate !== undefined ? detail.voice_rate : (selectedAgent.value?.voice_rate || 1.0);
    
    // 兼容处理：如果语速为0（通常是异常值），则设为1.0
    if (currentVoiceRate <= 0) currentVoiceRate = 1.0;

    console.log(`[TTS DEBUG] Initializing TTS for streaming: Voice='${currentVoiceName}', Lang='${currentVoiceLang}', Rate=${currentVoiceRate}`);

    // 预热语音引擎以在流式输出时解锁自动播放
    const warmUp = new SpeechSynthesisUtterance('');
    warmUp.volume = 0;
    synth.speak(warmUp);
  }

  try {
    const headers = getHeaders();
    // Prepare messages for context
    const currentHistory = chatHistories.value[historyKey] || [];
    // Only send completed messages, excluding the current assistant placeholder
    const messagesToSend = currentHistory
      .filter(m => (m.content || (m as any).Content) !== '_')
      .map(m => ({
        role: (m.role || (m as any).Role || 'user').toLowerCase(),
        content: m.content || (m as any).Content || ''
      }));

    const response = await fetch('/api/ai/chat/stream', {
      method: 'POST',
      headers,
      body: JSON.stringify({
        agent_id: selectedAgent.value.id,
        session_id: currentSessionId.value || '', // If empty, backend will create one
        messages: messagesToSend,
        context_length: contextLength.value
      })
    });

    if (!response.ok) throw new Error('Network response was not ok');

    const reader = response.body?.getReader();
    if (!reader) throw new Error('No reader');

    const decoder = new TextDecoder();
    let buffer = '';
    let displayContent = '';
    let isStreamingDone = false;

    // Dynamic typewriter effect with variable speed
    const updateDisplay = async () => {
      if (isStreamingDone && !displayContent) return;
      
      if (displayContent) {
        // Dynamic speed based on buffer length and punctuation
        let batchSize = 1;
        let delay = 30 + Math.random() * 40; // Base delay: 30-70ms

        // Speed up if buffer is large (catch up)
        if (displayContent.length > 50) {
          batchSize = Math.floor(displayContent.length / 15);
          delay = 10 + Math.random() * 20;
        } else if (displayContent.length > 20) {
          batchSize = 2;
          delay = 20 + Math.random() * 30;
        }

        // Slow down for punctuation to simulate thinking/breathing
        const firstChar = displayContent[0];
        if (['。', '！', '？', '；', '.', '!', '?', ';'].includes(firstChar)) {
          delay += 200 + Math.random() * 300; // Pause for punctuation
        } else if (['，', ',', ' ', '：', ':'].includes(firstChar)) {
          delay += 100 + Math.random() * 150; // Short pause for comma/space
        }

        const chunk = displayContent.slice(0, batchSize);
        displayContent = displayContent.slice(batchSize);
        
        if (assistantMsg.content.endsWith('_')) {
          assistantMsg.content = assistantMsg.content.slice(0, -1);
        }
        assistantMsg.content += chunk + '_';
        
        // Add to TTS buffer
        if (isAutoVoiceEnabled.value) {
          sentenceBuffer += chunk;
          if (!isSpeaking.value) {
            processSentenceBuffer();
          }
        }
        
        await nextTick(scrollToBottom);
        
        if (!isStreamingDone || displayContent) {
          setTimeout(updateDisplay, delay);
        }
      } else if (!isStreamingDone) {
        // Wait for more content from the stream
        setTimeout(updateDisplay, 50);
      }
    };
    updateDisplay();

    while (true) {
      const { done, value } = await reader.read();
      if (done) {
        isStreamingDone = true;
        // Wait for buffer to clear
        while (displayContent) {
          await new Promise(r => setTimeout(r, 20));
        }
        // Remove trailing _
        if (assistantMsg.content.endsWith('_')) {
          assistantMsg.content = assistantMsg.content.slice(0, -1);
        }
        
        // Final TTS check for any remaining content in buffer
        if (isAutoVoiceEnabled.value && sentenceBuffer.trim()) {
          processSentenceBuffer();
        }
        
        // Stream completed, refresh sessions list
        const sRes = await aiApi.getRecentSessions();
        const sData = sRes.data;
        if (sData.success) sessions.value = sData.data || [];
        break;
      }

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop() || '';

      for (const line of lines) {
        if (line.startsWith('data: ')) {
          const dataStr = line.slice(6);
          if (dataStr === '[DONE]') {
            isStreamingDone = true;
            break;
          }
          try {
            const data = JSON.parse(dataStr);
            if (data.session_id && !currentSessionId.value) {
              // ... existing session ID logic ...
              chatHistories.value[data.session_id] = chatHistories.value[historyKey];
              const oldKey = historyKey;
              currentSessionId.value = data.session_id;
              if (oldKey !== data.session_id) {
                delete chatHistories.value[oldKey];
              }
              aiApi.getRecentSessions().then(res => {
                const data = res.data;
                if (data.success) sessions.value = data.data || [];
              });
            }
            if (data.content) {
              displayContent += data.content;
            }
          } catch (e) {
            console.warn('Failed to parse SSE data:', dataStr);
          }
        }
      }
    }
  } catch (error) {
    console.error('Chat error:', error);
    if (assistantMsg.content.endsWith('_')) {
      assistantMsg.content = assistantMsg.content.slice(0, -1);
    }
    assistantMsg.content += `\n[${t('error')}: ${t('ai_save_failed')}]`;
  } finally {
    isGenerating.value = false;
    if (assistantMsg.content.endsWith('_')) {
      assistantMsg.content = assistantMsg.content.slice(0, -1);
    }
    await nextTick(scrollToBottom);
  }
};

const clearChat = () => {
  chatMessages.value = [];
};

// --- Helpers ---
const getModelName = (id: number) => {
  return models.value.find(m => m.id === id)?.model_name || t('unknown');
};

const getProviderName = (id: number) => {
  return providers.value.find(p => p.id === id)?.name || t('unknown');
};

const splitByHighlight = (text: string, highlight: string) => {
  if (!highlight) return [text];
  const parts = [];
  let current = text;
  while (current.includes(highlight)) {
    const idx = current.indexOf(highlight);
    if (idx > 0) parts.push(current.slice(0, idx));
    parts.push(highlight);
    current = current.slice(idx + highlight.length);
  }
  if (current) parts.push(current);
  return parts;
};

const filteredAgents = computed(() => {
  let list = [...agents.value];
  
  // Apply sorting: call_count DESC, then created_at DESC (if available)
  list.sort((a, b) => {
    if ((b.call_count || 0) !== (a.call_count || 0)) {
      return (b.call_count || 0) - (a.call_count || 0);
    }
    // Fallback to ID or date if counts are equal
    return b.id - a.id;
  });

  if (!searchQuery.value) return list;
  
  const q = searchQuery.value.toLowerCase();
  return list.filter(a => 
    a.name.toLowerCase().includes(q) || 
    a.description.toLowerCase().includes(q)
  );
});

const filteredSessions = computed(() => {
  if (!searchQuery.value) return sessions.value;
  const q = searchQuery.value.toLowerCase();
  return sessions.value.filter(s => 
    (s.topic && s.topic.toLowerCase().includes(q)) || 
    (s.agent?.name && s.agent.name.toLowerCase().includes(q)) ||
    (s.last_msg && s.last_msg.toLowerCase().includes(q))
  );
});

</script>

<template>
  <div class="h-full flex flex-col bg-[var(--bg-body)]">
    <!-- Header with Tabs -->
    <header class="flex items-center justify-between px-6 py-4 border-b border-[var(--border-color)] bg-[var(--bg-header)] backdrop-blur-md sticky top-0 z-20">
      <div class="flex items-center gap-4">
        <div class="p-2.5 rounded-2xl bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] shadow-lg shadow-[var(--matrix-color)]/5 border border-[var(--matrix-color)]/20">
          <Sparkles class="w-6 h-6" />
        </div>
        <div>
          <h1 class="text-xl sm:text-2xl font-black text-[var(--text-main)] tracking-tight uppercase italic leading-none">{{ t('ai_nexus') }}</h1>
          <p class="text-[var(--text-muted)] text-[10px] sm:text-xs font-bold tracking-[0.2em] uppercase mt-1 opacity-70">{{ t('system_control') }}</p>
        </div>
      </div>

      <nav class="flex bg-[var(--bg-body)]/50 p-1 rounded-2xl border border-[var(--border-color)]">
        <button 
          v-for="tab in ['sessions', 'agents', 'knowledge', 'models', 'providers']" 
          :key="tab"
          @click="activeTab = tab"
          :class="[
            'px-6 py-2 rounded-xl text-sm font-semibold transition-all duration-300 relative overflow-hidden',
            activeTab === tab 
              ? 'bg-[var(--matrix-color)] text-[var(--sidebar-text-active)] shadow-lg shadow-[var(--matrix-color)]/25' 
              : 'text-[var(--text-muted)] hover:text-[var(--text-main)] hover:bg-[var(--matrix-color)]/5'
          ]"
        >
          {{ tab === 'knowledge' ? '知识库' : t('ai_' + tab) }}
        </button>
      </nav>

      <div class="flex items-center gap-3">
        <button 
          v-if="activeTab === 'knowledge'"
          @click="kbFileInput?.click()"
          :disabled="isUploadingKB"
          class="flex items-center gap-2 px-5 py-2.5 bg-[var(--matrix-color)] hover:opacity-90 text-[var(--sidebar-text-active)] rounded-xl transition-all text-sm font-bold shadow-lg shadow-[var(--matrix-color)]/20 active:scale-95 disabled:opacity-50"
        >
          <Loader2 v-if="isUploadingKB" class="w-4 h-4 animate-spin" />
          <FileUp v-else class="w-4 h-4" />
          {{ isUploadingKB ? '上传中...' : '上传文档' }}
        </button>
        <input 
          ref="kbFileInput"
          type="file"
          class="hidden"
          @change="handleKBUpload"
          accept=".pdf,.doc,.docx,.txt,.md,.go,.js,.ts,.py,.java,.c,.cpp,.h"
        />
        <button 
          v-if="activeTab === 'agents'"
          @click="openAddAgent"
          class="flex items-center gap-2 px-5 py-2.5 bg-[var(--matrix-color)] hover:opacity-90 text-[var(--sidebar-text-active)] rounded-xl transition-all text-sm font-bold shadow-lg shadow-[var(--matrix-color)]/20 active:scale-95"
        >
          <Plus class="w-4 h-4" />
          {{ t('ai_create_new') }}
        </button>
        <button 
          v-if="activeTab === 'models'"
          @click="openAddModel"
          class="flex items-center gap-2 px-5 py-2.5 bg-[var(--matrix-color)] hover:opacity-90 text-[var(--sidebar-text-active)] rounded-xl transition-all text-sm font-bold shadow-lg shadow-[var(--matrix-color)]/20 active:scale-95"
        >
          <Plus class="w-4 h-4" />
          {{ t('ai_register') }}
        </button>
        <button 
          v-if="activeTab === 'providers'"
          @click="openAddProvider"
          class="flex items-center gap-2 px-5 py-2.5 bg-[var(--matrix-color)] hover:opacity-90 text-[var(--sidebar-text-active)] rounded-xl transition-all text-sm font-bold shadow-lg shadow-[var(--matrix-color)]/20 active:scale-95"
        >
          <Plus class="w-4 h-4" />
          {{ t('ai_register') }}
        </button>
      </div>
    </header>

    <!-- Main Content Area -->
    <main class="flex-1 overflow-hidden relative">
      <div v-if="isLoading" class="absolute inset-0 flex items-center justify-center bg-[var(--bg-body)]/60 z-30 backdrop-blur-sm">
        <div class="flex flex-col items-center gap-4">
          <div class="relative">
            <Loader2 class="w-12 h-12 text-[var(--matrix-color)] animate-spin" />
            <div class="absolute inset-0 blur-xl bg-[var(--matrix-color)]/20 animate-pulse"></div>
          </div>
          <span class="text-[var(--text-muted)] text-xs tracking-[0.3em] font-bold uppercase">{{ t('ai_initializing') }}</span>
        </div>
      </div>

      <!-- Agents & Sessions Tab -->
      <div v-if="activeTab === 'agents' || activeTab === 'sessions'" class="h-full flex divide-x divide-[var(--border-color)]">
        <!-- Sidebar -->
        <div class="w-80 flex flex-col bg-[var(--bg-card)]/30">
          <div class="p-5 border-b border-[var(--border-color)]">
            <div class="relative group">
              <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)] group-focus-within:text-[var(--matrix-color)] transition-colors" />
              <input 
                v-model="searchQuery"
                type="text" 
                :placeholder="t('search')"
                class="w-full bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-xl pl-11 pr-4 py-2.5 text-sm text-[var(--text-main)] placeholder:text-[var(--text-muted)]/50 focus:outline-none focus:border-[var(--matrix-color)]/50 focus:ring-4 focus:ring-[var(--matrix-color)]/5 transition-all"
              />
            </div>
          </div>

          <!-- Agent List -->
          <div v-if="activeTab === 'agents'" class="flex-1 overflow-y-auto custom-scrollbar p-3 space-y-2">
            <div 
              v-for="agent in filteredAgents" 
              :key="agent.id"
              @click="startNewChat(agent)"
              :class="[
                'p-4 cursor-pointer transition-all rounded-2xl flex flex-col gap-2 group relative overflow-hidden border',
                selectedAgent?.id === agent.id && !currentSessionId 
                  ? 'bg-[var(--matrix-color)]/10 border-[var(--matrix-color)]/30' 
                  : 'bg-transparent border-transparent hover:bg-[var(--matrix-color)]/5 hover:border-[var(--border-color)]'
              ]"
            >
              <div v-if="selectedAgent?.id === agent.id && !currentSessionId" class="absolute left-0 top-0 bottom-0 w-1 bg-[var(--matrix-color)]"></div>
              
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2 truncate">
                  <span class="font-bold text-sm text-[var(--text-main)] truncate">{{ agent.name }}</span>
                  <span v-if="agent.call_count > 0" class="flex-shrink-0 px-1.5 py-0.5 rounded text-[9px] font-black bg-[var(--matrix-color)] text-[var(--sidebar-text-active)] uppercase tracking-tighter">
                    {{ agent.call_count }}
                  </span>
                </div>
                <div class="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  <button @click.stop="openEditAgent(agent)" class="p-1.5 hover:bg-[var(--matrix-color)]/10 rounded-lg text-[var(--text-muted)] hover:text-[var(--matrix-color)] transition-colors" title="Edit"><Edit3 class="w-3.5 h-3.5" /></button>
                  <button @click.stop="deleteAgent(agent.id)" class="p-1.5 hover:bg-red-500/10 rounded-lg text-[var(--text-muted)] hover:text-red-500 transition-colors" title="Delete"><Trash2 class="w-3.5 h-3.5" /></button>
                </div>
              </div>
              <span class="text-xs text-[var(--text-muted)] line-clamp-1 leading-relaxed">{{ agent.description || t('no_description') }}</span>
              <div class="flex items-center gap-2 mt-1">
                <span class="px-2 py-0.5 rounded-md text-[10px] font-bold bg-[var(--bg-body)] text-[var(--text-muted)] border border-[var(--border-color)] uppercase tracking-wider">
                  {{ getModelName(agent.model_id) }}
                </span>
                <span v-if="agent.call_count > 0" class="flex items-center gap-1 text-[10px] text-[var(--text-muted)] font-bold">
                  <Activity class="w-3 h-3" />
                  {{ agent.call_count }}
                </span>
              </div>
            </div>
            <div v-if="filteredAgents.length === 0" class="p-12 text-center">
              <div class="w-12 h-12 rounded-2xl bg-[var(--bg-body)] flex items-center justify-center mx-auto mb-4 border border-[var(--border-color)] text-[var(--text-muted)] opacity-20">
                <Search class="w-6 h-6" />
              </div>
              <p class="text-[var(--text-muted)] text-sm font-medium">{{ t('ai_no_data') }}</p>
            </div>
          </div>

          <!-- Session List -->
          <div v-else-if="activeTab === 'sessions'" class="flex-1 overflow-y-auto custom-scrollbar p-3 space-y-2">
            <div 
              v-for="session in filteredSessions" 
              :key="session.id"
              @click="selectSession(session)"
              :class="[
                'p-4 cursor-pointer transition-all rounded-2xl flex flex-col gap-2 group relative overflow-hidden border',
                currentSessionId === session.session_id 
                  ? 'bg-[var(--matrix-color)]/10 border-[var(--matrix-color)]/30' 
                  : 'bg-transparent border-transparent hover:bg-[var(--matrix-color)]/5 hover:border-[var(--border-color)]'
              ]"
            >
              <div v-if="currentSessionId === session.session_id" class="absolute left-0 top-0 bottom-0 w-1 bg-[var(--matrix-color)]"></div>
              
              <div class="flex items-center justify-between">
                <span class="font-bold text-sm text-[var(--text-main)] truncate">{{ session.topic || t('ai_history') }}</span>
                <div class="flex items-center gap-1">
                  <button 
                    @click.stop="generateTitle(session)" 
                    class="p-1.5 hover:bg-[var(--matrix-color)]/10 rounded-lg text-[var(--text-muted)] hover:text-[var(--matrix-color)] transition-all"
                    :class="{'animate-spin': isGeneratingTitle}"
                    :title="t('ai_generate_title')"
                  >
                    <Wand2 class="w-3.5 h-3.5" />
                  </button>
                  <span class="text-[10px] text-[var(--text-muted)] font-mono">{{ new Date(session.updated_at).toLocaleDateString() }}</span>
                </div>
              </div>
              <span class="text-xs text-[var(--text-muted)] line-clamp-1 leading-relaxed">{{ session.last_msg || '...' }}</span>
              <div class="flex items-center gap-2 mt-1">
                <span class="px-2 py-0.5 rounded-md text-[10px] font-bold bg-[var(--bg-body)] text-[var(--matrix-color)] border border-[var(--matrix-color)]/20 uppercase tracking-wider">
                  {{ session.agent?.name || `智能体 #${session.agent_id}` }}
                </span>
              </div>
            </div>
            <div v-if="filteredSessions.length === 0" class="p-12 text-center">
              <div class="w-12 h-12 rounded-2xl bg-[var(--bg-body)] flex items-center justify-center mx-auto mb-4 border border-[var(--border-color)] text-[var(--text-muted)] opacity-20">
                <History class="w-6 h-6" />
              </div>
              <p class="text-[var(--text-muted)] text-sm font-medium">{{ t('ai_no_data') }}</p>
            </div>
          </div>
        </div>

        <!-- Chat Trial Area (Shared) -->
        <div class="flex-1 flex flex-col bg-[var(--bg-body)]" :class="{'wechat-style': chatStyle === 'wechat'}">
          <div v-if="selectedAgent" class="flex-1 flex flex-col overflow-hidden" :class="chatStyle === 'wechat' ? 'bg-[var(--wechat-bg)]' : ''">
            <!-- Chat Header -->
            <div class="px-6 py-4 border-b border-[var(--border-color)] flex items-center justify-between bg-[var(--bg-card)]/50 backdrop-blur-md">
              <div class="flex items-center gap-4">
                <div class="w-12 h-12 rounded-2xl bg-[var(--matrix-color)]/10 flex items-center justify-center text-[var(--matrix-color)] shadow-lg shadow-[var(--matrix-color)]/5 border border-[var(--matrix-color)]/20">
                  <BotIcon class="w-7 h-7" />
                </div>
                <div>
                  <h2 class="text-base font-bold text-[var(--text-main)]">{{ selectedAgent.name }}</h2>
                  <div class="flex items-center gap-2 mt-0.5">
                    <span class="flex h-2 w-2 relative">
                      <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-500 opacity-75"></span>
                      <span class="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
                    </span>
                    <span class="text-[10px] text-emerald-500 uppercase font-bold tracking-[0.1em]">{{ t('ai_ready_deployment') }}</span>
                  </div>
                </div>
              </div>
              <div class="flex items-center gap-4">
                <!-- 风格选择器 -->
                <div class="flex items-center bg-[var(--bg-body)]/50 rounded-xl p-1 border border-[var(--border-color)]">
                  <button 
                    @click="chatStyle = 'wechat'"
                    :class="[
                      'px-3 py-1.5 rounded-lg text-[10px] font-black tracking-widest uppercase transition-all',
                      chatStyle === 'wechat' ? 'bg-[var(--matrix-color)] text-[var(--sidebar-text-active)] shadow-lg' : 'text-[var(--text-muted)] hover:text-[var(--text-main)]'
                    ]"
                  >
                    WECHAT
                  </button>
                  <button 
                    @click="chatStyle = 'default'"
                    :class="[
                      'px-3 py-1.5 rounded-lg text-[10px] font-black tracking-widest uppercase transition-all',
                      chatStyle === 'default' ? 'bg-[var(--matrix-color)] text-[var(--sidebar-text-active)] shadow-lg' : 'text-[var(--text-muted)] hover:text-[var(--text-main)]'
                    ]"
                  >
                    MATRIX
                  </button>
                </div>

                <div class="flex items-center gap-2 px-3 py-1.5 bg-[var(--bg-body)]/50 rounded-xl border border-[var(--border-color)] group relative">
                  <Layers class="w-3.5 h-3.5 text-[var(--text-muted)]" />
                  <select 
                    v-model="contextLength"
                    class="bg-transparent text-[10px] font-bold uppercase tracking-wider text-[var(--text-main)] focus:outline-none cursor-pointer pr-1"
                  >
                    <option v-for="opt in contextOptions" :key="opt.value" :value="opt.value" class="bg-[var(--bg-card)] text-[var(--text-main)]">
                      {{ opt.label }}
                    </option>
                  </select>
                  <!-- Tooltip/Hint -->
                  <div class="absolute bottom-full right-0 mb-2 w-48 p-2 bg-black/80 text-[var(--sidebar-text)] text-[9px] rounded-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-50 leading-relaxed font-medium">
                    {{ t('ai_context_hint') || '上下文越长，消耗的算力资源越多，响应可能变慢。' }}
                  </div>
                </div>

                <!-- Voice Toggle -->
                <button 
                  @click="toggleAutoVoice"
                  :class="[
                    'p-2 rounded-xl transition-all border flex items-center gap-2',
                    isAutoVoiceEnabled 
                      ? 'bg-[var(--matrix-color)]/20 border-[var(--matrix-color)]/30 text-[var(--matrix-color)]' 
                      : 'bg-[var(--bg-body)]/50 border-[var(--border-color)] text-[var(--text-muted)] hover:text-[var(--text-main)]'
                  ]"
                  :title="isAutoVoiceEnabled ? t('ai_voice_on') || '语音播报已开启' : t('ai_voice_off') || '语音播报已关闭'"
                >
                  <component :is="isAutoVoiceEnabled ? Volume2 : VolumeX" class="w-4 h-4" />
                  <span v-if="isAutoVoiceEnabled && isSpeaking" class="flex gap-0.5 items-center">
                    <span class="w-0.5 h-2 bg-current animate-[voice_0.5s_ease-in-out_infinite]"></span>
                    <span class="w-0.5 h-3 bg-current animate-[voice_0.5s_ease-in-out_0.1s_infinite]"></span>
                    <span class="w-0.5 h-2 bg-current animate-[voice_0.5s_ease-in-out_0.2s_infinite]"></span>
                  </span>
                </button>
                <button 
                  v-if="currentSessionId"
                  @click="startNewChat(selectedAgent, true)" 
                  class="text-xs font-bold text-[var(--matrix-color)] hover:text-[var(--sidebar-text-active)] flex items-center gap-2 px-4 py-2 rounded-xl bg-[var(--matrix-color)]/10 hover:bg-[var(--matrix-color)] border border-[var(--matrix-color)]/20 transition-all active:scale-95 shadow-lg shadow-[var(--matrix-color)]/5"
                >
                  <Plus class="w-4 h-4" />
                  {{ t('ai_new_chat') }}
                </button>
                <button @click="clearChat" class="text-xs font-bold text-[var(--text-muted)] hover:text-[var(--text-main)] flex items-center gap-2 px-4 py-2 rounded-xl hover:bg-[var(--matrix-color)]/10 border border-transparent hover:border-[var(--matrix-color)]/20 transition-all active:scale-95">
                  <RefreshCw class="w-4 h-4" />
                  {{ t('clear') }}
                </button>
              </div>
            </div>

            <!-- Messages -->
            <div ref="scrollContainer" @scroll="handleScroll" class="flex-1 overflow-y-auto p-4 sm:p-8 space-y-6 custom-scrollbar transition-colors duration-500">
              <!-- Loading History Indicator -->
              <div v-if="isLoadingHistory" class="flex justify-center py-4">
                <Loader2 class="w-6 h-6 animate-spin text-[var(--matrix-color)] opacity-50" />
              </div>

              <div v-if="chatMessages.length === 0 && !isLoadingHistory" class="h-full flex flex-col items-center justify-center text-center space-y-6">
                <div class="relative group">
                  <div class="absolute inset-0 bg-[var(--matrix-color)]/20 blur-3xl rounded-full scale-150 animate-pulse"></div>
                  <div class="w-24 h-24 rounded-[2.5rem] bg-[var(--bg-card)] flex items-center justify-center text-[var(--matrix-color)] shadow-2xl border border-[var(--border-color)] relative z-10 transition-transform duration-700 group-hover:scale-110 group-hover:rotate-3">
                    <MessageSquare class="w-12 h-12" />
                  </div>
                </div>
                <div class="space-y-3 relative z-10">
                  <h3 class="text-[var(--text-main)] font-black text-xl uppercase italic tracking-tight">{{ t('ai_trial_chat') }}</h3>
                  <p class="text-[var(--text-muted)] text-sm max-w-sm mx-auto leading-relaxed font-bold tracking-wide uppercase opacity-40">{{ t('ai_chat_placeholder') }}</p>
                </div>
              </div>

              <div 
                v-for="(msg, idx) in chatMessages" 
                :key="idx"
                :class="[
                  'flex gap-3 items-start w-full animate-in fade-in slide-in-from-bottom-2 duration-300', 
                  (msg.role || (msg as any).Role) === 'user' ? 'flex-row-reverse' : 'flex-row'
                ]"
              >
                <!-- Avatar -->
                <div :class="[
                  'w-10 h-10 flex items-center justify-center flex-shrink-0 transition-all duration-300',
                  chatStyle === 'wechat' ? 'rounded-md shadow-none' : 'rounded-2xl shadow-lg border',
                  (msg.role || (msg as any).Role) === 'user' 
                    ? (chatStyle === 'wechat' ? 'bg-[var(--wechat-user-bubble)] text-[#000]/60' : 'bg-[var(--matrix-color)] border-[var(--matrix-color)]/20 text-[var(--sidebar-text-active)] shadow-[var(--matrix-color)]/20')
                    : (chatStyle === 'wechat' ? 'bg-[var(--wechat-bot-bubble)] text-[#000]/60' : 'bg-[var(--bg-card)] border-[var(--border-color)] text-[var(--matrix-color)] shadow-black/5')
                ]">
                  <User v-if="(msg.role || (msg as any).Role) === 'user'" class="w-5 h-5" />
                  <BotIcon v-else class="w-5 h-5" />
                </div>

                <!-- Bubble Container -->
                <div :class="['max-w-[70%] flex flex-col', (msg.role || (msg as any).Role) === 'user' ? 'items-end' : 'items-start']">
                  <!-- Bubble -->
                  <div class="relative group/bubble">
                    <div :class="[
                      'px-4 py-2.5 text-[15px] leading-relaxed relative transition-all duration-300',
                      chatStyle === 'wechat' 
                        ? ((msg.role || (msg as any).Role) === 'user' 
                            ? 'bg-[var(--wechat-user-bubble)] text-[var(--wechat-text)] rounded-lg rounded-tr-none shadow-sm' 
                            : 'bg-[var(--wechat-bot-bubble)] text-[var(--wechat-text)] rounded-lg rounded-tl-none shadow-sm')
                        : ((msg.role || (msg as any).Role) === 'user' 
                            ? 'bg-[var(--matrix-color)] text-[var(--sidebar-text-active)] border border-[var(--matrix-color)]/20 rounded-3xl rounded-tr-none shadow-lg' 
                            : 'bg-[var(--bg-card)] text-[var(--text-main)] border border-[var(--border-color)] rounded-3xl rounded-tl-none shadow-lg')
                    ]">
                      <!-- WeChat Style Arrow -->
                      <div v-if="chatStyle === 'wechat'" :class="[
                        'absolute top-3 w-0 h-0 border-8',
                        (msg.role || (msg as any).Role) === 'user' 
                          ? 'right-[-8px] border-l-[var(--wechat-user-bubble)] border-t-transparent border-b-transparent border-r-transparent' 
                          : 'left-[-8px] border-r-[var(--wechat-bot-bubble)] border-t-transparent border-b-transparent border-l-transparent'
                      ]"></div>

                      <div class="whitespace-pre-wrap break-words">
                        <template v-if="(msg.role || (msg as any).Role) === 'assistant' && isSpeaking && highlightedSentence && (msg.id === speakingMsgId || (msg as any).tempId === speakingMsgId)">
                          <span v-for="(part, pIdx) in splitByHighlight(msg.content || (msg as any).Content, highlightedSentence)" :key="pIdx" :class="{'bg-yellow-300 text-black rounded px-0.5': part === highlightedSentence}">
                            {{ part }}
                          </span>
                        </template>
                        <template v-else>
                          {{ msg.content || (msg as any).Content }}
                        </template>
                      </div>
                      
                      <!-- Loading Dots -->
                      <div v-if="isGenerating && idx === chatMessages.length - 1 && !(msg.content || (msg as any).Content)" class="flex gap-1.5 py-2">
                        <span class="w-2 h-2 bg-current rounded-full animate-bounce [animation-duration:1s]"></span>
                        <span class="w-2 h-2 bg-current rounded-full animate-bounce [animation-duration:1s] [animation-delay:0.2s]"></span>
                        <span class="w-2 h-2 bg-current rounded-full animate-bounce [animation-duration:1s] [animation-delay:0.4s]"></span>
                      </div>
                    </div>

                    <!-- Voice Action -->
                    <div 
                      v-if="(msg.role || (msg as any).Role) === 'assistant'"
                      class="absolute top-0 flex items-center gap-1 opacity-0 group-hover/bubble:opacity-100 transition-all"
                      :class="[
                        (msg.role || (msg as any).Role) === 'user' ? 'right-full mr-2' : 'left-full ml-2',
                        {'opacity-100': isSpeaking && (msg.id === speakingMsgId || (msg as any).tempId === speakingMsgId)}
                      ]"
                    >
                      <button 
                        @click="isSpeaking && (msg.id === speakingMsgId || (msg as any).tempId === speakingMsgId) 
                          ? stopSpeaking() 
                          : speak(msg.content || (msg as any).Content, selectedAgent?.voice_name, selectedAgent?.voice_lang, selectedAgent?.voice_rate || 1.0, msg.id || (msg as any).tempId)"
                        class="p-2 text-[var(--text-muted)] hover:text-[var(--matrix-color)] rounded-xl hover:bg-[var(--matrix-color)]/5 transition-all"
                        :class="{'text-[var(--matrix-color)]': isSpeaking && (msg.id === speakingMsgId || (msg as any).tempId === speakingMsgId)}"
                      >
                        <Volume2 v-if="!(isSpeaking && (msg.id === speakingMsgId || (msg as any).tempId === speakingMsgId))" class="w-4 h-4" />
                        <VolumeX v-else class="w-4 h-4 animate-pulse" />
                      </button>
                      
                      <button 
                        v-if="isSpeaking && (msg.id === speakingMsgId || (msg as any).tempId === speakingMsgId)"
                        @click.stop="isPaused ? resumeSpeaking() : pauseSpeaking()"
                        class="p-2 text-[var(--text-muted)] hover:text-[var(--matrix-color)] rounded-xl hover:bg-[var(--matrix-color)]/5 transition-all"
                      >
                        <Pause v-if="!isPaused" class="w-4 h-4" />
                        <Play v-else class="w-4 h-4" />
                      </button>
                    </div>
                  </div>

                  <!-- Meta Info -->
                  <div :class="[
                    'flex items-center gap-2 mt-1.5 px-1 opacity-40',
                    chatStyle === 'wechat' ? 'text-[var(--wechat-meta)] font-medium text-[11px]' : 'text-[var(--text-muted)] font-black text-[10px] uppercase tracking-[0.2em] italic'
                  ]">
                    <span v-if="chatStyle !== 'wechat'">{{ (msg.role || (msg as any).Role) === 'user' ? 'USER' : (agentDetails[selectedAgent?.id || 0]?.name || 'SYSTEM') }}</span>
                    <span v-if="chatStyle !== 'wechat'" class="w-0.5 h-0.5 rounded-full bg-current"></span>
                    <span>{{ new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) }}</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Input Area -->
            <div class="p-6 bg-[var(--bg-card)]/50 border-t border-[var(--border-color)] backdrop-blur-md">
              <div class="max-w-5xl mx-auto relative group">
                <textarea 
                  v-model="userInput"
                  @keydown.enter.exact.prevent="sendMessage"
                  :placeholder="t('ai_chat_placeholder')"
                    class="w-full bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-2xl pl-4 pr-14 py-4 text-sm text-[var(--text-main)] placeholder:text-[var(--text-muted)]/30 focus:outline-none focus:border-[var(--matrix-color)]/50 focus:ring-4 focus:ring-[var(--matrix-color)]/5 transition-all resize-none min-h-[56px] max-h-32 font-medium"
                  rows="1"
                ></textarea>
                <button 
                  @click="sendMessage"
                  :disabled="!userInput.trim() || isGenerating"
                  :class="[
                    'absolute right-3 bottom-3 p-2.5 rounded-xl transition-all active:scale-90',
                    userInput.trim() && !isGenerating 
                      ? 'bg-[var(--matrix-color)] text-[var(--sidebar-text-active)] shadow-lg shadow-[var(--matrix-color)]/30' 
                      : 'bg-[var(--bg-body)] text-[var(--text-muted)]/20 cursor-not-allowed border border-[var(--border-color)]'
                  ]"
                >
                  <Loader2 v-if="isGenerating" class="w-5 h-5 animate-spin" />
                  <Send v-else class="w-5 h-5" />
                </button>
              </div>
              <p class="text-[10px] text-center text-[var(--text-muted)] mt-4 uppercase tracking-[0.3em] font-bold opacity-30">{{ t('ai_secure_stream') }}</p>
            </div>
          </div>
          <div v-else class="flex-1 flex flex-col items-center justify-center text-center p-12 space-y-8 bg-[var(--bg-body)]">
            <div class="relative group">
              <div class="absolute inset-0 bg-[var(--matrix-color)]/20 blur-3xl rounded-full scale-150 opacity-50 group-hover:opacity-100 transition-opacity duration-1000 animate-pulse"></div>
              <div class="w-32 h-32 rounded-[3rem] bg-[var(--bg-card)] flex items-center justify-center border border-[var(--border-color)] shadow-2xl relative z-10 transition-all duration-700 group-hover:scale-110 group-hover:rotate-6">
                <Brain class="w-16 h-16 text-[var(--matrix-color)] opacity-20 group-hover:opacity-100 transition-all duration-700 group-hover:drop-shadow-[0_0_15px_var(--matrix-color)]" />
              </div>
            </div>
            <div class="space-y-4 relative z-10">
              <h2 class="text-3xl font-black text-[var(--text-main)] tracking-tight uppercase italic leading-none">{{ t('ai_nexus') }}</h2>
              <p class="text-[var(--text-muted)] text-sm max-w-sm mx-auto leading-relaxed font-bold tracking-widest uppercase opacity-40">{{ t('ai_task_desc') }}</p>
            </div>
            <button v-if="activeTab === 'agents'" @click="openAddAgent" class="px-10 py-4 bg-[var(--matrix-color)] text-[var(--sidebar-text-active)] border border-[var(--matrix-color)]/20 rounded-2xl transition-all text-sm font-black uppercase tracking-widest shadow-xl shadow-[var(--matrix-color)]/20 hover:scale-105 active:scale-95">
              {{ t('ai_create_new') }}
            </button>
            <button v-else-if="activeTab === 'sessions'" @click="activeTab = 'agents'" class="px-10 py-4 bg-[var(--matrix-color)] text-[var(--sidebar-text-active)] border border-[var(--matrix-color)]/20 rounded-2xl transition-all text-sm font-black uppercase tracking-widest shadow-xl shadow-[var(--matrix-color)]/20 hover:scale-105 active:scale-95">
              {{ t('ai_new_chat') }}
            </button>
          </div>
        </div>
      </div>

      <!-- Knowledge Tab -->
      <div v-if="activeTab === 'knowledge'" class="p-8 h-full overflow-y-auto custom-scrollbar bg-[var(--bg-body)]">
        <div class="max-w-6xl mx-auto">
          <div class="flex items-center justify-between mb-10">
            <div>
              <h2 class="text-3xl font-black text-[var(--text-main)] tracking-tight uppercase italic leading-none flex items-center gap-3">
                <Database class="w-8 h-8 text-[var(--matrix-color)]" />
                知识库管理
              </h2>
              <p class="text-[var(--text-muted)] text-xs font-bold tracking-[0.2em] uppercase mt-3 opacity-60">RAG KNOWLEDGE BASE / DOCUMENT INDEXING</p>
            </div>
            <div class="flex gap-4">
              <div class="relative group">
                <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)] group-focus-within:text-[var(--matrix-color)] transition-colors" />
                <input 
                  v-model="searchQuery"
                  type="text" 
                  placeholder="搜索文档..."
                  class="bg-[var(--bg-card)] border border-[var(--border-color)] rounded-xl pl-11 pr-4 py-2.5 text-sm text-[var(--text-main)] focus:outline-none focus:border-[var(--matrix-color)]/50 transition-all min-w-[240px]"
                />
              </div>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
            <!-- Upload Card -->
            <div 
              @click="kbFileInput?.click()"
              class="group cursor-pointer bg-[var(--bg-card)]/30 border-2 border-dashed border-[var(--border-color)] rounded-3xl p-8 flex flex-col items-center justify-center gap-4 hover:border-[var(--matrix-color)]/50 hover:bg-[var(--matrix-color)]/5 transition-all duration-500 min-h-[240px]"
            >
              <div class="w-16 h-16 rounded-2xl bg-[var(--matrix-color)]/10 flex items-center justify-center text-[var(--matrix-color)] group-hover:scale-110 group-hover:rotate-3 transition-all duration-500">
                <Plus v-if="!isUploadingKB" class="w-8 h-8" />
                <Loader2 v-else class="w-8 h-8 animate-spin" />
              </div>
              <div class="text-center">
                <h3 class="font-bold text-[var(--text-main)] uppercase tracking-widest">{{ isUploadingKB ? '正在上传...' : '上传新文档' }}</h3>
                <p class="text-xs text-[var(--text-muted)] mt-2 font-medium opacity-50 uppercase tracking-tighter">PDF, MD, DOCX, CODE, TXT...</p>
              </div>
            </div>

            <!-- Doc Cards -->
            <div 
              v-for="doc in knowledgeDocs.filter(d => !searchQuery || d.title.toLowerCase().includes(searchQuery.toLowerCase()))" 
              :key="doc.id"
              class="group bg-[var(--bg-card)] border border-[var(--border-color)] rounded-3xl p-6 hover:border-[var(--matrix-color)]/50 transition-all duration-500 relative overflow-hidden shadow-lg hover:shadow-[var(--matrix-color)]/10"
            >
              <div class="flex items-start justify-between mb-4 relative z-10">
                <div class="p-3 rounded-2xl bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] border border-[var(--border-color)]/20">
                  <FileText class="w-6 h-6" />
                </div>
                <button @click="deleteKnowledge(doc.id)" class="p-2.5 hover:bg-red-500/10 hover:text-red-500 text-[var(--text-muted)] rounded-xl transition-all opacity-0 group-hover:opacity-100">
                  <Trash2 class="w-4 h-4" />
                </button>
              </div>
              
              <div class="relative z-10">
                <h3 class="text-sm font-black text-[var(--text-main)] mb-2 truncate group-hover:text-[var(--matrix-color)] transition-colors">{{ doc.title }}</h3>
                <div class="flex flex-wrap gap-2 mt-4">
                  <span class="px-2 py-0.5 rounded-md text-[10px] font-bold bg-[var(--bg-body)] text-[var(--text-muted)] border border-[var(--border-color)] uppercase tracking-wider">
                    {{ doc.type }}
                  </span>
                  <span class="px-2 py-0.5 rounded-md text-[10px] font-bold bg-[var(--bg-body)] text-[var(--text-muted)] border border-[var(--border-color)] uppercase tracking-wider">
                    ID: {{ doc.id }}
                  </span>
                </div>
                <div class="mt-6 flex items-center justify-between text-[10px] text-[var(--text-muted)] font-bold uppercase tracking-widest opacity-50">
                  <span>{{ new Date(doc.created_at).toLocaleDateString() }}</span>
                  <span v-if="doc.uploader_id" class="flex items-center gap-1">
                    <User class="w-3 h-3" />
                    {{ doc.uploader_id }}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- Empty State -->
          <div v-if="knowledgeDocs.length === 0 && !isUploadingKB" class="mt-20 text-center py-20 bg-[var(--bg-card)]/10 rounded-[3rem] border border-dashed border-[var(--border-color)]">
            <div class="w-24 h-24 rounded-[2rem] bg-[var(--bg-body)] flex items-center justify-center mx-auto mb-6 border border-[var(--border-color)] text-[var(--text-muted)] opacity-20">
              <Database class="w-10 h-10" />
            </div>
            <h3 class="text-xl font-black text-[var(--text-main)] tracking-tight uppercase italic">{{ t('ai_no_data') }}</h3>
            <p class="text-[var(--text-muted)] text-sm mt-4 max-w-xs mx-auto font-medium leading-relaxed uppercase tracking-widest opacity-40">上传文档以开始构建机器人的本地知识库</p>
          </div>
        </div>
      </div>

      <!-- Models Tab -->
      <div v-if="activeTab === 'models'" class="p-8 h-full overflow-y-auto custom-scrollbar bg-[var(--bg-body)]">
        <div class="max-w-6xl mx-auto">
          <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-8">
            <div 
              v-for="model in models" 
              :key="model.id"
              class="group bg-[var(--bg-card)] border border-[var(--border-color)] rounded-3xl p-7 hover:border-[var(--matrix-color)]/50 transition-all duration-500 relative overflow-hidden shadow-lg hover:shadow-[var(--matrix-color)]/10"
            >
              <div class="absolute -right-8 -top-8 w-32 h-32 bg-[var(--matrix-color)]/5 rounded-full blur-3xl group-hover:bg-[var(--matrix-color)]/10 transition-colors duration-700"></div>
              
              <div class="flex items-start justify-between mb-6 relative z-10">
                <div class="p-3.5 rounded-2xl bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] border border-[var(--matrix-color)]/20 shadow-inner">
                  <Layers class="w-7 h-7" />
                </div>
                <div class="flex gap-2">
                  <button @click="openEditModel(model)" class="p-2.5 hover:bg-[var(--matrix-color)]/10 hover:text-[var(--matrix-color)] text-[var(--text-muted)] rounded-xl transition-all border border-transparent hover:border-[var(--matrix-color)]/20">
                    <Edit3 class="w-4.5 h-4.5" />
                  </button>
                  <button @click="deleteModel(model.id)" class="p-2.5 hover:bg-red-500/10 hover:text-red-500 text-[var(--text-muted)] rounded-xl transition-all border border-transparent hover:border-red-500/20">
                    <Trash2 class="w-4.5 h-4.5" />
                  </button>
                </div>
              </div>
              
              <div class="relative z-10">
                <h3 class="text-xl font-bold text-[var(--text-main)] mb-1.5">{{ model.model_name }}</h3>
                <p class="text-xs text-[var(--text-muted)] font-mono mb-6 opacity-60 tracking-wider">{{ model.model_id }}</p>
                
                <div class="space-y-4 pt-4 border-t border-[var(--border-color)]">
                  <div class="flex items-center justify-between">
                    <span class="text-[10px] text-[var(--text-muted)] font-bold uppercase tracking-[0.2em] opacity-50">{{ t('ai_provider') }}</span>
                    <span class="text-sm text-[var(--text-main)] font-bold">{{ getProviderName(model.provider_id) }}</span>
                  </div>
                  <div class="flex items-center justify-between">
                    <span class="text-[10px] text-[var(--text-muted)] font-bold uppercase tracking-[0.2em] opacity-50">{{ t('ai_context_size') }}</span>
                    <span class="text-sm text-[var(--matrix-color)] font-mono font-bold">{{ model.context_size.toLocaleString() }} <span class="text-[10px] opacity-70">TOKENS</span></span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Add Model Card -->
            <button 
              @click="openAddModel"
              class="group border-2 border-dashed border-[var(--border-color)] rounded-3xl p-8 hover:border-[var(--matrix-color)]/40 hover:bg-[var(--matrix-color)]/5 transition-all duration-500 flex flex-col items-center justify-center gap-6 min-h-[260px] relative overflow-hidden"
            >
              <div class="w-16 h-16 rounded-2xl bg-[var(--bg-card)] flex items-center justify-center border border-[var(--border-color)] group-hover:border-[var(--matrix-color)] group-hover:scale-110 group-hover:rotate-6 transition-all duration-500 shadow-xl">
                <Plus class="w-8 h-8 text-[var(--text-muted)] group-hover:text-[var(--matrix-color)]" />
              </div>
              <div class="text-center space-y-1">
                <span class="block text-sm font-bold text-[var(--text-muted)] group-hover:text-[var(--matrix-color)] tracking-[0.2em] uppercase transition-colors">{{ t('ai_register') }}</span>
                <span class="block text-xs font-medium text-[var(--text-muted)] opacity-40">{{ t('ai_models') }}</span>
              </div>
            </button>
          </div>
        </div>
      </div>

      <!-- Providers Tab -->
      <div v-if="activeTab === 'providers'" class="p-8 h-full overflow-y-auto custom-scrollbar bg-[var(--bg-body)]">
        <div class="max-w-6xl mx-auto">
          <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-8">
            <div 
              v-for="provider in providers" 
              :key="provider.id"
              class="group bg-[var(--bg-card)] border border-[var(--border-color)] rounded-3xl p-7 hover:border-[var(--matrix-color)]/50 transition-all duration-500 relative overflow-hidden shadow-lg hover:shadow-[var(--matrix-color)]/10"
            >
              <div class="absolute -right-8 -top-8 w-32 h-32 bg-[var(--matrix-color)]/5 rounded-full blur-3xl group-hover:bg-[var(--matrix-color)]/10 transition-colors duration-700"></div>
              
              <div class="flex items-start justify-between mb-6 relative z-10">
                <div class="p-3.5 rounded-2xl bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] border border-[var(--matrix-color)]/20 shadow-inner">
                  <Server class="w-7 h-7" />
                </div>
                <div class="flex gap-2">
                  <button @click="openEditProvider(provider)" class="p-2.5 hover:bg-[var(--matrix-color)]/10 hover:text-[var(--matrix-color)] text-[var(--text-muted)] rounded-xl transition-all border border-transparent hover:border-[var(--matrix-color)]/20">
                    <Edit3 class="w-4.5 h-4.5" />
                  </button>
                  <button @click="deleteProvider(provider.id)" class="p-2.5 hover:bg-red-500/10 hover:text-red-500 text-[var(--text-muted)] rounded-xl transition-all border border-transparent hover:border-red-500/20">
                    <Trash2 class="w-4.5 h-4.5" />
                  </button>
                </div>
              </div>
              
              <div class="relative z-10">
                <h3 class="text-xl font-bold text-[var(--text-main)] mb-2">{{ provider.name }}</h3>
                <div class="inline-flex items-center px-3 py-1 rounded-lg text-[10px] font-bold bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] border border-[var(--matrix-color)]/20 uppercase tracking-[0.1em] mb-6">
                  {{ provider.type }}
                </div>
                
                <div class="space-y-5 pt-4 border-t border-[var(--border-color)]">
                  <div class="flex flex-col gap-1.5">
                    <span class="text-[10px] text-[var(--text-muted)] font-bold uppercase tracking-[0.2em] opacity-50">{{ t('ai_base_url') }}</span>
                    <span class="text-xs text-[var(--text-main)] font-mono truncate bg-[var(--bg-body)]/50 p-2 rounded-lg border border-[var(--border-color)]">{{ provider.base_url || 'Default System Endpoint' }}</span>
                  </div>
                  <div class="flex flex-col gap-2">
                    <span class="text-[10px] text-[var(--text-muted)] font-bold uppercase tracking-[0.2em] opacity-50">{{ t('api_auth') }}</span>
                    <div class="flex items-center gap-3 bg-[var(--bg-body)]/50 p-2 rounded-lg border border-[var(--border-color)]">
                      <div class="flex gap-1.5">
                        <span v-for="i in 6" :key="i" class="w-1.5 h-1.5 rounded-full bg-[var(--matrix-color)]/30 animate-pulse" :style="{ animationDelay: i * 150 + 'ms' }"></span>
                      </div>
                      <span class="text-[10px] text-[var(--matrix-color)] font-bold tracking-widest opacity-50">ENCRYPTED</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Add Provider Card -->
            <button 
              @click="openAddProvider"
              class="group border-2 border-dashed border-[var(--border-color)] rounded-3xl p-8 hover:border-[var(--matrix-color)]/40 hover:bg-[var(--matrix-color)]/5 transition-all duration-500 flex flex-col items-center justify-center gap-6 min-h-[260px] relative overflow-hidden"
            >
              <div class="w-16 h-16 rounded-2xl bg-[var(--bg-card)] flex items-center justify-center border border-[var(--border-color)] group-hover:border-[var(--matrix-color)] group-hover:scale-110 group-hover:rotate-6 transition-all duration-500 shadow-xl">
                <Plus class="w-8 h-8 text-[var(--text-muted)] group-hover:text-[var(--matrix-color)]" />
              </div>
              <div class="text-center space-y-1">
                <span class="block text-sm font-bold text-[var(--text-muted)] group-hover:text-[var(--matrix-color)] tracking-[0.2em] uppercase transition-colors">{{ t('ai_register') }}</span>
                <span class="block text-xs font-medium text-[var(--text-muted)] opacity-40">{{ t('ai_provider') }}</span>
              </div>
            </button>
          </div>
        </div>
      </div>
    </main>

    <!-- Modals -->
    <!-- Agent Modal -->
    <div v-if="showAgentModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-[var(--bg-body)]/80 backdrop-blur-xl">
      <div class="bg-[var(--bg-card)] border border-[var(--border-color)] rounded-[2.5rem] w-full max-w-2xl shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-300">
        <div class="px-8 py-6 border-b border-[var(--border-color)] flex items-center justify-between bg-[var(--bg-header)] backdrop-blur-md">
          <div class="flex items-center gap-4">
            <div class="p-3 rounded-2xl bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] border border-[var(--matrix-color)]/20 shadow-inner">
              <BotIcon class="w-6 h-6" />
            </div>
            <h2 class="text-xl font-bold text-[var(--text-main)]">{{ editingAgent.id ? t('ai_edit') : t('ai_create_new') }}</h2>
          </div>
          <button @click="showAgentModal = false" class="p-2.5 hover:bg-[var(--bg-body)]/50 rounded-xl text-[var(--text-muted)] hover:text-[var(--text-main)] transition-all border border-transparent hover:border-[var(--border-color)]">
            <X class="w-6 h-6" />
          </button>
        </div>

        <div class="p-8 space-y-8 max-h-[75vh] overflow-y-auto custom-scrollbar">
          <!-- Error Alert -->
          <div v-if="errorMessage" class="p-4 rounded-2xl bg-red-500/10 border border-red-500/20 text-red-500 text-sm flex items-center gap-3 animate-shake">
            <div class="p-1.5 bg-red-500 rounded-lg text-[var(--sidebar-text)]">
              <X class="w-4 h-4" />
            </div>
            <span class="font-bold">{{ errorMessage }}</span>
          </div>

          <div class="grid grid-cols-2 gap-8">
            <div class="space-y-3 col-span-2">
              <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-[0.2em] opacity-60 px-1">{{ t('ai_agent_name') }}</label>
              <input v-model="editingAgent.name" type="text" class="w-full bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-2xl px-5 py-3.5 text-[var(--text-main)] font-medium focus:outline-none focus:border-[var(--matrix-color)]/50 focus:ring-4 focus:ring-[var(--matrix-color)]/5 transition-all" />
            </div>
            
            <div class="space-y-3 col-span-2">
              <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-[0.2em] opacity-60 px-1">{{ t('ai_description') }}</label>
              <input v-model="editingAgent.description" type="text" class="w-full bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-2xl px-5 py-3.5 text-[var(--text-main)] font-medium focus:outline-none focus:border-[var(--matrix-color)]/50 focus:ring-4 focus:ring-[var(--matrix-color)]/5 transition-all" />
            </div>

            <div class="space-y-3">
              <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-[0.2em] opacity-60 px-1">{{ t('ai_model') }}</label>
              <div class="relative">
                <select v-model="editingAgent.model_id" class="w-full bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-2xl px-5 py-3.5 text-[var(--text-main)] font-bold focus:outline-none focus:border-[var(--matrix-color)]/50 focus:ring-4 focus:ring-[var(--matrix-color)]/5 transition-all appearance-none cursor-pointer">
                  <option v-for="m in models" :key="m.id" :value="m.id" class="bg-[var(--bg-card)]">{{ m.model_name }}</option>
                </select>
                <ChevronRight class="absolute right-4 top-1/2 -translate-y-1/2 w-5 h-5 text-[var(--text-muted)] pointer-events-none rotate-90 opacity-40" />
              </div>
            </div>

            <div class="space-y-3">
              <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-[0.2em] opacity-60 px-1">{{ t('ai_temperature') }}</label>
              <input v-model.number="editingAgent.temperature" type="number" step="0.1" min="0" max="2" class="w-full bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-2xl px-5 py-3.5 text-[var(--text-main)] font-bold focus:outline-none focus:border-[var(--matrix-color)]/50 focus:ring-4 focus:ring-[var(--matrix-color)]/5 transition-all" />
            </div>

            <div class="space-y-3">
              <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-[0.2em] opacity-60 px-1">{{ t('ai_visibility') }}</label>
              <div class="relative">
                <select v-model="editingAgent.visibility" class="w-full bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-2xl px-5 py-3.5 text-[var(--text-main)] font-bold focus:outline-none focus:border-[var(--matrix-color)]/50 focus:ring-4 focus:ring-[var(--matrix-color)]/5 transition-all appearance-none cursor-pointer">
                  <option value="public" class="bg-[var(--bg-card)]">{{ t('ai_public') }}</option>
                  <option value="private" class="bg-[var(--bg-card)]">{{ t('ai_private') }}</option>
                  <option value="link_only" class="bg-[var(--bg-card)]">{{ t('ai_link_only') }}</option>
                </select>
                <ChevronRight class="absolute right-4 top-1/2 -translate-y-1/2 w-5 h-5 text-[var(--text-muted)] pointer-events-none rotate-90 opacity-40" />
              </div>
            </div>

            <div class="space-y-3">
              <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-[0.2em] opacity-60 px-1">{{ t('ai_revenue_rate') }}</label>
              <input v-model.number="editingAgent.revenue_rate" type="number" step="0.0001" min="0" class="w-full bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-2xl px-5 py-3.5 text-[var(--text-main)] font-bold focus:outline-none focus:border-[var(--matrix-color)]/50 focus:ring-4 focus:ring-[var(--matrix-color)]/5 transition-all" />
            </div>

            <!-- Voice Settings -->
            <div class="space-y-3 col-span-2 p-6 rounded-3xl bg-[var(--bg-body)]/30 border border-[var(--border-color)]">
              <div class="flex items-center justify-between mb-4">
                <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-[0.2em] opacity-60">{{ t('ai_voice_settings') || '语音设置' }}</label>
                <div class="flex items-center gap-2">
                  <span class="text-xs text-[var(--text-muted)]">{{ t('ai_voice_enable') || '开启语音播报' }}</span>
                  <button 
                    @click="editingAgent.is_voice = !editingAgent.is_voice"
                    :class="[
                      'w-10 h-5 rounded-full transition-all relative',
                      editingAgent.is_voice ? 'bg-[var(--matrix-color)]' : 'bg-[var(--border-color)]'
                    ]"
                  >
                    <div :class="['absolute top-1 w-3 h-3 bg-white rounded-full transition-all', editingAgent.is_voice ? 'right-1' : 'left-1']"></div>
                  </button>
                </div>
              </div>

              <div v-if="editingAgent.is_voice" class="grid grid-cols-2 gap-6">
                <div class="space-y-2">
                  <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-[0.2em] opacity-40 px-1">{{ t('ai_voice_role') || '语音角色' }}</label>
                  <div class="relative">
                    <select v-model="editingAgent.voice_id" class="w-full bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-xl px-4 py-2.5 text-xs text-[var(--text-main)] focus:outline-none focus:border-[var(--matrix-color)]/50 transition-all appearance-none cursor-pointer">
                      <option value="" class="bg-[var(--bg-card)]">默认中文女声</option>
                      <option v-for="v in voices" :key="v.voiceURI" :value="v.voiceURI" class="bg-[var(--bg-card)]">
                        {{ v.name }} ({{ v.lang }})
                      </option>
                    </select>
                    <ChevronRight class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)] pointer-events-none rotate-90 opacity-40" />
                  </div>
                </div>
                <div class="space-y-2">
                  <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-[0.2em] opacity-40 px-1">{{ t('ai_voice_rate') || '语速' }} ({{ editingAgent.voice_rate || 1.0 }})</label>
                  <input v-model.number="editingAgent.voice_rate" type="range" min="0.5" max="2" step="0.1" class="w-full accent-[var(--matrix-color)]" />
                </div>
              </div>
            </div>

            <div class="space-y-3 col-span-2">
              <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-[0.2em] opacity-60 px-1">{{ t('ai_system_prompt') }}</label>
              <textarea v-model="editingAgent.system_prompt" rows="6" class="w-full bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-2xl px-5 py-4 text-[var(--text-main)] font-medium focus:outline-none focus:border-[var(--matrix-color)]/50 focus:ring-4 focus:ring-[var(--matrix-color)]/5 transition-all resize-none font-mono text-sm leading-relaxed"></textarea>
            </div>
          </div>
        </div>

        <div class="px-8 py-6 border-t border-[var(--border-color)] bg-[var(--bg-header)] flex justify-end gap-4">
          <button @click="showAgentModal = false" class="px-8 py-3 rounded-2xl text-sm font-bold text-[var(--text-muted)] hover:text-[var(--text-main)] hover:bg-[var(--bg-body)]/50 transition-all border border-transparent hover:border-[var(--border-color)]">{{ t('ai_cancel') }}</button>
          <button @click="saveAgent" class="px-8 py-3 bg-[var(--matrix-color)] hover:opacity-90 text-[var(--sidebar-text-active)] rounded-2xl text-sm font-bold shadow-xl shadow-[var(--matrix-color)]/20 transition-all active:scale-95">{{ t('ai_save') }}</button>
        </div>
      </div>
    </div>

    <!-- Model Modal -->
    <div v-if="showModelModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-[var(--bg-body)]/80 backdrop-blur-xl">
      <div class="bg-[var(--bg-card)] border border-[var(--border-color)] rounded-[2.5rem] w-full max-w-md shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-300">
        <div class="px-8 py-6 border-b border-[var(--border-color)] flex items-center justify-between bg-[var(--bg-header)]">
          <h2 class="text-xl font-bold text-[var(--text-main)]">{{ editingModel.id ? t('ai_edit') : t('ai_register') }} {{ t('ai_models') }}</h2>
          <button @click="showModelModal = false" class="p-2.5 hover:bg-[var(--bg-body)]/50 rounded-xl text-[var(--text-muted)] hover:text-[var(--text-main)] transition-all"><X class="w-6 h-6" /></button>
        </div>
        <div class="p-8 space-y-6">
          <div v-if="errorMessage" class="p-4 rounded-2xl bg-red-500/10 border border-red-500/20 text-red-500 text-sm flex items-center gap-3">
            <X class="w-4 h-4" />
            <span class="font-bold">{{ errorMessage }}</span>
          </div>

          <div class="space-y-3">
            <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-[0.2em] opacity-60 px-1">{{ t('ai_provider') }}</label>
            <div class="relative">
              <select v-model="editingModel.provider_id" class="w-full bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-2xl px-5 py-3.5 text-[var(--text-main)] font-bold focus:outline-none focus:border-[var(--matrix-color)]/50 focus:ring-4 focus:ring-[var(--matrix-color)]/5 transition-all appearance-none cursor-pointer">
                <option v-for="p in providers" :key="p.id" :value="p.id" class="bg-[var(--bg-card)]">{{ p.name }}</option>
              </select>
              <ChevronRight class="absolute right-4 top-1/2 -translate-y-1/2 w-5 h-5 text-[var(--text-muted)] pointer-events-none rotate-90 opacity-40" />
            </div>
          </div>
          <div class="space-y-3">
            <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-[0.2em] opacity-60 px-1">{{ t('ai_model_name') }}</label>
            <input v-model="editingModel.model_name" type="text" placeholder="e.g. GPT-4o" class="w-full bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-2xl px-5 py-3.5 text-[var(--text-main)] font-medium focus:outline-none focus:border-[var(--matrix-color)]/50 focus:ring-4 focus:ring-[var(--matrix-color)]/5 transition-all" />
          </div>
          <div class="space-y-3">
            <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-[0.2em] opacity-60 px-1">{{ t('ai_model_id') }}</label>
            <input v-model="editingModel.model_id" type="text" placeholder="e.g. gpt-4o" class="w-full bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-2xl px-5 py-3.5 text-[var(--text-main)] font-mono focus:outline-none focus:border-[var(--matrix-color)]/50 focus:ring-4 focus:ring-[var(--matrix-color)]/5 transition-all" />
          </div>
          <div class="space-y-3">
            <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-[0.2em] opacity-60 px-1">{{ t('ai_context_size') }}</label>
            <input v-model.number="editingModel.context_size" type="number" class="w-full bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-2xl px-5 py-3.5 text-[var(--text-main)] font-bold focus:outline-none focus:border-[var(--matrix-color)]/50 focus:ring-4 focus:ring-[var(--matrix-color)]/5 transition-all" />
          </div>
        </div>
        <div class="px-8 py-6 border-t border-[var(--border-color)] bg-[var(--bg-header)] flex justify-end gap-4">
          <button @click="showModelModal = false" class="px-6 py-2 rounded-xl text-sm font-bold text-[var(--text-muted)] hover:text-[var(--text-main)] transition-all">{{ t('ai_cancel') }}</button>
          <button @click="saveModel" class="px-8 py-3 bg-[var(--matrix-color)] hover:opacity-90 text-[var(--sidebar-text-active)] rounded-2xl text-sm font-bold shadow-xl shadow-[var(--matrix-color)]/20 transition-all active:scale-95">{{ t('ai_register') }}</button>
        </div>
      </div>
    </div>

    <!-- Provider Modal -->
    <div v-if="showProviderModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-[var(--bg-body)]/80 backdrop-blur-xl">
      <div class="bg-[var(--bg-card)] border border-[var(--border-color)] rounded-[2.5rem] w-full max-w-md shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-300">
        <div class="px-8 py-6 border-b border-[var(--border-color)] flex items-center justify-between bg-[var(--bg-header)]">
          <h2 class="text-xl font-bold text-[var(--text-main)]">{{ editingProvider.id ? t('ai_edit') : t('ai_register') }} {{ t('ai_provider') }}</h2>
          <button @click="showProviderModal = false" class="p-2.5 hover:bg-[var(--bg-body)]/50 rounded-xl text-[var(--text-muted)] hover:text-[var(--text-main)] transition-all"><X class="w-6 h-6" /></button>
        </div>
        <div class="p-8 space-y-6">
          <div v-if="errorMessage" class="p-4 rounded-2xl bg-red-500/10 border border-red-500/20 text-red-500 text-sm flex items-center gap-3">
            <X class="w-4 h-4" />
            <span class="font-bold">{{ errorMessage }}</span>
          </div>

          <div class="space-y-3">
            <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-[0.2em] opacity-60 px-1">{{ t('ai_provider_name') }}</label>
            <input v-model="editingProvider.name" type="text" placeholder="e.g. OpenAI" class="w-full bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-2xl px-5 py-3.5 text-[var(--text-main)] font-medium focus:outline-none focus:border-[var(--matrix-color)]/50 focus:ring-4 focus:ring-[var(--matrix-color)]/5 transition-all" />
          </div>
          <div class="space-y-3">
            <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-[0.2em] opacity-60 px-1">{{ t('ai_provider_type') }}</label>
            <div class="relative">
              <select v-model="editingProvider.type" class="w-full bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-2xl px-5 py-3.5 text-[var(--text-main)] font-bold focus:outline-none focus:border-[var(--matrix-color)]/50 focus:ring-4 focus:ring-[var(--matrix-color)]/5 transition-all appearance-none cursor-pointer">
                <option value="openai" class="bg-[var(--bg-card)]">OpenAI Compatible</option>
                <option value="azure" class="bg-[var(--bg-card)]">Azure OpenAI</option>
                <option value="anthropic" class="bg-[var(--bg-card)]">Anthropic</option>
                <option value="google" class="bg-[var(--bg-card)]">Google Gemini</option>
              </select>
              <ChevronRight class="absolute right-4 top-1/2 -translate-y-1/2 w-5 h-5 text-[var(--text-muted)] pointer-events-none rotate-90 opacity-40" />
            </div>
          </div>
          <div class="space-y-3">
            <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-[0.2em] opacity-60 px-1">{{ t('ai_base_url') }}</label>
            <input v-model="editingProvider.base_url" type="text" placeholder="https://api.openai.com/v1" class="w-full bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-2xl px-5 py-3.5 text-[var(--text-main)] font-mono focus:outline-none focus:border-[var(--matrix-color)]/50 focus:ring-4 focus:ring-[var(--matrix-color)]/5 transition-all" />
          </div>
          <div class="space-y-3">
            <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-[0.2em] opacity-60 px-1">{{ t('ai_api_key') }}</label>
            <input v-model="editingProvider.api_key" type="password" placeholder="sk-..." class="w-full bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-2xl px-5 py-3.5 text-[var(--text-main)] font-mono focus:outline-none focus:border-[var(--matrix-color)]/50 focus:ring-4 focus:ring-[var(--matrix-color)]/5 transition-all" />
          </div>
        </div>
        <div class="px-8 py-6 border-t border-[var(--border-color)] bg-[var(--bg-header)] flex justify-end gap-4">
          <button @click="showProviderModal = false" class="px-6 py-2 rounded-xl text-sm font-bold text-[var(--text-muted)] hover:text-[var(--text-main)] transition-all">{{ t('ai_cancel') }}</button>
          <button @click="saveProvider" class="px-8 py-3 bg-[var(--matrix-color)] hover:opacity-90 text-[var(--sidebar-text-active)] rounded-2xl text-sm font-bold shadow-xl shadow-[var(--matrix-color)]/20 transition-all active:scale-95">{{ t('ai_register') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 微信风格变量 */
.wechat-style {
  --wechat-bg: #ededed;
  --wechat-user-bubble: #95ec69;
  --wechat-bot-bubble: #ffffff;
  --wechat-text: #191919;
  --wechat-meta: #b2b2b2;
}

.dark .wechat-style {
  --wechat-bg: #111111;
  --wechat-user-bubble: #2ba245;
  --wechat-bot-bubble: #2c2c2c;
  --wechat-text: #dfdfdf;
  --wechat-meta: #666666;
}

.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
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
  opacity: 0.5;
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-4px); }
  75% { transform: translateX(4px); }
}
.animate-shake {
  animation: shake 0.3s ease-in-out;
}

@keyframes voice {
  0%, 100% { height: 4px; opacity: 0.5; }
  50% { height: 12px; opacity: 1; }
}

textarea {
  scrollbar-width: none;
}
textarea::-webkit-scrollbar {
  display: none;
}
</style>

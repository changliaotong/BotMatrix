import api from './index';

/**
 * Unified AI system interface planning
 */
export interface TranslationRequest {
  text: string;
  target_lang: string;
  context?: string;
}

export interface TranslationResponse {
  translated_text: string;
  source_lang: string;
  usage?: {
    total_tokens: number;
  };
}

export interface LogAnalysisRequest {
  logs: string[];
  focus?: string;
}

export interface AIStatus {
  provider: string;
  model: string;
  connected: boolean;
}

export interface AIAgent {
  id: number;
  name: string;
  description: string;
  avatar?: string;
  category?: string;
  capabilities?: string[];
  tags?: string[];
  visibility: 'public' | 'private' | 'link_only';
  revenue_rate: number;
  owner_id: number;
  model_id: number;
  system_prompt: string;
  temperature: number;
  max_tokens: number;
  is_voice: boolean;
  voice_id: string;
  voice_name: string;
  voice_lang: string;
  voice_rate: number;
  call_count: number;
  created_at?: string;
  updated_at?: string;
}

export interface AISession {
  id: number;
  session_id: string;
  user_id: number;
  agent_id: number;
  topic: string;
  last_msg: string;
  platform: string;
  status: string;
  created_at: string;
  updated_at: string;
  agent?: AIAgent;
}

export interface AIChatMessage {
  id: number;
  session_id: string;
  role: 'system' | 'user' | 'assistant' | 'tool';
  content: string;
  created_at: string;
}

export const aiApi = {
  /**
   * 获取智能体列表
   */
  getAgents: () => 
    api.get<AIAgent[]>('/api/ai/agents'),

  /**
   * 获取智能体详情
   */
  getAgentDetail: (id: number) => 
    api.get<AIAgent>(`/api/ai/agents/${id}`),

  /**
   * 获取最近会话列表
   */
  getRecentSessions: () => 
    api.get<AISession[]>('/api/ai/sessions'),

  /**
   * 获取会话历史记录
   */
  getChatHistory: (sessionId: string) => 
    api.get<AIChatMessage[]>(`/api/ai/chat/history?session_id=${sessionId}`),

  /**
   * AI log analysis interface (reserved)
   */
  analyzeLogs: (data: LogAnalysisRequest) => 
    api.post('/api/ai/analyze-logs', data),

  /**
   * System diagnosis (reserved)
   */
  getSystemDiagnosis: () => 
    api.get('/api/ai/diagnosis'),

  /**
   * Get current AI system status
   */
  getStatus: () => 
    api.get<AIStatus>('/api/ai/status')
};

export default aiApi;

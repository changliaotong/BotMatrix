import api from './index';

/**
 * 统一 AI 系统接口规划
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

export const aiApi = {
  /**
   * 日志智能分析接口 (预留)
   */
  analyzeLogs: (data: LogAnalysisRequest) => 
    api.post('/api/ai/analyze-logs', data),

  /**
   * 系统状态诊断 (预留)
   */
  getSystemDiagnosis: () => 
    api.get('/api/ai/diagnosis'),

  /**
   * 获取当前 AI 系统状态
   */
  getStatus: () => 
    api.get<AIStatus>('/api/ai/status')
};

export default aiApi;

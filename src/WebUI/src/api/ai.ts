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

export const aiApi = {
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

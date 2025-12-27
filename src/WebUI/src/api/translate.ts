import api from './index';

export interface TranslationRequest {
  text: string;
  target_lang: string;
}

export interface TranslationResponse {
  success: boolean;
  data: {
    translated_text: string;
  };
  message?: string;
}

export const translateApi = {
  /**
   * 基础翻译接口 (目前对接 Azure)
   */
  translate: (data: TranslationRequest) => 
    api.post<TranslationResponse>('/api/translate', data),
};

export default translateApi;

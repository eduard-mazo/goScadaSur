import axios from 'axios';

const API_URL = 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

export interface AppConfig {
  app: {
    name: string;
    version: string;
    description: string;
  };
  files: {
    templates: string;
    dasip_mapping: string;
    output_dir: string;
    supported_input_formats: string[];
  };
  xml: {
    lang: string;
    version: string;
    indent: string;
  };
  logging: {
    level: string;
    timestamp_format: string;
  };
  database: {
    connection_timeout: number;
    csharp_executable: string;
  };
  processing: {
    parallel_enabled: boolean;
    max_workers: number;
    buffer_size: number;
  };
}

export interface TemplateStats {
  stats: {
    total: number;
    analog: number;
    discrete: number;
    breaker: number;
  };
  warnings: string[];
}

export interface DasipConfig {
  default_path: string;
  dasip_mapping: Record<string, string>;
}

export const getAppConfig = async () => {
  const response = await api.get<AppConfig>('/config');
  return response.data;
};

export const saveAppConfig = async (data: AppConfig) => {
  const response = await api.post('/config', data);
  return response.data;
};

export const getDasipConfig = async () => {
  const response = await api.get<DasipConfig>('/dasip');
  return response.data;
};

export const saveDasipConfig = async (data: DasipConfig) => {
  const response = await api.post('/dasip', data);
  return response.data;
};

export const getRawTemplates = async () => {
  const response = await api.get<{ raw: string }>('/templates/raw');
  return response.data.raw;
};

export const saveRawTemplates = async (raw: string) => {
  const response = await api.post('/templates/raw', { raw });
  return response.data;
};

export const getTemplateStats = async () => {
  const response = await api.get<TemplateStats>('/templates');
  return response.data;
};

export const searchStation = async (data: any) => {
  const response = await api.post('/search', data);
  return response.data;
};

export const runQuery = async (data: any) => {
  const response = await api.post('/query', data);
  return response.data;
};

export default api;

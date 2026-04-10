import axios from 'axios';

const API_URL = 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Interceptor para añadir el token JWT a todas las peticiones
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('gs_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
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
  postgres: {
    host: string;
    port: number;
    user: string;
    dbname: string;
    sslmode: string;
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

export interface SearchStationRequest {
  host: string;
  path: string;
  user: string;
  password?: string;
  aor: string;
}

export interface QueryRequest {
  query: string;
}

export interface LoginRequest {
  username: string;
  password?: string;
}

export interface LoginResponse {
  token: string;
  username: string;
  role: string;
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

export const searchStation = async (data: SearchStationRequest) => {
  const response = await api.post('/search', data);
  return response.data;
};

export const runQuery = async (data: QueryRequest) => {
  const response = await api.post('/query', data);
  return response.data;
};

export const login = async (data: LoginRequest) => {
  const response = await api.post<LoginResponse>('/auth/login', data);
  if (response.data.token) {
    localStorage.setItem('gs_token', response.data.token);
    localStorage.setItem('gs_user', JSON.stringify(response.data));
  }
  return response.data;
};

export const registerUser = async (data: Record<string, unknown>) => {
  const response = await api.post('/auth/register', data);
  return response.data;
};

export const logout = () => {
  localStorage.removeItem('gs_token');
  localStorage.removeItem('gs_user');
  window.location.reload();
};

export const uploadFile = async (file: File) => {
  const formData = new FormData();
  formData.append('file', file);
  const response = await api.post('/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
  return response.data;
};

export default api;

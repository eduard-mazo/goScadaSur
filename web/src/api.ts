import axios from "axios";

const API_URL = "http://localhost:8080/api";
const api = axios.create({
  baseURL: API_URL,
  headers: { "Content-Type": "application/json" },
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem("gs_token");
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

export interface AppConfig {
  app: { name: string; version: string; description: string };
  files: {
    templates: string;
    dasip_mapping: string;
    output_dir: string;
    supported_input_formats: string[];
  };
  xml: { lang: string; version: string; indent: string };
  logging: { level: string; timestamp_format: string };
  database: { connection_timeout: number; csharp_executable: string };
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
  stats: { total: number; analog: number; discrete: number; breaker: number };
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

export interface JobLogEntry {
  time: string;
  level: "info" | "warn" | "error";
  msg: string;
}
export interface UploadResponse {
  file: string;
  success: boolean;
  logs: JobLogEntry[];
  summary: { info: number; warn: number; error: number };
  files_generated: string[];
  error?: string;
}
export interface OutputFileItem {
  name: string;
  size: number;
  modified: string;
}

export const getAppConfig = async () =>
  (await api.get<AppConfig>("/config")).data;
export const saveAppConfig = async (data: AppConfig) =>
  (await api.post("/config", data)).data;
export const getDasipConfig = async () =>
  (await api.get<DasipConfig>("/dasip")).data;
export const saveDasipConfig = async (data: DasipConfig) =>
  (await api.post("/dasip", data)).data;
export const getRawTemplates = async () =>
  (await api.get<{ raw: string }>("/templates/raw")).data.raw;
export const saveRawTemplates = async (raw: string) =>
  (await api.post("/templates/raw", { raw })).data;
export const getTemplateStats = async () =>
  (await api.get<TemplateStats>("/templates")).data;
export const searchStation = async (data: SearchStationRequest) =>
  (await api.post("/search", data)).data;
export const runQuery = async (data: QueryRequest) =>
  (await api.post("/query", data)).data;

export const login = async (data: LoginRequest) => {
  const r = await api.post<LoginResponse>("/auth/login", data);
  if (r.data.token) {
    localStorage.setItem("gs_token", r.data.token);
    localStorage.setItem("gs_user", JSON.stringify(r.data));
  }
  return r.data;
};
export const registerUser = async (data: Record<string, unknown>) =>
  (await api.post("/auth/register", data)).data;
export const logout = () => {
  localStorage.removeItem("gs_token");
  localStorage.removeItem("gs_user");
  window.location.reload();
};

export const uploadFile = async (file: File): Promise<UploadResponse> => {
  const fd = new FormData();
  fd.append("file", file);
  const r = await api.post<UploadResponse>("/upload", fd, {
    headers: { "Content-Type": "multipart/form-data" },
  });
  return r.data;
};

export const listOutputFiles = async (): Promise<OutputFileItem[]> => {
  const r = await api.get<OutputFileItem[]>("/output");
  return r.data ?? [];
};

export const getOutputFile = async (name: string): Promise<string> => {
  const r = await api.get<string>("/output/file", {
    params: { name },
    responseType: "text",
    transformResponse: [(d) => d],
  });
  return r.data;
};

export default api;

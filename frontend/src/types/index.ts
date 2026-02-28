export interface Model {
  id: number;
  name: string;
  model: string;
  strategy: 'round_robin' | 'sequential';
  sources?: Source[];
}

export interface Source {
  id: number;
  external_model_id: number;
  name: string;
  api_url: string;
  api_key: string;
  model_name: string;
  billing_mode: 'count' | 'token';
  limit_count: number;
  limit_tokens: number;
  used_count: number;
  used_tokens: number;
  limit_reset_interval: number;
  limit_reset_time: string;
  last_reset_at: string;
  is_active: boolean;
}

export interface ApiKey {
  id: number;
  key: string;
  note: string;
  external_model_id: number | null;
  used_count: number;
  input_tokens: number;
  output_tokens: number;
  used_tokens: number;
}

export interface UsageStats {
  model_stats: ModelStat[];
  source_stats: SourceStat[];
  api_key_stats: ApiKeyStat[];
}

export interface ModelStat {
  id: number;
  name: string;
  model: string;
  strategy: string;
  source_count: number;
}

export interface SourceStat {
  id: number;
  name: string;
  model_name: string;
  billing_mode: 'count' | 'token';
  used_count: number;
  used_tokens: number;
  limit_count: number;
  limit_tokens: number;
  is_active: boolean;
  last_reset_at: string;
}

export interface ApiKeyStat {
  id: number;
  key: string;
  note: string;
  used_count: number;
  input_tokens: number;
  output_tokens: number;
  used_tokens: number;
}

export interface ServerInfo {
  proxy_url: string;
}

export interface LoginResponse {
  token: string;
}

export interface ImportResult {
  external_models_created: number;
  external_models_updated: number;
  sources_created: number;
  sources_updated: number;
  errors: string[];
  skipped: string[];
}

export interface ImportResponse {
  result: ImportResult;
}
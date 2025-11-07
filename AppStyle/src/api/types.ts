// API Types matching backend structures

export interface Document {
  id: string;
  uuid: string;
  filename: string;
  file_size: number;
  upload_date: string;
  chunk_count: number;
  content_type: string;
  first_chunk_id: string;
  last_chunk_id: string;
}

export interface DocumentStats {
  total_documents: number;
  total_chunks: number;
  total_size: number;
  oldest_upload?: string;
  newest_upload?: string;
}

export interface UploadResponse {
  message: string;
  num_files_succeeded: number;
  num_files_failed: number;
  successful_file_names: string[];
  failed_file_names: Record<string, string>;
  uploaded_documents: Document[];
}

export interface QuestionContext {
  text: string;
  title: string;
}

export interface QuestionRequest {
  question: string;
  model: string;
  uuid: string;
  apiKey?: string;
}

// Configuration types
export interface SystemConfig {
  llm_provider: string;
  openai_model?: string;
  openai_embedding?: string;
  ollama_host?: string;
  ollama_model?: string;
  ollama_embedding?: string;
  vector_db: string;
  qdrant_endpoint?: string;
  pinecone_endpoint?: string;
  port: string;
}

export interface OllamaModelDetails {
  format: string;
  family: string;
  parameter_size: string;
  quantization_level: string;
}

export interface OllamaModel {
  name: string;
  model: string;
  modified_at: string;
  size: number;
  digest: string;
  details: OllamaModelDetails;
}

export interface OllamaModelsResponse {
  models: OllamaModel[];
}

export interface ConnectionTestRequest {
  type: 'ollama' | 'qdrant' | 'pinecone';
  endpoint?: string;
  api_key?: string;
  model?: string;
}

export interface ConnectionTestResponse {
  success: boolean;
  message: string;
  details?: string;
}

// Helper to get or create UUID
export function getOrCreateUUID(): string {
  const stored = localStorage.getItem('askara_uuid');
  if (stored) return stored;

  const newUUID = crypto.randomUUID();
  localStorage.setItem('askara_uuid', newUUID);
  return newUUID;
}

// Askara Backend API Client

import type {
  Document,
  DocumentStats,
  UploadResponse,
  QuestionRequest,
  QuestionContext,
  SystemConfig,
  OllamaModelsResponse,
  ConnectionTestRequest,
  ConnectionTestResponse
} from './types';
import { getOrCreateUUID } from './types';

const API_BASE = '/api';

export class AskaraAPI {
  private uuid: string;

  constructor() {
    this.uuid = getOrCreateUUID();
  }

  // Get list of documents
  async getDocuments(): Promise<Document[]> {
    const response = await fetch(`${API_BASE}/documents?uuid=${this.uuid}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch documents: ${response.statusText}`);
    }
    const data = await response.json();
    return data.documents || [];
  }

  // Get document statistics
  async getDocumentStats(): Promise<DocumentStats> {
    const response = await fetch(`${API_BASE}/documents/stats?uuid=${this.uuid}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch stats: ${response.statusText}`);
    }
    const data = await response.json();
    return data.stats;
  }

  // Upload files
  async uploadFiles(files: File[]): Promise<UploadResponse> {
    console.log('[API] Uploading files:', files.map(f => ({ name: f.name, size: f.size })));
    console.log('[API] Using UUID:', this.uuid);

    const formData = new FormData();

    files.forEach((file) => {
      formData.append('files', file);  // Changed from 'file' to 'files' to match backend
    });

    formData.append('uuid', this.uuid);

    const response = await fetch('/upload', {
      method: 'POST',
      body: formData,
    });

    console.log('[API] Upload response status:', response.status, response.statusText);

    if (!response.ok) {
      const errorText = await response.text();
      console.error('[API] Upload failed:', errorText);
      throw new Error(`Upload failed: ${response.statusText} - ${errorText}`);
    }

    const result = await response.json();
    console.log('[API] Upload result:', result);
    return result;
  }

  // Delete document
  async deleteDocument(documentId: string): Promise<void> {
    const response = await fetch(
      `${API_BASE}/documents/${documentId}?uuid=${this.uuid}`,
      {
        method: 'DELETE',
      }
    );

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Failed to delete document: ${response.status} | ${errorText}`);
    }
  }

  // Ask question with streaming response
  async askQuestionStream(
    question: string,
    onChunk: (chunk: string) => void,
    onContext?: (contexts: QuestionContext[]) => void,
    onError?: (error: string) => void
  ): Promise<void> {
    const formData = new URLSearchParams();
    formData.append('question', question);
    formData.append('model', 'GPT Turbo');
    formData.append('uuid', this.uuid);

    console.log('[API] Sending question:', {
      question,
      model: 'GPT Turbo',
      uuid: this.uuid,
      formDataString: formData.toString()
    });

    const response = await fetch(`${API_BASE}/questions/stream`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: formData.toString(),
    });

    console.log('[API] Question response status:', response.status, response.statusText);

    if (!response.ok) {
      const errorText = await response.text();
      console.error('[API] Question failed:', errorText);
      throw new Error(`Failed to ask question: ${response.statusText} - ${errorText}`);
    }

    const reader = response.body?.getReader();
    if (!reader) {
      throw new Error('No response body');
    }

    const decoder = new TextDecoder();
    let buffer = '';

    try {
      while (true) {
        const { done, value } = await reader.read();

        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';

        // Parse SSE events - iterate with index to look ahead
        for (let i = 0; i < lines.length; i++) {
          const line = lines[i];

          if (line.startsWith('event:')) {
            const eventType = line.slice(6).trim();
            // Look ahead to next line for data
            const nextLine = lines[i + 1];

            if (nextLine?.startsWith('data:')) {
              const data = nextLine.slice(6); // Skip "data: " (5 chars + 1 space)
              i++; // Skip the data line we just processed

              if (eventType === 'chunk') {
                onChunk(data);
              } else if (eventType === 'context' && onContext) {
                try {
                  const contexts = JSON.parse(data);
                  onContext(contexts);
                } catch (e) {
                  console.error('Failed to parse context:', e);
                }
              } else if (eventType === 'error' && onError) {
                onError(data);
              } else if (eventType === 'done') {
                return; // Exit completely when done
              }
            }
          }
        }
      }
    } finally {
      reader.releaseLock();
    }
  }

  // Get system configuration
  async getConfig(): Promise<SystemConfig> {
    const response = await fetch(`${API_BASE}/config`);
    if (!response.ok) {
      throw new Error(`Failed to fetch config: ${response.statusText}`);
    }
    return await response.json();
  }

  // Get available Ollama models
  async getOllamaModels(): Promise<OllamaModelsResponse> {
    const response = await fetch(`${API_BASE}/config/ollama/models`);
    if (!response.ok) {
      throw new Error(`Failed to fetch Ollama models: ${response.statusText}`);
    }
    return await response.json();
  }

  // Test connection to a service
  async testConnection(request: ConnectionTestRequest): Promise<ConnectionTestResponse> {
    const response = await fetch(`${API_BASE}/config/test`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error(`Failed to test connection: ${response.statusText}`);
    }
    return await response.json();
  }
}

// Export singleton instance
export const api = new AskaraAPI();

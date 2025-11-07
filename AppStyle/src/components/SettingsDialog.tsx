import { useState, useEffect } from 'react';
import { Settings, Database, Cpu, HardDrive, CheckCircle2, XCircle, Loader2 } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from './ui/dialog';
import { Button } from './ui/button';
import { Label } from './ui/label';
import { Separator } from './ui/separator';
import { Badge } from './ui/badge';
import { Card } from './ui/card';
import { ScrollArea } from './ui/scroll-area';
import { toast } from 'sonner@2.0.3';
import { api } from '../api/client';
import type { SystemConfig, OllamaModel, ConnectionTestResponse } from '../api/types';

export function SettingsDialog() {
  const [open, setOpen] = useState(false);
  const [config, setConfig] = useState<SystemConfig | null>(null);
  const [ollamaModels, setOllamaModels] = useState<OllamaModel[]>([]);
  const [loading, setLoading] = useState(true);
  const [testing, setTesting] = useState<Record<string, boolean>>({});
  const [testResults, setTestResults] = useState<Record<string, ConnectionTestResponse>>({});

  useEffect(() => {
    if (open) {
      loadConfiguration();
    }
  }, [open]);

  const loadConfiguration = async () => {
    setLoading(true);
    try {
      const configData = await api.getConfig();
      setConfig(configData);

      // Load Ollama models if using Ollama
      if (configData.llm_provider === 'ollama' && configData.ollama_host) {
        try {
          const modelsData = await api.getOllamaModels();
          setOllamaModels(modelsData.models || []);
        } catch (error) {
          console.error('Failed to load Ollama models:', error);
        }
      }
    } catch (error) {
      toast.error('Failed to load configuration');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const testConnection = async (type: 'ollama' | 'qdrant' | 'pinecone', endpoint?: string) => {
    setTesting({ ...testing, [type]: true });
    try {
      const result = await api.testConnection({ type, endpoint });
      setTestResults({ ...testResults, [type]: result });
      toast[result.success ? 'success' : 'error'](result.message);
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : 'Connection test failed';
      setTestResults({
        ...testResults,
        [type]: { success: false, message: errorMsg },
      });
      toast.error(errorMsg);
    } finally {
      setTesting({ ...testing, [type]: false });
    }
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="ghost" size="icon">
          <Settings className="h-5 w-5" />
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl max-h-[90vh] flex flex-col p-0">
        <div className="p-6 pb-4">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Database className="h-5 w-5" />
              System Configuration
            </DialogTitle>
            <DialogDescription>
              View and test your Askara system configuration
            </DialogDescription>
          </DialogHeader>
        </div>

        {loading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : config ? (
          <ScrollArea className="flex-1 overflow-y-auto px-6">
            <div className="space-y-6 pb-4">
              {/* LLM Provider Section */}
              <div className="space-y-3">
                <div className="flex items-center gap-2">
                  <Cpu className="h-4 w-4 text-primary" />
                  <h3 className="font-semibold">LLM Provider</h3>
                </div>
                <Card className="p-4">
                  <div className="space-y-3">
                    <div className="flex justify-between items-center">
                      <Label className="text-muted-foreground">Provider</Label>
                      <Badge variant="secondary">{config.llm_provider.toUpperCase()}</Badge>
                    </div>

                    {config.llm_provider === 'ollama' && (
                      <>
                        <div className="flex justify-between items-center">
                          <Label className="text-muted-foreground">Ollama Host</Label>
                          <span className="text-sm font-mono">{config.ollama_host}</span>
                        </div>
                        <div className="flex justify-between items-center">
                          <Label className="text-muted-foreground">Chat Model</Label>
                          <Badge>{config.ollama_model}</Badge>
                        </div>
                        <div className="flex justify-between items-center">
                          <Label className="text-muted-foreground">Embedding Model</Label>
                          <Badge>{config.ollama_embedding}</Badge>
                        </div>
                        <Separator />
                        <div className="space-y-2">
                          <Button
                            onClick={() => testConnection('ollama', config.ollama_host)}
                            disabled={testing.ollama}
                            className="w-full"
                            variant="outline"
                          >
                            {testing.ollama ? (
                              <>
                                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                Testing...
                              </>
                            ) : (
                              'Test Ollama Connection'
                            )}
                          </Button>
                          {testResults.ollama && (
                            <div
                              className={`flex items-center gap-2 text-sm p-2 rounded ${
                                testResults.ollama.success
                                  ? 'bg-green-50 text-green-700 dark:bg-green-950 dark:text-green-300'
                                  : 'bg-red-50 text-red-700 dark:bg-red-950 dark:text-red-300'
                              }`}
                            >
                              {testResults.ollama.success ? (
                                <CheckCircle2 className="h-4 w-4" />
                              ) : (
                                <XCircle className="h-4 w-4" />
                              )}
                              <span>{testResults.ollama.message}</span>
                            </div>
                          )}
                        </div>
                      </>
                    )}

                    {config.llm_provider === 'openai' && (
                      <>
                        <div className="flex justify-between items-center">
                          <Label className="text-muted-foreground">Chat Model</Label>
                          <Badge>{config.openai_model}</Badge>
                        </div>
                        <div className="flex justify-between items-center">
                          <Label className="text-muted-foreground">Embedding Model</Label>
                          <Badge>{config.openai_embedding}</Badge>
                        </div>
                      </>
                    )}
                  </div>
                </Card>
              </div>

              {/* Available Ollama Models */}
              {ollamaModels.length > 0 && (
                <div className="space-y-3">
                  <div className="flex items-center gap-2">
                    <Cpu className="h-4 w-4 text-primary" />
                    <h3 className="font-semibold">Available Ollama Models ({ollamaModels.length})</h3>
                  </div>
                  <div className="space-y-2">
                    {ollamaModels.map((model) => (
                      <Card key={model.digest} className="p-3">
                        <div className="flex justify-between items-start mb-2">
                          <h4 className="font-medium">{model.name}</h4>
                          <Badge variant="outline">{formatBytes(model.size)}</Badge>
                        </div>
                        <div className="grid grid-cols-3 gap-2 text-xs text-muted-foreground">
                          <div>
                            <Label className="text-xs">Family</Label>
                            <p>{model.details.family}</p>
                          </div>
                          <div>
                            <Label className="text-xs">Parameters</Label>
                            <p>{model.details.parameter_size}</p>
                          </div>
                          <div>
                            <Label className="text-xs">Quantization</Label>
                            <p>{model.details.quantization_level}</p>
                          </div>
                        </div>
                      </Card>
                    ))}
                  </div>
                </div>
              )}

              {/* Vector Database Section */}
              <div className="space-y-3">
                <div className="flex items-center gap-2">
                  <HardDrive className="h-4 w-4 text-primary" />
                  <h3 className="font-semibold">Vector Database</h3>
                </div>
                <Card className="p-4">
                  <div className="space-y-3">
                    <div className="flex justify-between items-center">
                      <Label className="text-muted-foreground">Database</Label>
                      <Badge variant="secondary">{config.vector_db.toUpperCase()}</Badge>
                    </div>

                    {config.vector_db === 'qdrant' && config.qdrant_endpoint && (
                      <>
                        <div className="flex justify-between items-center">
                          <Label className="text-muted-foreground">Endpoint</Label>
                          <span className="text-sm font-mono">{config.qdrant_endpoint}</span>
                        </div>
                        <Separator />
                        <div className="space-y-2">
                          <Button
                            onClick={() => testConnection('qdrant', config.qdrant_endpoint)}
                            disabled={testing.qdrant}
                            className="w-full"
                            variant="outline"
                          >
                            {testing.qdrant ? (
                              <>
                                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                Testing...
                              </>
                            ) : (
                              'Test Qdrant Connection'
                            )}
                          </Button>
                          {testResults.qdrant && (
                            <div
                              className={`flex items-center gap-2 text-sm p-2 rounded ${
                                testResults.qdrant.success
                                  ? 'bg-green-50 text-green-700 dark:bg-green-950 dark:text-green-300'
                                  : 'bg-red-50 text-red-700 dark:bg-red-950 dark:text-red-300'
                              }`}
                            >
                              {testResults.qdrant.success ? (
                                <CheckCircle2 className="h-4 w-4" />
                              ) : (
                                <XCircle className="h-4 w-4" />
                              )}
                              <span>{testResults.qdrant.message}</span>
                            </div>
                          )}
                        </div>
                      </>
                    )}

                    {config.vector_db === 'pinecone' && config.pinecone_endpoint && (
                      <>
                        <div className="flex justify-between items-center">
                          <Label className="text-muted-foreground">Endpoint</Label>
                          <span className="text-sm font-mono">{config.pinecone_endpoint}</span>
                        </div>
                        <Separator />
                        <div className="space-y-2">
                          <Button
                            onClick={() => testConnection('pinecone', config.pinecone_endpoint)}
                            disabled={testing.pinecone}
                            className="w-full"
                            variant="outline"
                          >
                            {testing.pinecone ? (
                              <>
                                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                Testing...
                              </>
                            ) : (
                              'Test Pinecone Connection'
                            )}
                          </Button>
                          {testResults.pinecone && (
                            <div
                              className={`flex items-center gap-2 text-sm p-2 rounded ${
                                testResults.pinecone.success
                                  ? 'bg-green-50 text-green-700 dark:bg-green-950 dark:text-green-300'
                                  : 'bg-red-50 text-red-700 dark:bg-red-950 dark:text-red-300'
                              }`}
                            >
                              {testResults.pinecone.success ? (
                                <CheckCircle2 className="h-4 w-4" />
                              ) : (
                                <XCircle className="h-4 w-4" />
                              )}
                              <span>{testResults.pinecone.message}</span>
                            </div>
                          )}
                        </div>
                      </>
                    )}
                  </div>
                </Card>
              </div>

              {/* Application Settings */}
              <div className="space-y-3">
                <h3 className="font-semibold">Application</h3>
                <Card className="p-4">
                  <div className="flex justify-between items-center">
                    <Label className="text-muted-foreground">Server Port</Label>
                    <Badge variant="secondary">{config.port}</Badge>
                  </div>
                </Card>
              </div>

              <div className="rounded-lg bg-muted p-3 text-sm">
                <p className="text-muted-foreground">
                  <strong>Note:</strong> To change these settings, update your .env file and restart the application.
                </p>
              </div>
            </div>
          </ScrollArea>
        ) : (
          <div className="text-center py-12 text-muted-foreground">
            Failed to load configuration
          </div>
        )}

        <div className="flex justify-end p-6 pt-4 border-t">
          <Button onClick={() => setOpen(false)}>Close</Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

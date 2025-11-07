import React, { useState, useEffect } from 'react';
import Page from '../../Page';
import s from './index.less';

type Props = {
    history: Object,
};

const ConfigPage = (props: Props): React.Node => {
    const [config, setConfig] = useState(null);
    const [ollamaModels, setOllamaModels] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [testResults, setTestResults] = useState({});
    const [testing, setTesting] = useState({});

    useEffect(() => {
        loadConfiguration();
    }, []);

    const loadConfiguration = async () => {
        try {
            const response = await fetch('/api/config');
            if (!response.ok) {
                throw new Error('Failed to load configuration');
            }
            const data = await response.json();
            setConfig(data);

            // If using Ollama, load models
            if (data.llm_provider === 'ollama' || data.ollama_host) {
                loadOllamaModels();
            }
            setLoading(false);
        } catch (err) {
            setError(err.message);
            setLoading(false);
        }
    };

    const loadOllamaModels = async () => {
        try {
            const response = await fetch('/api/config/ollama/models');
            if (!response.ok) {
                console.error('Failed to load Ollama models');
                return;
            }
            const data = await response.json();
            setOllamaModels(data.models || []);
        } catch (err) {
            console.error('Error loading Ollama models:', err);
        }
    };

    const testConnection = async (type, endpoint = null, apiKey = null) => {
        setTesting({ ...testing, [type]: true });
        setTestResults({ ...testResults, [type]: null });

        try {
            const response = await fetch('/api/config/test', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ type, endpoint, api_key: apiKey }),
            });

            if (!response.ok) {
                throw new Error('Connection test failed');
            }

            const result = await response.json();
            setTestResults({ ...testResults, [type]: result });
        } catch (err) {
            setTestResults({
                ...testResults,
                [type]: { success: false, message: err.message },
            });
        } finally {
            setTesting({ ...testing, [type]: false });
        }
    };

    const formatBytes = (bytes) => {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
    };

    const goBack = () => {
        props.history.push('/');
    };

    if (loading) {
        return (
            <Page title="Configuration">
                <div className={s.configPage}>
                    <div className={s.loading}>Loading configuration...</div>
                </div>
            </Page>
        );
    }

    if (error) {
        return (
            <Page title="Configuration">
                <div className={s.configPage}>
                    <div className={s.error}>Error: {error}</div>
                </div>
            </Page>
        );
    }

    return (
        <Page title="Configuration">
            <div className={s.configPage}>
                <div className={s.header}>
                    <button onClick={goBack} className={s.backButton}>
                        ← Back to Home
                    </button>
                    <h1>System Configuration</h1>
                </div>

                {/* LLM Provider Section */}
                <div className={s.section}>
                    <h2>LLM Provider</h2>
                    <div className={s.configGroup}>
                        <div className={s.configItem}>
                            <label>Provider:</label>
                            <span className={s.value}>{config.llm_provider}</span>
                        </div>

                        {config.llm_provider === 'ollama' && (
                            <>
                                <div className={s.configItem}>
                                    <label>Ollama Host:</label>
                                    <span className={s.value}>{config.ollama_host}</span>
                                </div>
                                <div className={s.configItem}>
                                    <label>Chat Model:</label>
                                    <span className={s.value}>{config.ollama_model}</span>
                                </div>
                                <div className={s.configItem}>
                                    <label>Embedding Model:</label>
                                    <span className={s.value}>{config.ollama_embedding}</span>
                                </div>
                                <div className={s.testConnection}>
                                    <button
                                        onClick={() => testConnection('ollama', config.ollama_host)}
                                        disabled={testing.ollama}
                                        className={s.testButton}>
                                        {testing.ollama ? 'Testing...' : 'Test Ollama Connection'}
                                    </button>
                                    {testResults.ollama && (
                                        <div className={testResults.ollama.success ? s.success : s.failure}>
                                            {testResults.ollama.message}
                                            {testResults.ollama.details && (
                                                <div className={s.details}>{testResults.ollama.details}</div>
                                            )}
                                        </div>
                                    )}
                                </div>
                            </>
                        )}

                        {config.llm_provider === 'openai' && (
                            <>
                                <div className={s.configItem}>
                                    <label>Chat Model:</label>
                                    <span className={s.value}>{config.openai_model}</span>
                                </div>
                                <div className={s.configItem}>
                                    <label>Embedding Model:</label>
                                    <span className={s.value}>{config.openai_embedding}</span>
                                </div>
                            </>
                        )}
                    </div>
                </div>

                {/* Available Ollama Models */}
                {ollamaModels.length > 0 && (
                    <div className={s.section}>
                        <h2>Available Ollama Models ({ollamaModels.length})</h2>
                        <div className={s.modelList}>
                            {ollamaModels.map((model) => (
                                <div key={model.name} className={s.modelCard}>
                                    <div className={s.modelHeader}>
                                        <h3>{model.name}</h3>
                                        <span className={s.modelSize}>{formatBytes(model.size)}</span>
                                    </div>
                                    <div className={s.modelDetails}>
                                        <div className={s.modelDetail}>
                                            <label>Family:</label>
                                            <span>{model.details.family}</span>
                                        </div>
                                        <div className={s.modelDetail}>
                                            <label>Parameters:</label>
                                            <span>{model.details.parameter_size}</span>
                                        </div>
                                        <div className={s.modelDetail}>
                                            <label>Quantization:</label>
                                            <span>{model.details.quantization_level}</span>
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                )}

                {/* Vector Database Section */}
                <div className={s.section}>
                    <h2>Vector Database</h2>
                    <div className={s.configGroup}>
                        <div className={s.configItem}>
                            <label>Database:</label>
                            <span className={s.value}>{config.vector_db}</span>
                        </div>

                        {config.vector_db === 'qdrant' && (
                            <>
                                <div className={s.configItem}>
                                    <label>Endpoint:</label>
                                    <span className={s.value}>{config.qdrant_endpoint}</span>
                                </div>
                                <div className={s.testConnection}>
                                    <button
                                        onClick={() => testConnection('qdrant', config.qdrant_endpoint)}
                                        disabled={testing.qdrant}
                                        className={s.testButton}>
                                        {testing.qdrant ? 'Testing...' : 'Test Qdrant Connection'}
                                    </button>
                                    {testResults.qdrant && (
                                        <div className={testResults.qdrant.success ? s.success : s.failure}>
                                            {testResults.qdrant.message}
                                        </div>
                                    )}
                                </div>
                            </>
                        )}

                        {config.vector_db === 'pinecone' && (
                            <>
                                <div className={s.configItem}>
                                    <label>Endpoint:</label>
                                    <span className={s.value}>{config.pinecone_endpoint}</span>
                                </div>
                                <div className={s.testConnection}>
                                    <button
                                        onClick={() => testConnection('pinecone', config.pinecone_endpoint)}
                                        disabled={testing.pinecone}
                                        className={s.testButton}>
                                        {testing.pinecone ? 'Testing...' : 'Test Pinecone Connection'}
                                    </button>
                                    {testResults.pinecone && (
                                        <div className={testResults.pinecone.success ? s.success : s.failure}>
                                            {testResults.pinecone.message}
                                        </div>
                                    )}
                                </div>
                            </>
                        )}
                    </div>
                </div>

                {/* Application Settings */}
                <div className={s.section}>
                    <h2>Application</h2>
                    <div className={s.configGroup}>
                        <div className={s.configItem}>
                            <label>Server Port:</label>
                            <span className={s.value}>{config.port}</span>
                        </div>
                    </div>
                </div>

                <div className={s.footer}>
                    <p>
                        To change these settings, update your <code>.env</code> file and restart the application.
                    </p>
                </div>
            </div>
        </Page>
    );
};

export default ConfigPage;

import React, { useState } from 'react';
import { marked } from 'marked';
import { useDropzone } from 'react-dropzone';
import Page from '../../Page';
import PostAPI from '../../Util/PostAPI';
import Go from '../../Go';

import s from './index.less';

type Props = {
    history: Object,
};

const ContextSnippet = ({ context, index }) => {
    const [collapsed, setCollapsed] = useState(true);

    const toggleCollapse = () => {
        setCollapsed(!collapsed);
    };

    const preview =
        context.text.slice(0, 222) + (context.text.length > 222 ? '...' : '');

    return (
        <div className={s.contextSnippet}>
            <h4 onClick={toggleCollapse}>
                Context {index + 1}
                <span className={s.title}>
                    {context.title}
                    <span className={s.arrow}>{collapsed ? '▼' : '▲'}</span>
                </span>
            </h4>
            <div className={collapsed ? s.collapsed : s.expanded}>
                {collapsed ? preview : context.text}
            </div>
        </div>
    );
};

const ResponseDisplay = ({ response, streamingAnswer }) => {
    if (!response && !streamingAnswer) return null;

    const answerToDisplay = streamingAnswer || (response ? response.answer : '');
    const renderedMarkdown = marked(answerToDisplay);

    return (
        <div className={s.responseDisplay}>
            <div
                className={s.answer}
                dangerouslySetInnerHTML={{ __html: renderedMarkdown }}
            />
            {response && response.context && response.context.length > 0 && (
                <div className={s.contexts}>
                    {response.context.map((context, index) => (
                        <ContextSnippet
                            key={index}
                            context={context}
                            index={index}
                        />
                    ))}
                </div>
            )}
        </div>
    );
};

const LandingPage = (props: Props): React.Node => {
    const [question, setQuestion] = useState('');
    const [uploadedFiles, setUploadedFiles] = useState([]);
    const [failedFiles, setFailedFiles] = useState({});
    const [errorMessage, setErrorMessage] = useState('');
    const [response, setResponse] = useState(null);
    const [loading, setLoading] = useState(false);
    const [streamingAnswer, setStreamingAnswer] = useState('');
    const [useStreaming, setUseStreaming] = useState(true);

    React.useEffect(() => {}, []);

    const handleAskQuestion = () => {
        // Perform question submission here
        // Set loading state to true
        setLoading(true);
        setErrorMessage('');
        setStreamingAnswer('');
        console.log('Asking question:', question);

        if (useStreaming) {
            // Use streaming API
            let contextData = [];
            PostAPI.questions
                .submitQuestionStream(
                    question,
                    'GPT Turbo',
                    (chunk) => {
                        // Append each chunk to the streaming answer
                        setStreamingAnswer((prev) => prev + chunk);
                    },
                    (context) => {
                        // Set context data when received
                        contextData = context;
                        setResponse({ answer: '', context: contextData, tokens: 0 });
                    }
                )
                .promise.then(() => {
                    console.log('Streaming completed');
                    setLoading(false);
                })
                .catch((error) => {
                    console.log('Error asking question:', error);
                    setErrorMessage(
                        error ? error.message || JSON.stringify(error) : 'Error asking question'
                    );
                    setResponse(null);
                    setStreamingAnswer('');
                    setLoading(false);
                });
        } else {
            // Use regular non-streaming API
            PostAPI.questions
                .submitQuestion(question, 'GPT Turbo')
                .promise.then((response) => {
                    console.log('Response:', response);
                    setResponse(response);
                    setLoading(false);
                })
                .catch((error) => {
                    console.log('Error asking question:', error);
                    setErrorMessage(
                        error ? JSON.stringify(error) : 'Error asking question'
                    );
                    setResponse(null);
                    setLoading(false);
                });
        }
    };

	React.useEffect(() => {
    const handleKeyDown = (e) => {
        if (e.key === 'Enter' && e.ctrlKey && !loading && question) {
            handleAskQuestion();
        }
    };

    document.addEventListener('keydown', handleKeyDown);

    return () => {
        document.removeEventListener('keydown', handleKeyDown);
    };
}, [loading, question]);

    const handleFileUpload = (files) => {
        setUploadedFiles([]);
        setFailedFiles({});
        setErrorMessage('');
        setLoading(true);

        PostAPI.upload
            .uploadFiles(files)
            .promise.then((response) => {
                console.log('Response:', response);
                setUploadedFiles(response.successful_file_names ?? []);
                setFailedFiles(response.failed_file_names ?? {});
                setLoading(false);
            })
            .catch((error) => {
                console.log('Error uploading files:', error);
                setErrorMessage(
                    error?.message
                        ? 'Error: ' + error?.message
                        : 'Error uploading files'
                );
                setUploadedFiles([]);
                setFailedFiles({});
                setLoading(false);
            });
    };

    const { getRootProps, getInputProps } = useDropzone({
        onDrop: handleFileUpload,
        multiple: true,
    });

    const failedFilenames = Object.keys(failedFiles);

    return (
        <Page
            pageClass={s.page}
            contentBackgroundClass={s.background}
            contentPreferredWidth={1300}
            contentClass={s.pageContent}>
            <div className={s.text}>
                <h1>Askara - AI-Powered Knowledge Base</h1>
                {errorMessage && <div className={s.error}>{errorMessage}</div>}
                <div className={s.workArea}>
                    <div className={s.leftColumn}>
                        <div
                            {...getRootProps()}
                            className={[
                                s.dropzone,
                                loading ? s.dropzoneLoading : '',
                            ].join(' ')}>
                            <input {...getInputProps()} disabled={loading} />
                            <p>
                                {loading
                                    ? 'Loading...'
                                    : 'Drag and drop files to add to the knowledgebase here, or click to select files'}
                            </p>
                        </div>
                        <div className={s.questionInput}>
                            <textarea
                                value={question}
                                onChange={(e) => setQuestion(e.target.value)}
                                placeholder="Enter your question here..."
                                className={s.textarea}
                            />
                            <div style={{ display: 'flex', gap: '10px', alignItems: 'center', marginTop: '10px' }}>
                                <button
                                    onClick={handleAskQuestion}
                                    disabled={loading || !question}
                                    className={
                                        loading
                                            ? s.askQuestionDisabled
                                            : s.askQuestion
                                    }>
                                    Submit
                                </button>
                                <label style={{ display: 'flex', alignItems: 'center', gap: '5px', cursor: 'pointer' }}>
                                    <input
                                        type="checkbox"
                                        checked={useStreaming}
                                        onChange={(e) => setUseStreaming(e.target.checked)}
                                        disabled={loading}
                                    />
                                    <span>Enable Streaming</span>
                                </label>
                            </div>
                            {loading && <div className={s.loader} />}
                        </div>
                        {response?.tokens && (
                            <div className={s.tokenCount}>
                                TOTAL TOKENS USED:{' '}
                                <span style={{ fontWeight: 900 }}>
                                    {response.tokens}
                                </span>
                            </div>
                        )}
                        {(response || streamingAnswer) && (
                            <ResponseDisplay
                                response={response}
                                streamingAnswer={streamingAnswer}
                            />
                        )}
                        <div style={{ height: 32 }} />
                        <div className={s.fileList}>
                            {uploadedFiles.length > 0 && (
                                <div className={s.successfulFiles}>
                                    <h4>Uploaded Files:</h4>
                                    <ul>
                                        {uploadedFiles.map((file, index) => (
                                            <li key={index}>{file}</li>
                                        ))}
                                    </ul>
                                </div>
                            )}
                            {failedFilenames.length > 0 && (
                                <div className={s.failedFiles}>
                                    <h4>Failed Files:</h4>
                                    <ul>
                                        {failedFilenames.map((file, index) => (
                                            <li key={index}>
                                                {file}
                                                <div>
                                                    Reason: {failedFiles[file]}
                                                </div>
                                            </li>
                                        ))}
                                    </ul>
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            </div>
        </Page>
    );
};

export default LandingPage;

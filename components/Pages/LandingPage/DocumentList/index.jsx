import React, { useState, useEffect } from 'react';
import PostAPI from '../../../Util/PostAPI';
import s from './index.less';

const formatFileSize = (bytes) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
};

const formatDate = (dateString) => {
    const date = new Date(dateString);
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString();
};

const DeleteConfirmDialog = ({ document, onConfirm, onCancel }) => {
    return (
        <div className={s.modalOverlay} onClick={onCancel}>
            <div className={s.modal} onClick={(e) => e.stopPropagation()}>
                <h3>Confirm Deletion</h3>
                <p>Are you sure you want to delete this document?</p>
                <div className={s.documentInfo}>
                    <strong>{document.filename}</strong>
                    <div className={s.documentMeta}>
                        {formatFileSize(document.file_size)} • {document.chunk_count} chunks
                    </div>
                </div>
                <p className={s.warning}>
                    This action cannot be undone. All vectors and metadata will be permanently removed.
                </p>
                <div className={s.modalButtons}>
                    <button className={s.cancelButton} onClick={onCancel}>
                        Cancel
                    </button>
                    <button className={s.deleteButton} onClick={onConfirm}>
                        Delete
                    </button>
                </div>
            </div>
        </div>
    );
};

const DocumentList = ({ onDocumentsChange }) => {
    const [documents, setDocuments] = useState([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [deletingId, setDeletingId] = useState(null);
    const [confirmDelete, setConfirmDelete] = useState(null);
    const [stats, setStats] = useState(null);

    const loadDocuments = () => {
        setLoading(true);
        setError('');

        PostAPI.documents.listDocuments()
            .promise.then((response) => {
                console.log('Documents loaded:', response);
                setDocuments(response.documents || []);
                setLoading(false);
                if (onDocumentsChange) {
                    onDocumentsChange(response.documents || []);
                }
            })
            .catch((err) => {
                console.error('Error loading documents:', err);
                setError('Failed to load documents');
                setLoading(false);
            });
    };

    const loadStats = () => {
        PostAPI.documents.getStats()
            .promise.then((response) => {
                console.log('Stats loaded:', response);
                setStats(response.stats);
            })
            .catch((err) => {
                console.error('Error loading stats:', err);
            });
    };

    useEffect(() => {
        loadDocuments();
        loadStats();
    }, []);

    const handleDeleteClick = (doc) => {
        setConfirmDelete(doc);
    };

    const handleDeleteConfirm = () => {
        if (!confirmDelete) return;

        setDeletingId(confirmDelete.id);
        setError('');

        PostAPI.documents.deleteDocument(confirmDelete.id)
            .promise.then(() => {
                console.log('Document deleted');
                setDeletingId(null);
                setConfirmDelete(null);
                loadDocuments();
                loadStats();
            })
            .catch((err) => {
                console.error('Error deleting document:', err);
                setError('Failed to delete document: ' + err.message);
                setDeletingId(null);
                setConfirmDelete(null);
            });
    };

    const handleDeleteCancel = () => {
        setConfirmDelete(null);
    };

    const handleRefresh = () => {
        loadDocuments();
        loadStats();
    };

    if (loading && documents.length === 0) {
        return (
            <div className={s.documentList}>
                <div className={s.header}>
                    <h3>My Documents</h3>
                </div>
                <div className={s.loading}>Loading documents...</div>
            </div>
        );
    }

    return (
        <div className={s.documentList}>
            <div className={s.header}>
                <h3>My Documents</h3>
                <button className={s.refreshButton} onClick={handleRefresh} disabled={loading}>
                    {loading ? 'Loading...' : 'Refresh'}
                </button>
            </div>

            {error && <div className={s.error}>{error}</div>}

            {stats && (
                <div className={s.stats}>
                    <div className={s.statItem}>
                        <span className={s.statLabel}>Total Documents:</span>
                        <span className={s.statValue}>{stats.total_documents}</span>
                    </div>
                    <div className={s.statItem}>
                        <span className={s.statLabel}>Total Chunks:</span>
                        <span className={s.statValue}>{stats.total_chunks}</span>
                    </div>
                    <div className={s.statItem}>
                        <span className={s.statLabel}>Total Size:</span>
                        <span className={s.statValue}>{formatFileSize(stats.total_size)}</span>
                    </div>
                </div>
            )}

            {documents.length === 0 ? (
                <div className={s.empty}>
                    No documents uploaded yet. Upload some files to get started!
                </div>
            ) : (
                <div className={s.table}>
                    <div className={s.tableHeader}>
                        <div className={s.col1}>Filename</div>
                        <div className={s.col2}>Upload Date</div>
                        <div className={s.col3}>Size</div>
                        <div className={s.col4}>Chunks</div>
                        <div className={s.col5}>Actions</div>
                    </div>
                    {documents.map((doc) => (
                        <div key={doc.id} className={s.tableRow}>
                            <div className={s.col1} title={doc.filename}>
                                {doc.filename}
                            </div>
                            <div className={s.col2}>{formatDate(doc.upload_date)}</div>
                            <div className={s.col3}>{formatFileSize(doc.file_size)}</div>
                            <div className={s.col4}>{doc.chunk_count}</div>
                            <div className={s.col5}>
                                <button
                                    className={s.deleteBtn}
                                    onClick={() => handleDeleteClick(doc)}
                                    disabled={deletingId === doc.id}>
                                    {deletingId === doc.id ? 'Deleting...' : 'Delete'}
                                </button>
                            </div>
                        </div>
                    ))}
                </div>
            )}

            {confirmDelete && (
                <DeleteConfirmDialog
                    document={confirmDelete}
                    onConfirm={handleDeleteConfirm}
                    onCancel={handleDeleteCancel}
                />
            )}
        </div>
    );
};

export default DocumentList;

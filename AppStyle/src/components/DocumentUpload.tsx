import { useState, useRef, useEffect } from 'react';
import { Upload, File, X, Loader2, CheckCircle2, AlertCircle } from 'lucide-react';
import { Button } from './ui/button';
import { Card } from './ui/card';

interface DocumentUploadProps {
  onUpload: (files: File[]) => Promise<void>;
}

export function DocumentUpload({ onUpload }: DocumentUploadProps) {
  const [isDragging, setIsDragging] = useState(false);
  const [selectedFiles, setSelectedFiles] = useState<File[]>([]);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [uploadStatus, setUploadStatus] = useState('');
  const [uploadComplete, setUploadComplete] = useState(false);
  const [uploadError, setUploadError] = useState('');
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    
    const files = Array.from(e.dataTransfer.files);
    setSelectedFiles(files);
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const files = Array.from(e.target.files);
      setSelectedFiles(files);
    }
  };

  const handleUpload = async () => {
    if (selectedFiles.length > 0 && !isUploading) {
      setIsUploading(true);
      setUploadProgress(0);
      setUploadStatus('Preparing upload...');
      setUploadComplete(false);
      setUploadError('');

      // Simulate progress for better UX
      const progressInterval = setInterval(() => {
        setUploadProgress(prev => {
          if (prev < 90) return prev + 1;
          return prev;
        });
      }, 200);

      try {
        // Update status messages during upload
        setTimeout(() => setUploadStatus('Uploading files...'), 500);
        setTimeout(() => setUploadStatus('Processing document with ML Worker...'), 2000);
        setTimeout(() => setUploadStatus('Extracting text and generating embeddings...'), 10000);
        setTimeout(() => setUploadStatus('Almost done, finalizing...'), 30000);

        await onUpload(selectedFiles);

        clearInterval(progressInterval);
        setUploadProgress(100);
        setUploadStatus('Upload complete!');
        setUploadComplete(true);

        // Clear files after successful upload
        setTimeout(() => {
          setSelectedFiles([]);
          setUploadProgress(0);
          setUploadStatus('');
          setUploadComplete(false);
          if (fileInputRef.current) {
            fileInputRef.current.value = '';
          }
        }, 2000);
      } catch (error) {
        clearInterval(progressInterval);
        console.error('Upload error in component:', error);
        setUploadError(error instanceof Error ? error.message : 'Upload failed');
        setUploadStatus('Upload failed');
        setUploadProgress(0);
      } finally {
        setIsUploading(false);
      }
    }
  };

  const removeFile = (index: number) => {
    setSelectedFiles(selectedFiles.filter((_, i) => i !== index));
  };

  return (
    <div className="space-y-4 p-4 md:p-6">
      <div
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        className={`border-2 border-dashed rounded-lg p-8 md:p-12 text-center transition-colors ${
          isDragging
            ? 'border-primary bg-accent'
            : 'border-muted-foreground/25 hover:border-muted-foreground/50'
        }`}
      >
        <Upload className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
        <h3 className="mb-2">Upload Documents</h3>
        <p className="text-muted-foreground mb-4">
          Drag and drop files here, or click to browse
        </p>
        <input
          ref={fileInputRef}
          type="file"
          multiple
          onChange={handleFileSelect}
          className="hidden"
          accept=".pdf,.doc,.docx,.txt"
        />
        <Button
          onClick={() => fileInputRef.current?.click()}
          variant="outline"
        >
          Browse Files
        </Button>
      </div>

      {selectedFiles.length > 0 && (
        <Card className="p-4">
          <h4 className="mb-3">Selected Files ({selectedFiles.length})</h4>
          <div className="space-y-2 mb-4 max-h-[40vh] overflow-y-auto pr-2">
            {selectedFiles.map((file, index) => (
              <div
                key={index}
                className="flex items-center gap-3 p-2 rounded-lg bg-muted"
              >
                <File className="h-4 w-4 flex-shrink-0" />
                <span className="flex-1 truncate">{file.name}</span>
                <span className="text-muted-foreground text-sm flex-shrink-0">
                  {(file.size / 1024).toFixed(1)} KB
                </span>
                <Button
                  size="icon"
                  variant="ghost"
                  onClick={() => removeFile(index)}
                  className="h-8 w-8 flex-shrink-0"
                >
                  <X className="h-4 w-4" />
                </Button>
              </div>
            ))}
          </div>
          {/* Progress Bar */}
          {(isUploading || uploadComplete || uploadError) && (
            <div className="mb-4 space-y-2">
              <div className="flex items-center justify-between text-sm">
                <div className="flex items-center gap-2">
                  {isUploading && <Loader2 className="h-4 w-4 animate-spin text-primary" />}
                  {uploadComplete && <CheckCircle2 className="h-4 w-4 text-green-500" />}
                  {uploadError && <AlertCircle className="h-4 w-4 text-red-500" />}
                  <span className={uploadComplete ? 'text-green-600' : uploadError ? 'text-red-600' : ''}>
                    {uploadStatus}
                  </span>
                </div>
                {isUploading && <span className="text-muted-foreground">{uploadProgress}%</span>}
              </div>
              <div className="w-full bg-muted rounded-full h-2 overflow-hidden">
                <div
                  className={`h-full transition-all duration-300 ${
                    uploadComplete ? 'bg-green-500' :
                    uploadError ? 'bg-red-500' :
                    'bg-primary'
                  }`}
                  style={{ width: `${uploadProgress}%` }}
                />
              </div>
              {uploadError && (
                <p className="text-sm text-red-600">{uploadError}</p>
              )}
            </div>
          )}

          <Button
            onClick={handleUpload}
            disabled={isUploading}
            size="default"
            className="w-full"
          >
            {isUploading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Uploading...
              </>
            ) : uploadComplete ? (
              <>
                <CheckCircle2 className="mr-2 h-4 w-4" />
                Upload Complete
              </>
            ) : (
              <>
                <Upload className="mr-2 h-4 w-4" />
                Upload {selectedFiles.length} {selectedFiles.length === 1 ? 'File' : 'Files'}
              </>
            )}
          </Button>
        </Card>
      )}
    </div>
  );
}

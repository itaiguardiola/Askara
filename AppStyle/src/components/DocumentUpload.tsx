import { useState, useRef } from 'react';
import { Upload, File, X, Loader2 } from 'lucide-react';
import { Button } from './ui/button';
import { Card } from './ui/card';

interface DocumentUploadProps {
  onUpload: (files: File[]) => Promise<void>;
}

export function DocumentUpload({ onUpload }: DocumentUploadProps) {
  const [isDragging, setIsDragging] = useState(false);
  const [selectedFiles, setSelectedFiles] = useState<File[]>([]);
  const [isUploading, setIsUploading] = useState(false);
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
      try {
        await onUpload(selectedFiles);
        setSelectedFiles([]);
        if (fileInputRef.current) {
          fileInputRef.current.value = '';
        }
      } catch (error) {
        console.error('Upload error in component:', error);
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
          <div className="space-y-2 mb-4">
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
          <Button
            onClick={handleUpload}
            disabled={isUploading}
            size="default"
          >
            {isUploading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Uploading...
              </>
            ) : (
              <>
                Upload {selectedFiles.length} {selectedFiles.length === 1 ? 'File' : 'Files'}
              </>
            )}
          </Button>
        </Card>
      )}
    </div>
  );
}

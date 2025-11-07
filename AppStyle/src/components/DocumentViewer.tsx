import { FileText, Download, X } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from './ui/dialog';
import { Button } from './ui/button';
import { ScrollArea } from './ui/scroll-area';
import { Badge } from './ui/badge';
import type { Document } from './DocumentLibrary';

interface DocumentViewerProps {
  document: Document | null;
  onClose: () => void;
}

export function DocumentViewer({ document, onClose }: DocumentViewerProps) {
  if (!document) return null;

  const formatDate = (date: Date) => {
    return new Intl.DateTimeFormat('en-US', {
      month: 'long',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }).format(date);
  };

  const formatSize = (bytes: number) => {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  };

  // Mock document content - in production, this would be actual file content
  const mockContent = `This is a preview of ${document.name}

In a production environment, this viewer would display:
- Full text content for TXT files
- Rendered PDFs using a PDF viewer library
- Word document content using a document parser
- Markdown rendering for .md files

File Information:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
• File Name: ${document.name}
• File Type: ${document.type.toUpperCase()}
• File Size: ${formatSize(document.size)}
• Uploaded: ${formatDate(document.uploadDate)}

Sample Content:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.

Section 1: Introduction
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.

Section 2: Main Content
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

• Key Point 1: Important information about the document
• Key Point 2: Another crucial detail
• Key Point 3: Additional relevant content
• Key Point 4: Supporting evidence and data

Section 3: Conclusion
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

To integrate real document viewing:
1. For PDFs: Use libraries like 'react-pdf' or 'pdfjs-dist'
2. For Word docs: Use 'mammoth' or similar parsers
3. For text files: Fetch and display raw content
4. For images: Display using standard img tags

The document content would be extracted during upload and stored for both viewing and AI processing.`;

  return (
    <Dialog open={document !== null} onOpenChange={onClose}>
      <DialogContent className="max-w-3xl max-h-[90vh] flex flex-col">
        <DialogHeader>
          <div className="flex items-start justify-between gap-4">
            <div className="flex-1 min-w-0">
              <DialogTitle className="flex items-center gap-2 mb-2">
                <FileText className="h-5 w-5 flex-shrink-0" />
                <span className="truncate">{document.name}</span>
              </DialogTitle>
              <div className="flex flex-wrap gap-2">
                <Badge variant="secondary">{document.type.toUpperCase()}</Badge>
                <Badge variant="outline">{formatSize(document.size)}</Badge>
                <Badge variant="outline">{formatDate(document.uploadDate)}</Badge>
              </div>
            </div>
            <Button
              variant="outline"
              size="icon"
              className="flex-shrink-0"
              title="Download (mock)"
            >
              <Download className="h-4 w-4" />
            </Button>
          </div>
        </DialogHeader>

        <ScrollArea className="flex-1 -mx-6 px-6">
          <div className="py-4">
            <pre className="whitespace-pre-wrap break-words font-mono text-sm">
              {mockContent}
            </pre>
          </div>
        </ScrollArea>
      </DialogContent>
    </Dialog>
  );
}

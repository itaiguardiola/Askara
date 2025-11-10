import { useState } from 'react';
import { Search, FileText, Calendar, Trash2, Eye, MessageCircle } from 'lucide-react';
import { Input } from './ui/input';
import { Card } from './ui/card';
import { Button } from './ui/button';
import { Checkbox } from './ui/checkbox';
import { Badge } from './ui/badge';
import { ScrollArea } from './ui/scroll-area';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from './ui/alert-dialog';

export interface Document {
  id: string;
  name: string;
  size: number;
  uploadDate: Date;
  type: string;
}

interface DocumentLibraryProps {
  documents: Document[];
  selectedDocuments: string[];
  onSelectDocument: (id: string) => void;
  onDeleteDocument: (id: string) => void;
  onViewDocument?: (document: Document) => void;
  onChatDocument?: (document: Document) => void;
}

export function DocumentLibrary({
  documents,
  selectedDocuments,
  onSelectDocument,
  onDeleteDocument,
  onViewDocument,
  onChatDocument,
}: DocumentLibraryProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [deleteId, setDeleteId] = useState<string | null>(null);

  const filteredDocuments = documents.filter((doc) =>
    doc.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const formatDate = (date: Date) => {
    return new Intl.DateTimeFormat('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    }).format(date);
  };

  const formatSize = (bytes: number) => {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  };

  return (
    <div className="flex flex-col h-full">
      <div className="p-4 md:p-6 border-b">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search documents..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>
        {selectedDocuments.length > 0 && (
          <div className="mt-3">
            <Badge variant="secondary">
              {selectedDocuments.length} document{selectedDocuments.length !== 1 ? 's' : ''} selected
            </Badge>
          </div>
        )}
      </div>

      <ScrollArea className="flex-1">
        <div className="p-4 md:p-6 space-y-3">
          {filteredDocuments.length === 0 ? (
            <div className="text-center py-12">
              <FileText className="mx-auto h-12 w-12 text-muted-foreground mb-3" />
              <p className="text-muted-foreground">
                {searchQuery ? 'No documents found' : 'No documents uploaded yet'}
              </p>
            </div>
          ) : (
            filteredDocuments.map((doc) => (
              <Card
                key={doc.id}
                className={`p-4 transition-all hover:shadow-md ${
                  selectedDocuments.includes(doc.id)
                    ? 'ring-2 ring-primary bg-accent/50'
                    : ''
                }`}
              >
                <div className="flex items-start gap-3">
                  <Checkbox
                    checked={selectedDocuments.includes(doc.id)}
                    onCheckedChange={() => onSelectDocument(doc.id)}
                    className="mt-1"
                  />
                  <div
                    className="flex-1 min-w-0 cursor-pointer"
                    onClick={() => onSelectDocument(doc.id)}
                  >
                    <div className="flex items-start justify-between gap-2 mb-2">
                      <div className="flex items-center gap-2 min-w-0 flex-1">
                        <FileText className="h-5 w-5 flex-shrink-0 text-primary" />
                        <h4 className="truncate">{doc.name}</h4>
                      </div>
                      <div className="flex gap-1 flex-shrink-0">
                        {onViewDocument && (
                          <Button
                            size="icon"
                            variant="ghost"
                            onClick={(e) => {
                              e.stopPropagation();
                              onViewDocument(doc);
                            }}
                            className="h-8 w-8"
                            title="View document"
                          >
                            <Eye className="h-4 w-4" />
                          </Button>
                        )}
                        {onChatDocument && (
                          <Button
                            size="icon"
                            variant="ghost"
                            onClick={(e) => {
                              e.stopPropagation();
                              onChatDocument(doc);
                            }}
                            className="h-8 w-8"
                            title="Chat with this document"
                          >
                            <MessageCircle className="h-4 w-4" />
                          </Button>
                        )}
                        <Button
                          size="icon"
                          variant="ghost"
                          onClick={(e) => {
                            e.stopPropagation();
                            setDeleteId(doc.id);
                          }}
                          className="h-8 w-8"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                    <div className="flex items-center gap-4 text-muted-foreground text-sm">
                      <span className="flex items-center gap-1">
                        <Calendar className="h-3 w-3" />
                        {formatDate(doc.uploadDate)}
                      </span>
                      <span>{formatSize(doc.size)}</span>
                      <span className="uppercase">{doc.type}</span>
                    </div>
                  </div>
                </div>
              </Card>
            ))
          )}
        </div>
      </ScrollArea>

      <AlertDialog open={deleteId !== null} onOpenChange={() => setDeleteId(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Document?</AlertDialogTitle>
            <AlertDialogDescription>
              This action cannot be undone. This will permanently delete the document.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                if (deleteId) {
                  onDeleteDocument(deleteId);
                  setDeleteId(null);
                }
              }}
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

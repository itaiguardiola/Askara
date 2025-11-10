import { useState, useEffect } from "react";
import {
  Upload,
  Library,
  MessageSquare,
  Menu,
} from "lucide-react";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "./components/ui/tabs";
import { DocumentUpload } from "./components/DocumentUpload";
import {
  DocumentLibrary,
  type Document,
} from "./components/DocumentLibrary";
import { ChatInterface } from "./components/ChatInterface";
import { SettingsDialog } from "./components/SettingsDialog";
import { DocumentViewer } from "./components/DocumentViewer";
import { DarkModeToggle } from "./components/DarkModeToggle";
import { Button } from "./components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetTrigger,
} from "./components/ui/sheet";
import { toast } from "sonner@2.0.3";
import { Toaster } from "./components/ui/sonner";
import { api } from "./api/client";
import type { Document as ApiDocument } from "./api/types";

export default function App() {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [selectedDocuments, setSelectedDocuments] = useState<
    string[]
  >([]);
  const [activeTab, setActiveTab] = useState("upload");
  const [isMobileMenuOpen, setIsMobileMenuOpen] =
    useState(false);
  const [viewingDocument, setViewingDocument] = useState<Document | null>(null);
  const [isDarkMode, setIsDarkMode] = useState(() => {
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem('darkMode');
      return saved ? JSON.parse(saved) : window.matchMedia('(prefers-color-scheme: dark)').matches;
    }
    return false;
  });

  useEffect(() => {
    if (isDarkMode) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
    localStorage.setItem('darkMode', JSON.stringify(isDarkMode));
  }, [isDarkMode]);

  // Load documents on mount
  useEffect(() => {
    const loadDocuments = async () => {
      try {
        const docs = await api.getDocuments();
        const mappedDocs: Document[] = docs.map((doc: ApiDocument) => ({
          id: doc.id,
          name: doc.filename,
          size: doc.file_size,
          uploadDate: new Date(doc.upload_date),
          type: doc.content_type.split('/')[1] || 'unknown',
        }));
        setDocuments(mappedDocs);
      } catch (error) {
        console.error('Failed to load documents:', error);
        toast.error('Failed to load documents');
      }
    };
    loadDocuments();
  }, []);

  const handleUpload = async (files: File[]) => {
    console.log('[handleUpload] Starting upload for files:', files.map(f => f.name));
    try {
      const result = await api.uploadFiles(files);
      console.log('[handleUpload] API response:', result);

      // Map API documents to local document format
      const newDocuments: Document[] = result.uploaded_documents.map((doc: ApiDocument) => ({
        id: doc.id,
        name: doc.filename,
        size: doc.file_size,
        uploadDate: new Date(doc.upload_date),
        type: doc.content_type.split('/')[1] || 'unknown',
      }));

      console.log('[handleUpload] Mapped documents:', newDocuments);
      setDocuments((prev) => {
        const updated = [...prev, ...newDocuments];
        console.log('[handleUpload] Updated documents state:', updated);
        return updated;
      });

      toast.success(
        `${result.num_files_succeeded} document${result.num_files_succeeded !== 1 ? "s" : ""} uploaded successfully`,
      );

      if (result.num_files_failed > 0) {
        const failedFiles = Object.keys(result.failed_file_names).join(', ');
        toast.error(`Failed to upload: ${failedFiles}`);
      }

      setActiveTab("library");
    } catch (error) {
      console.error('[handleUpload] Error:', error);
      toast.error(`Upload failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
      throw error;
    }
  };

  const handleSelectDocument = (id: string) => {
    setSelectedDocuments((prev) =>
      prev.includes(id)
        ? prev.filter((docId) => docId !== id)
        : [...prev, id],
    );
  };

  const handleDeleteDocument = async (id: string) => {
    try {
      await api.deleteDocument(id);
      setDocuments((prev) => prev.filter((doc) => doc.id !== id));
      setSelectedDocuments((prev) =>
        prev.filter((docId) => docId !== id),
      );
      toast.success("Document deleted successfully");
    } catch (error) {
      toast.error(`Delete failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  };

  const handleChatDocument = (doc: Document) => {
    // Select the document if not already selected
    setSelectedDocuments((prev) =>
      prev.includes(doc.id) ? prev : [...prev, doc.id],
    );
    // Switch to chat tab
    setActiveTab("chat");
  };

  const getSelectedDocs = () =>
    documents.filter((doc) =>
      selectedDocuments.includes(doc.id),
    );

  return (
    <div className="h-screen flex flex-col bg-background">
      <header className="border-b p-2 flex items-center justify-between bg-card">
        <div className="flex items-center gap-2">
          <div className="h-8 w-8 rounded-lg bg-primary flex items-center justify-center">
            <MessageSquare className="h-5 w-5 text-primary-foreground" />
          </div>
          <div>
            <h1 className="text-base leading-none">Askara</h1>
            <p className="text-xs text-muted-foreground">
              Document Q&A Assistant
            </p>
          </div>
        </div>

        <div className="flex items-center gap-1">
          <SettingsDialog />
          <DarkModeToggle isDark={isDarkMode} onToggle={() => setIsDarkMode(!isDarkMode)} />
          <Sheet
            open={isMobileMenuOpen}
            onOpenChange={setIsMobileMenuOpen}
          >
            <SheetTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="md:hidden"
              >
                <Menu className="h-5 w-5" />
              </Button>
            </SheetTrigger>
          <SheetContent
            side="right"
            className="w-[280px] sm:w-[350px] p-0"
          >
            <div className="p-4 border-b">
              <h2>Quick Stats</h2>
            </div>
            <div className="p-4 space-y-4">
              <div className="p-4 rounded-lg bg-accent">
                <p className="text-sm text-muted-foreground">
                  Total Documents
                </p>
                <p className="text-2xl">{documents.length}</p>
              </div>
              <div className="p-4 rounded-lg bg-accent">
                <p className="text-sm text-muted-foreground">
                  Selected
                </p>
                <p className="text-2xl">
                  {selectedDocuments.length}
                </p>
              </div>
              <div className="p-4 rounded-lg bg-accent">
                <p className="text-sm text-muted-foreground">
                  Total Size
                </p>
                <p className="text-2xl">
                  {(
                    documents.reduce(
                      (acc, doc) => acc + doc.size,
                      0,
                    ) / 1024
                  ).toFixed(0)}{" "}
                  KB
                </p>
              </div>
            </div>
          </SheetContent>
          </Sheet>
        </div>
      </header>

      <main className="flex-1 overflow-hidden p-2">
        <Tabs
          value={activeTab}
          onValueChange={setActiveTab}
          className="h-full flex flex-col"
        >
          <TabsList className="w-full grid grid-cols-3 mb-2">
            <TabsTrigger value="upload" className="gap-2">
              <Upload className="h-4 w-4" />
              <span className="hidden sm:inline">Upload</span>
            </TabsTrigger>
            <TabsTrigger value="library" className="gap-2">
              <Library className="h-4 w-4" />
              <span className="hidden sm:inline">Library</span>
              {documents.length > 0 && (
                <span className="ml-1">
                  ({documents.length})
                </span>
              )}
            </TabsTrigger>
            <TabsTrigger value="chat" className="gap-2">
              <MessageSquare className="h-4 w-4" />
              <span className="hidden sm:inline">Chat</span>
              {selectedDocuments.length > 0 && (
                <span className="ml-1">
                  ({selectedDocuments.length})
                </span>
              )}
            </TabsTrigger>
          </TabsList>

          <div className="flex-1 overflow-hidden">
            <TabsContent value="upload" className="h-full m-0">
              <DocumentUpload onUpload={handleUpload} />
            </TabsContent>

            <TabsContent value="library" className="h-full m-0">
              <DocumentLibrary
                documents={documents}
                selectedDocuments={selectedDocuments}
                onSelectDocument={handleSelectDocument}
                onDeleteDocument={handleDeleteDocument}
                onViewDocument={setViewingDocument}
                onChatDocument={handleChatDocument}
              />
            </TabsContent>

            <TabsContent value="chat" className="h-full m-0">
              <ChatInterface
                selectedDocuments={getSelectedDocs()}
              />
            </TabsContent>
          </div>
        </Tabs>
      </main>

      <Toaster />
      <DocumentViewer
        document={viewingDocument}
        onClose={() => setViewingDocument(null)}
      />
    </div>
  );
}
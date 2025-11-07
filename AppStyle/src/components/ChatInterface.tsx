import { useState, useRef, useEffect } from 'react';
import { Send, Bot, User, FileText, Sparkles, Download } from 'lucide-react';
import { Button } from './ui/button';
import { Textarea } from './ui/textarea';
import { ScrollArea } from './ui/scroll-area';
import { Card } from './ui/card';
import { Badge } from './ui/badge';
import { Avatar, AvatarFallback } from './ui/avatar';
import { SuggestedQuestions } from './SuggestedQuestions';
import { toast } from 'sonner@2.0.3';
import type { Document } from './DocumentLibrary';
import { api } from '../api/client';

export interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
}

interface ChatInterfaceProps {
  selectedDocuments: Document[];
}

export function ChatInterface({ selectedDocuments }: ChatInterfaceProps) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim() || selectedDocuments.length === 0 || isLoading) return;

    const userMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: input.trim(),
      timestamp: new Date(),
    };

    setMessages((prev) => [...prev, userMessage]);
    setInput('');
    setIsLoading(true);

    let responseText = '';

    try {
      await api.askQuestionStream(
        userMessage.content,
        // onChunk callback
        (chunk) => {
          responseText += chunk;
          // Update the last message with accumulated response
          setMessages((prev) => {
            const newMessages = [...prev];
            const lastMsg = newMessages[newMessages.length - 1];
            if (lastMsg && lastMsg.role === 'assistant') {
              lastMsg.content = responseText;
            } else {
              newMessages.push({
                id: (Date.now() + 1).toString(),
                role: 'assistant',
                content: responseText,
                timestamp: new Date(),
              });
            }
            return newMessages;
          });
        },
        // onContext callback (optional)
        (contexts) => {
          console.log('Retrieved contexts:', contexts);
        },
        // onError callback
        (error) => {
          console.error('Streaming error:', error);
          toast.error(`Error: ${error}`);
        }
      );
    } catch (error) {
      console.error('Error asking question:', error);
      toast.error('Failed to get response');
    } finally {
      setIsLoading(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e);
    }
  };

  const exportChat = () => {
    if (messages.length === 0) {
      toast.error('No messages to export');
      return;
    }

    const chatContent = messages
      .map((msg) => {
        const role = msg.role === 'user' ? 'You' : 'Assistant';
        const time = msg.timestamp.toLocaleString();
        return `[${time}] ${role}:\n${msg.content}\n`;
      })
      .join('\n---\n\n');

    const header = `Askara Chat Export\nDocuments: ${selectedDocuments.map((d) => d.name).join(', ')}\nExported: ${new Date().toLocaleString()}\n\n${'='.repeat(60)}\n\n`;
    
    const blob = new Blob([header + chatContent], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `askara-chat-${Date.now()}.txt`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    
    toast.success('Chat exported successfully');
  };

  if (selectedDocuments.length === 0) {
    return (
      <div className="flex items-center justify-center h-full p-4">
        <Card className="p-8 md:p-12 text-center max-w-md">
          <FileText className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
          <h3 className="mb-2">No Documents Selected</h3>
          <p className="text-muted-foreground">
            Select one or more documents from the library to start asking questions.
          </p>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full overflow-hidden">
      <div className="p-4 md:p-6 border-b flex-shrink-0">
        <div className="flex items-center justify-between gap-2 mb-2">
          <div className="flex items-center gap-2">
            <Sparkles className="h-5 w-5 text-primary" />
            <h3>Ask Questions</h3>
          </div>
          {messages.length > 0 && (
            <Button
              variant="outline"
              size="sm"
              onClick={exportChat}
              className="gap-2"
            >
              <Download className="h-4 w-4" />
              <span className="hidden sm:inline">Export</span>
            </Button>
          )}
        </div>
        <div className="flex flex-wrap gap-2">
          {selectedDocuments.map((doc) => (
            <Badge key={doc.id} variant="secondary" className="gap-1">
              <FileText className="h-3 w-3" />
              <span className="truncate max-w-[150px]">{doc.name}</span>
            </Badge>
          ))}
        </div>
      </div>

      <ScrollArea className="flex-1 overflow-y-auto" ref={scrollRef}>
        <div className="space-y-4 max-w-3xl mx-auto p-4 md:p-6">
          {messages.length === 0 ? (
            <div>
              <SuggestedQuestions onSelectQuestion={setInput} />
              <div className="text-center py-12">
                <Bot className="mx-auto h-12 w-12 text-muted-foreground mb-3" />
                <p className="text-muted-foreground">
                  Start a conversation by asking a question about your documents
                </p>
              </div>
            </div>
          ) : (
            messages.map((message) => (
              <div
                key={message.id}
                className={`flex gap-3 ${
                  message.role === 'user' ? 'justify-end' : 'justify-start'
                }`}
              >
                {message.role === 'assistant' && (
                  <Avatar className="h-8 w-8 flex-shrink-0">
                    <AvatarFallback className="bg-primary text-primary-foreground">
                      <Bot className="h-4 w-4" />
                    </AvatarFallback>
                  </Avatar>
                )}
                <Card
                  className={`p-3 md:p-4 max-w-[85%] md:max-w-[75%] ${
                    message.role === 'user'
                      ? 'bg-primary text-primary-foreground'
                      : ''
                  }`}
                >
                  <p className="whitespace-pre-wrap break-words">{message.content}</p>
                  <p
                    className={`text-xs mt-2 ${
                      message.role === 'user'
                        ? 'text-primary-foreground/70'
                        : 'text-muted-foreground'
                    }`}
                  >
                    {message.timestamp.toLocaleTimeString([], {
                      hour: '2-digit',
                      minute: '2-digit',
                    })}
                  </p>
                </Card>
                {message.role === 'user' && (
                  <Avatar className="h-8 w-8 flex-shrink-0">
                    <AvatarFallback>
                      <User className="h-4 w-4" />
                    </AvatarFallback>
                  </Avatar>
                )}
              </div>
            ))
          )}
          {isLoading && (
            <div className="flex gap-3">
              <Avatar className="h-8 w-8 flex-shrink-0">
                <AvatarFallback className="bg-primary text-primary-foreground">
                  <Bot className="h-4 w-4" />
                </AvatarFallback>
              </Avatar>
              <Card className="p-4">
                <div className="flex gap-1">
                  <span className="w-2 h-2 bg-muted-foreground rounded-full animate-bounce" />
                  <span className="w-2 h-2 bg-muted-foreground rounded-full animate-bounce [animation-delay:0.2s]" />
                  <span className="w-2 h-2 bg-muted-foreground rounded-full animate-bounce [animation-delay:0.4s]" />
                </div>
              </Card>
            </div>
          )}
        </div>
      </ScrollArea>

      <div className="p-4 md:p-6 border-t flex-shrink-0">
        <form onSubmit={handleSubmit} className="flex gap-2">
          <Textarea
            ref={textareaRef}
            placeholder="Ask a question about your documents..."
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            className="min-h-[60px] max-h-[120px] resize-none flex-1"
            disabled={isLoading}
          />
          <Button
            type="submit"
            size="icon"
            disabled={!input.trim() || isLoading}
            className="h-[60px] w-[60px] flex-shrink-0"
          >
            <Send className="h-5 w-5" />
          </Button>
        </form>
        <p className="text-xs text-muted-foreground mt-2">
          Press Enter to send, Shift+Enter for new line
        </p>
      </div>
    </div>
  );
}

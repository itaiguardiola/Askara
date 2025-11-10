import { Lightbulb } from 'lucide-react';
import { Button } from './ui/button';

interface SuggestedQuestionsProps {
  onSelectQuestion: (question: string) => void;
}

const SUGGESTED_QUESTIONS = [
  'Summarize the main points',
  'What are the key takeaways?',
  'List important information',
  'What topics are covered?',
  'Compare main themes',
  'What are the conclusions?',
];

export function SuggestedQuestions({ onSelectQuestion }: SuggestedQuestionsProps) {
  return (
    <div className="flex items-center gap-2 overflow-x-auto pb-1 scrollbar-hide">
      <Lightbulb className="h-4 w-4 text-muted-foreground flex-shrink-0" />
      {SUGGESTED_QUESTIONS.map((question, index) => (
        <Button
          key={index}
          variant="outline"
          size="sm"
          onClick={() => onSelectQuestion(question)}
          className="text-xs h-auto py-1.5 px-3 rounded-full whitespace-nowrap flex-shrink-0 hover:bg-primary hover:text-primary-foreground transition-colors"
        >
          {question}
        </Button>
      ))}
    </div>
  );
}

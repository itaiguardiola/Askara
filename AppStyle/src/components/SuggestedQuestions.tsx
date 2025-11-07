import { Lightbulb } from 'lucide-react';
import { Button } from './ui/button';
import { Card } from './ui/card';

interface SuggestedQuestionsProps {
  onSelectQuestion: (question: string) => void;
}

const SUGGESTED_QUESTIONS = [
  'Summarize the main points of these documents',
  'What are the key takeaways?',
  'List the most important information',
  'What topics are covered?',
  'Compare the main themes across documents',
  'What are the conclusions or recommendations?',
];

export function SuggestedQuestions({ onSelectQuestion }: SuggestedQuestionsProps) {
  return (
    <Card className="p-4 mb-4">
      <div className="flex items-center gap-2 mb-3">
        <Lightbulb className="h-4 w-4 text-primary" />
        <h4 className="text-sm">Suggested Questions</h4>
      </div>
      <div className="flex flex-wrap gap-2">
        {SUGGESTED_QUESTIONS.map((question, index) => (
          <Button
            key={index}
            variant="outline"
            size="sm"
            onClick={() => onSelectQuestion(question)}
            className="text-xs h-auto py-1.5 px-3"
          >
            {question}
          </Button>
        ))}
      </div>
    </Card>
  );
}

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Trash2 } from 'lucide-react';
import { apiClient } from '@/lib/apiClient';
import type { Todo } from '@/types/todo.types';
import { toast } from 'sonner';
import { cn } from '@/lib/utils';
import { Card } from '@/components/ui/card';
import { Checkbox } from '@/components/ui/checkbox';
import { Button } from '@/components/ui/button';

interface TodoItemProps {
    todo: Todo;
}

export function TodoItem({ todo }: TodoItemProps) {
    const queryClient = useQueryClient();

    const toggleMutation = useMutation({
        mutationFn: async () => {
            const response = await apiClient.put(`/tasks/${todo.id}`, {
                completed: !todo.completed,
            });
            return response.data;
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['todos'] });
        },
        onError: () => {
            toast.error('Failed to update task');
        },
    });

    const deleteMutation = useMutation({
        mutationFn: async () => {
            await apiClient.delete(`/tasks/${todo.id}`);
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['todos'] });
            toast.success('Task deleted');
        },
        onError: () => {
            toast.error('Failed to delete task');
        },
    });

    return (
        <Card className="p-4">
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-3 flex-1">
                    <Checkbox
                        checked={todo.completed}
                        onCheckedChange={() => toggleMutation.mutate()}
                        disabled={toggleMutation.isPending}
                    />
                    <div className="flex flex-col">
                        <span
                            className={cn(
                                "text-lg font-medium transition-all",
                                todo.completed ? "text-muted-foreground line-through" : "text-foreground"
                            )}
                        >
                            {todo.title}
                        </span>
                        {todo.description && (
                            <p className={cn("text-sm text-muted-foreground", todo.completed && "line-through opacity-70")}>
                                {todo.description}
                            </p>
                        )}
                    </div>
                </div>
                <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => deleteMutation.mutate()}
                    disabled={deleteMutation.isPending}
                    className="text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors"
                    aria-label="Delete todo"
                >
                    <Trash2 className="w-5 h-5" />
                </Button>
            </div>
        </Card>
    );
}

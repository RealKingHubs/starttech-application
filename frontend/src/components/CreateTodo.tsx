import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus } from 'lucide-react';
import { apiClient } from '@/lib/apiClient';
import { toast } from 'sonner';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Card } from '@/components/ui/card';

const createTodoSchema = z.object({
    title: z.string().min(1, 'Title is required'),
    description: z.string().optional(),
});

type CreateTodoFormValues = z.infer<typeof createTodoSchema>;

export function CreateTodo() {
    const queryClient = useQueryClient();

    const {
        register,
        handleSubmit,
        reset,
        formState: { errors, isSubmitting },
    } = useForm<CreateTodoFormValues>({
        resolver: zodResolver(createTodoSchema),
    });

    const createMutation = useMutation({
        mutationFn: async (data: CreateTodoFormValues) => {
            const response = await apiClient.post('/tasks', data);
            return response.data;
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['todos'] });
            reset();
            toast.success('Task added');
        },
        onError: () => {
            toast.error('Failed to create task');
        },
    });

    const onSubmit = (data: CreateTodoFormValues) => {
        createMutation.mutate(data);
    };

    return (
        <Card className="mb-8 p-4">
            <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-4">
                <div className="flex gap-2">
                    <div className="flex-1">
                        <Input
                            type="text"
                            placeholder="What needs to be done?"
                            {...register('title')}
                        />
                        {errors.title && (
                            <p className="text-destructive text-xs mt-1">{errors.title.message}</p>
                        )}
                    </div>
                    <Button
                        type="submit"
                        disabled={isSubmitting || createMutation.isPending}
                        className="gap-2 shrink-0"
                    >
                        <Plus className="w-5 h-5" />
                        Add Task
                    </Button>
                </div>
                <Textarea
                    placeholder="Description (optional)"
                    {...register('description')}
                />
            </form>
        </Card>
    );
}

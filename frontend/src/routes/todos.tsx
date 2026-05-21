import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '@/lib/apiClient';
import type { Todo } from '@/types/todo.types';
import { TodoItem } from '@/components/TodoItem';
import { CreateTodo } from '@/components/CreateTodo';
import { Loader2 } from 'lucide-react';
import { useAuth } from '@/hooks/useAuth';
import { useEffect } from 'react';

export const Route = createFileRoute('/todos')({
    component: Todos,
});

function Todos() {
    const { user, isLoading: isAuthLoading } = useAuth();
    const navigate = useNavigate();

    useEffect(() => {
        if (!isAuthLoading && !user) {
            navigate({ to: '/login' });
        }
    }, [user, isAuthLoading, navigate]);

    const { data: todos, isLoading: isTodosLoading, error } = useQuery({
        queryKey: ['todos'],
        queryFn: async () => {
            const response = await apiClient.get('/tasks');
            return response.data as Todo[];
        },
        enabled: !!user,
    });

    if (isAuthLoading || (isTodosLoading && !error)) {
        return (
            <div className="flex h-screen items-center justify-center">
                <Loader2 className="h-8 w-8 animate-spin text-indigo-600" />
            </div>
        );
    }

    if (!user) {
        return null; // Will redirect
    }

    return (
        <div className="min-h-screen bg-gray-50 py-8">
            <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8">
                <div className="mb-8">
                    <h1 className="text-3xl font-bold text-gray-900">My Tasks</h1>
                    <p className="mt-2 text-gray-600">
                        Welcome back, {user.firstName}! Here's what you need to do.
                    </p>
                </div>

                <CreateTodo />

                <div className="space-y-4">
                    {todos?.length === 0 ? (
                        <div className="text-center py-12 bg-white rounded-lg border border-gray-200 border-dashed">
                            <p className="text-gray-500">No todos yet. Add one above!</p>
                        </div>
                    ) : (
                        todos?.map((todo) => (
                            <TodoItem key={todo.id} todo={todo} />
                        ))
                    )}
                </div>
            </div>
        </div>
    );
}

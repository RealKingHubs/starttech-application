export interface Todo {
    id: string;
    title: string;
    description: string;
    completed: boolean;
    userId: string;
    createdAt: string;
    updatedAt: string;
}

export interface CreateTodoDTO {
    title: string;
    description?: string;
}

export interface UpdateTodoDTO {
    title?: string;
    description?: string;
    completed?: boolean;
}

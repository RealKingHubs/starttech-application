export interface User {
    id: string;
    firstName: string;
    lastName: string;
    username: string;
}

export interface LoginUserDTO {
    username: string;
    password: string;
}

export interface UpdateUserDTO {
    firstName?: string;
    lastName?: string;
    username?: string;
}

export interface ChangePasswordDTO {
    oldPassword: string;
    newPassword: string;
}

export interface HealthStatus {
    database: 'ok' | 'down';
    cache: 'ok' | 'down' | 'disabled';
}

export interface AuthContextType {
    user: User | null;
    setUser: React.Dispatch<React.SetStateAction<User | null>>;
    isAuthenticated: boolean;
    logout: () => void;
    isLoading: boolean;
}

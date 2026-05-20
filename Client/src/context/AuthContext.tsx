import React, { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient, clearAuthToken, getAuthToken } from '@/lib/apiClient';
import type { User } from '@/types/auth.types';
import { AuthContext } from '@/hooks/useAuth';


export const AuthProvider = ({ children }: { children: React.ReactNode }) => {
    const [user, setUser] = useState<User | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const queryClient = useQueryClient();

    // Query to fetch the current user
    const { refetch } = useQuery({
        queryKey: ['currentUser'],
        queryFn: async () => {
            try {
                const { data } = await apiClient.get('/users/me');
                setUser(data);
                return data;
            } catch (error) {
                console.error('Failed to fetch current user:', error);
                clearAuthToken();
                setUser(null);
                return null;
            }
        },
        enabled: false,
        retry: false,
    });

    // Mutation to log out the user
    const { mutate: logoutUser } = useMutation({
        mutationFn: () => apiClient.post('/auth/logout'),
        onSuccess: () => {
            clearAuthToken();
            setUser(null);
            queryClient.setQueryData(['currentUser'], null);
            queryClient.clear();
        },
        onError: () => {
            clearAuthToken();
            setUser(null);
            queryClient.setQueryData(['currentUser'], null);
            queryClient.clear();
        },
    });

    const logout = () => logoutUser();

    // On component mount, check if the user is logged in
    useEffect(() => {
        const checkUserStatus = async () => {
            if (!getAuthToken()) {
                setUser(null);
                setIsLoading(false);
                return;
            }

            setIsLoading(true);
            await refetch();
            setIsLoading(false);
        };
        checkUserStatus();
    }, [refetch]);

    return (
        <AuthContext.Provider value={{ user, setUser, isAuthenticated: !!user, logout, isLoading }}>
            {children}
        </AuthContext.Provider>
    );
};

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createFileRoute, Link } from '@tanstack/react-router';
import { apiClient } from '@/lib/apiClient';
import { useAuth } from '@/hooks/useAuth';
import type { UpdateUserDTO } from '@/types/auth.types';
import { toast } from 'sonner';
import { User, KeyRound, Trash2, Save } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import { Alert, AlertDescription } from '@/components/ui/alert';

const updateProfileSchema = z.object({
  firstName: z.string().min(1, 'First name is required'),
  lastName: z.string().min(1, 'Last name is required'),
  username: z.string().min(3, 'Username must be at least 3 characters'),
});

type UpdateProfileFormValues = z.infer<typeof updateProfileSchema>;

export const Route = createFileRoute('/profile')({
  component: ProfilePage,
});

function ProfilePage() {
  const { user, setUser } = useAuth();
  const queryClient = useQueryClient();
  const [isEditing, setIsEditing] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isDirty },
    reset,
  } = useForm<UpdateProfileFormValues>({
    resolver: zodResolver(updateProfileSchema),
    defaultValues: {
      firstName: user?.firstName || '',
      lastName: user?.lastName || '',
      username: user?.username || '',
    },
  });

  const updateProfileMutation = useMutation({
    mutationFn: async (data: UpdateUserDTO) => {
      const response = await apiClient.put('/users/me', data);
      return response.data;
    },
    onSuccess: (data) => {
      setUser(data.user || data);
      queryClient.setQueryData(['currentUser'], data.user || data);
      toast.success('Profile updated successfully');
      setIsEditing(false);
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    onError: (err: any) => {
      const message = err.response?.data?.error || 'Failed to update profile';
      toast.error(message);
    },
  });

  const onSubmit = (data: UpdateProfileFormValues) => {
    updateProfileMutation.mutate(data);
  };

  const handleCancel = () => {
    reset();
    setIsEditing(false);
  };

  if (!user) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <p className="text-muted-foreground">Loading...</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 flex items-center gap-2">
            <User className="h-8 w-8 text-indigo-600" />
            My Profile
          </h1>
          <p className="mt-2 text-muted-foreground">Manage your account settings and preferences</p>
        </div>

        {/* Profile Information Card */}
        <Card className="mb-6">
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Personal Information</CardTitle>
                <CardDescription>Update your profile details</CardDescription>
              </div>
              {!isEditing && (
                <Button
                  variant="outline"
                  onClick={() => setIsEditing(true)}
                >
                  Edit Profile
                </Button>
              )}
            </div>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit(onSubmit)}>
              <div className="space-y-4">
                {/* First Name */}
                <div>
                  <Label htmlFor="firstName">First Name</Label>
                  <Input
                    id="firstName"
                    type="text"
                    disabled={!isEditing}
                    {...register('firstName')}
                  />
                  {errors.firstName && (
                    <p className="text-destructive text-xs mt-1">{errors.firstName.message}</p>
                  )}
                </div>

                {/* Last Name */}
                <div>
                  <Label htmlFor="lastName">Last Name</Label>
                  <Input
                    id="lastName"
                    type="text"
                    disabled={!isEditing}
                    {...register('lastName')}
                  />
                  {errors.lastName && (
                    <p className="text-destructive text-xs mt-1">{errors.lastName.message}</p>
                  )}
                </div>

                {/* Username */}
                <div>
                  <Label htmlFor="username">Username</Label>
                  <Input
                    id="username"
                    type="text"
                    disabled={!isEditing}
                    {...register('username')}
                  />
                  {errors.username && (
                    <p className="text-destructive text-xs mt-1">{errors.username.message}</p>
                  )}
                </div>
              </div>

              {/* Action Buttons */}
              {isEditing && (
                <div className="flex gap-3 mt-6">
                  <Button
                    type="submit"
                    disabled={!isDirty || updateProfileMutation.isPending}
                    className="gap-2"
                  >
                    <Save className="h-4 w-4" />
                    {updateProfileMutation.isPending ? 'Saving...' : 'Save Changes'}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={handleCancel}
                  >
                    Cancel
                  </Button>
                </div>
              )}
            </form>
          </CardContent>
        </Card>

        {/* Quick Actions */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* Change Password */}
          <Link to="/change-password">
            <Card className="hover:border-primary transition-colors cursor-pointer h-full">
              <CardContent className="flex items-center gap-3 p-4">
                <div className="p-2 bg-primary/10 rounded-lg">
                  <KeyRound className="h-6 w-6 text-primary" />
                </div>
                <div>
                  <h3 className="font-medium">Change Password</h3>
                  <p className="text-sm text-muted-foreground">Update your account password</p>
                </div>
              </CardContent>
            </Card>
          </Link>

          {/* Delete Account */}
          <Card
            className="hover:border-destructive transition-colors cursor-pointer h-full"
            onClick={() => setShowDeleteModal(true)}
          >
            <CardContent className="flex items-center gap-3 p-4">
              <div className="p-2 bg-destructive/10 rounded-lg">
                <Trash2 className="h-6 w-6 text-destructive" />
              </div>
              <div>
                <h3 className="font-medium">Delete Account</h3>
                <p className="text-sm text-muted-foreground">Permanently delete your account</p>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Delete Account Dialog */}
        <DeleteAccountDialog
          open={showDeleteModal}
          onOpenChange={setShowDeleteModal}
        />
      </div>
    </div>
  );
}

// Delete Account Dialog Component
function DeleteAccountDialog({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const { logout } = useAuth();
  const [password, setPassword] = useState('');
  const [confirmed, setConfirmed] = useState(false);

  const deleteAccountMutation = useMutation({
    mutationFn: async () => {
      await apiClient.delete('/users/me');
    },
    onSuccess: () => {
      toast.success('Account deleted successfully');
      logout();
      window.location.href = '/login';
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    onError: (err: any) => {
      const message = err.response?.data?.error || 'Failed to delete account';
      toast.error(message);
    },
  });

  const handleDelete = () => {
    if (!confirmed || !password) {
      toast.error('Please confirm and enter your password');
      return;
    }
    deleteAccountMutation.mutate();
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete Account</DialogTitle>
          <DialogDescription>
            This action cannot be undone. This will permanently delete your account and remove all your data.
          </DialogDescription>
        </DialogHeader>

        <Alert variant="destructive">
          <AlertDescription>
            <p className="font-medium mb-1">⚠️ Warning: This action cannot be undone!</p>
            <p className="text-sm">
              Deleting your account will permanently remove all your data, including tasks and profile information.
            </p>
          </AlertDescription>
        </Alert>

        <div className="space-y-4">
          <div className="flex items-start space-x-2">
            <Checkbox
              id="confirm-delete"
              checked={confirmed}
              onCheckedChange={(checked) => setConfirmed(checked as boolean)}
            />
            <Label
              htmlFor="confirm-delete"
              className="text-sm leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
            >
              I understand this action cannot be undone and all my data will be permanently deleted
            </Label>
          </div>

          <div>
            <Label htmlFor="password-confirm">Enter your password to confirm</Label>
            <Input
              id="password-confirm"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Your password"
            />
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={deleteAccountMutation.isPending}
          >
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={handleDelete}
            disabled={!confirmed || !password || deleteAccountMutation.isPending}
          >
            {deleteAccountMutation.isPending ? 'Deleting...' : 'Delete My Account'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

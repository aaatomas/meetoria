import {
  Box, Typography, Card, CardContent, TextField, Button, Stack,
} from '@mui/material';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';
import { useNavigate } from 'react-router-dom';

const orgSchema = z.object({
  name: z.string().min(2),
  slug: z.string().min(2).regex(/^[a-z0-9-]+$/),
  timezone: z.string().min(1),
  email: z.string().email().optional().or(z.literal('')),
});

type OrgForm = z.infer<typeof orgSchema>;

export function SettingsPage() {
  const orgId = localStorage.getItem('organization_id');
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  const { control, handleSubmit } = useForm<OrgForm>({
    resolver: zodResolver(orgSchema),
    defaultValues: { name: '', slug: '', timezone: 'Europe/Vilnius', email: '' },
  });

  const createOrg = useMutation({
    mutationFn: (data: OrgForm) => api.post('/organizations', data),
    onSuccess: (res) => {
      localStorage.setItem('organization_id', res.data.id);
      queryClient.invalidateQueries({ queryKey: ['organizations'] });
      navigate('/dashboard');
    },
  });

  return (
    <Box>
      <Typography variant="h5" fontWeight={700} gutterBottom>Settings</Typography>

      {!orgId && (
        <Card sx={{ maxWidth: 500, mt: 2 }}>
          <CardContent>
            <Typography variant="h6" gutterBottom>Create Your Organization</Typography>
            <Typography variant="body2" color="text.secondary" mb={3}>
              Set up your business to start managing appointments.
            </Typography>
            <form onSubmit={handleSubmit((d) => createOrg.mutate(d))}>
              <Stack spacing={2}>
                <Controller name="name" control={control} render={({ field }) => <TextField {...field} label="Business Name" fullWidth />} />
                <Controller name="slug" control={control} render={({ field }) => <TextField {...field} label="URL Slug" helperText="Lowercase letters, numbers, hyphens" fullWidth />} />
                <Controller name="timezone" control={control} render={({ field }) => <TextField {...field} label="Timezone" fullWidth />} />
                <Controller name="email" control={control} render={({ field }) => <TextField {...field} label="Business Email" fullWidth />} />
                <Button type="submit" variant="contained" disabled={createOrg.isPending}>Create Organization</Button>
              </Stack>
            </form>
          </CardContent>
        </Card>
      )}

      {orgId && (
        <Typography color="text.secondary" mt={2}>
          Organization settings and working hours configuration coming soon.
        </Typography>
      )}
    </Box>
  );
}

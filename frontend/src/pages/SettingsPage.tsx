import { useState } from 'react';
import {
  Box,
  Typography,
  Card,
  CardContent,
  TextField,
  Button,
  Stack,
  Alert,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Chip,
  IconButton,
  Link,
} from '@mui/material';
import { Add, Settings as SettingsIcon } from '@mui/icons-material';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  api,
  getApiErrorMessage,
  type Organization,
  type PaginatedResponse,
} from '../api/client';
import { OrganizationSettingsDialog } from '../components/settings/OrganizationSettingsDialog';

const orgSchema = z.object({
  name: z.string().min(2),
});

type OrgForm = z.infer<typeof orgSchema>;

export function SettingsPage() {
  const queryClient = useQueryClient();
  const activeOrgId = localStorage.getItem('organization_id');
  const [settingsOrg, setSettingsOrg] = useState<Organization | null>(null);
  const [createOpen, setCreateOpen] = useState(false);

  const { data: orgs = [], isLoading } = useQuery({
    queryKey: ['organizations'],
    queryFn: async () => {
      const { data } = await api.get<PaginatedResponse<Organization>>('/organizations', {
        params: { limit: 100 },
      });
      return data.data;
    },
  });

  const { control, handleSubmit, reset } = useForm<OrgForm>({
    resolver: zodResolver(orgSchema),
    defaultValues: { name: '' },
  });

  const createOrg = useMutation({
    mutationFn: (data: OrgForm) => api.post('/organizations', data),
    onSuccess: (res) => {
      queryClient.invalidateQueries({ queryKey: ['organizations'] });
      localStorage.setItem('organization_id', res.data.id);
      reset();
      setCreateOpen(false);
      window.location.reload();
    },
  });

  const openCreateDialog = () => {
    reset();
    setCreateOpen(true);
  };

  const closeCreateDialog = () => {
    reset();
    setCreateOpen(false);
  };

  const switchOrg = (orgId: string) => {
    localStorage.setItem('organization_id', orgId);
    window.location.reload();
  };

  return (
    <Box>
      <Stack direction="row" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h5" fontWeight={700}>
          Organizations
        </Typography>
        {orgs.length > 0 && (
          <Button variant="contained" startIcon={<Add />} onClick={openCreateDialog}>
            Add organization
          </Button>
        )}
      </Stack>

      {isLoading && <Typography color="text.secondary">Loading organizations…</Typography>}

      {!isLoading && orgs.length === 0 && (
        <Card sx={{ maxWidth: 480 }}>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              No organizations yet
            </Typography>
            <Typography variant="body2" color="text.secondary" mb={2}>
              Create your first business to start managing appointments.
            </Typography>
            <Button variant="contained" startIcon={<Add />} onClick={openCreateDialog}>
              Create organization
            </Button>
          </CardContent>
        </Card>
      )}

      {orgs.length > 0 && (
        <TableContainer component={Card}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Slug</TableCell>
                <TableCell>Timezone</TableCell>
                <TableCell>Public booking</TableCell>
                <TableCell align="right">Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {orgs.map((org) => {
                const isActive = org.id === activeOrgId;
                const bookingUrl = `${window.location.origin}/book/${org.slug}`;
                return (
                  <TableRow key={org.id} hover selected={isActive}>
                    <TableCell>
                      <Stack direction="row" spacing={1} alignItems="center">
                        <Typography fontWeight={isActive ? 600 : 400}>{org.name}</Typography>
                        {isActive && <Chip label="Active" size="small" color="primary" />}
                      </Stack>
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2" color="text.secondary">
                        {org.slug}
                      </Typography>
                    </TableCell>
                    <TableCell>{org.timezone}</TableCell>
                    <TableCell>
                      <Link href={bookingUrl} target="_blank" rel="noopener" variant="body2">
                        /book/{org.slug}
                      </Link>
                    </TableCell>
                    <TableCell align="right">
                      <Stack direction="row" spacing={1} justifyContent="flex-end">
                        {!isActive && (
                          <Button size="small" onClick={() => switchOrg(org.id)}>
                            Switch
                          </Button>
                        )}
                        <IconButton
                          size="small"
                          aria-label={`Settings for ${org.name}`}
                          onClick={() => setSettingsOrg(org)}
                        >
                          <SettingsIcon fontSize="small" />
                        </IconButton>
                      </Stack>
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      <OrganizationSettingsDialog
        org={settingsOrg}
        open={!!settingsOrg}
        onClose={() => setSettingsOrg(null)}
      />

      <Dialog open={createOpen} onClose={closeCreateDialog} maxWidth="sm" fullWidth>
        <DialogTitle>Create organization</DialogTitle>
        <form onSubmit={handleSubmit((d) => createOrg.mutate(d))}>
          <DialogContent>
            <Stack spacing={2}>
              <Controller
                name="name"
                control={control}
                render={({ field }) => (
                  <TextField
                    {...field}
                    label="Business Name"
                    fullWidth
                    autoFocus
                    helperText="Other details like URL slug, timezone, and currency can be configured later in settings."
                  />
                )}
              />
              {createOrg.isError && <Alert severity="error">{getApiErrorMessage(createOrg.error)}</Alert>}
            </Stack>
          </DialogContent>
          <DialogActions>
            <Button onClick={closeCreateDialog}>Cancel</Button>
            <Button type="submit" variant="contained" disabled={createOrg.isPending}>
              Create
            </Button>
          </DialogActions>
        </form>
      </Dialog>
    </Box>
  );
}

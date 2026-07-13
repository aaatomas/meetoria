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
import { Add, EditOutlined, Settings as SettingsIcon } from '@mui/icons-material';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  api,
  Branch,
  checkBranchDeletion,
  createBranch,
  deleteBranch,
  getApiErrorMessage,
  listBranches,
  setDefaultBranch,
  setActiveOrganizationId,
  updateBranch,
  type Organization,
  type PaginatedResponse,
} from '../api/client';
import { OrganizationSettingsDialog } from '../components/settings/OrganizationSettingsDialog';
import { ConfirmDeleteDialog } from '../components/common/ConfirmDeleteDialog';
import { EditDialogTitle } from '../components/EditDialogTitle';
import { PhoneField } from '../components/common/PhoneField';
import { formatPhoneDisplay, optionalPhoneField } from '../utils/phoneUtils';

const orgSchema = z.object({
  name: z.string().min(2),
});

const branchSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  address: z.string().optional(),
  city: z.string().optional(),
  country: z.string().optional(),
  phone: optionalPhoneField,
  email: z.string().email().optional().or(z.literal('')),
  is_active: z.boolean(),
});

type OrgForm = z.infer<typeof orgSchema>;
type BranchForm = z.infer<typeof branchSchema>;

export function SettingsPage() {
  const queryClient = useQueryClient();
  const activeOrgId = localStorage.getItem('organization_id');
  const [settingsOrg, setSettingsOrg] = useState<Organization | null>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const [branchDialogOpen, setBranchDialogOpen] = useState(false);
  const [editingBranch, setEditingBranch] = useState<Branch | null>(null);
  const [confirmDeleteOpen, setConfirmDeleteOpen] = useState(false);
  const [branchError, setBranchError] = useState<string | null>(null);

  const { data: orgs = [], isLoading } = useQuery({
    queryKey: ['organizations'],
    queryFn: async () => {
      const { data } = await api.get<PaginatedResponse<Organization>>('/organizations', {
        params: { limit: 100 },
      });
      return data.data;
    },
  });

  const { data: branches = [], isLoading: branchesLoading } = useQuery({
    queryKey: ['branches', activeOrgId],
    queryFn: () => listBranches(activeOrgId!),
    enabled: !!activeOrgId,
  });

  const { data: branchDeletionCheck, isLoading: branchDeletionCheckLoading } = useQuery({
    queryKey: ['branch-deletion-check', activeOrgId, editingBranch?.id],
    queryFn: () => checkBranchDeletion(activeOrgId!, editingBranch!.id),
    enabled: confirmDeleteOpen && !!activeOrgId && !!editingBranch,
  });

  const { control, handleSubmit, reset } = useForm<OrgForm>({
    resolver: zodResolver(orgSchema),
    defaultValues: { name: '' },
  });

  const {
    control: branchControl,
    handleSubmit: handleBranchSubmit,
    reset: resetBranch,
  } = useForm<BranchForm>({
    resolver: zodResolver(branchSchema),
    defaultValues: { name: '', address: '', city: '', country: '', phone: '', email: '', is_active: true },
  });

  const createOrg = useMutation({
    mutationFn: (data: OrgForm) => api.post('/organizations', data),
    onSuccess: (res) => {
      queryClient.invalidateQueries({ queryKey: ['organizations'] });
      setActiveOrganizationId(res.data.id);
      reset();
      setCreateOpen(false);
      window.location.reload();
    },
  });

  const saveBranch = useMutation({
    mutationFn: (data: BranchForm) => {
      const payload = {
        name: data.name,
        address: data.address || undefined,
        city: data.city || undefined,
        country: data.country || undefined,
        phone: data.phone || undefined,
        email: data.email || undefined,
        is_active: data.is_active,
      };
      if (editingBranch) {
        return updateBranch(activeOrgId!, editingBranch.id, payload);
      }
      return createBranch(activeOrgId!, payload);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['branches', activeOrgId] });
      closeBranchDialog();
    },
    onError: (error) => setBranchError(getApiErrorMessage(error)),
  });

  const deleteBranchMutation = useMutation({
    mutationFn: (branchId: string) => deleteBranch(activeOrgId!, branchId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['branches', activeOrgId] });
      queryClient.invalidateQueries({ queryKey: ['branches-by-org'] });
      setConfirmDeleteOpen(false);
      closeBranchDialog();
    },
    onError: (error) => setBranchError(getApiErrorMessage(error)),
  });

  const setDefaultBranchMutation = useMutation({
    mutationFn: (branchId: string) => setDefaultBranch(activeOrgId!, branchId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['branches', activeOrgId] });
      queryClient.invalidateQueries({ queryKey: ['branches-by-org'] });
    },
    onError: (error) => setBranchError(getApiErrorMessage(error)),
  });

  const openCreateDialog = () => {
    reset();
    setCreateOpen(true);
  };

  const closeCreateDialog = () => {
    reset();
    setCreateOpen(false);
  };

  const openBranchDialog = (branch?: Branch) => {
    setBranchError(null);
    setEditingBranch(branch ?? null);
    resetBranch(
      branch
        ? {
            name: branch.name,
            address: branch.address ?? '',
            city: branch.city ?? '',
            country: branch.country ?? '',
            phone: branch.phone ? formatPhoneDisplay(branch.phone) : '',
            email: branch.email ?? '',
            is_active: branch.is_active,
          }
        : { name: '', address: '', city: '', country: '', phone: '', email: '', is_active: true },
    );
    setBranchDialogOpen(true);
  };

  const closeBranchDialog = () => {
    setBranchDialogOpen(false);
    setEditingBranch(null);
    setBranchError(null);
    setConfirmDeleteOpen(false);
    resetBranch({ name: '', address: '', city: '', country: '', phone: '', email: '', is_active: true });
  };

  const branchDeleteMessage = branchDeletionCheckLoading
    ? 'Checking related bookings and employees…'
    : branchDeletionCheck?.can_delete
      ? `Delete ${editingBranch?.name}? This cannot be undone.`
      : branchDeletionCheck?.message ?? 'Cannot delete this location.';

  const switchOrg = (orgId: string) => {
    setActiveOrganizationId(orgId);
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
            Add another business
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

      {activeOrgId && orgs.length > 0 && (
        <Box mt={4}>
          <Stack direction="row" justifyContent="space-between" alignItems="center" mb={2}>
            <Typography variant="h5" fontWeight={700}>
              Locations
            </Typography>
            <Button variant="contained" startIcon={<Add />} onClick={() => openBranchDialog()}>
              Add location
            </Button>
          </Stack>

          {branchError && !branchDialogOpen && (
            <Alert severity="error" sx={{ mb: 2 }} onClose={() => setBranchError(null)}>
              {branchError}
            </Alert>
          )}

          {branchesLoading && <Typography color="text.secondary">Loading locations…</Typography>}

          {!branchesLoading && branches.length === 0 && (
            <Card>
              <CardContent>
                <Typography color="text.secondary">
                  No locations yet. Add your first branch or office location.
                </Typography>
              </CardContent>
            </Card>
          )}

          {branches.length > 0 && (
            <TableContainer component={Card}>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>Name</TableCell>
                    <TableCell>Address</TableCell>
                    <TableCell>City</TableCell>
                    <TableCell>Status</TableCell>
                    <TableCell align="right">Actions</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {branches.map((branch) => (
                    <TableRow key={branch.id} hover>
                      <TableCell>
                        <Stack direction="row" spacing={1} alignItems="center">
                          <Typography fontWeight={branch.is_default ? 600 : 400}>{branch.name}</Typography>
                          {branch.is_default && <Chip label="Default" size="small" />}
                        </Stack>
                      </TableCell>
                      <TableCell>{branch.address || '—'}</TableCell>
                      <TableCell>{branch.city || '—'}</TableCell>
                      <TableCell>
                        <Chip
                          label={branch.is_active ? 'Active' : 'Inactive'}
                          color={branch.is_active ? 'success' : 'default'}
                          size="small"
                        />
                      </TableCell>
                      <TableCell align="right">
                        {!branch.is_default && branch.is_active && (
                          <Button
                            size="small"
                            onClick={() => setDefaultBranchMutation.mutate(branch.id)}
                            disabled={setDefaultBranchMutation.isPending}
                          >
                            Set default
                          </Button>
                        )}
                        <IconButton size="small" aria-label={`Edit ${branch.name}`} onClick={() => openBranchDialog(branch)}>
                          <EditOutlined fontSize="small" />
                        </IconButton>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </Box>
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

      <Dialog open={branchDialogOpen} onClose={closeBranchDialog} maxWidth="sm" fullWidth>
        <form onSubmit={handleBranchSubmit((d) => saveBranch.mutate(d))}>
          <EditDialogTitle
            title={editingBranch ? 'Edit location' : 'Add location'}
            showActive={!!editingBranch}
            activeDisabled={!!editingBranch?.is_default}
            control={branchControl}
          />
          <DialogContent>
            <Stack spacing={2}>
              {branchError && <Alert severity="error">{branchError}</Alert>}
              {editingBranch?.is_default && (
                <Alert severity="info">Default locations cannot be deactivated or deleted.</Alert>
              )}
              <Controller
                name="name"
                control={branchControl}
                render={({ field, fieldState }) => (
                  <TextField {...field} label="Name" fullWidth required error={!!fieldState.error} helperText={fieldState.error?.message} />
                )}
              />
              <Controller
                name="address"
                control={branchControl}
                render={({ field }) => <TextField {...field} label="Address" fullWidth />}
              />
              <Controller
                name="city"
                control={branchControl}
                render={({ field }) => <TextField {...field} label="City" fullWidth />}
              />
              <Controller
                name="country"
                control={branchControl}
                render={({ field }) => <TextField {...field} label="Country" fullWidth />}
              />
              <Controller
                name="phone"
                control={branchControl}
                render={({ field, fieldState }) => (
                  <PhoneField
                    {...field}
                    label="Phone"
                    fullWidth
                    error={!!fieldState.error}
                    helperText={fieldState.error?.message}
                  />
                )}
              />
              <Controller
                name="email"
                control={branchControl}
                render={({ field }) => <TextField {...field} label="Email" fullWidth />}
              />
            </Stack>
          </DialogContent>
          <DialogActions sx={{ justifyContent: 'space-between', px: 3, pb: 2 }}>
            {editingBranch && !editingBranch.is_default ? (
              <Button type="button" color="error" onClick={() => setConfirmDeleteOpen(true)}>
                Delete
              </Button>
            ) : (
              <span />
            )}
            <Box sx={{ display: 'flex', gap: 1 }}>
              <Button type="button" onClick={closeBranchDialog}>Cancel</Button>
              <Button type="submit" variant="contained" disabled={saveBranch.isPending}>
                {saveBranch.isPending ? 'Saving…' : editingBranch ? 'Save' : 'Create'}
              </Button>
            </Box>
          </DialogActions>
        </form>
      </Dialog>

      <ConfirmDeleteDialog
        open={confirmDeleteOpen}
        title="Delete location"
        message={branchDeleteMessage}
        loading={deleteBranchMutation.isPending || branchDeletionCheckLoading}
        confirmDisabled={branchDeletionCheckLoading || !branchDeletionCheck?.can_delete}
        onCancel={() => setConfirmDeleteOpen(false)}
        onConfirm={() => editingBranch && deleteBranchMutation.mutate(editingBranch.id)}
      />
    </Box>
  );
}

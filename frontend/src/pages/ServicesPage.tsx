import {
  Alert,
  Box,
  Button,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  IconButton,
  ListItemIcon,
  Menu,
  MenuItem,
  Paper,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from '@mui/material';
import { Add, EditOutlined, MoreVert } from '@mui/icons-material';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useState } from 'react';
import {
  api,
  checkServiceDeletion,
  deleteService,
  getActiveBranchId,
  getApiErrorMessage,
  listServices,
  Service,
  resolveActiveBranchId,
  updateService,
  type Organization,
} from '../api/client';
import { ConfirmDeleteDialog } from '../components/common/ConfirmDeleteDialog';
import { EditDialogTitle } from '../components/EditDialogTitle';
import { formatPrice } from '../utils/formatCurrency';

const schema = z.object({
  name: z.string().min(1),
  description: z.string().optional(),
  duration_minutes: z.coerce.number().min(5).max(480),
  price: z.coerce.number().min(0),
  is_active: z.boolean(),
});

type FormData = z.infer<typeof schema>;

export function ServicesPage() {
  const orgId = localStorage.getItem('organization_id')!;
  const branchId = getActiveBranchId();
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Service | null>(null);
  const [menuAnchor, setMenuAnchor] = useState<null | HTMLElement>(null);
  const [menuService, setMenuService] = useState<Service | null>(null);
  const [confirmDeleteOpen, setConfirmDeleteOpen] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const queryClient = useQueryClient();

  const { data: org } = useQuery({
    queryKey: ['organization', orgId],
    queryFn: async () => (await api.get<Organization>(`/organizations/${orgId}`)).data,
    enabled: !!orgId,
  });

  const currency = org?.currency?.trim() || 'EUR';

  const { data: services, isLoading } = useQuery({
    queryKey: ['services', orgId, branchId],
    queryFn: () => listServices(orgId, branchId),
    enabled: !!orgId && !!branchId,
  });

  const { data: deletionCheck, isLoading: deletionCheckLoading } = useQuery({
    queryKey: ['service-deletion-check', orgId, editing?.id],
    queryFn: () => checkServiceDeletion(orgId, editing!.id),
    enabled: confirmDeleteOpen && !!editing,
  });

  const { control, handleSubmit, reset } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: { name: '', description: '', duration_minutes: 30, price: 0, is_active: true },
  });

  const closeDialog = () => {
    setOpen(false);
    setEditing(null);
    setConfirmDeleteOpen(false);
    setSubmitError(null);
    reset({ name: '', description: '', duration_minutes: 30, price: 0, is_active: true });
  };

  const openCreateDialog = () => {
    setEditing(null);
    setSubmitError(null);
    reset({ name: '', description: '', duration_minutes: 30, price: 0, is_active: true });
    setOpen(true);
  };

  const openEditDialog = (service: Service) => {
    setEditing(service);
    setSubmitError(null);
    reset({
      name: service.name,
      description: service.description ?? '',
      duration_minutes: service.duration_minutes,
      price: service.price,
      is_active: service.is_active,
    });
    setOpen(true);
  };

  const saveMutation = useMutation({
    mutationFn: async (data: FormData) => {
      if (editing) {
        return updateService(orgId, editing.id, {
          name: data.name,
          description: data.description,
          duration_minutes: data.duration_minutes,
          price: data.price,
          is_active: data.is_active,
        });
      }
      const activeBranchId = branchId ?? (await resolveActiveBranchId(orgId));
      if (!activeBranchId) {
        throw new Error('No location selected. Choose a location in the header.');
      }
      const { data: created } = await api.post<Service>(`/organizations/${orgId}/services`, {
        name: data.name,
        description: data.description,
        duration_minutes: data.duration_minutes,
        price: data.price,
        branch_id: activeBranchId,
      });
      return created;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['services'] });
      closeDialog();
    },
    onError: (error) => setSubmitError(getApiErrorMessage(error)),
  });

  const deleteMutation = useMutation({
    mutationFn: (serviceId: string) => deleteService(orgId, serviceId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['services'] });
      closeDialog();
    },
    onError: (error) => setSubmitError(getApiErrorMessage(error)),
  });

  const openMenu = (event: React.MouseEvent<HTMLElement>, service: Service) => {
    setMenuAnchor(event.currentTarget);
    setMenuService(service);
  };

  const closeMenu = () => {
    setMenuAnchor(null);
    setMenuService(null);
  };

  const deleteMessage = deletionCheckLoading
    ? 'Checking related bookings…'
    : deletionCheck?.can_delete
      ? `Delete "${editing?.name}"? This cannot be undone.`
      : deletionCheck?.message ?? 'Cannot delete this service.';

  if (!branchId) {
    return (
      <Box>
        <Typography variant="h5" fontWeight={700} gutterBottom>Services</Typography>
        <Alert severity="info">Select a location in the header to manage branch services.</Alert>
      </Box>
    );
  }

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" mb={3}>
        <Typography variant="h5" fontWeight={700}>Services</Typography>
        <Button variant="contained" startIcon={<Add />} onClick={openCreateDialog}>Add Service</Button>
      </Box>

      {submitError && !open && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setSubmitError(null)}>
          {submitError}
        </Alert>
      )}

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Name</TableCell>
              <TableCell>Duration</TableCell>
              <TableCell>Price</TableCell>
              <TableCell>Status</TableCell>
              <TableCell width={56} />
            </TableRow>
          </TableHead>
          <TableBody>
            {isLoading ? (
              <TableRow><TableCell colSpan={5}>Loading...</TableCell></TableRow>
            ) : (
              services?.map((s) => (
                <TableRow key={s.id}>
                  <TableCell>{s.name}</TableCell>
                  <TableCell>{s.duration_minutes} min</TableCell>
                  <TableCell>{formatPrice(s.price, currency)}</TableCell>
                  <TableCell>
                    <Chip
                      label={s.is_active ? 'Active' : 'Inactive'}
                      color={s.is_active ? 'success' : 'default'}
                      size="small"
                    />
                  </TableCell>
                  <TableCell align="right">
                    <IconButton size="small" aria-label={`Actions for ${s.name}`} onClick={(e) => openMenu(e, s)}>
                      <MoreVert fontSize="small" />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>

      <Menu anchorEl={menuAnchor} open={Boolean(menuAnchor)} onClose={closeMenu}>
        <MenuItem
          onClick={() => {
            if (menuService) openEditDialog(menuService);
            closeMenu();
          }}
        >
          <ListItemIcon><EditOutlined fontSize="small" /></ListItemIcon>
          Edit
        </MenuItem>
      </Menu>

      <Dialog open={open} onClose={closeDialog} maxWidth="sm" fullWidth>
        <EditDialogTitle title={editing ? 'Edit Service' : 'Add Service'} showActive={!!editing} control={control} />
        <form onSubmit={handleSubmit((d) => saveMutation.mutate(d))}>
          <DialogContent>
            <Stack spacing={2} mt={1}>
              {submitError && <Alert severity="error">{submitError}</Alert>}
              <Controller name="name" control={control} render={({ field }) => <TextField {...field} label="Name" fullWidth required />} />
              <Controller name="description" control={control} render={({ field }) => <TextField {...field} label="Description" multiline rows={2} fullWidth />} />
              <Controller name="duration_minutes" control={control} render={({ field }) => <TextField {...field} label="Duration (minutes)" type="number" fullWidth required />} />
              <Controller name="price" control={control} render={({ field }) => <TextField {...field} label={`Price (${currency})`} type="number" fullWidth required />} />
            </Stack>
          </DialogContent>
          <DialogActions sx={{ justifyContent: 'space-between', px: 3, pb: 2 }}>
            {editing ? (
              <Button color="error" onClick={() => setConfirmDeleteOpen(true)}>
                Delete
              </Button>
            ) : (
              <span />
            )}
            <Box>
              <Button onClick={closeDialog}>Cancel</Button>
              <Button type="submit" variant="contained" disabled={saveMutation.isPending} sx={{ ml: 1 }}>
                {saveMutation.isPending ? 'Saving…' : editing ? 'Save' : 'Create'}
              </Button>
            </Box>
          </DialogActions>
        </form>
      </Dialog>

      <ConfirmDeleteDialog
        open={confirmDeleteOpen}
        title="Delete service"
        message={deleteMessage}
        loading={deleteMutation.isPending || deletionCheckLoading}
        confirmDisabled={deletionCheckLoading || !deletionCheck?.can_delete}
        onCancel={() => setConfirmDeleteOpen(false)}
        onConfirm={() => editing && deleteMutation.mutate(editing.id)}
      />
    </Box>
  );
}

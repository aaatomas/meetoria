import {
  Alert,
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  ListItemIcon,
  Menu,
  MenuItem,
  Paper,
  Snackbar,
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
import { Add, EditOutlined, EmailOutlined, MoreVert, SmsOutlined } from '@mui/icons-material';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useState } from 'react';
import {
  api,
  checkCustomerDeletion,
  Customer,
  deleteCustomer,
  getApiErrorMessage,
  PaginatedResponse,
  sendCustomerEmail,
  sendCustomerSms,
  updateCustomer,
} from '../api/client';
import { ConfirmDeleteDialog } from '../components/common/ConfirmDeleteDialog';

const schema = z.object({
  first_name: z.string().min(1),
  last_name: z.string().min(1),
  email: z.string().email().optional().or(z.literal('')),
  phone: z.string().optional(),
  notes: z.string().optional(),
});

type FormData = z.infer<typeof schema>;

export function CustomersPage() {
  const orgId = localStorage.getItem('organization_id')!;
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Customer | null>(null);
  const [menuAnchor, setMenuAnchor] = useState<null | HTMLElement>(null);
  const [selectedCustomer, setSelectedCustomer] = useState<Customer | null>(null);
  const [confirmDeleteOpen, setConfirmDeleteOpen] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [notificationMessage, setNotificationMessage] = useState<{ severity: 'success' | 'error'; text: string } | null>(null);
  const queryClient = useQueryClient();

  const { data: customers, isLoading } = useQuery({
    queryKey: ['customers', orgId],
    queryFn: async () => (await api.get<PaginatedResponse<Customer>>(`/organizations/${orgId}/customers`)).data.data,
    enabled: !!orgId,
  });

  const { data: deletionCheck, isLoading: deletionCheckLoading } = useQuery({
    queryKey: ['customer-deletion-check', orgId, editing?.id],
    queryFn: () => checkCustomerDeletion(orgId, editing!.id),
    enabled: confirmDeleteOpen && !!editing,
  });

  const { control, handleSubmit, reset } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: { first_name: '', last_name: '', email: '', phone: '', notes: '' },
  });

  const closeDialog = () => {
    setOpen(false);
    setEditing(null);
    setConfirmDeleteOpen(false);
    setSubmitError(null);
    reset({ first_name: '', last_name: '', email: '', phone: '', notes: '' });
  };

  const openCreateDialog = () => {
    setEditing(null);
    setSubmitError(null);
    reset({ first_name: '', last_name: '', email: '', phone: '', notes: '' });
    setOpen(true);
  };

  const openEditDialog = (customer: Customer) => {
    setEditing(customer);
    setSubmitError(null);
    reset({
      first_name: customer.first_name,
      last_name: customer.last_name,
      email: customer.email ?? '',
      phone: customer.phone ?? '',
      notes: customer.notes ?? '',
    });
    setOpen(true);
  };

  const saveMutation = useMutation({
    mutationFn: async (data: FormData) => {
      const payload = {
        first_name: data.first_name,
        last_name: data.last_name,
        email: data.email || undefined,
        phone: data.phone || undefined,
        notes: data.notes || undefined,
      };
      if (editing) {
        return updateCustomer(orgId, editing.id, payload);
      }
      return api.post(`/organizations/${orgId}/customers`, payload);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['customers'] });
      closeDialog();
    },
    onError: (error) => setSubmitError(getApiErrorMessage(error)),
  });

  const deleteMutation = useMutation({
    mutationFn: (customerId: string) => deleteCustomer(orgId, customerId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['customers'] });
      closeDialog();
    },
    onError: (error) => {
      setNotificationMessage({ severity: 'error', text: getApiErrorMessage(error) });
    },
  });

  const sendSmsMutation = useMutation({
    mutationFn: (customerId: string) => sendCustomerSms(orgId, customerId),
    onSuccess: () => {
      setNotificationMessage({ severity: 'success', text: 'SMS queued for delivery.' });
    },
    onError: (error) => {
      setNotificationMessage({ severity: 'error', text: getApiErrorMessage(error) });
    },
  });

  const sendEmailMutation = useMutation({
    mutationFn: (customerId: string) => sendCustomerEmail(orgId, customerId),
    onSuccess: () => {
      setNotificationMessage({ severity: 'success', text: 'Email queued for delivery.' });
    },
    onError: (error) => {
      setNotificationMessage({ severity: 'error', text: getApiErrorMessage(error) });
    },
  });

  const openCustomerMenu = (event: React.MouseEvent<HTMLElement>, customer: Customer) => {
    setMenuAnchor(event.currentTarget);
    setSelectedCustomer(customer);
  };

  const closeCustomerMenu = () => {
    setMenuAnchor(null);
    setSelectedCustomer(null);
  };

  const deleteMessage = deletionCheckLoading
    ? 'Checking related bookings…'
    : deletionCheck?.can_delete
      ? `Delete ${editing?.first_name} ${editing?.last_name}? This cannot be undone.`
      : deletionCheck?.message ?? 'Cannot delete this customer.';

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" mb={3}>
        <Typography variant="h5" fontWeight={700}>Customers</Typography>
        <Button variant="contained" startIcon={<Add />} onClick={openCreateDialog}>Add Customer</Button>
      </Box>

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Name</TableCell>
              <TableCell>Email</TableCell>
              <TableCell>Phone</TableCell>
              <TableCell align="right">Bookings</TableCell>
              <TableCell align="right">Cancellations</TableCell>
              <TableCell width={56} />
            </TableRow>
          </TableHead>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={6}>Loading...</TableCell>
              </TableRow>
            ) : (
              customers?.map((c) => (
                <TableRow key={c.id}>
                  <TableCell>{c.first_name} {c.last_name}</TableCell>
                  <TableCell>{c.email || '—'}</TableCell>
                  <TableCell>{c.phone || '—'}</TableCell>
                  <TableCell align="right">{c.bookings_count ?? 0}</TableCell>
                  <TableCell align="right">{c.cancellations_count ?? 0}</TableCell>
                  <TableCell align="right">
                    <IconButton
                      size="small"
                      aria-label={`Actions for ${c.first_name} ${c.last_name}`}
                      onClick={(event) => openCustomerMenu(event, c)}
                    >
                      <MoreVert fontSize="small" />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>

      <Menu
        anchorEl={menuAnchor}
        open={Boolean(menuAnchor)}
        onClose={closeCustomerMenu}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
        transformOrigin={{ vertical: 'top', horizontal: 'right' }}
      >
        <MenuItem
          onClick={() => {
            if (selectedCustomer) openEditDialog(selectedCustomer);
            closeCustomerMenu();
          }}
        >
          <ListItemIcon><EditOutlined fontSize="small" /></ListItemIcon>
          Edit
        </MenuItem>
        <MenuItem
          disabled={!selectedCustomer?.phone || sendSmsMutation.isPending}
          onClick={() => {
            if (!selectedCustomer) return;
            setNotificationMessage(null);
            sendSmsMutation.mutate(selectedCustomer.id);
            closeCustomerMenu();
          }}
        >
          <ListItemIcon><SmsOutlined fontSize="small" /></ListItemIcon>
          {sendSmsMutation.isPending ? 'Sending SMS...' : 'Send SMS'}
        </MenuItem>
        <MenuItem
          disabled={!selectedCustomer?.email || sendEmailMutation.isPending}
          onClick={() => {
            if (!selectedCustomer) return;
            setNotificationMessage(null);
            sendEmailMutation.mutate(selectedCustomer.id);
            closeCustomerMenu();
          }}
        >
          <ListItemIcon><EmailOutlined fontSize="small" /></ListItemIcon>
          {sendEmailMutation.isPending ? 'Sending email...' : 'Send email'}
        </MenuItem>
      </Menu>

      <Snackbar
        open={notificationMessage !== null}
        autoHideDuration={5000}
        onClose={() => setNotificationMessage(null)}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        {notificationMessage ? (
          <Alert severity={notificationMessage.severity} onClose={() => setNotificationMessage(null)}>
            {notificationMessage.text}
          </Alert>
        ) : undefined}
      </Snackbar>

      <Dialog open={open} onClose={closeDialog} maxWidth="sm" fullWidth>
        <DialogTitle>{editing ? 'Edit Customer' : 'Add Customer'}</DialogTitle>
        <form onSubmit={handleSubmit((d) => saveMutation.mutate(d))}>
          <DialogContent>
            <Stack spacing={2} mt={1}>
              {submitError && <Alert severity="error">{submitError}</Alert>}
              <Controller name="first_name" control={control} render={({ field }) => <TextField {...field} label="First Name" fullWidth required />} />
              <Controller name="last_name" control={control} render={({ field }) => <TextField {...field} label="Last Name" fullWidth required />} />
              <Controller name="email" control={control} render={({ field }) => <TextField {...field} label="Email" type="email" fullWidth />} />
              <Controller name="phone" control={control} render={({ field }) => <TextField {...field} label="Phone (E.164)" placeholder="+37060000000" fullWidth />} />
              <Controller name="notes" control={control} render={({ field }) => <TextField {...field} label="Notes" multiline rows={2} fullWidth />} />
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
        title="Delete customer"
        message={deleteMessage}
        loading={deleteMutation.isPending || deletionCheckLoading}
        confirmDisabled={deletionCheckLoading || !deletionCheck?.can_delete}
        onCancel={() => setConfirmDeleteOpen(false)}
        onConfirm={() => editing && deleteMutation.mutate(editing.id)}
      />
    </Box>
  );
}

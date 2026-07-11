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
import { Add, EmailOutlined, MoreVert, SmsOutlined } from '@mui/icons-material';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useState } from 'react';
import {
  api,
  Customer,
  getApiErrorMessage,
  PaginatedResponse,
  sendCustomerEmail,
  sendCustomerSms,
} from '../api/client';

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
  const [menuAnchor, setMenuAnchor] = useState<null | HTMLElement>(null);
  const [selectedCustomer, setSelectedCustomer] = useState<Customer | null>(null);
  const [notificationMessage, setNotificationMessage] = useState<{ severity: 'success' | 'error'; text: string } | null>(null);
  const queryClient = useQueryClient();

  const { data: customers, isLoading } = useQuery({
    queryKey: ['customers', orgId],
    queryFn: async () => (await api.get<PaginatedResponse<Customer>>(`/organizations/${orgId}/customers`)).data.data,
    enabled: !!orgId,
  });

  const { control, handleSubmit, reset } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: { first_name: '', last_name: '', email: '', phone: '', notes: '' },
  });

  const createMutation = useMutation({
    mutationFn: (data: FormData) => api.post(`/organizations/${orgId}/customers`, data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['customers'] }); setOpen(false); reset(); },
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

  const menuOpen = Boolean(menuAnchor);

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" mb={3}>
        <Typography variant="h5" fontWeight={700}>Customers</Typography>
        <Button variant="contained" startIcon={<Add />} onClick={() => setOpen(true)}>Add Customer</Button>
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
            </TableRow>
          </TableHead>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={5}>Loading...</TableCell>
              </TableRow>
            ) : (
              customers?.map((c) => (
                <TableRow key={c.id}>
                  <TableCell>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                      <Typography component="span" noWrap>
                        {c.first_name} {c.last_name}
                      </Typography>
                      <IconButton
                        size="small"
                        aria-label={`Actions for ${c.first_name} ${c.last_name}`}
                        onClick={(event) => openCustomerMenu(event, c)}
                        sx={{ flexShrink: 0 }}
                      >
                        <MoreVert fontSize="small" />
                      </IconButton>
                    </Box>
                  </TableCell>
                  <TableCell>{c.email || '—'}</TableCell>
                  <TableCell>{c.phone || '—'}</TableCell>
                  <TableCell align="right">{c.bookings_count ?? 0}</TableCell>
                  <TableCell align="right">{c.cancellations_count ?? 0}</TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>

      <Menu
        anchorEl={menuAnchor}
        open={menuOpen}
        onClose={closeCustomerMenu}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'left' }}
        transformOrigin={{ vertical: 'top', horizontal: 'left' }}
      >
        <MenuItem
          disabled={!selectedCustomer?.phone || sendSmsMutation.isPending}
          onClick={() => {
            if (!selectedCustomer) {
              return;
            }
            setNotificationMessage(null);
            sendSmsMutation.mutate(selectedCustomer.id);
            closeCustomerMenu();
          }}
        >
          <ListItemIcon>
            <SmsOutlined fontSize="small" />
          </ListItemIcon>
          {sendSmsMutation.isPending ? 'Sending SMS...' : 'Send SMS'}
        </MenuItem>
        <MenuItem
          disabled={!selectedCustomer?.email || sendEmailMutation.isPending}
          onClick={() => {
            if (!selectedCustomer) {
              return;
            }
            setNotificationMessage(null);
            sendEmailMutation.mutate(selectedCustomer.id);
            closeCustomerMenu();
          }}
        >
          <ListItemIcon>
            <EmailOutlined fontSize="small" />
          </ListItemIcon>
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

      <Dialog open={open} onClose={() => setOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Add Customer</DialogTitle>
        <form onSubmit={handleSubmit((d) => createMutation.mutate(d))}>
          <DialogContent>
            <Stack spacing={2} mt={1}>
              <Controller name="first_name" control={control} render={({ field }) => <TextField {...field} label="First Name" fullWidth />} />
              <Controller name="last_name" control={control} render={({ field }) => <TextField {...field} label="Last Name" fullWidth />} />
              <Controller name="email" control={control} render={({ field }) => <TextField {...field} label="Email" type="email" fullWidth />} />
              <Controller name="phone" control={control} render={({ field }) => <TextField {...field} label="Phone (E.164)" placeholder="+37060000000" fullWidth />} />
              <Controller name="notes" control={control} render={({ field }) => <TextField {...field} label="Notes" multiline rows={2} fullWidth />} />
            </Stack>
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setOpen(false)}>Cancel</Button>
            <Button type="submit" variant="contained">Create</Button>
          </DialogActions>
        </form>
      </Dialog>
    </Box>
  );
}

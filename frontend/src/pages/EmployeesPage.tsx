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
import { Add, EditOutlined, MoreVert, PhotoCamera } from '@mui/icons-material';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useRef, useState } from 'react';
import {
  api,
  checkEmployeeDeletion,
  deleteEmployee,
  Employee,
  getApiErrorMessage,
  PaginatedResponse,
  updateEmployee,
  uploadEmployeeAvatar,
} from '../api/client';
import { ConfirmDeleteDialog } from '../components/common/ConfirmDeleteDialog';
import { EditDialogTitle } from '../components/EditDialogTitle';
import { EmployeeAvatar } from '../components/employees/EmployeeAvatar';

const schema = z.object({
  first_name: z.string().min(1, 'First name is required'),
  last_name: z.string().min(1, 'Last name is required'),
  email: z.string().email().optional().or(z.literal('')),
  phone: z.string().optional(),
  title: z.string().optional(),
  is_active: z.boolean(),
});

type FormData = z.infer<typeof schema>;

function AvatarPicker({
  previewUrl,
  firstName,
  lastName,
  onSelect,
}: {
  previewUrl?: string;
  firstName: string;
  lastName: string;
  onSelect: (file: File) => void;
}) {
  const inputRef = useRef<HTMLInputElement>(null);

  return (
    <Box display="flex" alignItems="center" gap={2}>
      <EmployeeAvatar
        firstName={firstName || 'N'}
        lastName={lastName || 'A'}
        avatarUrl={previewUrl}
        size={72}
      />
      <Box>
        <Button
          variant="outlined"
          startIcon={<PhotoCamera />}
          onClick={() => inputRef.current?.click()}
        >
          Choose Photo
        </Button>
        <Typography variant="caption" display="block" color="text.secondary" mt={0.5}>
          JPEG, PNG, or WebP up to 5MB
        </Typography>
        <input
          ref={inputRef}
          type="file"
          accept="image/jpeg,image/png,image/webp"
          hidden
          onChange={(event) => {
            const file = event.target.files?.[0];
            if (file) onSelect(file);
            event.target.value = '';
          }}
        />
      </Box>
    </Box>
  );
}

export function EmployeesPage() {
  const orgId = localStorage.getItem('organization_id')!;
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Employee | null>(null);
  const [menuAnchor, setMenuAnchor] = useState<null | HTMLElement>(null);
  const [menuEmployee, setMenuEmployee] = useState<Employee | null>(null);
  const [confirmDeleteOpen, setConfirmDeleteOpen] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [avatarFile, setAvatarFile] = useState<File | null>(null);
  const [avatarPreview, setAvatarPreview] = useState<string | undefined>();
  const queryClient = useQueryClient();

  const { data: employees, isLoading } = useQuery({
    queryKey: ['employees', orgId],
    queryFn: async () => (await api.get<PaginatedResponse<Employee>>(`/organizations/${orgId}/employees`)).data.data,
    enabled: !!orgId,
  });

  const { data: deletionCheck, isLoading: deletionCheckLoading } = useQuery({
    queryKey: ['employee-deletion-check', orgId, editing?.id],
    queryFn: () => checkEmployeeDeletion(orgId, editing!.id),
    enabled: confirmDeleteOpen && !!editing,
  });

  const { control, handleSubmit, reset, watch } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: { first_name: '', last_name: '', email: '', phone: '', title: '', is_active: true },
  });

  const firstName = watch('first_name');
  const lastName = watch('last_name');

  const resetDialog = () => {
    reset({ first_name: '', last_name: '', email: '', phone: '', title: '', is_active: true });
    setSubmitError(null);
    setEditing(null);
    setConfirmDeleteOpen(false);
    setAvatarFile(null);
    if (avatarPreview) URL.revokeObjectURL(avatarPreview);
    setAvatarPreview(undefined);
  };

  const openCreateDialog = () => {
    resetDialog();
    setOpen(true);
  };

  const openEditDialog = (employee: Employee) => {
    setEditing(employee);
    setSubmitError(null);
    setAvatarFile(null);
    if (avatarPreview) URL.revokeObjectURL(avatarPreview);
    setAvatarPreview(undefined);
    reset({
      first_name: employee.first_name,
      last_name: employee.last_name,
      email: employee.email ?? '',
      phone: employee.phone ?? '',
      title: employee.title ?? '',
      is_active: employee.is_active,
    });
    setOpen(true);
  };

  const closeDialog = () => {
    setOpen(false);
    resetDialog();
  };

  const handleAvatarSelect = (file: File) => {
    setAvatarFile(file);
    if (avatarPreview) URL.revokeObjectURL(avatarPreview);
    setAvatarPreview(URL.createObjectURL(file));
  };

  const saveMutation = useMutation({
    mutationFn: async (data: FormData) => {
      const payload = {
        first_name: data.first_name,
        last_name: data.last_name,
        email: data.email || undefined,
        phone: data.phone || undefined,
        title: data.title || undefined,
        is_active: data.is_active,
      };

      if (editing) {
        const employee = await updateEmployee(orgId, editing.id, payload);
        if (avatarFile) {
          await uploadEmployeeAvatar(orgId, employee.id, avatarFile);
        }
        return employee;
      }

      const { is_active: _isActive, ...createData } = payload;
      const { data: employee } = await api.post<Employee>(`/organizations/${orgId}/employees`, createData);
      return employee;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['employees'] });
      closeDialog();
    },
    onError: (error) => setSubmitError(getApiErrorMessage(error)),
  });

  const deleteMutation = useMutation({
    mutationFn: (employeeId: string) => deleteEmployee(orgId, employeeId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['employees'] });
      closeDialog();
    },
    onError: (error) => setSubmitError(getApiErrorMessage(error)),
  });

  const onSubmit = handleSubmit(
    (data) => {
      setSubmitError(null);
      saveMutation.mutate(data);
    },
    () => setSubmitError('Please fill in all required fields.'),
  );

  const openMenu = (event: React.MouseEvent<HTMLElement>, employee: Employee) => {
    setMenuAnchor(event.currentTarget);
    setMenuEmployee(employee);
  };

  const closeMenu = () => {
    setMenuAnchor(null);
    setMenuEmployee(null);
  };

  const dialogAvatarUrl = avatarPreview ?? (editing ? editing.avatar_url : undefined);

  const deleteMessage = deletionCheckLoading
    ? 'Checking related bookings…'
    : deletionCheck?.can_delete
      ? `Delete ${editing?.first_name} ${editing?.last_name}? This cannot be undone.`
      : deletionCheck?.message ?? 'Cannot delete this employee.';

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" mb={3}>
        <Typography variant="h5" fontWeight={700}>Employees</Typography>
        <Button variant="contained" startIcon={<Add />} onClick={openCreateDialog}>Add Employee</Button>
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
              <TableCell width={72}>Photo</TableCell>
              <TableCell>Name</TableCell>
              <TableCell>Title</TableCell>
              <TableCell>Status</TableCell>
              <TableCell width={56} />
            </TableRow>
          </TableHead>
          <TableBody>
            {isLoading ? (
              <TableRow><TableCell colSpan={5}>Loading...</TableCell></TableRow>
            ) : employees?.map((employee) => (
              <TableRow key={employee.id}>
                <TableCell>
                  <EmployeeAvatar
                    firstName={employee.first_name}
                    lastName={employee.last_name}
                    avatarUrl={employee.avatar_url}
                    color={employee.color}
                    cacheKey={employee.updated_at}
                  />
                </TableCell>
                <TableCell>{employee.first_name} {employee.last_name}</TableCell>
                <TableCell>{employee.title}</TableCell>
                <TableCell>
                  <Chip
                    label={employee.is_active ? 'Active' : 'Inactive'}
                    color={employee.is_active ? 'success' : 'default'}
                    size="small"
                  />
                </TableCell>
                <TableCell align="right">
                  <IconButton
                    size="small"
                    aria-label={`Actions for ${employee.first_name} ${employee.last_name}`}
                    onClick={(e) => openMenu(e, employee)}
                  >
                    <MoreVert fontSize="small" />
                  </IconButton>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      <Menu anchorEl={menuAnchor} open={Boolean(menuAnchor)} onClose={closeMenu}>
        <MenuItem
          onClick={() => {
            if (menuEmployee) openEditDialog(menuEmployee);
            closeMenu();
          }}
        >
          <ListItemIcon><EditOutlined fontSize="small" /></ListItemIcon>
          Edit
        </MenuItem>
      </Menu>

      <Dialog open={open} onClose={closeDialog} maxWidth="sm" fullWidth>
        <EditDialogTitle title={editing ? 'Edit Employee' : 'Add Employee'} showActive={!!editing} control={control} />
        <Box component="form" onSubmit={onSubmit}>
          <DialogContent>
            <Stack spacing={2} mt={1}>
              {submitError && <Alert severity="error">{submitError}</Alert>}
              {editing && (
                <AvatarPicker
                  previewUrl={dialogAvatarUrl}
                  firstName={firstName}
                  lastName={lastName}
                  onSelect={handleAvatarSelect}
                />
              )}
              <Controller name="first_name" control={control} render={({ field }) => (
                <TextField {...field} label="First Name" fullWidth required />
              )} />
              <Controller name="last_name" control={control} render={({ field }) => (
                <TextField {...field} label="Last Name" fullWidth required />
              )} />
              <Controller name="title" control={control} render={({ field }) => (
                <TextField {...field} label="Title" fullWidth />
              )} />
              <Controller name="email" control={control} render={({ field }) => (
                <TextField {...field} label="Email" fullWidth />
              )} />
              <Controller name="phone" control={control} render={({ field }) => (
                <TextField {...field} label="Phone" fullWidth />
              )} />
            </Stack>
          </DialogContent>
          <DialogActions sx={{ justifyContent: 'space-between', px: 3, pb: 2 }}>
            {editing ? (
              <Button type="button" color="error" onClick={() => setConfirmDeleteOpen(true)}>
                Delete
              </Button>
            ) : (
              <span />
            )}
            <Box>
              <Button type="button" onClick={closeDialog}>Cancel</Button>
              <Button type="submit" variant="contained" disabled={saveMutation.isPending} sx={{ ml: 1 }}>
                {saveMutation.isPending ? 'Saving…' : editing ? 'Save' : 'Create'}
              </Button>
            </Box>
          </DialogActions>
        </Box>
      </Dialog>

      <ConfirmDeleteDialog
        open={confirmDeleteOpen}
        title="Delete employee"
        message={deleteMessage}
        loading={deleteMutation.isPending || deletionCheckLoading}
        confirmDisabled={deletionCheckLoading || !deletionCheck?.can_delete}
        onCancel={() => setConfirmDeleteOpen(false)}
        onConfirm={() => editing && deleteMutation.mutate(editing.id)}
      />
    </Box>
  );
}

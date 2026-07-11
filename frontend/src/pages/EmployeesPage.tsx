import {
  Box,
  Typography,
  Button,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Stack,
  Chip,
  IconButton,
  Alert,
} from '@mui/material';
import { Add, PhotoCamera } from '@mui/icons-material';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useRef, useState } from 'react';
import {
  api,
  Employee,
  PaginatedResponse,
  uploadEmployeeAvatar,
  getApiErrorMessage,
} from '../api/client';
import { EmployeeAvatar } from '../components/employees/EmployeeAvatar';

const schema = z.object({
  first_name: z.string().min(1, 'First name is required'),
  last_name: z.string().min(1, 'Last name is required'),
  email: z.string().email().optional().or(z.literal('')),
  phone: z.string().optional(),
  title: z.string().optional(),
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
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [avatarFile, setAvatarFile] = useState<File | null>(null);
  const [avatarPreview, setAvatarPreview] = useState<string | undefined>();
  const [uploadingEmployeeId, setUploadingEmployeeId] = useState<string | null>(null);
  const rowInputRef = useRef<HTMLInputElement>(null);
  const queryClient = useQueryClient();

  const { data: employees, isLoading } = useQuery({
    queryKey: ['employees', orgId],
    queryFn: async () => (await api.get<PaginatedResponse<Employee>>(`/organizations/${orgId}/employees`)).data.data,
    enabled: !!orgId,
  });

  const { control, handleSubmit, reset, watch } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: { first_name: '', last_name: '', email: '', phone: '', title: '' },
  });

  const firstName = watch('first_name');
  const lastName = watch('last_name');

  const resetDialog = () => {
    reset();
    setSubmitError(null);
    setAvatarFile(null);
    if (avatarPreview) URL.revokeObjectURL(avatarPreview);
    setAvatarPreview(undefined);
  };

  const openDialog = () => {
    resetDialog();
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

  const createMutation = useMutation({
    mutationFn: async (data: FormData) => {
      const { data: employee } = await api.post<Employee>(`/organizations/${orgId}/employees`, data);
      if (avatarFile) {
        await uploadEmployeeAvatar(orgId, employee.id, avatarFile);
      }
      return employee;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['employees'] });
      closeDialog();
    },
    onError: (error) => setSubmitError(getApiErrorMessage(error)),
  });

  const avatarMutation = useMutation({
    mutationFn: ({ employeeId, file }: { employeeId: string; file: File }) =>
      uploadEmployeeAvatar(orgId, employeeId, file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['employees'] });
      setUploadingEmployeeId(null);
    },
    onError: (error) => {
      setSubmitError(getApiErrorMessage(error));
      setUploadingEmployeeId(null);
    },
  });

  const onSubmit = handleSubmit(
    (data) => {
      setSubmitError(null);
      createMutation.mutate(data);
    },
    () => setSubmitError('Please fill in all required fields.'),
  );

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" mb={3}>
        <Typography variant="h5" fontWeight={700}>Employees</Typography>
        <Button variant="contained" startIcon={<Add />} onClick={openDialog}>Add Employee</Button>
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
            </TableRow>
          </TableHead>
          <TableBody>
            {isLoading ? (
              <TableRow><TableCell colSpan={4}>Loading...</TableCell></TableRow>
            ) : employees?.map((employee) => (
              <TableRow key={employee.id}>
                <TableCell>
                  <Box position="relative" display="inline-flex">
                    <EmployeeAvatar
                      firstName={employee.first_name}
                      lastName={employee.last_name}
                      avatarUrl={employee.avatar_url}
                      color={employee.color}
                      cacheKey={employee.updated_at}
                    />
                    <IconButton
                      size="small"
                      aria-label="Change photo"
                      disabled={avatarMutation.isPending && uploadingEmployeeId === employee.id}
                      onClick={() => {
                        setSubmitError(null);
                        setUploadingEmployeeId(employee.id);
                        rowInputRef.current?.click();
                      }}
                      sx={{
                        position: 'absolute',
                        right: -8,
                        bottom: -8,
                        bgcolor: 'background.paper',
                        boxShadow: 1,
                        '&:hover': { bgcolor: 'background.paper' },
                      }}
                    >
                      <PhotoCamera fontSize="small" />
                    </IconButton>
                  </Box>
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
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      <input
        ref={rowInputRef}
        type="file"
        accept="image/jpeg,image/png,image/webp"
        hidden
        onChange={(event) => {
          const file = event.target.files?.[0];
          if (file && uploadingEmployeeId) {
            avatarMutation.mutate({ employeeId: uploadingEmployeeId, file });
          }
          event.target.value = '';
        }}
      />

      <Dialog open={open} onClose={closeDialog} maxWidth="sm" fullWidth>
        <DialogTitle>Add Employee</DialogTitle>
        <Box component="form" onSubmit={onSubmit}>
          <DialogContent>
            <Stack spacing={2} mt={1}>
              {submitError && <Alert severity="error">{submitError}</Alert>}
              <AvatarPicker
                previewUrl={avatarPreview}
                firstName={firstName}
                lastName={lastName}
                onSelect={handleAvatarSelect}
              />
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
          <DialogActions>
            <Button type="button" onClick={closeDialog}>Cancel</Button>
            <Button type="submit" variant="contained" disabled={createMutation.isPending}>
              {createMutation.isPending ? 'Creating...' : 'Create'}
            </Button>
          </DialogActions>
        </Box>
      </Dialog>
    </Box>
  );
}

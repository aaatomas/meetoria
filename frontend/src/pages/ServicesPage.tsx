import {
  Box, Typography, Button, Table, TableBody, TableCell, TableContainer,
  TableHead, TableRow, Paper, Dialog, DialogTitle, DialogContent, DialogActions,
  TextField, Stack,
} from '@mui/material';
import { Add } from '@mui/icons-material';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useState } from 'react';
import { api, Service, PaginatedResponse } from '../api/client';
import { formatPrice } from '../utils/formatCurrency';

const schema = z.object({
  name: z.string().min(1),
  description: z.string().optional(),
  duration_minutes: z.coerce.number().min(5).max(480),
  price: z.coerce.number().min(0),
  currency: z.string().length(3),
  category: z.string().optional(),
});

type FormData = z.infer<typeof schema>;

export function ServicesPage() {
  const orgId = localStorage.getItem('organization_id')!;
  const [open, setOpen] = useState(false);
  const queryClient = useQueryClient();

  const { data: services, isLoading } = useQuery({
    queryKey: ['services', orgId],
    queryFn: async () => (await api.get<PaginatedResponse<Service>>(`/organizations/${orgId}/services`)).data.data,
    enabled: !!orgId,
  });

  const { control, handleSubmit, reset } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: { name: '', description: '', duration_minutes: 30, price: 0, currency: 'EUR', category: '' },
  });

  const createMutation = useMutation({
    mutationFn: (data: FormData) => api.post(`/organizations/${orgId}/services`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['services'] });
      setOpen(false);
      reset();
    },
  });

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" mb={3}>
        <Typography variant="h5" fontWeight={700}>Services</Typography>
        <Button variant="contained" startIcon={<Add />} onClick={() => setOpen(true)}>Add Service</Button>
      </Box>
      <TableContainer component={Paper}>
        <Table>
          <TableHead><TableRow><TableCell>Name</TableCell><TableCell>Duration</TableCell><TableCell>Price</TableCell><TableCell>Category</TableCell></TableRow></TableHead>
          <TableBody>
            {isLoading ? <TableRow><TableCell colSpan={4}>Loading...</TableCell></TableRow> :
              services?.map((s) => (
                <TableRow key={s.id}>
                  <TableCell>{s.name}</TableCell>
                  <TableCell>{s.duration_minutes} min</TableCell>
                  <TableCell>{formatPrice(s.price, s.currency)}</TableCell>
                  <TableCell>{s.category}</TableCell>
                </TableRow>
              ))}
          </TableBody>
        </Table>
      </TableContainer>
      <Dialog open={open} onClose={() => setOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Add Service</DialogTitle>
        <form onSubmit={handleSubmit((d) => createMutation.mutate(d))}>
          <DialogContent><Stack spacing={2} mt={1}>
            <Controller name="name" control={control} render={({ field }) => <TextField {...field} label="Name" fullWidth />} />
            <Controller name="description" control={control} render={({ field }) => <TextField {...field} label="Description" multiline rows={2} fullWidth />} />
            <Controller name="duration_minutes" control={control} render={({ field }) => <TextField {...field} label="Duration (minutes)" type="number" fullWidth />} />
            <Controller name="price" control={control} render={({ field }) => <TextField {...field} label="Price" type="number" fullWidth />} />
            <Controller name="currency" control={control} render={({ field }) => <TextField {...field} label="Currency" fullWidth />} />
            <Controller name="category" control={control} render={({ field }) => <TextField {...field} label="Category" fullWidth />} />
          </Stack></DialogContent>
          <DialogActions><Button onClick={() => setOpen(false)}>Cancel</Button><Button type="submit" variant="contained">Create</Button></DialogActions>
        </form>
      </Dialog>
    </Box>
  );
}

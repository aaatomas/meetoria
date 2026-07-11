import { Grid, Typography, Card, CardContent, Box, Skeleton, Alert } from '@mui/material';
import { useQuery } from '@tanstack/react-query';
import { api, DashboardStats, getApiErrorMessage } from '../api/client';
import { formatPrice } from '../utils/formatCurrency';
import dayjs from 'dayjs';
import { BusiestHoursHeatmap } from '../components/dashboard/BusiestHoursHeatmap';
import { PopularServicesList } from '../components/dashboard/PopularServicesList';

function StatCard({ title, value, subtitle }: { title: string; value: string | number; subtitle?: string }) {
  return (
    <Card>
      <CardContent>
        <Typography variant="body2" color="text.secondary" gutterBottom>{title}</Typography>
        <Typography variant="h4" fontWeight={700}>{value}</Typography>
        {subtitle && <Typography variant="caption" color="text.secondary">{subtitle}</Typography>}
      </CardContent>
    </Card>
  );
}

export function DashboardPage() {
  const orgId = localStorage.getItem('organization_id');
  const from = dayjs().startOf('month').format('YYYY-MM-DD');
  const to = dayjs().endOf('month').format('YYYY-MM-DD');

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ['dashboard', orgId, from, to],
    queryFn: async () => {
      const { data } = await api.get<DashboardStats>(`/organizations/${orgId}/analytics/dashboard`, {
        params: { from, to },
      });
      return data;
    },
    enabled: !!orgId,
  });

  if (!orgId) {
    return (
      <Box>
        <Typography variant="h5" gutterBottom>Welcome to Meetoria</Typography>
        <Typography color="text.secondary">Create an organization to get started.</Typography>
      </Box>
    );
  }

  if (isLoading) {
    return (
      <Grid container spacing={3}>
        {[1, 2, 3, 4].map((i) => (
          <Grid size={{ xs: 12, sm: 6, md: 3 }} key={i}><Skeleton variant="rounded" height={120} /></Grid>
        ))}
        <Grid size={{ xs: 12, md: 6 }}><Skeleton variant="rounded" height={280} /></Grid>
        <Grid size={{ xs: 12, md: 6 }}><Skeleton variant="rounded" height={280} /></Grid>
      </Grid>
    );
  }

  if (isError) {
    return (
      <Box>
        <Typography variant="h5" fontWeight={700} gutterBottom>Dashboard</Typography>
        <Alert severity="error">{getApiErrorMessage(error)}</Alert>
      </Box>
    );
  }

  return (
    <Box>
      <Typography variant="h5" fontWeight={700} gutterBottom>Dashboard</Typography>
      <Typography variant="body2" color="text.secondary" mb={3}>
        {dayjs(from).format('MMMM YYYY')} overview
      </Typography>

      <Grid container spacing={3}>
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <StatCard title="Total Bookings" value={data?.total_bookings ?? 0} />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <StatCard title="Completed" value={data?.completed_bookings ?? 0} />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <StatCard title="Revenue" value={formatPrice(data?.revenue ?? 0, 'EUR')} />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <StatCard title="New Customers" value={data?.new_customers ?? 0} />
        </Grid>

        <Grid size={{ xs: 12, md: 6 }}>
          <Card variant="outlined" sx={{ height: '100%' }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>Popular Services</Typography>
              <PopularServicesList services={data?.popular_services ?? []} />
            </CardContent>
          </Card>
        </Grid>

        <Grid size={{ xs: 12, md: 6 }}>
          <Card variant="outlined" sx={{ height: '100%' }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>Busiest Hours</Typography>
              <Typography variant="caption" color="text.secondary" display="block" mb={1.5}>
                Bookings by day and 2-hour slot
              </Typography>
              <BusiestHoursHeatmap heatmapData={data?.hourly_heatmap ?? []} />
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
}

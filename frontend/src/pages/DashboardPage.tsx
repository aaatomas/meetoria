import { Grid, Typography, Card, CardContent, Box, Skeleton, Alert, Stack } from '@mui/material';
import { useQuery } from '@tanstack/react-query';
import { useState, type ReactNode } from 'react';
import {
  api,
  DashboardStats,
  getActiveBranchId,
  getApiErrorMessage,
  listBranches,
  type DashboardScope,
  type MetricTrend,
  type Organization,
} from '../api/client';
import { formatPrice } from '../utils/formatCurrency';
import dayjs from 'dayjs';
import { BusiestDaysChart } from '../components/dashboard/BusiestDaysChart';
import { BusiestHoursHeatmap } from '../components/dashboard/BusiestHoursHeatmap';
import { PopularServicesList } from '../components/dashboard/PopularServicesList';
import { StatTrend } from '../components/dashboard/StatTrend';
import { DashboardScopeBadge } from '../components/dashboard/DashboardScopeBadge';

const DASHBOARD_SCOPE_KEY = 'dashboard_scope';

function readDashboardScope(): DashboardScope {
  const saved = localStorage.getItem(DASHBOARD_SCOPE_KEY);
  return saved === 'organization' ? 'organization' : 'branch';
}

function StatCard({
  title,
  value,
  trend,
}: {
  title: string;
  value: string | number;
  trend?: MetricTrend;
}) {
  return (
    <Card>
      <CardContent>
        <Stack direction="row" justifyContent="space-between" alignItems="flex-start" spacing={1} mb={1}>
          <Typography variant="body2" color="text.secondary">{title}</Typography>
          <StatTrend trend={trend} />
        </Stack>
        <Typography variant="h4" fontWeight={700}>{value}</Typography>
      </CardContent>
    </Card>
  );
}

function buildScopeBadgeLabel(scope: DashboardScope, branchName?: string) {
  if (scope === 'branch') {
    return branchName ?? 'Branch';
  }
  return 'Organization';
}

const DASHBOARD_CHART_HEIGHT = 300;

const dashboardChartCardSx = {
  height: '100%',
  display: 'flex',
  flexDirection: 'column',
} as const;

const dashboardChartContentSx = {
  flex: 1,
  display: 'flex',
  flexDirection: 'column',
  pb: 2,
} as const;

const dashboardChartBodySx = {
  flex: 1,
  minHeight: DASHBOARD_CHART_HEIGHT,
  display: 'flex',
  flexDirection: 'column',
} as const;

function DashboardChartCard({
  title,
  subtitle,
  children,
}: {
  title: string;
  subtitle?: string;
  children: ReactNode;
}) {
  return (
    <Card variant="outlined" sx={{ ...dashboardChartCardSx, width: '100%' }}>
      <CardContent sx={dashboardChartContentSx}>
        <Typography variant="h6" gutterBottom>{title}</Typography>
        {subtitle && (
          <Typography variant="caption" color="text.secondary" display="block" mb={1}>
            {subtitle}
          </Typography>
        )}
        <Box sx={dashboardChartBodySx}>{children}</Box>
      </CardContent>
    </Card>
  );
}

export function DashboardPage() {
  const orgId = localStorage.getItem('organization_id');
  const branchId = getActiveBranchId();
  const [scope, setScope] = useState<DashboardScope>(readDashboardScope);
  const from = dayjs().startOf('month').format('YYYY-MM-DD');
  const to = dayjs().endOf('month').format('YYYY-MM-DD');

  const { data: org } = useQuery({
    queryKey: ['organization', orgId],
    queryFn: async () => (await api.get<Organization>(`/organizations/${orgId}`)).data,
    enabled: !!orgId,
  });

  const { data: branches = [] } = useQuery({
    queryKey: ['branches', orgId],
    queryFn: () => listBranches(orgId!),
    enabled: !!orgId,
  });

  const activeBranch = branches.find((branch) => branch.id === branchId);
  const currency = org?.currency?.trim() || 'EUR';
  const effectiveScope: DashboardScope = scope === 'branch' && !branchId ? 'organization' : scope;

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ['dashboard', orgId, effectiveScope, branchId, from, to],
    queryFn: async () => {
      const { data } = await api.get<DashboardStats>(`/organizations/${orgId}/analytics/dashboard`, {
        params: {
          from,
          to,
          scope: effectiveScope,
          ...(effectiveScope === 'branch' && branchId ? { branch_id: branchId } : {}),
        },
      });
      return data;
    },
    enabled: !!orgId,
  });

  const toggleScope = () => {
    const nextScope: DashboardScope = effectiveScope === 'organization' ? 'branch' : 'organization';
    if (nextScope === 'branch' && !branchId) {
      return;
    }
    setScope(nextScope);
    localStorage.setItem(DASHBOARD_SCOPE_KEY, nextScope);
  };

  if (!orgId) {
    return (
      <Box>
        <Typography variant="h5" gutterBottom>Welcome to Meetoria</Typography>
        <Typography color="text.secondary">Create an organization to get started.</Typography>
      </Box>
    );
  }

  if (effectiveScope === 'branch' && !branchId) {
    return (
      <Box>
        <Typography variant="h5" fontWeight={700} gutterBottom>Dashboard</Typography>
        <Alert severity="info">Select a location in the header to view branch-level stats.</Alert>
      </Box>
    );
  }

  if (isLoading) {
    return (
      <Grid container spacing={3}>
        {[1, 2, 3, 4].map((i) => (
          <Grid size={{ xs: 12, sm: 6, md: 3 }} key={i}><Skeleton variant="rounded" height={120} /></Grid>
        ))}
        {[1, 2, 3].map((i) => (
          <Grid size={{ xs: 12, md: 4 }} key={i}>
            <Skeleton variant="rounded" height={DASHBOARD_CHART_HEIGHT + 88} />
          </Grid>
        ))}
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

  const branchName = data?.branch_name ?? activeBranch?.name;
  const scopeBadgeLabel = buildScopeBadgeLabel(effectiveScope, branchName);

  return (
    <Box>
      <Stack direction="row" alignItems="center" spacing={1} mb={0.5} useFlexGap flexWrap="wrap">
        <Typography variant="h5" fontWeight={700}>Dashboard</Typography>
        <DashboardScopeBadge
          scope={effectiveScope}
          label={scopeBadgeLabel}
          disabled={effectiveScope === 'organization' && !branchId}
          onClick={toggleScope}
        />
      </Stack>
      <Typography variant="body2" color="text.secondary" mb={3}>
        {dayjs(from).format('MMMM YYYY')} overview{org?.name ? ` · ${org.name}` : ''}
      </Typography>

      <Grid container spacing={3}>
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <StatCard title="Total Bookings" value={data?.total_bookings ?? 0} trend={data?.trends?.total_bookings} />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <StatCard title="Completed" value={data?.completed_bookings ?? 0} trend={data?.trends?.completed_bookings} />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <StatCard title="Revenue" value={formatPrice(data?.revenue ?? 0, currency)} trend={data?.trends?.revenue} />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <StatCard
            title="New Customers"
            value={data?.new_customers ?? 0}
            trend={data?.trends?.new_customers}
          />
        </Grid>

        <Grid size={{ xs: 12, md: 4 }} sx={{ display: 'flex' }}>
          <DashboardChartCard title="Popular Services">
            <Box sx={{ flex: 1, overflow: 'auto', pr: 0.5 }}>
              <PopularServicesList
                services={data?.popular_services ?? []}
                currency={currency}
                showBranch={effectiveScope === 'organization'}
                compact
              />
            </Box>
          </DashboardChartCard>
        </Grid>

        <Grid size={{ xs: 12, md: 4 }} sx={{ display: 'flex' }}>
          <DashboardChartCard title="Busiest Days" subtitle="Bookings by day of week">
            <BusiestDaysChart days={data?.busiest_days ?? []} compact />
          </DashboardChartCard>
        </Grid>

        <Grid size={{ xs: 12, md: 4 }} sx={{ display: 'flex' }}>
          <DashboardChartCard title="Busiest Hours" subtitle="Bookings by day and 2-hour slot">
            <BusiestHoursHeatmap heatmapData={data?.hourly_heatmap ?? []} compact />
          </DashboardChartCard>
        </Grid>
      </Grid>
    </Box>
  );
}

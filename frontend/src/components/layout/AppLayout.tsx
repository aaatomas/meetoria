import {
  AppBar,
  Box,
  Drawer,
  IconButton,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Toolbar,
  Typography,
  Avatar,
  Menu,
  MenuItem,
  Divider,
} from '@mui/material';
import {
  Dashboard,
  CalendarMonth,
  People,
  Badge,
  Spa,
  Settings,
  Menu as MenuIcon,
  Logout,
} from '@mui/icons-material';
import { useEffect, useMemo, useState } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { useAuth } from '../../auth/AuthProvider';
import {
  api,
  Branch,
  listBranches,
  locationKey,
  Organization,
  parseLocationKey,
  pickActiveBranch,
  setActiveBranchId,
  setActiveLocation,
} from '../../api/client';
import { LocationSelect } from './LocationSelect';

const DRAWER_WIDTH = 260;
const CONTENT_GUTTER = 1.5;

const navItems = [
  { label: 'Dashboard', path: '/dashboard', icon: <Dashboard /> },
  { label: 'Bookings', path: '/bookings', icon: <CalendarMonth /> },
  { label: 'Customers', path: '/customers', icon: <People /> },
  { label: 'Employees', path: '/employees', icon: <Badge /> },
  { label: 'Services', path: '/services', icon: <Spa /> },
  { label: 'Settings', path: '/settings', icon: <Settings /> },
];

function resolveBranchId(branches: Branch[], preferredId?: string | null): string {
  return pickActiveBranch(branches, preferredId)?.id ?? '';
}

function readStoredLocationKey(): string {
  const orgId = localStorage.getItem('organization_id');
  const branchId = localStorage.getItem('branch_id');
  return orgId && branchId ? locationKey(orgId, branchId) : '';
}

export function AppLayout() {
  const [mobileOpen, setMobileOpen] = useState(false);
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [activeLocationKey, setActiveLocationKey] = useState(readStoredLocationKey);
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const { data: orgs } = useQuery({
    queryKey: ['organizations'],
    queryFn: async () => {
      const { data } = await api.get<{ data: Organization[] }>('/organizations');
      return data.data;
    },
  });

  const { data: branchesByOrg } = useQuery({
    queryKey: ['branches-by-org', orgs?.map((o) => o.id).join(',')],
    queryFn: async () => {
      if (!orgs?.length) return {} as Record<string, Branch[]>;
      const pairs = await Promise.all(
        orgs.map(async (org) => [org.id, await listBranches(org.id)] as const),
      );
      return Object.fromEntries(pairs);
    },
    enabled: !!orgs?.length,
  });

  const parsedLocation = parseLocationKey(activeLocationKey);
  const selectedOrgId = parsedLocation?.orgId || localStorage.getItem('organization_id') || orgs?.[0]?.id || '';
  const branchesForOrg = branchesByOrg?.[selectedOrgId] ?? [];
  const storedBranchId = parsedLocation?.branchId || localStorage.getItem('branch_id');
  const selectedBranchId = resolveBranchId(branchesForOrg, storedBranchId);

  const selectedLocationKey =
    selectedOrgId && selectedBranchId ? locationKey(selectedOrgId, selectedBranchId) : activeLocationKey;

  const locationOptions = useMemo(() => {
    if (!orgs?.length || !branchesByOrg) return [];
    return orgs.flatMap((org) => {
      const branches = (branchesByOrg[org.id] ?? []).filter((branch) => branch.is_active);
      return branches.map((branch) => ({
        org,
        branch,
        key: locationKey(org.id, branch.id),
      }));
    });
  }, [orgs, branchesByOrg]);

  const selectedLocation = locationOptions.find((option) => option.key === selectedLocationKey);

  useEffect(() => {
    if (!orgs?.length || !branchesByOrg) return;

    const orgId = localStorage.getItem('organization_id') || orgs[0].id;
    const branches = branchesByOrg[orgId] ?? [];
    const storedId = localStorage.getItem('branch_id');
    const branchId = resolveBranchId(branches, storedId);
    if (!orgId || !branchId) return;

    if (!localStorage.getItem('organization_id')) {
      localStorage.setItem('organization_id', orgId);
    }
    if (storedId !== branchId) {
      setActiveBranchId(branchId);
    }

    setActiveLocationKey(locationKey(orgId, branchId));
  }, [orgs, branchesByOrg]);

  const handleLocationChange = (value: string) => {
    const parsed = parseLocationKey(value);
    if (!parsed) return;
    setActiveLocation(parsed.orgId, parsed.branchId);
    window.location.reload();
  };

  const drawer = (
    <>
      <Toolbar />
      <List sx={{ px: 1, pt: 1 }}>
        {navItems.map((item) => (
          <ListItemButton
            key={item.path}
            selected={location.pathname === item.path}
            onClick={() => {
              navigate(item.path);
              setMobileOpen(false);
            }}
            sx={{ borderRadius: 2, mb: 0.5 }}
          >
            <ListItemIcon sx={{ minWidth: 40 }}>{item.icon}</ListItemIcon>
            <ListItemText primary={item.label} />
          </ListItemButton>
        ))}
      </List>
    </>
  );

  return (
    <Box sx={{ display: 'flex', minHeight: '100vh', bgcolor: 'background.default' }}>
      <AppBar
        position="fixed"
        sx={{
          zIndex: (t) => t.zIndex.drawer + 1,
          bgcolor: 'background.paper',
          color: 'text.primary',
          boxShadow: 1,
          width: { md: `calc(100% - ${DRAWER_WIDTH}px)` },
          ml: { md: `${DRAWER_WIDTH}px` },
        }}
      >
        <Toolbar>
          <IconButton
            edge="start"
            onClick={() => setMobileOpen(!mobileOpen)}
            sx={{ mr: 2, display: { md: 'none' } }}
          >
            <MenuIcon />
          </IconButton>
          <Typography variant="h6" sx={{ fontWeight: 700, color: 'primary.main' }}>
            Meetoria
          </Typography>
          <Box sx={{ flexGrow: 1 }} />
          {locationOptions.length > 0 && (
            <LocationSelect
              options={locationOptions}
              value={selectedLocationKey}
              selected={selectedLocation}
              onChange={handleLocationChange}
            />
          )}
          <IconButton onClick={(e) => setAnchorEl(e.currentTarget)}>
            <Avatar sx={{ width: 36, height: 36, bgcolor: 'primary.main' }}>
              {user?.name?.[0]?.toUpperCase() || 'U'}
            </Avatar>
          </IconButton>
          <Menu anchorEl={anchorEl} open={Boolean(anchorEl)} onClose={() => setAnchorEl(null)}>
            <MenuItem disabled>{user?.email}</MenuItem>
            <Divider />
            <MenuItem onClick={logout}>
              <ListItemIcon><Logout fontSize="small" /></ListItemIcon>
              Logout
            </MenuItem>
          </Menu>
        </Toolbar>
      </AppBar>

      <Box
        component="nav"
        sx={{ width: { md: DRAWER_WIDTH }, flexShrink: { md: 0 } }}
      >
        <Drawer
          variant="temporary"
          open={mobileOpen}
          onClose={() => setMobileOpen(false)}
          ModalProps={{ keepMounted: true }}
          sx={{
            display: { xs: 'block', md: 'none' },
            '& .MuiDrawer-paper': {
              width: DRAWER_WIDTH,
              boxSizing: 'border-box',
              borderRight: '1px solid',
              borderColor: 'divider',
            },
          }}
        >
          {drawer}
        </Drawer>
        <Drawer
          variant="permanent"
          sx={{
            display: { xs: 'none', md: 'block' },
            '& .MuiDrawer-paper': {
              width: DRAWER_WIDTH,
              boxSizing: 'border-box',
              borderRight: '1px solid',
              borderColor: 'divider',
            },
          }}
          open
        >
          {drawer}
        </Drawer>
      </Box>

      <Box
        component="main"
        sx={{
          flexGrow: 1,
          minWidth: 0,
          display: 'flex',
          flexDirection: 'column',
          pt: `${64 + CONTENT_GUTTER * 8}px`,
          px: CONTENT_GUTTER,
          pb: CONTENT_GUTTER,
          height: '100vh',
          boxSizing: 'border-box',
        }}
      >
        <Outlet />
      </Box>
    </Box>
  );
}

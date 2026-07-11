import {
  Box,
  Button,
  Container,
  Grid,
  Typography,
  Card,
  CardContent,
  Stack,
} from '@mui/material';
import {
  CalendarMonth,
  Analytics,
  Notifications,
  Groups,
} from '@mui/icons-material';
import { useAuth } from '../auth/AuthProvider';
import { useNavigate } from 'react-router-dom';

interface LandingPageProps {
  onLogin?: () => void;
}

const features = [
  { icon: <CalendarMonth fontSize="large" />, title: 'Smart Scheduling', desc: 'Prevent double bookings with real-time availability and timezone support.' },
  { icon: <Groups fontSize="large" />, title: 'Multi-Tenant', desc: 'Manage multiple locations and teams from a single platform.' },
  { icon: <Notifications fontSize="large" />, title: 'Automated Notifications', desc: 'SMS and email reminders keep customers informed automatically.' },
  { icon: <Analytics fontSize="large" />, title: 'Business Analytics', desc: 'Track revenue, utilization, and customer growth with actionable insights.' },
];

export function LandingPage({ onLogin }: LandingPageProps) {
  const { isAuthenticated, login } = useAuth();
  const navigate = useNavigate();

  const handleGetStarted = () => {
    if (isAuthenticated) {
      navigate('/dashboard');
    } else if (onLogin) {
      onLogin();
    } else {
      login();
    }
  };

  return (
    <Box>
      <Box
        sx={{
          background: 'linear-gradient(135deg, #6C3AED 0%, #EC4899 100%)',
          color: 'white',
          py: { xs: 8, md: 12 },
        }}
      >
        <Container maxWidth="lg">
          <Typography variant="h2" fontWeight={800} gutterBottom>
            Meetoria
          </Typography>
          <Typography variant="h5" sx={{ opacity: 0.9, mb: 4, maxWidth: 600 }}>
            Schedule Smarter. Grow Faster.
          </Typography>
          <Typography variant="body1" sx={{ opacity: 0.85, mb: 4, maxWidth: 540 }}>
            The all-in-one appointment platform for hair salons, beauty studios, and service businesses worldwide.
          </Typography>
          <Stack direction="row" spacing={2}>
            <Button variant="contained" size="large" onClick={handleGetStarted} sx={{ bgcolor: 'white', color: 'primary.main', '&:hover': { bgcolor: 'grey.100' } }}>
              {isAuthenticated ? 'Go to Dashboard' : 'Get Started Free'}
            </Button>
            {!isAuthenticated && (
              <Button variant="outlined" size="large" onClick={login} sx={{ borderColor: 'white', color: 'white' }}>
                Sign In
              </Button>
            )}
          </Stack>
        </Container>
      </Box>

      <Container maxWidth="lg" sx={{ py: 8 }}>
        <Grid container spacing={4}>
          {features.map((feature) => (
            <Grid size={{ xs: 12, sm: 6, md: 3 }} key={feature.title}>
              <Card sx={{ height: '100%' }}>
                <CardContent>
                  <Box color="primary.main" mb={2}>{feature.icon}</Box>
                  <Typography variant="h6" gutterBottom>{feature.title}</Typography>
                  <Typography variant="body2" color="text.secondary">{feature.desc}</Typography>
                </CardContent>
              </Card>
            </Grid>
          ))}
        </Grid>
      </Container>
    </Box>
  );
}

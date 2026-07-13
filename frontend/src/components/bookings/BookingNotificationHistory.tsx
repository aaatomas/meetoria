import EmailOutlinedIcon from '@mui/icons-material/EmailOutlined';
import SmsOutlinedIcon from '@mui/icons-material/SmsOutlined';
import {
  Box,
  Chip,
  CircularProgress,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
} from '@mui/material';
import dayjs from 'dayjs';
import type { Notification } from '../../api/client';

function formatNotificationTemplate(template: string): string {
  return template.replace(/_/g, ' ').replace(/\b\w/g, (char) => char.toUpperCase());
}

function formatNotificationStatus(status: Notification['status']): string {
  return status.replace(/_/g, ' ');
}

function getNotificationStatusColors(status: Notification['status']) {
  switch (status) {
    case 'sent':
    case 'delivered':
      return { background: 'success.light', foreground: 'success.dark' };
    case 'failed':
      return { background: 'error.light', foreground: 'error.dark' };
    case 'queued':
      return { background: 'info.light', foreground: 'info.dark' };
    default:
      return { background: 'grey.200', foreground: 'text.primary' };
  }
}

function formatTimestamp(value?: string): string {
  if (!value) {
    return '—';
  }
  return dayjs(value).format('D MMM YYYY, HH:mm');
}

interface BookingNotificationHistoryProps {
  notifications: Notification[];
  isLoading: boolean;
  errorMessage?: string | null;
}

export function BookingNotificationHistory({
  notifications,
  isLoading,
  errorMessage,
}: BookingNotificationHistoryProps) {
  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" py={4}>
        <CircularProgress size={28} />
      </Box>
    );
  }

  if (errorMessage) {
    return (
      <Typography color="error" variant="body2">
        {errorMessage}
      </Typography>
    );
  }

  if (notifications.length === 0) {
    return (
      <Typography color="text.secondary" variant="body2">
        No notifications have been sent for this booking yet.
      </Typography>
    );
  }

  return (
    <TableContainer>
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell>Channel</TableCell>
            <TableCell>Template</TableCell>
            <TableCell>Recipient</TableCell>
            <TableCell>Status</TableCell>
            <TableCell>Created</TableCell>
            <TableCell>Sent</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {notifications.map((notification) => {
            const statusColors = getNotificationStatusColors(notification.status);
            return (
              <TableRow key={notification.id} hover>
                <TableCell>
                  <Stack direction="row" alignItems="center" spacing={0.75}>
                    {notification.channel === 'sms' ? (
                      <SmsOutlinedIcon fontSize="small" color="action" />
                    ) : (
                      <EmailOutlinedIcon fontSize="small" color="action" />
                    )}
                    <Typography variant="body2" sx={{ textTransform: 'uppercase' }}>
                      {notification.channel}
                    </Typography>
                  </Stack>
                </TableCell>
                <TableCell>{formatNotificationTemplate(notification.template)}</TableCell>
                <TableCell>{notification.recipient}</TableCell>
                <TableCell>
                  <Chip
                    size="small"
                    label={formatNotificationStatus(notification.status)}
                    sx={{
                      bgcolor: statusColors.background,
                      color: statusColors.foreground,
                      fontWeight: 600,
                      textTransform: 'capitalize',
                    }}
                  />
                </TableCell>
                <TableCell>{formatTimestamp(notification.created_at)}</TableCell>
                <TableCell>{formatTimestamp(notification.sent_at)}</TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </TableContainer>
  );
}

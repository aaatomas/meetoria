import { Box } from '@mui/material';
import type { Employee } from '../../api/client';
import { EmployeeAvatar } from '../employees/EmployeeAvatar';

interface RenderEmployeeLabelOptions {
  avatarSize?: number;
  showInactive?: boolean;
}

export function renderEmployeeLabel(
  employee: Employee,
  { avatarSize = 24, showInactive = true }: RenderEmployeeLabelOptions = {},
) {
  return (
    <Box display="flex" alignItems="center" gap={1} minWidth={0}>
      <EmployeeAvatar
        firstName={employee.first_name}
        lastName={employee.last_name}
        avatarUrl={employee.avatar_url}
        color={employee.color}
        size={avatarSize}
        cacheKey={employee.updated_at}
      />
      <Box component="span" sx={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
        {employee.first_name} {employee.last_name}
        {showInactive && !employee.is_active ? ' (inactive)' : ''}
      </Box>
    </Box>
  );
}

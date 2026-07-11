import { Avatar } from '@mui/material';
import { resolveUploadUrl } from '../../api/client';

interface EmployeeAvatarProps {
  firstName: string;
  lastName: string;
  avatarUrl?: string;
  color?: string;
  size?: number;
  cacheKey?: string;
}

export function EmployeeAvatar({
  firstName,
  lastName,
  avatarUrl,
  color,
  size = 40,
  cacheKey,
}: EmployeeAvatarProps) {
  const initials = `${firstName?.[0] ?? ''}${lastName?.[0] ?? ''}`.toUpperCase() || '?';
  const src = resolveUploadUrl(avatarUrl);
  const srcWithCache = src && cacheKey ? `${src}?v=${encodeURIComponent(cacheKey)}` : src;

  return (
    <Avatar
      src={srcWithCache}
      alt={`${firstName} ${lastName}`}
      sx={{
        width: size,
        height: size,
        bgcolor: color || 'primary.main',
        fontSize: size * 0.4,
      }}
    >
      {initials}
    </Avatar>
  );
}

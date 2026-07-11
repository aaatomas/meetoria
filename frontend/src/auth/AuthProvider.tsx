import { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { CircularProgress, Box } from '@mui/material';
import { keycloak, initKeycloak, login, logout, getToken } from './keycloak';

interface AuthContextType {
  isAuthenticated: boolean;
  isLoading: boolean;
  user: { email?: string; name?: string } | null;
  login: () => void;
  logout: () => void;
  getToken: () => string | undefined;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [user, setUser] = useState<{ email?: string; name?: string } | null>(null);

  useEffect(() => {
    initKeycloak()
      .then((authenticated) => {
        setIsAuthenticated(authenticated);
        if (authenticated && keycloak.tokenParsed) {
          setUser({
            email: keycloak.tokenParsed.email,
            name: keycloak.tokenParsed.name || keycloak.tokenParsed.preferred_username,
          });
        }
      })
      .finally(() => setIsLoading(false));

    keycloak.onTokenExpired = () => {
      keycloak.updateToken(30);
    };
  }, []);

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="100vh">
        <CircularProgress />
      </Box>
    );
  }

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated,
        isLoading,
        user,
        login,
        logout,
        getToken,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}

import Keycloak from 'keycloak-js';

const keycloakConfig = {
  url: import.meta.env.VITE_KEYCLOAK_URL || 'http://localhost:8080',
  realm: import.meta.env.VITE_KEYCLOAK_REALM || 'meetoria',
  clientId: import.meta.env.VITE_KEYCLOAK_CLIENT_ID || 'meetoria-web',
};

export const keycloak = new Keycloak(keycloakConfig);

export async function initKeycloak(): Promise<boolean> {
  const authenticated = await keycloak.init({
    onLoad: 'check-sso',
    pkceMethod: 'S256',
    checkLoginIframe: false,
  });
  return authenticated;
}

export function getToken(): string | undefined {
  return keycloak.token;
}

export function login(): void {
  keycloak.login();
}

export function logout(): void {
  keycloak.logout({ redirectUri: window.location.origin });
}

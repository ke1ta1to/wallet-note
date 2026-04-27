interface ViteTypeOptions {
  strictImportMetaEnv: unknown;
}

interface ImportMetaEnv {
  readonly VITE_COGNITO_USER_POOL_ID: string;
  readonly VITE_COGNITO_CLIENT_ID: string;
  readonly VITE_COGNITO_DOMAIN: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

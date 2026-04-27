import path from "node:path";
import { tanstackRouter } from "@tanstack/router-plugin/vite";
import react from "@vitejs/plugin-react";
import { defineConfig, loadEnv } from "vite";

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const apiProxyTarget = env.API_PROXY_TARGET;

  return {
    plugins: [
      tanstackRouter({
        target: "react",
        autoCodeSplitting: true,
      }),
      react(),
    ],
    resolve: {
      alias: {
        "@": path.resolve(import.meta.dirname, "./src"),
      },
    },
    server: {
      proxy: {
        "/api": apiProxyTarget
          ? {
              // Forward /api/* to the deployed API (e.g. CloudFront). API
              // Gateway's JWT Authorizer validates the Bearer token, so no
              // x-amzn-request-context injection here.
              target: apiProxyTarget,
              changeOrigin: true,
              secure: true,
            }
          : {
              // Forward /api/* to the local Go server. Decode the Cognito JWT
              // from the Authorization header and inject x-amzn-request-context
              // so the Go app sees the same shape API Gateway HTTP API delivers
              // in production. Single auth code path; no LOCAL_DEV branch in
              // web-server.
              target: "http://localhost:8080",
              changeOrigin: true,
              rewrite: (p) => p.replace(/^\/api/, ""),
              configure: (proxy) => {
                proxy.on("proxyReq", (proxyReq, req) => {
                  // Defensive: skip injection when no Bearer is present.
                  // The Go auth middleware will return 401, mirroring API
                  // Gateway's behavior in prod.
                  const auth = (req.headers.authorization ?? "").replace(
                    "Bearer ",
                    "",
                  );
                  if (!auth) return;
                  const payload = JSON.parse(
                    Buffer.from(auth.split(".")[1], "base64url").toString(),
                  );
                  const requestContext = {
                    authorizer: { jwt: { claims: payload } },
                  };
                  proxyReq.setHeader(
                    "x-amzn-request-context",
                    JSON.stringify(requestContext),
                  );
                });
              },
            },
      },
    },
  };
});

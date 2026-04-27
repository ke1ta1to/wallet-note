import path from "node:path";
import { tanstackRouter } from "@tanstack/router-plugin/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

// https://vite.dev/config/
export default defineConfig({
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
      // Forward /api/* to the local Go server. Decode the Cognito JWT from the
      // Authorization header and inject x-amzn-request-context so the Go app
      // sees the same shape API Gateway HTTP API delivers in production. This
      // keeps a single auth code path; no LOCAL_DEV branch in web-server.
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
        rewrite: (p) => p.replace(/^\/api/, ""),
        configure: (proxy) => {
          proxy.on("proxyReq", (proxyReq, req) => {
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
});

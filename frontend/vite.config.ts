import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import wails from "@wailsio/runtime/plugins/vite";

export default defineConfig({
  plugins: [
    react(),
    wails("./bindings"),
    {
      name: "development-csp",
      apply: "serve",
      transformIndexHtml: (html) =>
        html
          .replace("script-src 'self'", "script-src 'self' 'unsafe-inline'")
          .replace(
            "connect-src 'self'",
            "connect-src 'self' ws://127.0.0.1:9245 http://127.0.0.1:9245",
          ),
    },
  ],
  server: { host: "127.0.0.1", port: 9245, strictPort: true },
});

import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, ".", "");
  return {
    base: "/agent/",
    plugins: [react()],
    build: {
      rollupOptions: {
        // 设计预览独立构建，不进入正式应用入口和路由。
        input: { app: "index.html", uiPreview: "ui-preview.html" },
      },
    },
    server: {
      port: 5174,
      proxy: {
        "/api": {
          target: env.VITE_PROXY_TARGET || "http://127.0.0.1:8888",
          rewrite: (path) => path.replace(/^\/api/, ""),
        },
      },
    },
  };
});

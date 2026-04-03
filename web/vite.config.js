import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
export default defineConfig({
    base: "/",
    plugins: [react(), tailwindcss()],
    server: {
        port: 3000,
        strictPort: true,
        proxy: {
            "/api": "http://localhost:8000",
            "/login": "http://localhost:8000",
            "/logout": "http://localhost:8000",
            "/auth": "http://localhost:8000",
        },
    },
});

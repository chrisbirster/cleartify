import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
export default defineConfig({
    plugins: [react(), tailwindcss()],
    test: {
        environment: "jsdom",
        setupFiles: ["./test/setup.ts"],
        include: ["src/**/*.test.{ts,tsx}"],
        exclude: ["test/smoke/**"],
        coverage: {
            provider: "v8",
            all: true,
            include: ["src/**/*.{ts,tsx}"],
            exclude: ["src/index.css", "src/lib/types.ts"],
            reporter: ["text", "json-summary"],
            reportsDirectory: "./coverage",
            thresholds: {
                lines: 90,
                functions: 90,
                branches: 90,
                statements: 90,
            },
        },
    },
});

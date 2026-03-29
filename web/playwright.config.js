import { defineConfig } from "@playwright/test";
export default defineConfig({
    testDir: "./test/smoke",
    use: {
        baseURL: "http://127.0.0.1:3000",
    },
    webServer: {
        command: "node ./node_modules/vite/bin/vite.js --port 3000 --strictPort --host 127.0.0.1",
        port: 3000,
        reuseExistingServer: !process.env.CI,
        timeout: 120000,
    },
});

import {defineConfig} from '@playwright/test';
export default defineConfig({
 testDir:'./e2e',
 testMatch:'*.spec.ts',
 timeout:90000,
 expect:{timeout:10000},
 workers:1,
 retries:0,
 use:{browserName:'chromium',headless:true,baseURL:'http://127.0.0.1:3000',trace:'off',video:'off',screenshot:'off'},
 reporter:[['list']],
});

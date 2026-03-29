import { expect, test } from "@playwright/test";

import { appDataFixture } from "../fixtures";

test("loads the app shell with mocked phase zero API data", async ({ page }) => {
  await page.route("**/api/auth/session", async (route) => {
    await route.fulfill({ json: appDataFixture.session });
  });
  await page.route("**/api/signals", async (route) => {
    await route.fulfill({ json: appDataFixture.signals });
  });
  await page.route("**/api/clusters", async (route) => {
    await route.fulfill({ json: appDataFixture.clusters });
  });
  await page.route("**/api/work-objects", async (route) => {
    await route.fulfill({ json: appDataFixture.workObjects });
  });

  await page.goto("/app/");

  await expect(page.getByRole("heading", { name: "Today" })).toBeVisible();
  await expect(page.getByRole("link", { name: "Context Inbox" })).toBeVisible();
  await expect(page.getByText("Demo User")).toBeVisible();
});

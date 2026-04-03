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

  await page.goto("/");

  await expect(page.getByRole("heading", { name: "Today" })).toBeVisible();
  await expect(page.getByRole("link", { name: "Context Inbox" })).toBeVisible();
  await expect(page.getByText("Demo User")).toBeVisible();
});

test("runs the hero path without runtime errors and shows release queue semantics clearly", async ({ page }) => {
  const pageErrors: string[] = [];
  const consoleErrors: string[] = [];
  const workObjects = JSON.parse(JSON.stringify(appDataFixture.workObjects));
  const persistedDraftPlan = {
    workObjectId: "wo_onboarding_refresh",
    overview:
      'This draft plan turns "Guided onboarding refresh" into executable work using the current summary, goal, and prototype context.',
    implementationSteps: [
      "Review the Work Object goal.",
      "Translate the current summary into the smallest useful implementation slice.",
    ],
    edgeCases: ["Prototype drift after review."],
    openQuestions: ["Which prototype states must ship first?"],
    createdAt: "2026-03-31T09:00:00Z",
    updatedAt: "2026-03-31T09:05:00Z",
  };

  page.on("pageerror", (error) => {
    pageErrors.push(error.message);
  });
  page.on("console", (message) => {
    if (message.type() === "error") {
      consoleErrors.push(message.text());
    }
  });

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
    await route.fulfill({ json: workObjects });
  });
  await page.route("**/api/work-objects/wo_onboarding_refresh/plan", async (route) => {
    if (route.request().method() === "GET") {
      await route.fulfill({ json: persistedDraftPlan });
      return;
    }

    await route.fulfill({ json: persistedDraftPlan });
  });
  await page.route("**/api/work-objects/wo_onboarding_refresh/release-ready", async (route) => {
    const updatedWorkObject = {
      ...workObjects.find((entry: (typeof workObjects)[number]) => entry.id === "wo_onboarding_refresh"),
      releaseReadyAt: "2026-03-31T17:00:00Z",
      updatedAt: "2026-03-31T17:00:00Z",
    };
    const nextWorkObjects = workObjects.map((entry: (typeof workObjects)[number]) =>
      entry.id === "wo_onboarding_refresh" ? updatedWorkObject : entry,
    );
    workObjects.splice(0, workObjects.length, ...nextWorkObjects);

    await route.fulfill({ json: updatedWorkObject });
  });

  await page.goto("/work/wo_onboarding_refresh");
  await expect(page.getByRole("heading", { name: "Guided onboarding refresh", level: 1 })).toBeVisible();

  await page.goto("/plans/wo_onboarding_refresh");
  await expect(page.getByRole("heading", { name: "Draft plan" })).toBeVisible();
  await expect(page.getByText(/This draft plan turns "Guided onboarding refresh"/)).toBeVisible();

  await page.goto("/agents/wo_onboarding_refresh");
  await expect(page.getByRole("heading", { name: "Agent Workspace" })).toBeVisible();
  await expect(page.getByText(/Create execution artifacts/)).toBeVisible();

  await page.goto("/review/wo_onboarding_refresh");
  await expect(page.getByRole("heading", { name: "Review", level: 1 })).toBeVisible();
  await expect(page.getByText(/sent into the release queue/i)).toBeVisible();
  await page.getByRole("button", { name: "Mark release-ready" }).click();
  await expect(page.getByText(/Release-ready on Mar 31, 2026/i)).toBeVisible();

  await page.goto("/releases");
  await expect(page.getByRole("heading", { name: "Release queue" })).toBeVisible();
  await expect(page.getByText(/appear here after Review marks them release-ready/i)).toBeVisible();
  await expect(page.getByText("Guided onboarding refresh")).toBeVisible();

  expect(pageErrors).toEqual([]);
  expect(consoleErrors).toEqual([]);
});

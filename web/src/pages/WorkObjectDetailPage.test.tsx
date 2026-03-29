import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { WorkObjectDetailPage } from "./WorkObjectDetailPage";
import { appDataFixture } from "../../test/fixtures";

describe("WorkObjectDetailPage", () => {
  it("renders the selected work object", () => {
    render(
      <MemoryRouter initialEntries={["/work/wo_mobile_perf"]}>
        <Routes>
          <Route path="/work/:workObjectId" element={<WorkObjectDetailPage workObjects={appDataFixture.workObjects} />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "Speed up mobile dashboard rendering" })).toBeInTheDocument();
    expect(screen.getByText("Virtualize heavy dashboard sections")).toBeInTheDocument();
    expect(screen.getByText("Mobile optimization implementation plan")).toBeInTheDocument();
  });

  it("renders a not found empty state", () => {
    render(
      <MemoryRouter initialEntries={["/work/missing"]}>
        <Routes>
          <Route path="/work/:workObjectId" element={<WorkObjectDetailPage workObjects={appDataFixture.workObjects} />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "Work Object not found" })).toBeInTheDocument();
  });
});

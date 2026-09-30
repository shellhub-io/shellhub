import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { getConfig, defaultConfig, type Edition } from "@/env";
import AuthLayout from "../AuthLayout";

const mockGetConfig = vi.mocked(getConfig);

function renderLayout(edition: Edition) {
  mockGetConfig.mockReturnValue({
    ...defaultConfig,
    edition,
    version: "v0.99.0",
  });

  return render(
    <MemoryRouter initialEntries={["/login"]}>
      <Routes>
        <Route element={<AuthLayout />}>
          <Route path="/login" element={<p>sign-in screen</p>} />
        </Route>
      </Routes>
    </MemoryRouter>,
  );
}

describe("AuthLayout", () => {
  it("shows the docs, version and community line under the screen on community", () => {
    renderLayout("community");

    expect(screen.getByText("sign-in screen")).toBeInTheDocument();
    expect(screen.getByText("v0.99.0")).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: /documentation/i }),
    ).toBeInTheDocument();
  });

  it.each<Edition>(["enterprise", "cloud"])(
    "ends at the screen on %s",
    (edition) => {
      renderLayout(edition);

      expect(screen.getByText("sign-in screen")).toBeInTheDocument();
      expect(screen.queryByText("v0.99.0")).not.toBeInTheDocument();
      expect(
        screen.queryByRole("link", { name: /documentation/i }),
      ).not.toBeInTheDocument();
    },
  );
});

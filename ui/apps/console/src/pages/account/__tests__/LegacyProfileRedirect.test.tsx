import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes, useLocation } from "react-router-dom";
import LegacyProfileRedirect from "../LegacyProfileRedirect";

function Pathname() {
  return <output aria-label="Current page">{useLocation().pathname}</output>;
}

function renderAt(path: string) {
  render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route path="/profile/*" element={<LegacyProfileRedirect />} />
        <Route path="/account/*" element={<Pathname />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("LegacyProfileRedirect", () => {
  it.each([
    ["/profile", "/account"],
    ["/profile/security", "/account/security"],
  ])("sends %s to %s", (from, to) => {
    renderAt(from);

    expect(screen.getByLabelText("Current page")).toHaveTextContent(to);
  });
});

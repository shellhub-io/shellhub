import { describe, it, expect, afterEach } from "vitest";
import { render, cleanup } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { routes } from "@/routes";
import App from "@/App";

afterEach(cleanup);

describe("routes smoke", () => {
  it("exports exactly 12 route entries", () => {
    // exactly 12 — update this assertion when a route is added or removed
    expect(routes).toHaveLength(12);
  });

  it("route paths are unique", () => {
    const paths = routes.map((r) => r.path);
    expect(new Set(paths).size).toBe(paths.length);
  });

  for (const { path } of routes) {
    it(`mounts without throwing: ${path}`, () => {
      expect(() => {
        render(
          <MemoryRouter initialEntries={[path]}>
            <App />
          </MemoryRouter>,
        );
      }).not.toThrow();
    });
  }
});

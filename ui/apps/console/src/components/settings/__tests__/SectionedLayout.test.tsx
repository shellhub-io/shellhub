import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { Cog6ToothIcon, KeyIcon } from "@heroicons/react/24/outline";

const viewport = await vi.hoisted(async () =>
  (await import("@/tests/viewport")).installViewport(),
);

import SectionedLayout from "../SectionedLayout";

function renderAt(path: string) {
  render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route
          path="/prefs"
          element={
            <SectionedLayout
              base="/prefs"
              icon={<Cog6ToothIcon />}
              title="Preferences"
              description="How things behave"
              sections={[
                { to: "general", label: "General", icon: Cog6ToothIcon },
                { to: "keys", label: "Keys", icon: KeyIcon },
              ]}
            />
          }
        >
          <Route path="general" element={<p>General content</p>} />
          <Route path="keys" element={<p>Keys content</p>} />
        </Route>
      </Routes>
    </MemoryRouter>,
  );
}

describe("SectionedLayout", () => {
  describe("on a wide window", () => {
    beforeEach(() => {
      viewport.wide = true;
    });

    it("opens the first section from the base path", () => {
      renderAt("/prefs");

      expect(screen.getByText("General content")).toBeInTheDocument();
    });

    it("shows the header and the menu beside the section", () => {
      renderAt("/prefs/keys");

      expect(
        screen.getByRole("heading", { name: "Preferences" }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("navigation", { name: "Preferences sections" }),
      ).toBeInTheDocument();
      expect(screen.getByText("Keys content")).toBeInTheDocument();
    });

    it("moves between sections from the menu", async () => {
      renderAt("/prefs/general");

      const user = userEvent.setup();
      await user.click(screen.getByRole("link", { name: "Keys" }));

      expect(screen.getByText("Keys content")).toBeInTheDocument();
    });
  });

  describe("on a narrow window", () => {
    beforeEach(() => {
      viewport.wide = false;
    });

    it("shows the menu alone at the base path", () => {
      renderAt("/prefs");

      expect(
        screen.getByRole("heading", { name: "Preferences" }),
      ).toBeInTheDocument();
      expect(screen.getByRole("link", { name: "General" })).toBeInTheDocument();
      expect(screen.queryByText("General content")).not.toBeInTheDocument();
    });

    it("shows a section alone, with a way back instead of the header", () => {
      renderAt("/prefs/keys");

      expect(screen.getByText("Keys content")).toBeInTheDocument();
      expect(
        screen.queryByRole("heading", { name: "Preferences" }),
      ).not.toBeInTheDocument();
      expect(
        screen.queryByRole("navigation", { name: "Preferences sections" }),
      ).not.toBeInTheDocument();
    });

    it("goes back to the menu", async () => {
      renderAt("/prefs/keys");

      const user = userEvent.setup();
      await user.click(screen.getByRole("link", { name: "Preferences" }));

      expect(screen.queryByText("Keys content")).not.toBeInTheDocument();
      expect(screen.getByRole("link", { name: "Keys" })).toBeInTheDocument();
    });
  });
});

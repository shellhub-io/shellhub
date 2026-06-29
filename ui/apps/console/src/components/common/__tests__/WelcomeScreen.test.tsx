import { describe, it, expect, vi, beforeEach } from "vitest";
import { render } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

import WelcomeScreen from "../WelcomeScreen";
import AuthFooterLinks from "../AuthFooterLinks";

vi.mock("@shellhub/design-system/primitives", async (importOriginal) => {
  const original =
    await importOriginal<typeof import("@shellhub/design-system/primitives")>();
  return {
    ...original,
    GithubIcon: vi.fn((props) => original.GithubIcon(props)),
  };
});

vi.mock("@/env", () => ({
  getConfig: vi.fn(() => ({ cloud: false })),
}));

function renderWelcomeScreen() {
  return render(
    <MemoryRouter>
      <WelcomeScreen namespaceName="my-namespace" />
    </MemoryRouter>,
  );
}

function renderAuthFooterLinks() {
  return render(
    <MemoryRouter>
      <AuthFooterLinks />
    </MemoryRouter>,
  );
}

describe("WelcomeScreen", () => {
  it("hides the decorative ConnectionGrid from assistive technology", () => {
    const { container } = renderWelcomeScreen();
    const grid = container.querySelector(".connection-grid");
    expect(grid).not.toBeNull();
    expect(grid).toHaveAttribute("aria-hidden", "true");
  });

  it("renders the GitHub link using DS GithubIcon instead of a bare inline SVG", async () => {
    const { GithubIcon } = await import("@shellhub/design-system/primitives");

    renderWelcomeScreen();

    expect(GithubIcon).toHaveBeenCalled();
  });
});

describe("AuthFooterLinks", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders the GitHub link using DS GithubIcon instead of a bare inline SVG", async () => {
    const { GithubIcon } = await import("@shellhub/design-system/primitives");

    renderAuthFooterLinks();

    expect(GithubIcon).toHaveBeenCalled();
  });
});

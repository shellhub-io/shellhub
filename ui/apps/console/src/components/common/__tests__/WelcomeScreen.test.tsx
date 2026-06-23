import { describe, it, expect } from "vitest";
import { render } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

import WelcomeScreen from "../WelcomeScreen";

function renderWelcomeScreen() {
  return render(
    <MemoryRouter>
      <WelcomeScreen namespaceName="my-namespace" />
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
});

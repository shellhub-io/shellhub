import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";

import SettingsRow from "@/components/common/SettingsRow";

describe("SettingsRow", () => {
  it("renders icon", () => {
    render(
      <SettingsRow
        icon={<svg data-testid="row-icon" />}
        title="Two-Factor Auth"
        description="Protect your account."
      >
        <button type="button">Enable</button>
      </SettingsRow>,
    );
    expect(screen.getByTestId("row-icon")).toBeInTheDocument();
  });

  it("renders title", () => {
    render(
      <SettingsRow
        icon={<svg />}
        title="Two-Factor Auth"
        description="Protect your account."
      >
        <button type="button">Enable</button>
      </SettingsRow>,
    );
    expect(screen.getByText("Two-Factor Auth")).toBeInTheDocument();
  });

  it("renders description", () => {
    render(
      <SettingsRow
        icon={<svg />}
        title="Two-Factor Auth"
        description="Protect your account."
      >
        <button type="button">Enable</button>
      </SettingsRow>,
    );
    expect(screen.getByText("Protect your account.")).toBeInTheDocument();
  });

  it("renders children", () => {
    render(
      <SettingsRow
        icon={<svg />}
        title="Two-Factor Auth"
        description="Protect your account."
      >
        <button type="button">Enable</button>
      </SettingsRow>,
    );
    expect(screen.getByRole("button", { name: "Enable" })).toBeInTheDocument();
  });

  it("renders badge when provided", () => {
    render(
      <SettingsRow
        icon={<svg />}
        title="Two-Factor Auth"
        description="Protect your account."
        badge={<span data-testid="badge-node">Recommended</span>}
      >
        <button type="button">Enable</button>
      </SettingsRow>,
    );
    expect(screen.getByTestId("badge-node")).toBeInTheDocument();
    expect(screen.getByText("Recommended")).toBeInTheDocument();
  });

  it("does not render badge node when omitted", () => {
    render(
      <SettingsRow
        icon={<svg />}
        title="Two-Factor Auth"
        description="Protect your account."
      >
        <button type="button">Enable</button>
      </SettingsRow>,
    );
    expect(screen.queryByTestId("badge-node")).not.toBeInTheDocument();
  });
});

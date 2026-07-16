import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";

import { ClipboardProvider } from "../ClipboardProvider";
import IdentityCard from "../IdentityCard";

function renderCard(props: Partial<Parameters<typeof IdentityCard>[0]> = {}) {
  return render(
    <ClipboardProvider>
      <IdentityCard uid="abc-def-123" mac="aa:bb:cc:dd" remoteAddr="192.168.1.1" {...props} />
    </ClipboardProvider>,
  );
}

describe("IdentityCard", () => {
  it("renders the Identity heading", () => {
    renderCard();

    expect(
      screen.getByRole("heading", { name: /identity/i }),
    ).toBeInTheDocument();
  });

  it("renders UID truncated to 8 characters", () => {
    renderCard();

    expect(screen.getByText("abc-def-")).toBeInTheDocument();
  });

  it("renders MAC Address", () => {
    renderCard();

    expect(screen.getByText("aa:bb:cc:dd")).toBeInTheDocument();
  });

  it("renders Remote Address", () => {
    renderCard();

    expect(screen.getByText("192.168.1.1")).toBeInTheDocument();
  });

  it("shows copy buttons for UID and MAC", () => {
    renderCard();

    const copyButtons = screen.getAllByRole("button", { name: /copy/i });
    expect(copyButtons).toHaveLength(2);
  });

  it("shows em-dash for empty MAC and Remote Address", () => {
    renderCard({ mac: "", remoteAddr: "" });

    const dashes = screen.getAllByText("—");
    expect(dashes).toHaveLength(2);
  });
});

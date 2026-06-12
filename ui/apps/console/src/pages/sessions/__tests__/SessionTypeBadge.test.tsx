import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import SessionTypeBadge from "../SessionTypeBadge";

describe("SessionTypeBadge", () => {
  it("renders 'sftp' for subsystem type", () => {
    render(<SessionTypeBadge types={["subsystem"]} />);
    expect(screen.getByText("sftp")).toBeInTheDocument();
  });

  it("renders 'exec' for exec type", () => {
    render(<SessionTypeBadge types={["exec"]} />);
    expect(screen.getByText("exec")).toBeInTheDocument();
  });

  it("renders 'shell' for shell type", () => {
    render(<SessionTypeBadge types={["shell"]} />);
    expect(screen.getByText("shell")).toBeInTheDocument();
  });

  it("renders 'shell' for pty-req type", () => {
    render(<SessionTypeBadge types={["pty-req"]} />);
    expect(screen.getByText("shell")).toBeInTheDocument();
  });

  it("renders nothing for unknown types", () => {
    const { container } = render(<SessionTypeBadge types={["unknown"]} />);
    expect(container.firstChild).toBeNull();
  });

  it("renders nothing for empty types array", () => {
    const { container } = render(<SessionTypeBadge types={[]} />);
    expect(container.firstChild).toBeNull();
  });

  it("prefers 'sftp' when subsystem type is present alongside others", () => {
    render(<SessionTypeBadge types={["subsystem", "shell"]} />);
    expect(screen.getByText("sftp")).toBeInTheDocument();
    expect(screen.queryByText("shell")).not.toBeInTheDocument();
  });
});

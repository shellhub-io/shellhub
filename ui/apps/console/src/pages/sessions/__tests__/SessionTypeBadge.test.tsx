import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import SessionTypeBadge from "../SessionTypeBadge";

describe("SessionTypeBadge", () => {
  it.each([
    [["subsystem"], "sftp"],
    [["exec"], "exec"],
    [["shell"], "shell"],
    [["pty-req"], "shell"],
  ])("renders %j as '%s'", (types, label) => {
    render(<SessionTypeBadge types={types} />);
    expect(screen.getByText(label)).toBeInTheDocument();
  });

  it.each([[["unknown"]], [[] as string[]]])(
    "renders nothing for %j",
    (types) => {
      const { container } = render(<SessionTypeBadge types={types} />);
      expect(container.firstChild).toBeNull();
    },
  );

  it("prefers 'sftp' when subsystem type is present alongside others", () => {
    render(<SessionTypeBadge types={["subsystem", "shell"]} />);
    expect(screen.getByText("sftp")).toBeInTheDocument();
    expect(screen.queryByText("shell")).not.toBeInTheDocument();
  });
});

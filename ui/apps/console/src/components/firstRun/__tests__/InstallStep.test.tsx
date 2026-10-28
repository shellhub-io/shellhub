import type { ComponentProps } from "react";
import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { ClipboardProvider } from "@/components/common/ClipboardProvider";
import InstallStep from "../InstallStep";

const { origin } = window.location;

const renderInstallStep = (props: ComponentProps<typeof InstallStep> = {}) =>
  render(<InstallStep {...props} />, { wrapper: ClipboardProvider });

describe("InstallStep", () => {
  it.each([
    {
      namespace: "a new",
      props: {},
      waiting:
        "Waiting for acceptance... (code expires in 10 minutes; a server that renews it keeps it valid while this command waits, up to an hour)",
    },
    {
      namespace: "a legacy",
      props: { legacy: true, tenant: "tenant-1" },
      waiting: "Waiting for acceptance... (code expires in 10 minutes)",
    },
  ])(
    "shows the shellhub-agent login output for $namespace namespace",
    ({ props, waiting }) => {
      renderInstallStep(props);

      expect(
        screen.getByText(
          "To accept this device, open the ShellHub console and enter this code:",
        ),
      ).toBeInTheDocument();
      expect(screen.getByText("XXXX-XXXX")).toBeInTheDocument();
      expect(
        screen.getByText(
          `Enter it at ${origin}/accept-device, or open the link:`,
        ),
      ).toBeInTheDocument();
      expect(
        screen.getByText(`${origin}/accept-device?code=`),
      ).toBeInTheDocument();
      expect(screen.getByText("XXXXXXXX")).toBeInTheDocument();
      expect(screen.getByText(waiting)).toBeInTheDocument();
    },
  );

  it("shows the plain install command and no enrolment log for a new namespace", () => {
    renderInstallStep();

    expect(
      screen.getByText(`curl -sSf ${origin}/install.sh | sh`),
    ).toBeInTheDocument();
    expect(
      screen.queryByText(/not enrolled in any namespace/),
    ).not.toBeInTheDocument();
  });

  it("shows the tenant install command and the login command for a legacy namespace", () => {
    renderInstallStep({ legacy: true, tenant: "tenant-1" });

    expect(
      screen.getByText(
        `curl -sSf ${origin}/install.sh | TENANT_ID=tenant-1 sh`,
      ),
    ).toBeInTheDocument();
    expect(screen.getByText("shellhub-agent login")).toBeInTheDocument();
  });
});

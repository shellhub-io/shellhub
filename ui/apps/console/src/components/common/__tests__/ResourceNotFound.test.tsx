import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { CpuChipIcon, ShieldExclamationIcon } from "@heroicons/react/24/outline";
import ResourceNotFound from "../ResourceNotFound";

function renderNotFound(props: {
  icon: React.ComponentType<React.SVGProps<SVGSVGElement>>;
  resource: string;
  backTo: string;
} = {
  icon: CpuChipIcon,
  resource: "Device",
  backTo: "/devices",
}) {
  return render(
    <MemoryRouter>
      <ResourceNotFound {...props} />
    </MemoryRouter>,
  );
}

describe("ResourceNotFound", () => {
  it("announces itself as a status region for screen readers", () => {
    renderNotFound();

    expect(screen.getByRole("status")).toBeInTheDocument();
  });

  it('renders "{resource} not found" heading and "Back to {resource}s" link', () => {
    renderNotFound();

    expect(screen.getByText("Device not found")).toBeInTheDocument();
    const link = screen.getByRole("link", { name: "Back to devices" });
    expect(link).toHaveAttribute("href", "/devices");
  });

  it("handles multi-word resource names", () => {
    renderNotFound({
      icon: ShieldExclamationIcon,
      resource: "Firewall rule",
      backTo: "/admin/firewall-rules",
    });

    expect(screen.getByText("Firewall rule not found")).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: "Back to firewall rules" }),
    ).toHaveAttribute("href", "/admin/firewall-rules");
  });
});

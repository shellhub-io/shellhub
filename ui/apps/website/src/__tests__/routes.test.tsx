import { describe, it, expect, afterEach } from "vitest";
import { render, cleanup, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { routes } from "@/routes";
import App from "@/App";

afterEach(cleanup);

const pageHeadings = [
  ["/", "All your devices. One gateway."],
  ["/getting-started", "Get started with ShellHub"],
  ["/enterprise", "Device management built for scale and control"],
  ["/pricing", "Simple, transparent pricing"],
  ["/features", "Everything you need to manage remote devices"],
  ["/how-it-works", "From anywhere to any device"],
  ["/integrations", "Works with your existing tools"],
  ["/use-cases/iot-embedded", "IoT & Embedded Device Management"],
  ["/use-cases/edge-computing", "Edge Computing Access"],
  ["/use-cases/remote-support", "Remote Support & Troubleshooting"],
  ["/use-cases/devops-ci-cd", "DevOps & CI/CD Automation"],
  ["/use-cases/container-management", "SSH into Docker containers without docker exec"],
];

describe("routes smoke", () => {
  it("covers every declared route once", () => {
    const paths = routes.map((r) => r.path);
    expect(new Set(paths).size).toBe(paths.length);
    expect(pageHeadings.map(([path]) => path).sort()).toEqual(paths.sort());
  });

  it.each(pageHeadings)("%s shows its page", (path, heading) => {
    render(
      <MemoryRouter initialEntries={[path]}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { level: 1, name: heading })).toBeInTheDocument();
  });
});

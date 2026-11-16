import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import PlaintextServerError from "../PlaintextServerError";

describe("PlaintextServerError", () => {
  it.each([
    "http://shellhub.example.com",
    "http://192.0.2.10:8080",
    "http://127.example.com",
  ])("explains that the agent refuses %s", (origin) => {
    render(<PlaintextServerError origin={origin} />);

    const alert = screen.getByRole("alert");
    expect(alert).toHaveTextContent(/the agent\s+refuses to connect/);
    expect(alert).toHaveTextContent("SHELLHUB_ALLOW_INSECURE_SERVER=true");
  });

  it.each([
    "https://shellhub.example.com",
    "http://localhost:8080",
    "http://LOCALHOST",
    "http://shellhub.localhost",
    "http://127.0.0.1",
    "http://[::1]:8080",
  ])("stays hidden for %s", (origin) => {
    const { container } = render(<PlaintextServerError origin={origin} />);

    expect(container).toBeEmptyDOMElement();
  });
});

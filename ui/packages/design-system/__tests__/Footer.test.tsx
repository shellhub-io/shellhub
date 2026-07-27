import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { Footer } from "../components/Footer";

describe("Footer", () => {
  describe("app=website (default)", () => {
    it("renders website links via linkComponent", () => {
      render(<Footer />);
      const link = screen.getByRole("link", { name: "Features" });
      expect(link).not.toHaveAttribute("target");
    });

    it("renders docs link as external", () => {
      render(<Footer />);
      const link = screen.getByRole("link", { name: "Documentation" });
      expect(link).toHaveAttribute("target", "_blank");
      expect(link).toHaveAttribute("rel", "noopener noreferrer");
    });
  });

  describe("app=docs", () => {
    it("renders website links as external", () => {
      render(<Footer app="docs" />);
      const link = screen.getByRole("link", { name: "Features" });
      expect(link).toHaveAttribute("target", "_blank");
    });

    it("renders docs link via linkComponent", () => {
      render(<Footer app="docs" />);
      const link = screen.getByRole("link", { name: "Documentation" });
      expect(link).not.toHaveAttribute("target");
    });
  });

  it("renders the GitHub social link as external", () => {
    render(<Footer />);
    const link = screen.getByRole("link", { name: "GitHub" });
    expect(link).toHaveAttribute("href", "https://github.com/shellhub-io/shellhub");
    expect(link).toHaveAttribute("target", "_blank");
  });

  it("uses linkComponent for same-app links", () => {
    function CustomLink({
      href,
      className,
      children,
    }: {
      href: string;
      className?: string;
      children: React.ReactNode;
    }) {
      return (
        <a href={href} className={className} data-router="true">
          {children}
        </a>
      );
    }

    render(<Footer linkComponent={CustomLink} />);

    expect(screen.getByRole("link", { name: "Features" })).toHaveAttribute(
      "data-router",
      "true",
    );
    expect(screen.getByRole("link", { name: "Documentation" })).not.toHaveAttribute(
      "data-router",
    );
  });

  it("renders the copyright text with the current year", () => {
    render(<Footer />);
    const year = new Date().getFullYear();
    expect(
      screen.getByText(new RegExp(`${year}.*ShellHub.*All rights reserved`)),
    ).toBeInTheDocument();
  });
});

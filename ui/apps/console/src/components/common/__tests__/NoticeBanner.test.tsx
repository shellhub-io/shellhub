import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import NoticeBanner from "@/components/common/NoticeBanner";

const getStrip = (container: HTMLElement): HTMLElement =>
  container.querySelector('[role="alert"],[role="status"]') as HTMLElement;

describe("NoticeBanner", () => {
  describe("visibility toggle", () => {
    it("collapsed (visible=false) has grid-rows-[0fr], aria-hidden=true, and inert on the outer wrapper", () => {
      const { container } = render(
        <NoticeBanner visible={false} severity="error">
          Message
        </NoticeBanner>,
      );
      const outer = container.firstChild as HTMLElement;
      expect(outer.className).toContain("grid-rows-[0fr]");
      expect(outer).toHaveAttribute("aria-hidden", "true");
      expect(outer).toHaveAttribute("inert");
    });

    it("expanded (visible=true) has grid-rows-[1fr], no aria-hidden, and no inert on the outer wrapper", () => {
      const { container } = render(
        <NoticeBanner visible={true} severity="error">
          Message
        </NoticeBanner>,
      );
      const outer = container.firstChild as HTMLElement;
      expect(outer.className).toContain("grid-rows-[1fr]");
      expect(outer).not.toHaveAttribute("aria-hidden");
      expect(outer).not.toHaveAttribute("inert");
    });
  });

  describe("severity color classes", () => {
    it.each<["error" | "warning", string]>([
      ["error", "bg-accent-red"],
      ["warning", "bg-accent-yellow"],
    ])(
      "severity=%s applies the correct background class to the strip",
      (severity, expected) => {
        const { container } = render(
          <NoticeBanner visible={true} severity={severity}>
            Message
          </NoticeBanner>,
        );
        const strip = getStrip(container);
        expect(strip.className).toContain(expected);
      },
    );
  });

  describe("severity role / aria-live", () => {
    it("severity=error uses role=alert and aria-live=assertive on the strip", () => {
      const { container } = render(
        <NoticeBanner visible={true} severity="error">
          Message
        </NoticeBanner>,
      );
      const strip = getStrip(container);
      expect(strip).toHaveAttribute("role", "alert");
      expect(strip).toHaveAttribute("aria-live", "assertive");
    });

    it("severity=warning uses role=status and aria-live=polite on the strip", () => {
      const { container } = render(
        <NoticeBanner visible={true} severity="warning">
          Message
        </NoticeBanner>,
      );
      const strip = getStrip(container);
      expect(strip).toHaveAttribute("role", "status");
      expect(strip).toHaveAttribute("aria-live", "polite");
    });
  });

  describe("align prop", () => {
    it("align defaults to start (justify-start on strip)", () => {
      const { container } = render(
        <NoticeBanner visible={true} severity="error">
          Message
        </NoticeBanner>,
      );
      const strip = getStrip(container);
      expect(strip.className).toContain("justify-start");
    });

    it("align=center applies justify-center on strip", () => {
      const { container } = render(
        <NoticeBanner visible={true} severity="error" align="center">
          Message
        </NoticeBanner>,
      );
      const strip = getStrip(container);
      expect(strip.className).toContain("justify-center");
    });
  });

  describe("children rendering", () => {
    it("renders plain string children", () => {
      render(
        <NoticeBanner visible={true} severity="error">
          Something went wrong
        </NoticeBanner>,
      );
      expect(screen.getByText("Something went wrong")).toBeInTheDocument();
    });
  });

  describe("message styling", () => {
    it("message <p> always uses text-xs (flat size, no override possible)", () => {
      const { container } = render(
        <NoticeBanner visible={true} severity="error">
          Message
        </NoticeBanner>,
      );
      const strip = getStrip(container);
      const p = strip.querySelector("p") as HTMLElement;
      expect(p.className).toContain("text-xs");
      expect(p.className).not.toContain("text-2xs");
    });
  });
});

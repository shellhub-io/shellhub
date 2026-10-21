import { describe, it, expect } from "vitest";
import { render } from "@testing-library/react";
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
});

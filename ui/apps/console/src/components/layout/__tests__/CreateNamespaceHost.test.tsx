import { describe, it, expect, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { createTestWrapper } from "@/tests/wrapper";
import { setEdition } from "@/tests/edition";
import { useCreateNamespaceStore } from "@/stores/createNamespaceStore";
import CreateNamespaceHost from "../CreateNamespaceHost";

function renderHost() {
  useCreateNamespaceStore.setState({ open: true });
  return render(<CreateNamespaceHost />, {
    wrapper: createTestWrapper({ initialEntries: ["/"] }),
  });
}

describe("CreateNamespaceHost", () => {
  beforeEach(() => {
    useCreateNamespaceStore.setState({ open: false });
  });

  it("offers the upgrade on community, where namespaces come from the CLI", () => {
    setEdition("community");
    renderHost();

    expect(
      screen.queryByRole("heading", { name: "Create namespace" }),
    ).not.toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: /see pricing/i }),
    ).toBeInTheDocument();
  });

  it.each(["enterprise", "cloud"] as const)(
    "opens the create form on %s",
    (edition) => {
      setEdition(edition);
      renderHost();

      expect(
        screen.getByRole("heading", { name: "Create namespace" }),
      ).toBeInTheDocument();
    },
  );
});

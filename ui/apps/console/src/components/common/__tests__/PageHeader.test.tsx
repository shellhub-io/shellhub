import { describe, it, expect } from "vitest";
import { render } from "@testing-library/react";

import PageHeader from "@/components/common/PageHeader";

describe("PageHeader", () => {
  it("leaves the overline out when there is none", () => {
    const { container } = render(
      <PageHeader icon={<svg />} title="All Devices" />,
    );
    expect(container).toHaveTextContent(/^All Devices$/);
  });
});

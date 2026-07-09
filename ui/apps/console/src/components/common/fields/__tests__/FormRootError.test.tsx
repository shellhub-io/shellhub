import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import FormRootError from "@/components/common/fields/FormRootError";

describe("FormRootError", () => {
  it("renders nothing when there is no message", () => {
    const { container } = render(<FormRootError />);
    expect(container).toBeEmptyDOMElement();
  });

  it("renders the message as an alert", () => {
    render(<FormRootError message="Something went wrong" />);

    const alert = screen.getByRole("alert");
    expect(alert).toHaveTextContent("Something went wrong");
  });
});

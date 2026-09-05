import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { UserIcon } from "@heroicons/react/24/outline";

import RadioGroupField from "@/components/common/fields/RadioGroupField";
import RadioCard from "@/components/common/fields/RadioCard";

describe("RadioCard", () => {
  it("exposes each option under its own title, not the title plus description", () => {
    render(
      <RadioGroupField label="Add this key for" value="self" onChange={vi.fn()}>
        <RadioCard
          value="self"
          icon={<UserIcon />}
          label="Myself"
          description="The key becomes your own identity."
        />
        <RadioCard
          value="service-account"
          icon={<UserIcon />}
          label="A new service account"
          description="A non-human identity for an automated system, separate from you."
        />
      </RadioGroupField>,
    );

    expect(
      screen.getByRole("radio", { name: "A new service account" }),
    ).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: "Myself" })).toBeInTheDocument();
  });

  it("selects the option a click on its radio targets", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();

    render(
      <RadioGroupField label="Add this key for" value="self" onChange={onChange}>
        <RadioCard
          value="self"
          icon={<UserIcon />}
          label="Myself"
          description="The key becomes your own identity."
        />
        <RadioCard
          value="service-account"
          icon={<UserIcon />}
          label="A new service account"
          description="A non-human identity for an automated system, separate from you."
        />
      </RadioGroupField>,
    );

    await user.click(screen.getByRole("radio", { name: "A new service account" }));

    expect(onChange).toHaveBeenCalledWith("service-account");
  });
});

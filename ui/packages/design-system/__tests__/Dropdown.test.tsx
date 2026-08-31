import { useState } from "react";
import { describe, it, expect, vi } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Dropdown } from "../primitives/Dropdown";

function MenuDropdown({
  onEdit,
  onDelete,
}: {
  onEdit?: () => void;
  onDelete?: () => void;
}) {
  return (
    <Dropdown>
      <Dropdown.Trigger>
        <button type="button">Actions</button>
      </Dropdown.Trigger>
      <Dropdown.Panel>
        <Dropdown.Item onSelect={onEdit} label="Edit">
          Edit
        </Dropdown.Item>
        <Dropdown.Item onSelect={onDelete} variant="danger" label="Delete">
          Delete
        </Dropdown.Item>
      </Dropdown.Panel>
    </Dropdown>
  );
}

function MenuWithDisabled() {
  return (
    <Dropdown>
      <Dropdown.Trigger>
        <button type="button">Actions</button>
      </Dropdown.Trigger>
      <Dropdown.Panel>
        <Dropdown.Item label="Edit">Edit</Dropdown.Item>
        <Dropdown.Item label="Copy" disabled>
          Copy
        </Dropdown.Item>
        <Dropdown.Item label="Delete">Delete</Dropdown.Item>
      </Dropdown.Panel>
    </Dropdown>
  );
}

function ComboboxDropdown({ onSelect }: { onSelect?: (v: string) => void }) {
  return (
    <Dropdown mode="combobox">
      <Dropdown.Input placeholder="Search..." aria-label="Search items" />
      <Dropdown.Panel>
        <Dropdown.Item label="Apple" onSelect={() => onSelect?.("apple")}>
          Apple
        </Dropdown.Item>
        <Dropdown.Item label="Banana" onSelect={() => onSelect?.("banana")}>
          Banana
        </Dropdown.Item>
        <Dropdown.Item label="Cherry" onSelect={() => onSelect?.("cherry")}>
          Cherry
        </Dropdown.Item>
      </Dropdown.Panel>
    </Dropdown>
  );
}

function AnchoredCombobox({ onSelect }: { onSelect?: (v: string) => void }) {
  return (
    <Dropdown mode="combobox">
      <Dropdown.Anchor className="border">
        <Dropdown.Input placeholder="Search..." aria-label="Search items" />
      </Dropdown.Anchor>
      <Dropdown.Panel>
        <Dropdown.Item label="Apple" onSelect={() => onSelect?.("apple")}>
          Apple
        </Dropdown.Item>
        <Dropdown.Item label="Banana" onSelect={() => onSelect?.("banana")}>
          Banana
        </Dropdown.Item>
      </Dropdown.Panel>
    </Dropdown>
  );
}

function MultiSelectCombobox({ onSelect }: { onSelect?: (v: string) => void }) {
  const [selected, setSelected] = useState<string[]>([]);
  const [search, setSearch] = useState("");
  const all = ["Apple", "Banana", "Cherry"];
  const filtered = all.filter(
    (i) =>
      i.toLowerCase().includes(search.toLowerCase()) && !selected.includes(i),
  );

  return (
    <Dropdown mode="combobox">
      <Dropdown.Anchor role="presentation" className="flex flex-wrap">
        {selected.map((s) => (
          <span key={s} data-testid={`chip-${s}`}>
            {s}
            <button
              type="button"
              onClick={() => setSelected((prev) => prev.filter((p) => p !== s))}
            >
              ×
            </button>
          </span>
        ))}
        <Dropdown.Input
          id="multi-combo"
          value={search}
          onChange={(val) => setSearch(val)}
          placeholder="Search..."
          aria-label="Multi select"
        />
      </Dropdown.Anchor>
      <Dropdown.Panel>
        {filtered.map((item) => (
          <Dropdown.Item
            key={item}
            label={item}
            closeOnSelect={false}
            onSelect={() => {
              setSelected((prev) => [...prev, item]);
              setSearch("");
              onSelect?.(item);
            }}
          >
            {item}
          </Dropdown.Item>
        ))}
      </Dropdown.Panel>
    </Dropdown>
  );
}

function ContentDropdown() {
  return (
    <Dropdown mode="content">
      <Dropdown.Trigger>
        <button type="button">Open panel</button>
      </Dropdown.Trigger>
      <Dropdown.Panel aria-label="Info panel">
        <p>Custom content here</p>
      </Dropdown.Panel>
    </Dropdown>
  );
}

describe("Dropdown", () => {
  describe("menu mode", () => {
    it("opens on trigger click and shows items", async () => {
      const user = userEvent.setup();
      render(<MenuDropdown />);

      expect(screen.queryByRole("menu")).not.toBeInTheDocument();

      await user.click(screen.getByRole("button", { name: "Actions" }));

      expect(screen.getByRole("menu")).toBeInTheDocument();
      expect(screen.getAllByRole("menuitem")).toHaveLength(2);
    });

    it("opens on Enter key when trigger is focused", async () => {
      const user = userEvent.setup();
      render(<MenuDropdown />);

      screen.getByRole("button", { name: "Actions" }).focus();
      await user.keyboard("{Enter}");

      expect(screen.getByRole("menu")).toBeInTheDocument();
    });

    it("closes on second trigger click", async () => {
      const user = userEvent.setup();
      render(<MenuDropdown />);

      const trigger = screen.getByRole("button", { name: "Actions" });
      await user.click(trigger);
      expect(screen.getByRole("menu")).toBeInTheDocument();

      await user.click(trigger);
      expect(screen.queryByRole("menu")).not.toBeInTheDocument();
    });

    it("closes on Escape", async () => {
      const user = userEvent.setup();
      render(<MenuDropdown />);

      await user.click(screen.getByRole("button", { name: "Actions" }));
      expect(screen.getByRole("menu")).toBeInTheDocument();

      await user.keyboard("{Escape}");
      expect(screen.queryByRole("menu")).not.toBeInTheDocument();
    });

    it("fires onSelect and closes when an item is clicked", async () => {
      const user = userEvent.setup();
      const onEdit = vi.fn();
      render(<MenuDropdown onEdit={onEdit} />);

      await user.click(screen.getByRole("button", { name: "Actions" }));
      await user.click(screen.getByRole("menuitem", { name: "Edit" }));

      expect(onEdit).toHaveBeenCalledOnce();
      expect(screen.queryByRole("menu")).not.toBeInTheDocument();
    });

    it("navigates items with ArrowDown/ArrowUp", async () => {
      const user = userEvent.setup();
      render(<MenuDropdown />);

      await user.click(screen.getByRole("button", { name: "Actions" }));

      await user.keyboard("{ArrowDown}");
      expect(screen.getByRole("menuitem", { name: "Edit" })).toHaveFocus();

      await user.keyboard("{ArrowDown}");
      expect(screen.getByRole("menuitem", { name: "Delete" })).toHaveFocus();

      await user.keyboard("{ArrowUp}");
      expect(screen.getByRole("menuitem", { name: "Edit" })).toHaveFocus();
    });

    it("selects item with Enter", async () => {
      const user = userEvent.setup();
      const onEdit = vi.fn();
      render(<MenuDropdown onEdit={onEdit} />);

      await user.click(screen.getByRole("button", { name: "Actions" }));
      await user.keyboard("{ArrowDown}");
      await user.keyboard("{Enter}");

      expect(onEdit).toHaveBeenCalledOnce();
    });

    it("skips disabled items during keyboard navigation", async () => {
      const user = userEvent.setup();
      render(<MenuWithDisabled />);

      await user.click(screen.getByRole("button", { name: "Actions" }));

      await user.keyboard("{ArrowDown}");
      expect(screen.getByRole("menuitem", { name: "Edit" })).toHaveFocus();

      await user.keyboard("{ArrowDown}");
      expect(screen.getByRole("menuitem", { name: "Delete" })).toHaveFocus();
    });

    it("renders disabled item with aria-disabled", async () => {
      const user = userEvent.setup();
      render(<MenuWithDisabled />);

      await user.click(screen.getByRole("button", { name: "Actions" }));

      expect(screen.getByRole("menuitem", { name: "Copy" })).toHaveAttribute(
        "aria-disabled",
        "true",
      );
    });

    it("sets aria-haspopup and aria-expanded on the trigger", async () => {
      const user = userEvent.setup();
      render(<MenuDropdown />);

      const trigger = screen.getByRole("button", { name: "Actions" });
      expect(trigger).toHaveAttribute("aria-haspopup", "menu");
      expect(trigger).toHaveAttribute("aria-expanded", "false");

      await user.click(trigger);
      expect(trigger).toHaveAttribute("aria-expanded", "true");
    });

    it("returns focus to trigger on close", async () => {
      const user = userEvent.setup();
      render(<MenuDropdown />);

      const trigger = screen.getByRole("button", { name: "Actions" });
      await user.click(trigger);
      await user.keyboard("{Escape}");

      expect(trigger).toHaveFocus();
    });
  });

  describe("combobox mode", () => {
    it("opens panel on input focus", async () => {
      const user = userEvent.setup();
      render(<ComboboxDropdown />);

      expect(screen.queryByRole("listbox")).not.toBeInTheDocument();

      await user.click(screen.getByRole("combobox", { name: "Search items" }));

      expect(screen.getByRole("listbox")).toBeInTheDocument();
      expect(screen.getAllByRole("option")).toHaveLength(3);
    });

    it("keeps focus in the input while navigating with arrows", async () => {
      const user = userEvent.setup();
      render(<ComboboxDropdown />);

      const input = screen.getByRole("combobox", { name: "Search items" });
      await user.click(input);

      await user.keyboard("{ArrowDown}");
      expect(input).toHaveFocus();
      expect(screen.getByRole("option", { name: "Apple" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });

    it("highlights first item when typing", async () => {
      const user = userEvent.setup();
      render(<ComboboxDropdown />);

      const input = screen.getByRole("combobox", { name: "Search items" });
      await user.click(input);
      await user.type(input, "a");

      expect(screen.getByRole("option", { name: "Apple" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });

    it("selects item with Enter on the highlighted option", async () => {
      const user = userEvent.setup();
      const onSelect = vi.fn();
      render(<ComboboxDropdown onSelect={onSelect} />);

      await user.click(screen.getByRole("combobox", { name: "Search items" }));
      await user.keyboard("{ArrowDown}");
      await user.keyboard("{Enter}");

      expect(onSelect).toHaveBeenCalledWith("apple");
    });

    it("closes on Escape", async () => {
      const user = userEvent.setup();
      render(<ComboboxDropdown />);

      await user.click(screen.getByRole("combobox", { name: "Search items" }));
      expect(screen.getByRole("listbox")).toBeInTheDocument();

      await user.keyboard("{Escape}");
      expect(screen.queryByRole("listbox")).not.toBeInTheDocument();
    });

    it("sets aria-expanded on the input", async () => {
      const user = userEvent.setup();
      render(<ComboboxDropdown />);

      const input = screen.getByRole("combobox", { name: "Search items" });
      expect(input).toHaveAttribute("aria-expanded", "false");

      await user.click(input);
      expect(input).toHaveAttribute("aria-expanded", "true");
    });

    it("arrow navigation works with Anchor wrapper", async () => {
      const user = userEvent.setup();
      const onSelect = vi.fn();
      render(<AnchoredCombobox onSelect={onSelect} />);

      const input = screen.getByRole("combobox", { name: "Search items" });
      await user.click(input);
      expect(screen.getByRole("listbox")).toBeInTheDocument();

      await user.keyboard("{ArrowDown}");
      expect(input).toHaveFocus();
      expect(screen.getByRole("option", { name: "Apple" })).toHaveAttribute(
        "aria-selected",
        "true",
      );

      await user.keyboard("{ArrowDown}");
      expect(screen.getByRole("option", { name: "Banana" })).toHaveAttribute(
        "aria-selected",
        "true",
      );

      await user.keyboard("{Enter}");
      expect(onSelect).toHaveBeenCalledWith("banana");
    });

    it("multi-select with closeOnSelect=false keeps panel open and navigable", async () => {
      const user = userEvent.setup();
      const onSelect = vi.fn();
      render(<MultiSelectCombobox onSelect={onSelect} />);

      const input = screen.getByRole("combobox", { name: "Multi select" });
      await user.click(input);

      expect(screen.getAllByRole("option")).toHaveLength(3);

      await user.keyboard("{ArrowDown}");
      expect(screen.getByRole("option", { name: "Apple" })).toHaveAttribute(
        "aria-selected",
        "true",
      );

      await user.keyboard("{Enter}");
      expect(onSelect).toHaveBeenCalledWith("Apple");
      expect(screen.getByRole("listbox")).toBeInTheDocument();
      expect(screen.getAllByRole("option")).toHaveLength(2);
    });
  });

  describe("content mode", () => {
    it("opens on trigger click and shows arbitrary content", async () => {
      const user = userEvent.setup();
      render(<ContentDropdown />);

      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();

      await user.click(screen.getByRole("button", { name: "Open panel" }));

      const dialog = screen.getByRole("dialog", { name: "Info panel" });
      expect(dialog).toBeInTheDocument();
      expect(
        within(dialog).getByText("Custom content here"),
      ).toBeInTheDocument();
    });

    it("closes on Escape", async () => {
      const user = userEvent.setup();
      render(<ContentDropdown />);

      await user.click(screen.getByRole("button", { name: "Open panel" }));
      expect(screen.getByRole("dialog")).toBeInTheDocument();

      await user.keyboard("{Escape}");
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });

    it("sets aria-haspopup='dialog' on the trigger", () => {
      render(<ContentDropdown />);

      expect(
        screen.getByRole("button", { name: "Open panel" }),
      ).toHaveAttribute("aria-haspopup", "dialog");
    });

    it("returns focus to trigger on close", async () => {
      const user = userEvent.setup();
      render(<ContentDropdown />);

      const trigger = screen.getByRole("button", { name: "Open panel" });
      await user.click(trigger);
      await user.keyboard("{Escape}");

      expect(trigger).toHaveFocus();
    });
  });

  describe("controlled open state", () => {
    it("respects external open/onOpenChange", async () => {
      const user = userEvent.setup();
      const onOpenChange = vi.fn();

      render(
        <Dropdown open={true} onOpenChange={onOpenChange} mode="content">
          <Dropdown.Trigger>
            <button type="button">Trigger</button>
          </Dropdown.Trigger>
          <Dropdown.Panel aria-label="Controlled">
            <p>Content</p>
          </Dropdown.Panel>
        </Dropdown>,
      );

      expect(screen.getByRole("dialog")).toBeInTheDocument();

      await user.keyboard("{Escape}");
      expect(onOpenChange).toHaveBeenCalledWith(
        false,
        expect.anything(),
        expect.anything(),
      );
    });
  });
});

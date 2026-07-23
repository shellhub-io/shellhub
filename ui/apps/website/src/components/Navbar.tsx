import { useState } from "react";
import { Link } from "react-router-dom";
import { cn } from "@shellhub/design-system/cn";
import { Dropdown } from "@shellhub/design-system/primitives";
import {
  ArrowTopRightOnSquareIcon,
  ChevronDownIcon,
} from "@heroicons/react/24/outline";
import { navEntries, type DropdownItem } from "./navData";

interface DropdownGroupProps {
  label: string;
  items: DropdownItem[];
}

function ItemLink({
  item,
  className,
  onClick,
}: {
  item: DropdownItem;
  className?: string;
  onClick?: () => void;
}) {
  const classes = cn(
    "group flex items-center gap-3 rounded-xl px-2.5 py-2.5 transition-colors hover:bg-hover-medium",
    className,
  );

  const content = (
    <>
      <div className="mt-0.5 flex size-8 items-center justify-center rounded-lg border border-white/[.07] bg-white/5 transition-colors group-hover:border-white/15">
        {item.icon}
      </div>
      <div>
        <span className="text-[13px] font-medium text-text-primary">
          {item.label}
        </span>
        <p className="text-[11px] text-text-muted">{item.desc}</p>
      </div>
    </>
  );

  return item.href.startsWith("/") ? (
    <Link to={item.href} className={classes} onClick={onClick}>
      {content}
    </Link>
  ) : (
    <a
      href={item.href}
      className={classes}
      target="_blank"
      rel="noopener noreferrer"
      onClick={onClick}
    >
      {content}
      <ArrowTopRightOnSquareIcon className="ml-auto size-4 text-text-muted group-hover:text-text-primary" />
    </a>
  );
}

function DesktopDropdown({ label, items }: DropdownGroupProps) {
  const [open, setOpen] = useState(false);

  return (
    <Dropdown mode="content" placement="bottom-start" open={open} onOpenChange={setOpen}>
      <Dropdown.Trigger>
        <button
          type="button"
          className={cn(
            "flex items-center gap-1 rounded-lg px-3 py-1.5 text-[13px] font-medium transition-colors",
            open ? "bg-hover-strong text-text-primary" : "text-text-secondary",
          )}
        >
          {label}
          <ChevronDownIcon
            className={cn("size-3 transition-transform", open && "rotate-180")}
            aria-hidden="true"
          />
        </button>
      </Dropdown.Trigger>

      <Dropdown.Panel className="w-80 p-1.5">
        {items.map((item) => (
          <ItemLink
            key={item.label}
            item={item}
            onClick={() => setOpen(false)}
          />
        ))}
      </Dropdown.Panel>
    </Dropdown>
  );
}

export function MobileDropdown({
  label,
  items,
  onNavigate,
}: DropdownGroupProps & { onNavigate?: () => void }) {
  const [open, setOpen] = useState(false);

  return (
    <div>
      <button
        type="button"
        aria-expanded={open}
        onClick={() => setOpen(!open)}
        className="flex w-full items-center justify-between rounded-lg px-3 py-2 text-[13px] font-medium text-text-secondary"
      >
        {label}
        <ChevronDownIcon
          className={cn("size-3 transition-transform", open && "rotate-180")}
          aria-hidden="true"
        />
      </button>

      {open && (
        <div className="mb-1 ml-2 mt-1 space-y-0.5 border-l border-border pl-2">
          {items.map((item) => (
            <ItemLink
              key={item.label}
              item={item}
              className="px-3 py-2"
              onClick={onNavigate}
            />
          ))}
        </div>
      )}
    </div>
  );
}

export function Navbar() {
  return (
    <nav
      data-testid="desktop-nav"
      className="hidden items-center gap-0.5 lg:flex"
    >
      {navEntries.map((entry) =>
        entry.kind === "link" ? (
          <Link
            key={entry.label}
            to={entry.href}
            className="rounded-lg px-3 py-1.5 text-[13px] font-medium text-text-secondary transition-colors hover:text-text-primary"
          >
            {entry.label}
          </Link>
        ) : (
          <DesktopDropdown
            key={entry.label}
            label={entry.label}
            items={entry.items}
          />
        ),
      )}
    </nav>
  );
}

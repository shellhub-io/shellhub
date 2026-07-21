import { useState, useEffect } from "react";
import { Link, useLocation } from "react-router-dom";
import { cn } from "@shellhub/design-system/cn";
import { ShellHubLogo } from "@shellhub/design-system/primitives";
import { ActionButton } from "@/components/marketing";
import {
  ChevronDownIcon,
  Bars3Icon,
  XMarkIcon,
} from "@heroicons/react/24/outline";
import { C } from "./constants";
import { loginUrl, signupUrl } from "@/links";
import {
  productCols,
  solutionsCols,
  resourcesCols,
  simpleLinks,
} from "./navData";
import type { MenuItem, MenuSection } from "./navData";

// ─────────────────────────────────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────────────────────────────────

type ActiveMenu = "product" | "solutions" | "resources" | null;

// ─────────────────────────────────────────────────────────────────────────────
// Sub-components
// ─────────────────────────────────────────────────────────────────────────────

function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <p
      className="text-[9.5px] font-bold uppercase tracking-[0.13em] mb-3 px-2.5"
      style={{ color: C.textMuted }}
    >
      {children}
    </p>
  );
}

function MegaMenuItem({ item }: { item: MenuItem }) {
  const inner = (
    <>
      <div className="w-8 h-8 rounded-lg flex items-center justify-center shrink-0 mt-0.5 transition-all duration-150 bg-white/5 border border-white/[0.07] group-hover:border-white/15">
        {item.icon}
      </div>
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <span
            className="text-[13px] font-medium leading-snug transition-colors duration-100"
            style={{ color: C.text }}
          >
            {item.label}
          </span>
          {item.badge && (
            <span
              className="text-[9px] font-bold px-1.5 py-0.5 rounded-full"
              style={{
                background:
                  item.badge === "Free" ? `${C.greenDim}` : `${C.primaryDim}`,
                color: item.badge === "Free" ? C.green : C.primary,
              }}
            >
              {item.badge}
            </span>
          )}
        </div>
        <p
          className="text-[11px] leading-snug mt-0.5 transition-colors duration-100"
          style={{ color: C.textMuted }}
        >
          {item.desc}
        </p>
      </div>
    </>
  );

  const sharedClass =
    "group flex items-start gap-3 px-2.5 py-2.5 rounded-xl transition-colors duration-100";
  const hoverHandlers = {
    onMouseEnter: (e: React.MouseEvent<HTMLElement>) =>
      (e.currentTarget.style.background = "rgba(255,255,255,0.05)"),
    onMouseLeave: (e: React.MouseEvent<HTMLElement>) =>
      (e.currentTarget.style.background = ""),
  };

  const isInternal = item.href.startsWith("/");
  const isExternal = item.href.startsWith("http");

  if (isInternal) {
    return (
      <Link to={item.href} className={sharedClass} {...hoverHandlers}>
        {inner}
      </Link>
    );
  }
  if (isExternal) {
    return (
      <a
        href={item.href}
        className={sharedClass}
        target="_blank"
        rel="noopener noreferrer"
        {...hoverHandlers}
      >
        {inner}
      </a>
    );
  }
  return (
    <span
      className={cn(sharedClass, "cursor-default")}
      aria-disabled="true"
      {...hoverHandlers}
    >
      {inner}
    </span>
  );
}

// ─── Full-width Mega Menu Panel ───────────────────────────────────────────────

function FullWidthMenu({
  id,
  active,
  cols,
}: {
  id: ActiveMenu;
  active: ActiveMenu;
  cols: MenuSection[];
}) {
  const isOpen = active === id;
  const colCount = cols.length;

  return (
    <div
      className={cn(
        "absolute top-full left-0 right-0 transition-all duration-[170ms] ease-out z-50",
        isOpen
          ? "opacity-100 translate-y-0 pointer-events-auto"
          : "opacity-0 -translate-y-1 pointer-events-none",
      )}
      style={{
        background: C.bg,
        borderBottom: `1px solid ${C.border}`,
      }}
    >
      <div className="max-w-7xl mx-auto px-8 py-6">
        <div
          className="grid gap-x-8 gap-y-0"
          style={{ gridTemplateColumns: `repeat(${colCount}, 1fr)` }}
        >
          {cols.map((section) => (
            <div key={section.title}>
              <SectionLabel>{section.title}</SectionLabel>
              <div className="space-y-0.5">
                {section.items.map((item) => (
                  <MegaMenuItem key={item.label} item={item} />
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

// ─── Mobile Dropdown ──────────────────────────────────────────────────────────

function MobileDropdown({
  label,
  cols,
}: {
  label: string;
  cols: MenuSection[];
}) {
  const [open, setOpen] = useState(false);
  const items = cols.flatMap((s) => s.items);

  return (
    <div>
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className="w-full flex items-center justify-between px-3 py-2 rounded-lg text-[13px] font-medium transition-all"
        style={{ color: C.textSec }}
      >
        {label}
        <ChevronDownIcon
          className={cn(
            "w-3 h-3 transition-transform duration-200",
            open && "rotate-180",
          )}
          aria-hidden="true"
        />
      </button>
      {open && (
        <div
          className="ml-2 mt-1 mb-1 pl-2 space-y-0.5"
          style={{ borderLeft: `1px solid ${C.border}` }}
        >
          {items.map((item) => {
            const itemInner = (
              <>
                <div
                  className="w-6 h-6 rounded-md flex items-center justify-center shrink-0"
                  style={{
                    background: "rgba(255,255,255,0.04)",
                    border: "1px solid rgba(255,255,255,0.07)",
                  }}
                >
                  {item.icon}
                </div>
                <div>
                  <div className="text-[13px]" style={{ color: C.textSec }}>
                    {item.label}
                  </div>
                  <div
                    className="text-[10px] leading-snug"
                    style={{ color: C.textMuted }}
                  >
                    {item.desc}
                  </div>
                </div>
              </>
            );
            const itemClass =
              "flex items-center gap-2.5 px-3 py-2 rounded-lg transition-all";
            const itemHover = {
              onMouseEnter: (e: React.MouseEvent<HTMLElement>) =>
                (e.currentTarget.style.background = "rgba(255,255,255,0.03)"),
              onMouseLeave: (e: React.MouseEvent<HTMLElement>) =>
                (e.currentTarget.style.background = ""),
            };
            const isInternal = item.href.startsWith("/");
            const isExternal = item.href.startsWith("http");

            if (isInternal) {
              return (
                <Link
                  key={item.label}
                  to={item.href}
                  className={itemClass}
                  {...itemHover}
                >
                  {itemInner}
                </Link>
              );
            }
            if (isExternal) {
              return (
                <a
                  key={item.label}
                  href={item.href}
                  className={itemClass}
                  target="_blank"
                  rel="noopener noreferrer"
                  {...itemHover}
                >
                  {itemInner}
                </a>
              );
            }
            return (
              <span
                key={item.label}
                className={cn(itemClass, "cursor-default")}
                aria-disabled="true"
                {...itemHover}
              >
                {itemInner}
              </span>
            );
          })}
        </div>
      )}
    </div>
  );
}

// ─────────────────────────────────────────────────────────────────────────────
// Navbar
// ─────────────────────────────────────────────────────────────────────────────

export function Navbar({
  navSolid,
  mobileMenu,
  setMobileMenu,
}: {
  navSolid: boolean;
  mobileMenu: boolean;
  setMobileMenu: (v: boolean) => void;
}) {
  const [activeMenu, setActiveMenu] = useState<ActiveMenu>(null);
  const { pathname } = useLocation();

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      const nav = document.getElementById("shellhub-nav");
      if (nav && !nav.contains(e.target as Node)) setActiveMenu(null);
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  const closeMenus = () => {
    setActiveMenu(null);
    setMobileMenu(false);
  };

  const navBg = activeMenu ? C.bg : navSolid ? `${C.bg}b8` : "transparent";

  const menuItems: { id: ActiveMenu; label: string }[] = [
    { id: "product", label: "Product" },
    { id: "solutions", label: "Solutions" },
    { id: "resources", label: "Resources" },
  ];

  return (
    <>
      {/* Backdrop */}
      <div
        className={cn(
          "fixed inset-0 top-14 z-40 transition-all duration-200",
          activeMenu
            ? "opacity-100 pointer-events-auto"
            : "opacity-0 pointer-events-none",
        )}
        style={{ background: "rgba(0,0,0,0.3)", backdropFilter: "blur(2px)" }}
        onClick={() => setActiveMenu(null)}
      />

      {/* eslint-disable-next-line jsx-a11y/no-noninteractive-element-interactions -- nav onClick delegates to child links only */}
      <nav
        id="shellhub-nav"
        className="fixed top-0 left-0 right-0 z-50 h-14 transition-all duration-300"
        onClick={(e) => {
          if ((e.target as HTMLElement).closest("a")) closeMenus();
        }}
        style={{
          background: navBg,
          backdropFilter:
            !activeMenu && navSolid ? "blur(24px) saturate(180%)" : "none",
          borderBottom: `1px solid ${navSolid || activeMenu ? C.border : "transparent"}`,
          boxShadow:
            navSolid || activeMenu ? "0 1px 30px rgba(0,0,0,0.3)" : "none",
        }}
      >
        <div className="max-w-7xl mx-auto px-8 h-full flex items-center justify-between">
          {/* Logo */}
          <Link to="/" className="shrink-0">
            <ShellHubLogo className="h-8" />
          </Link>

          {/* Mobile toggle */}
          <button
            type="button"
            data-testid="mobile-nav-toggle"
            className="lg:hidden p-2 transition-colors"
            style={{ color: C.textSec }}
            onClick={() => setMobileMenu(!mobileMenu)}
          >
            {mobileMenu ? (
              <XMarkIcon className="w-5 h-5" aria-hidden="true" />
            ) : (
              <Bars3Icon className="w-5 h-5" aria-hidden="true" />
            )}
          </button>

          {/* Desktop nav */}
          <div
            data-testid="desktop-nav"
            className="hidden lg:flex items-center gap-0.5"
          >
            {menuItems.map(({ id, label }) => (
              <button
                type="button"
                key={id}
                aria-expanded={activeMenu === id}
                onClick={() => setActiveMenu(activeMenu === id ? null : id)}
                className="flex items-center gap-1 px-3 py-1.5 rounded-lg text-[13px] font-medium transition-all duration-150"
                style={{
                  color: activeMenu === id ? C.text : C.textSec,
                  background: activeMenu === id ? "rgba(255,255,255,0.06)" : "",
                }}
              >
                {label}
                <ChevronDownIcon
                  className={cn(
                    "w-3 h-3 transition-transform duration-200",
                    activeMenu === id && "rotate-180",
                  )}
                  aria-hidden="true"
                />
              </button>
            ))}

            {simpleLinks.map((l) => (
              <Link
                key={l.label}
                to={l.href}
                className="px-3 py-1.5 rounded-lg text-[13px] font-medium transition-all"
                style={{ color: C.textSec }}
                onMouseEnter={(e) => (e.currentTarget.style.color = C.text)}
                onMouseLeave={(e) => (e.currentTarget.style.color = C.textSec)}
              >
                {l.label}
              </Link>
            ))}
          </div>

          <div className="hidden lg:flex items-center gap-2">
            <ActionButton
              action={{ label: "Log In", href: loginUrl, external: true }}
              variant="outline"
              size="md"
            />
            <ActionButton
              action={{
                label: "Sign Up Free",
                href: signupUrl,
                external: true,
              }}
              size="md"
              glow={false}
              iconRight={null}
            />
          </div>
        </div>

        {/* Full-width mega menus */}
        <FullWidthMenu id="product" active={activeMenu} cols={productCols} />
        <FullWidthMenu
          id="solutions"
          active={activeMenu}
          cols={solutionsCols}
        />
        <FullWidthMenu
          id="resources"
          active={activeMenu}
          cols={resourcesCols}
        />

        {/* Mobile nav */}
        <div
          data-testid="mobile-nav"
          className={cn(
            mobileMenu ? "flex" : "hidden",
            "lg:hidden absolute top-14 left-0 right-0 flex-col gap-0.5 items-stretch p-3 border-b shadow-xl",
          )}
          style={{
            background: `${C.surface}f8`,
            backdropFilter: "blur(20px)",
            borderColor: C.border,
          }}
        >
          <MobileDropdown
            key={mobileMenu ? pathname : "closed"}
            label="Product"
            cols={productCols}
          />
          <MobileDropdown
            key={mobileMenu ? pathname + "-sol" : "closed-sol"}
            label="Solutions"
            cols={solutionsCols}
          />
          <MobileDropdown
            key={mobileMenu ? pathname + "-res" : "closed-res"}
            label="Resources"
            cols={resourcesCols}
          />
          {simpleLinks.map((l) => (
            <Link
              key={l.label}
              to={l.href}
              className="px-3 py-2 rounded-lg text-[13px] font-medium transition-all"
              style={{ color: C.textSec }}
            >
              {l.label}
            </Link>
          ))}
          <div
            className="pt-2 mt-1 flex flex-col gap-2"
            style={{ borderTop: `1px solid ${C.border}` }}
          >
            <ActionButton
              action={{ label: "Log In", href: loginUrl, external: true }}
              variant="outline"
              size="md"
              fullWidth
            />
            <ActionButton
              action={{
                label: "Sign Up Free",
                href: signupUrl,
                external: true,
              }}
              size="md"
              glow={false}
              fullWidth
              iconRight={null}
            />
          </div>
        </div>
      </nav>
    </>
  );
}

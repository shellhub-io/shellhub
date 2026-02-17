import { useState, useEffect } from "react";
import { C } from "./constants";

// ─────────────────────────────────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────────────────────────────────

interface MenuItem {
  label: string;
  href: string;
  desc: string;
  icon: React.ReactNode;
  badge?: string;
}

interface MenuSection {
  title: string;
  items: MenuItem[];
}

type ActiveMenu = "product" | "solutions" | "resources" | null;

// ─────────────────────────────────────────────────────────────────────────────
// Menu Data
// ─────────────────────────────────────────────────────────────────────────────

const productCols: MenuSection[] = [
  {
    title: "Access & Security",
    items: [
      {
        label: "SSH Gateway",
        href: "/v2/features",
        desc: "Centralized SSH access via reverse tunnel — no VPN",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.primary}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <rect x="2" y="3" width="20" height="14" rx="2" />
            <path d="M8 21h8M12 17v4" />
            <path d="m9 9 2 2 4-4" />
          </svg>
        ),
      },
      {
        label: "MFA & RBAC",
        href: "/v2/features",
        desc: "Multi-factor auth and role-based access control",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.blue}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
            <path d="m9 12 2 2 4-4" />
          </svg>
        ),
      },
      {
        label: "Firewall Rules",
        href: "/v2/features",
        desc: "IP filtering, port restrictions and access policies",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.red}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M4 15s1-1 4-1 5 2 8 2 4-1 4-1V3s-1 1-4 1-5-2-8-2-4 1-4 1z" />
            <line x1="4" y1="22" x2="4" y2="15" />
          </svg>
        ),
      },
      {
        label: "Public Keys",
        href: "/v2/features",
        desc: "Centralized SSH key management across your fleet",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.cyan}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <circle cx="7.5" cy="15.5" r="5.5" />
            <path d="m21 2-9.6 9.6M15.5 7.5l3 3L22 7l-3-3" />
          </svg>
        ),
      },
    ],
  },
  {
    title: "Device Management",
    items: [
      {
        label: "Fleet Manager",
        href: "/v2/features",
        desc: "Monitor and manage thousands of devices at once",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.green}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <rect x="2" y="7" width="20" height="14" rx="2" />
            <path d="M16 7V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v2" />
            <line x1="12" y1="12" x2="12" y2="16" />
            <line x1="10" y1="14" x2="14" y2="14" />
          </svg>
        ),
      },
      {
        label: "Namespaces",
        href: "/v2/features",
        desc: "Isolate teams, projects and devices logically",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.primary}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M2 9V5a2 2 0 0 1 2-2h4" />
            <path d="M2 15v4a2 2 0 0 0 2 2h4" />
            <path d="M22 9V5a2 2 0 0 0-2-2h-4" />
            <path d="M22 15v4a2 2 0 0 1-2 2h-4" />
            <rect x="8" y="8" width="8" height="8" rx="1" />
          </svg>
        ),
      },
      {
        label: "Remote Exec",
        href: "/v2/features",
        desc: "Run commands across device groups simultaneously",
        badge: "New",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.yellow}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <polyline points="4 17 10 11 4 5" />
            <line x1="12" y1="19" x2="20" y2="19" />
          </svg>
        ),
      },
      {
        label: "Audit Logs",
        href: "/v2/features",
        desc: "Immutable history of sessions, commands and events",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.cyan}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
            <polyline points="14 2 14 8 20 8" />
            <line x1="16" y1="13" x2="8" y2="13" />
            <line x1="16" y1="17" x2="8" y2="17" />
          </svg>
        ),
      },
    ],
  },
  {
    title: "Observability",
    items: [
      {
        label: "Session Replay",
        href: "/v2/features",
        desc: "Watch and replay recorded SSH sessions",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.blue}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <circle cx="12" cy="12" r="10" />
            <polygon points="10 8 16 12 10 16 10 8" />
          </svg>
        ),
      },
      {
        label: "Getting Started",
        href: "/v2/getting-started",
        desc: "Deploy ShellHub in minutes, self-hosted or cloud",
        badge: "Free",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.green}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M5 12h14M12 5l7 7-7 7" />
          </svg>
        ),
      },
      {
        label: "Changelog",
        href: "#",
        desc: "Release notes and version history",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.primary}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M12 20h9" />
            <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
          </svg>
        ),
      },
    ],
  },
];

const solutionsCols: MenuSection[] = [
  {
    title: "Infrastructure",
    items: [
      {
        label: "IoT & Embedded",
        href: "/v2/use-cases/iot-embedded",
        desc: "Manage fleets of IoT and embedded Linux devices",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.green}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <rect x="4" y="4" width="16" height="16" rx="2" />
            <rect x="9" y="9" width="6" height="6" />
            <line x1="9" y1="1" x2="9" y2="4" />
            <line x1="15" y1="1" x2="15" y2="4" />
            <line x1="9" y1="20" x2="9" y2="23" />
            <line x1="15" y1="20" x2="15" y2="23" />
            <line x1="20" y1="9" x2="23" y2="9" />
            <line x1="20" y1="15" x2="23" y2="15" />
            <line x1="1" y1="9" x2="4" y2="9" />
            <line x1="1" y1="15" x2="4" y2="15" />
          </svg>
        ),
      },
      {
        label: "Edge Computing",
        href: "/v2/use-cases/edge-computing",
        desc: "Reach distributed edge servers across any topology",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.blue}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M22 12h-4l-3 9L9 3l-3 9H2" />
          </svg>
        ),
      },
      {
        label: "Container Management",
        href: "/v2/use-cases/container-management",
        desc: "SSH directly into Docker containers from anywhere",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.cyan}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" />
            <polyline points="3.27 6.96 12 12.01 20.73 6.96" />
            <line x1="12" y1="22.08" x2="12" y2="12" />
          </svg>
        ),
      },
    ],
  },
  {
    title: "Teams",
    items: [
      {
        label: "Remote Support",
        href: "/v2/use-cases/remote-support",
        desc: "Session recording, replay and audit trail",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.yellow}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M15 10l4.553-2.276A1 1 0 0 1 21 8.618v6.764a1 1 0 0 1-1.447.894L15 14v-4z" />
            <rect x="3" y="6" width="12" height="12" rx="2" />
          </svg>
        ),
      },
      {
        label: "DevOps & CI/CD",
        href: "/v2/use-cases/devops-ci-cd",
        desc: "Automate with Ansible, Terraform and CI pipelines",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.primary}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <polyline points="16 18 22 12 16 6" />
            <polyline points="8 6 2 12 8 18" />
            <line x1="14" y1="4" x2="10" y2="20" />
          </svg>
        ),
      },
      {
        label: "Getting Started",
        href: "/v2/getting-started",
        desc: "Deploy ShellHub in minutes — self-hosted or cloud",
        badge: "Free",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.green}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M5 12h14M12 5l7 7-7 7" />
          </svg>
        ),
      },
    ],
  },
];

const resourcesCols: MenuSection[] = [
  {
    title: "Spotlight",
    items: [
      {
        label: "Documentation",
        href: "/v2/docs/",
        desc: "Full reference, guides and API docs",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.blue}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20" />
            <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z" />
          </svg>
        ),
      },
      {
        label: "Quick Start",
        href: "/v2/getting-started",
        desc: "Get a device connected in under 5 minutes",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.green}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M4.5 16.5c-1.5 1.26-2 5-2 5s3.74-.5 5-2c.71-.84.7-2.13-.09-2.91a2.18 2.18 0 0 0-2.91-.09z" />
            <path d="m12 15-3-3a22 22 0 0 1 2-3.95A12.88 12.88 0 0 1 22 2c0 2.72-.78 7.5-6 11a22.35 22.35 0 0 1-4 2z" />
          </svg>
        ),
      },
      {
        label: "Blog",
        href: "#",
        desc: "Tutorials, use cases and product updates",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.yellow}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
            <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
          </svg>
        ),
      },
      {
        label: "Changelog",
        href: "#",
        desc: "See what shipped in every release",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.primary}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M12 20h9" />
            <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
          </svg>
        ),
      },
    ],
  },
  {
    title: "Docs & Community",
    items: [
      {
        label: "API Reference",
        href: "/v2/docs/",
        desc: "REST API endpoints and authentication",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.cyan}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <polyline points="16 18 22 12 16 6" />
            <polyline points="8 6 2 12 8 18" />
          </svg>
        ),
      },
      {
        label: "GitHub",
        href: "https://github.com/shellhub-io/shellhub",
        desc: "Source code, issues and contributions",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.textSec}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M9 19c-5 1.5-5-2.5-7-3m14 6v-3.87a3.37 3.37 0 0 0-.94-2.61c3.14-.35 6.44-1.54 6.44-7A5.44 5.44 0 0 0 20 4.77 5.07 5.07 0 0 0 19.91 1S18.73.65 16 2.48a13.38 13.38 0 0 0-7 0C6.27.65 5.09 1 5.09 1A5.07 5.07 0 0 0 5 4.77a5.44 5.44 0 0 0-1.5 3.78c0 5.42 3.3 6.61 6.44 7A3.37 3.37 0 0 0 9 18.13V22" />
          </svg>
        ),
      },
      {
        label: "Discord",
        href: "#",
        desc: "Chat with the ShellHub community",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.blue}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
          </svg>
        ),
      },
      {
        label: "Forum",
        href: "#",
        desc: "Ask questions and share knowledge",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.primary}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
            <circle cx="9" cy="7" r="4" />
            <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
            <path d="M16 3.13a4 4 0 0 1 0 7.75" />
          </svg>
        ),
      },
    ],
  },
  {
    title: "Self-Hosted",
    items: [
      {
        label: "Deploy Guide",
        href: "/v2/docs/",
        desc: "Docker Compose, Kubernetes, bare metal",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.green}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" />
          </svg>
        ),
      },
      {
        label: "Enterprise",
        href: "/v2/enterprise",
        desc: "SAML, LDAP, audit trails and SLA support",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.yellow}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
            <polyline points="9 22 9 12 15 12 15 22" />
          </svg>
        ),
      },
      {
        label: "Security",
        href: "#",
        desc: "CVEs, responsible disclosure and trust",
        icon: (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke={C.red}
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
          </svg>
        ),
      },
    ],
  },
];

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
  return (
    <a
      href={item.href}
      className="group flex items-start gap-3 px-2.5 py-2.5 rounded-xl transition-colors duration-100"
      style={{ ["--hover-bg" as string]: "rgba(255,255,255,0.05)" }}
      onMouseEnter={(e) =>
        (e.currentTarget.style.background = "rgba(255,255,255,0.05)")
      }
      onMouseLeave={(e) => (e.currentTarget.style.background = "")}
    >
      <div
        className="w-8 h-8 rounded-lg flex items-center justify-center shrink-0 mt-0.5 transition-all duration-150"
        style={{
          background: "rgba(255,255,255,0.04)",
          border: "1px solid rgba(255,255,255,0.07)",
        }}
        onMouseEnter={(e) =>
          ((e.currentTarget as HTMLElement).style.borderColor =
            "rgba(255,255,255,0.12)")
        }
      >
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
    </a>
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
      className={`absolute top-full left-0 right-0 transition-all duration-[170ms] ease-out z-50 ${
        isOpen
          ? "opacity-100 translate-y-0 pointer-events-auto"
          : "opacity-0 -translate-y-1 pointer-events-none"
      }`}
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
        onClick={() => setOpen(!open)}
        className="w-full flex items-center justify-between px-3 py-2 rounded-lg text-[13px] font-medium transition-all"
        style={{ color: C.textSec }}
      >
        {label}
        <svg
          className={`w-3 h-3 transition-transform duration-200 ${open ? "rotate-180" : ""}`}
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={2.5}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="m19.5 8.25-7.5 7.5-7.5-7.5"
          />
        </svg>
      </button>
      {open && (
        <div
          className="ml-2 mt-1 mb-1 pl-2 space-y-0.5"
          style={{ borderLeft: `1px solid ${C.border}` }}
        >
          {items.map((item) => (
            <a
              key={item.label}
              href={item.href}
              className="flex items-center gap-2.5 px-3 py-2 rounded-lg transition-all"
              onMouseEnter={(e) =>
                (e.currentTarget.style.background = "rgba(255,255,255,0.03)")
              }
              onMouseLeave={(e) => (e.currentTarget.style.background = "")}
            >
              <div
                className="w-6 h-6 rounded-md flex items-center justify-center shrink-0"
                style={{
                  background: "rgba(255,255,255,0.04)",
                  border: `1px solid rgba(255,255,255,0.07)`,
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
            </a>
          ))}
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

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      const nav = document.getElementById("shellhub-nav");
      if (nav && !nav.contains(e.target as Node)) setActiveMenu(null);
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  const navBg = activeMenu ? C.bg : navSolid ? `${C.bg}b8` : "transparent";

  const simpleLinks = [
    { label: "Enterprise", href: "/v2/enterprise" },
    { label: "Pricing", href: "/v2/pricing" },
  ];

  const menuItems: { id: ActiveMenu; label: string }[] = [
    { id: "product", label: "Product" },
    { id: "solutions", label: "Solutions" },
    { id: "resources", label: "Resources" },
  ];

  return (
    <>
      {/* Backdrop */}
      <div
        className={`fixed inset-0 top-14 z-40 transition-all duration-200 ${
          activeMenu
            ? "opacity-100 pointer-events-auto"
            : "opacity-0 pointer-events-none"
        }`}
        style={{ background: "rgba(0,0,0,0.3)", backdropFilter: "blur(2px)" }}
        onClick={() => setActiveMenu(null)}
      />

      <nav
        id="shellhub-nav"
        className="fixed top-0 left-0 right-0 z-50 h-14 transition-all duration-300"
        style={{
          background: navBg,
          backdropFilter:
            !activeMenu && navSolid ? "blur(24px) saturate(180%)" : "none",
          borderBottom: `1px solid ${navSolid || activeMenu ? C.border : "transparent"}`,
          boxShadow:
            navSolid || activeMenu ? `0 1px 30px rgba(0,0,0,0.3)` : "none",
        }}
      >
        <div className="max-w-7xl mx-auto px-8 h-full flex items-center justify-between">
          {/* Logo */}
          <a href="/v2/" className="shrink-0">
            <img src="/v2/logo-inverted.png" alt="ShellHub" className="h-8" />
          </a>

          {/* Mobile toggle */}
          <button
            className="lg:hidden p-2 transition-colors"
            style={{ color: C.textSec }}
            onClick={() => setMobileMenu(!mobileMenu)}
          >
            <svg
              className="w-5 h-5"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={2}
            >
              {mobileMenu ? (
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M6 18 18 6M6 6l12 12"
                />
              ) : (
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M3.75 6.75h16.5M3.75 12h16.5M3.75 17.25h16.5"
                />
              )}
            </svg>
          </button>

          {/* Desktop nav */}
          <div className="hidden lg:flex items-center gap-0.5">
            {menuItems.map(({ id, label }) => (
              <button
                key={id}
                onClick={() => setActiveMenu(activeMenu === id ? null : id)}
                className="flex items-center gap-1 px-3 py-1.5 rounded-lg text-[13px] font-medium transition-all duration-150"
                style={{
                  color: activeMenu === id ? C.text : C.textSec,
                  background: activeMenu === id ? "rgba(255,255,255,0.06)" : "",
                }}
              >
                {label}
                <svg
                  className={`w-3 h-3 transition-transform duration-200 ${activeMenu === id ? "rotate-180" : ""}`}
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  strokeWidth={2.5}
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="m19.5 8.25-7.5 7.5-7.5-7.5"
                  />
                </svg>
              </button>
            ))}

            {simpleLinks.map((l) => (
              <a
                key={l.label}
                href={l.href}
                className="px-3 py-1.5 rounded-lg text-[13px] font-medium transition-all"
                style={{ color: C.textSec }}
                onMouseEnter={(e) => (e.currentTarget.style.color = C.text)}
                onMouseLeave={(e) => (e.currentTarget.style.color = C.textSec)}
              >
                {l.label}
              </a>
            ))}
          </div>

          {/* CTA */}
          <div className="hidden lg:flex items-center gap-2">
            <a
              href="#"
              className="px-4 py-2 text-[13px] font-medium rounded-lg transition-all"
              style={{ color: C.textSec, border: `1px solid ${C.border}` }}
              onMouseEnter={(e) => {
                e.currentTarget.style.color = C.text;
                e.currentTarget.style.borderColor = C.borderLight;
                e.currentTarget.style.background = "rgba(255,255,255,0.04)";
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.color = C.textSec;
                e.currentTarget.style.borderColor = C.border;
                e.currentTarget.style.background = "";
              }}
            >
              Log In
            </a>
            <a
              href="#"
              className="px-4 py-2 text-[13px] font-semibold rounded-lg transition-all"
              style={{ background: C.primary, color: "#111214" }}
              onMouseEnter={(e) => (e.currentTarget.style.opacity = "0.88")}
              onMouseLeave={(e) => (e.currentTarget.style.opacity = "1")}
            >
              Sign Up Free
            </a>
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
          className={`${mobileMenu ? "flex" : "hidden"} lg:hidden absolute top-14 left-0 right-0 flex-col gap-0.5 items-stretch p-3 border-b shadow-xl`}
          style={{
            background: `${C.surface}f8`,
            backdropFilter: "blur(20px)",
            borderColor: C.border,
          }}
        >
          <MobileDropdown label="Product" cols={productCols} />
          <MobileDropdown label="Solutions" cols={solutionsCols} />
          <MobileDropdown label="Resources" cols={resourcesCols} />
          {simpleLinks.map((l) => (
            <a
              key={l.label}
              href={l.href}
              className="px-3 py-2 rounded-lg text-[13px] font-medium transition-all"
              style={{ color: C.textSec }}
            >
              {l.label}
            </a>
          ))}
          <div
            className="pt-2 mt-1 flex flex-col gap-2"
            style={{ borderTop: `1px solid ${C.border}` }}
          >
            <a
              href="#"
              className="px-4 py-2 text-center text-[13px] font-medium rounded-lg"
              style={{ color: C.textSec, border: `1px solid ${C.border}` }}
            >
              Log In
            </a>
            <a
              href="#"
              className="px-4 py-2 text-center text-[13px] font-semibold rounded-lg"
              style={{ background: C.primary, color: "#111214" }}
            >
              Sign Up Free
            </a>
          </div>
        </div>
      </nav>
    </>
  );
}

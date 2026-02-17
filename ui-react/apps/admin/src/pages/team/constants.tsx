import { ROLES } from "./helpers";

/* ─── Constants ─── */

const ROLE_STYLES: Record<
  string,
  { bg: string; text: string; border: string }
> = {
  owner: {
    bg: "bg-accent-yellow/10",
    text: "text-accent-yellow",
    border: "border-accent-yellow/20",
  },
  administrator: {
    bg: "bg-primary/10",
    text: "text-primary",
    border: "border-primary/20",
  },
  operator: {
    bg: "bg-accent-green/10",
    text: "text-accent-green",
    border: "border-accent-green/20",
  },
  observer: {
    bg: "bg-hover-medium",
    text: "text-text-muted",
    border: "border-border",
  },
};

const ROLE_META: Record<string, { icon: string; summary: string }> = {
  administrator: {
    icon: "M9 12.75 11.25 15 15 9.75m-3-7.036A11.959 11.959 0 0 1 3.598 6 11.99 11.99 0 0 0 3 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285Z",
    summary: "Full access — manage devices, members, keys, and firewall rules",
  },
  operator: {
    icon: "M11.42 15.17 17.25 21A2.652 2.652 0 0 0 21 17.25l-5.877-5.877M11.42 15.17l2.496-3.03c.317-.384.74-.626 1.208-.766M11.42 15.17l-4.655 5.653a2.548 2.548 0 1 1-3.586-3.586l6.837-5.63m5.108-.233c.55-.164 1.163-.188 1.743-.14a4.5 4.5 0 0 0 4.486-6.336l-3.276 3.277a3.004 3.004 0 0 1-2.25-2.25l3.276-3.276a4.5 4.5 0 0 0-6.336 4.486c.091 1.076-.071 2.264-.904 2.95l-.102.085",
    summary: "Manage devices and tags, connect via SSH, view sessions",
  },
  observer: {
    icon: "M2.036 12.322a1.012 1.012 0 0 1 0-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178ZM15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z",
    summary: "Read-only — connect to devices, view details and sessions",
  },
};

/* ─── Role Badge ─── */

export function RoleBadge({ role }: { role: string }) {
  const style = ROLE_STYLES[role] ?? ROLE_STYLES.observer;
  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 text-2xs font-mono font-semibold rounded border ${style.bg} ${style.text} ${style.border}`}
    >
      {role}
    </span>
  );
}

/* ─── Role Selector ─── */

export function RoleSelector({
  value,
  onChange,
}: {
  value: string;
  onChange: (v: string) => void;
}) {
  return (
    <div className="space-y-1.5">
      {ROLES.map((role) => {
        const meta = ROLE_META[role];
        const selected = value === role;
        return (
          <button
            key={role}
            type="button"
            onClick={() => onChange(role)}
            className={`flex items-center gap-3 w-full px-3 py-2.5 rounded-lg border text-left transition-all ${
              selected
                ? "bg-primary/[0.06] border-primary/30 ring-1 ring-primary/10"
                : "bg-card border-border hover:border-border-light hover:bg-hover-subtle"
            }`}
          >
            <svg
              className={`w-4 h-4 shrink-0 transition-colors ${selected ? "text-primary" : "text-text-muted/50"}`}
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={1.5}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d={meta.icon}
              />
            </svg>
            <div className="min-w-0 flex-1">
              <span
                className={`text-sm font-medium capitalize ${selected ? "text-text-primary" : "text-text-secondary"}`}
              >
                {role}
              </span>
              <p
                className={`text-2xs mt-0.5 leading-relaxed ${selected ? "text-text-secondary" : "text-text-muted"}`}
              >
                {meta.summary}
              </p>
            </div>
          </button>
        );
      })}
    </div>
  );
}

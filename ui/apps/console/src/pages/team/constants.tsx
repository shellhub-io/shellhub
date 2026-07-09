import {
  ExclamationCircleIcon,
  EyeIcon,
  ShieldCheckIcon,
  WrenchScrewdriverIcon,
} from "@heroicons/react/24/outline";
import type { ComponentType, SVGProps } from "react";
import {
  useController,
  type Control,
  type FieldValues,
  type Path,
} from "react-hook-form";
import RadioCard from "@/components/common/fields/RadioCard";
import RadioGroupField from "@/components/common/fields/RadioGroupField";
import { Badge, type BadgeColor } from "@shellhub/design-system/primitives";
import { ROLES, type AssignableRole } from "./helpers";

/** Roles that map directly to a Badge palette color. */
const ROLE_COLOR: Record<string, BadgeColor> = {
  owner: "yellow",
  administrator: "primary",
  operator: "green",
};

/** Fallback inline styles for roles that have no palette equivalent (observer). */
const ROLE_NEUTRAL_STYLE =
  "inline-flex items-center px-2 py-0.5 text-2xs font-mono font-semibold rounded border bg-hover-medium text-text-muted border-border";

const ROLE_META: Record<
  string,
  { icon: ComponentType<SVGProps<SVGSVGElement>>; summary: string }
> = {
  administrator: {
    icon: ShieldCheckIcon,
    summary: "Full access — manage devices, members, keys, and firewall rules",
  },
  operator: {
    icon: WrenchScrewdriverIcon,
    summary: "Manage devices and tags, connect via SSH, view sessions",
  },
  observer: {
    icon: EyeIcon,
    summary: "Read-only — connect to devices, view details and sessions",
  },
};

/** Small destructive badge shown next to expired API keys or invitations. */
export function ExpiredBadge() {
  return (
    <Badge color="red" shape="pill">
      <ExclamationCircleIcon className="w-2.5 h-2.5" strokeWidth={2} />
      Expired
    </Badge>
  );
}

export function RoleBadge({ role }: { role: string }) {
  const color = ROLE_COLOR[role];
  if (color) {
    return (
      <Badge color={color} shape="pill">
        {role}
      </Badge>
    );
  }
  // observer and unknown roles use a neutral style not in the Badge palette
  return <span className={ROLE_NEUTRAL_STYLE}>{role}</span>;
}

export function RoleSelector({
  label = "Role",
  value,
  onChange,
}: {
  label?: string;
  value: AssignableRole;
  onChange: (v: AssignableRole) => void;
}) {
  return (
    <RadioGroupField label={label} value={value} onChange={onChange}>
      {ROLES.map((role) => {
        const meta = ROLE_META[role];
        const { icon: Icon } = meta;
        return (
          <RadioCard
            key={role}
            value={role}
            icon={<Icon className="w-4 h-4" />}
            label={role.charAt(0).toUpperCase() + role.slice(1)}
            description={meta.summary}
          />
        );
      })}
    </RadioGroupField>
  );
}

/** React Hook Form binding for {@link RoleSelector}. */
export function FormRoleSelector<T extends FieldValues>({
  control,
  name,
  label,
}: {
  control: Control<T>;
  name: Path<T>;
  label?: string;
}) {
  const { field } = useController({ name, control });
  return (
    <RoleSelector label={label} value={field.value} onChange={field.onChange} />
  );
}

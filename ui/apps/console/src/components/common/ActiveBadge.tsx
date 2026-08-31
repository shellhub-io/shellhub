import { Badge } from "@shellhub/design-system/primitives";

/**
 * The active/inactive badge. Inactive is yellow rather than red: it is a state, not a fault.
 */
export default function ActiveBadge({ active }: { active: boolean }) {
  return (
    <Badge color={active ? "green" : "yellow"}>
      {active ? "Active" : "Inactive"}
    </Badge>
  );
}

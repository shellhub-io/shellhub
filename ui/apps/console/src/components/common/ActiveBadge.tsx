import { Badge } from "@shellhub/design-system/primitives";

export default function ActiveBadge({ active }: { active: boolean }) {
  return (
    <Badge color={active ? "green" : "yellow"}>
      {active ? "Active" : "Inactive"}
    </Badge>
  );
}

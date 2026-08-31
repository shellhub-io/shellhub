import { Badge } from "@shellhub/design-system/primitives";

interface SessionTypeBadgeProps {
  types: string[];
}

/**
 * The badge for what a session did — a file transfer, a single command, or an ordinary shell,
 * which gets no badge.
 */
export default function SessionTypeBadge({ types }: SessionTypeBadgeProps) {
  if (types.includes("subsystem"))
    return <Badge color="cyan" shape="pill">sftp</Badge>;

  if (types.includes("exec"))
    return <Badge color="yellow" shape="pill">exec</Badge>;

  if (types.includes("shell") || types.includes("pty-req"))
    return <Badge color="primary" shape="pill">shell</Badge>;

  return null;
}

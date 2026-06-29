import { Badge, DockerIcon } from "@shellhub/design-system/primitives";
import { ServerStackIcon } from "@heroicons/react/24/outline";

export default function PlatformBadge({ platform }: { platform: string }) {
  if (platform === "docker") {
    return (
      <Badge color="blue">
        <DockerIcon className="w-2.5 h-2.5" />
        Docker
      </Badge>
    );
  }
  return (
    <Badge color="green">
      <ServerStackIcon className="w-2.5 h-2.5" strokeWidth={2} />
      Native
    </Badge>
  );
}

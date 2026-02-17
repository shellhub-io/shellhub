import { ServerStackIcon } from "@heroicons/react/24/outline";
import { DockerIcon } from "../icons";

export default function PlatformBadge({ platform }: { platform: string }) {
  if (platform === "docker") {
    return (
      <span className="inline-flex items-center gap-1 px-1.5 py-0.5 bg-accent-blue/10 text-accent-blue text-2xs rounded font-medium">
        <DockerIcon className="w-2.5 h-2.5" />
        Docker
      </span>
    );
  }
  return (
    <span className="inline-flex items-center gap-1 px-1.5 py-0.5 bg-accent-green/10 text-accent-green text-2xs rounded font-medium">
      <ServerStackIcon className="w-2.5 h-2.5" strokeWidth={2} />
      Native
    </span>
  );
}

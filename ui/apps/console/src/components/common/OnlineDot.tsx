import { StatusDot } from "@shellhub/design-system/primitives";

export default function OnlineDot({ online }: { online?: boolean }) {
  return <StatusDot online={online} className="mx-auto" />;
}

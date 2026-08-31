import { StatusDot } from "@shellhub/design-system/primitives";

/**
 * The online dot as a table cell — centred, since it is the whole content of its column.
 */
export default function OnlineDot({ online }: { online?: boolean }) {
  return <StatusDot online={online} className="mx-auto" />;
}

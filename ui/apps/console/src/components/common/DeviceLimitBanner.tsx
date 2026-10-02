import NoticeBanner from "@/components/common/NoticeBanner";
import { useDeviceCapacity } from "@/hooks/useDeviceCapacity";

/**
 * Warns when the namespace is at or near its licensed device limit, before an enrolment starts
 * failing rather than after.
 */
export default function DeviceLimitBanner() {
  const { capacity, isLoading, isError } = useDeviceCapacity();

  const over = capacity?.state === "over";
  const approaching = capacity?.state === "approaching";

  const visible = !isLoading && !isError && (over || approaching);

  const severity = over ? "error" : "warning";

  const message = over
    ? "You've reached your licensed device limit. New devices can't connect until you contact the ShellHub team to raise the limit or remove some."
    : "You're approaching your licensed device limit. Contact the ShellHub team to raise it before new devices are blocked.";

  return (
    <NoticeBanner visible={visible} severity={severity}>
      {message}
    </NoticeBanner>
  );
}

import NoticeBanner from "@/components/common/NoticeBanner";
import { useAdminLicense } from "@/hooks/useAdminLicense";
import { useAdminStats } from "@/hooks/useAdminStats";

export default function DeviceLimitBanner() {
  const {
    data: license,
    isLoading: licenseLoading,
    isError: licenseError,
  } = useAdminLicense();
  const {
    stats,
    isLoading: statsLoading,
    isError: statsError,
  } = useAdminStats();

  const cap = license?.features.devices;
  const registered = stats?.registered_devices;

  const over =
    typeof cap === "number" &&
    cap >= 0 &&
    typeof registered === "number" &&
    Number.isFinite(registered) &&
    registered >= cap;

  const approaching =
    typeof cap === "number" &&
    cap > 0 &&
    !over &&
    typeof registered === "number" &&
    Number.isFinite(registered) &&
    registered / cap >= 0.9;

  const visible =
    !licenseLoading &&
    !licenseError &&
    !statsLoading &&
    !statsError &&
    license != null &&
    (over || approaching);

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

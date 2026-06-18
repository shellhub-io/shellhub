import { useMemo } from "react";
import { useAdminLicense } from "@/hooks/useAdminLicense";
import NoticeBanner from "@/components/common/NoticeBanner";

export default function LicenseBanner() {
  const { data, isLoading, isError, dataUpdatedAt } = useAdminLicense();

  // data === undefined: query not enabled (non-admin) — banner stays hidden
  // data === null:      no license installed (400 normalized by hook)
  // data object:        license found, check flags
  const noLicense = data === null;
  const expired = data != null && data.expired && !data.grace_period;
  const gracePeriod = data != null && data.expired && data.grace_period;
  const aboutToExpire = data != null && !data.expired && data.about_to_expire;

  const daysUntilExpiration = useMemo(() => {
    if (data == null || data.expires_at <= 0) return null;
    const days = Math.ceil((data.expires_at - dataUpdatedAt / 1000) / 86400);
    // treat zero or negative as null so the fallback copy is shown
    return days > 0 ? days : null;
  }, [data, dataUpdatedAt]);

  const visible =
    !isLoading &&
    !isError &&
    (noLicense || expired || gracePeriod || aboutToExpire);

  const severity = noLicense || expired ? "error" : "warning";

  const message = (() => {
    if (noLicense)
      return "No license installed. This instance needs a valid license to run properly — contact the ShellHub team to get one.";
    if (expired)
      return "Your license has expired. This instance won't function properly until it's renewed — contact the ShellHub team.";
    if (gracePeriod)
      return "Your license has expired and is in its grace period. Contact the ShellHub team to renew it before this instance stops working.";
    if (aboutToExpire) {
      return daysUntilExpiration !== null
        ? `Your license expires in ${daysUntilExpiration} day${daysUntilExpiration === 1 ? "" : "s"}. Contact the ShellHub team to renew it before it lapses.`
        : "Your license is about to expire. Contact the ShellHub team to renew it before it lapses.";
    }
    return "";
  })();

  return (
    <NoticeBanner visible={visible} severity={severity}>
      {message}
    </NoticeBanner>
  );
}

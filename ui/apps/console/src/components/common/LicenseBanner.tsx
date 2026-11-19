import { useMemo } from "react";
import { useAdminLicense } from "@/hooks/useAdminLicense";
import NoticeBanner from "@/components/common/NoticeBanner";

/**
 * Warns about the licence — missing, expiring, expired, inside its grace period, or not covering
 * the hostname this instance was reached under. Each state reads differently, because an instance
 * in its grace period is still working.
 */
export default function LicenseBanner() {
  const { data, isLoading, isError, dataUpdatedAt } = useAdminLicense();

  const noLicense = data === null;
  const expired = data != null && data.expired && !data.grace_period;
  const hostRefused = data != null && data.domain_allowed === false;
  const gracePeriod = data != null && data.expired && data.grace_period;
  const aboutToExpire = data != null && !data.expired && data.about_to_expire;

  const daysUntilExpiration = useMemo(() => {
    if (data == null || data.expires_at <= 0) return null;
    const days = Math.ceil((data.expires_at - dataUpdatedAt / 1000) / 86400);
    return days > 0 ? days : null;
  }, [data, dataUpdatedAt]);

  const visible =
    !isLoading &&
    !isError &&
    (noLicense || expired || hostRefused || gracePeriod || aboutToExpire);

  const severity = noLicense || expired || hostRefused ? "error" : "warning";

  const message = (() => {
    if (noLicense)
      return "No license installed. This instance needs a valid license to run properly — contact the ShellHub team to get one.";
    if (expired)
      return "Your license has expired. This instance won't function properly until it's renewed — contact the ShellHub team.";
    if (hostRefused)
      return `This license is for ${data.domains?.join(", ")}, but this instance was reached as ${data.request_host}. Devices cannot connect until they reach it under a licensed name or the license is replaced.`;
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

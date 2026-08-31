import NoticeBanner from "@/components/common/NoticeBanner";
import { useConnectivityStore } from "@/stores/connectivityStore";

/**
 * The banner shown while the API is unreachable. It stays out of the way rather than blocking,
 * since a brief outage should not throw away what the user was doing.
 */
export default function ConnectivityBanner() {
  const apiReachable = useConnectivityStore((s) => s.apiReachable);

  return (
    <NoticeBanner
      visible={!apiReachable}
      severity="error"
      align="center"
    >
      API unreachable — reconnecting automatically
    </NoticeBanner>
  );
}

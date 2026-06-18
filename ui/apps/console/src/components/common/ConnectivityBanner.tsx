import NoticeBanner from "@/components/common/NoticeBanner";
import { useConnectivityStore } from "@/stores/connectivityStore";

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

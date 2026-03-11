import { useConnectivityStore } from "../../stores/connectivityStore";

export default function ConnectivityBanner() {
  const apiReachable = useConnectivityStore((s) => s.apiReachable);

  return (
    <div
      className={`grid transition-[grid-template-rows] duration-300 ease-out ${
        apiReachable ? "grid-rows-[0fr]" : "grid-rows-[1fr]"
      }`}
    >
      <div className="overflow-hidden">
        <div className="bg-accent-red/[0.06] border-b border-accent-red/10 px-5 py-1.5 flex items-center justify-center gap-2">
          <span className="inline-flex rounded-full h-1.5 w-1.5 bg-accent-red shrink-0" />
          <p className="text-2xs font-mono text-accent-red whitespace-nowrap">
            API unreachable — reconnecting automatically
          </p>
        </div>
      </div>
    </div>
  );
}

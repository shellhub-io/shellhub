import { XMarkIcon, CommandLineIcon } from "@heroicons/react/24/outline";
import { useTerminalStore } from "../../stores/terminalStore";

export default function TerminalTaskbar() {
  const { sessions, restore, close } = useTerminalStore();
  const minimized = sessions.filter((s) => s.state === "minimized");

  if (minimized.length === 0) return null;

  return (
    <div className="fixed bottom-0 left-[220px] right-0 z-40 h-11 flex items-center gap-1.5 px-3 bg-surface border-t border-border animate-slide-up">
      {minimized.map((s) => {
        const isConnected = s.connectionStatus === "connected";

        return (
          <div
            key={s.id}
            className={`flex items-center gap-2 pl-3 pr-1.5 py-1.5 border rounded-lg transition-all duration-150 cursor-pointer group animate-fade-in ${
              isConnected
                ? "bg-accent-green/[0.06] border-accent-green/25 hover:border-accent-green/40 hover:bg-accent-green/[0.1]"
                : "bg-card border-border hover:border-primary/30 hover:bg-primary/[0.04]"
            }`}
            onClick={() => restore(s.id)}
          >
            <span
              className={`shrink-0 w-1.5 h-1.5 rounded-full transition-colors duration-300 ${
                isConnected
                  ? "bg-accent-green shadow-[0_0_4px_rgba(130,165,104,0.6)]"
                  : s.connectionStatus === "connecting"
                    ? "bg-accent-yellow animate-pulse"
                    : "bg-accent-red"
              }`}
            />
            <CommandLineIcon
              className={`w-3.5 h-3.5 transition-colors duration-150 ${
                isConnected
                  ? "text-accent-green group-hover:text-accent-green"
                  : "text-text-muted group-hover:text-primary"
              }`}
            />
            <span
              className={`text-xs font-medium transition-colors duration-150 max-w-[160px] truncate ${
                isConnected
                  ? "text-accent-green"
                  : "text-text-secondary group-hover:text-text-primary"
              }`}
            >
              {s.deviceName}
            </span>
            <button
              onClick={(e) => {
                e.stopPropagation();
                close(s.id);
              }}
              className="p-0.5 rounded text-text-muted/50 hover:text-accent-red hover:bg-accent-red/10 transition-all duration-150"
            >
              <XMarkIcon className="w-3.5 h-3.5" strokeWidth={2} />
            </button>
          </div>
        );
      })}
    </div>
  );
}

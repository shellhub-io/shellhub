import { Outlet } from "react-router-dom";
import AmbientBackground from "../common/AmbientBackground";

/**
 * The shell of the unauthenticated screens: centred card over the ambient background.
 */
export default function LoginLayout() {
  return (
    <div className="relative min-h-screen flex items-center justify-center bg-background overflow-hidden">
      <AmbientBackground />
      <div className="relative z-raised w-full px-8 py-12 flex flex-col items-center">
        <Outlet />
      </div>
    </div>
  );
}

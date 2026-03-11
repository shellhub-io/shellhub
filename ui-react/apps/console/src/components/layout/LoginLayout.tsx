import { Outlet } from "react-router-dom";
import AmbientBackground from "../common/AmbientBackground";

export default function LoginLayout() {
  return (
    <div className="relative min-h-screen flex items-center justify-center bg-background overflow-hidden">
      <AmbientBackground />
      <div className="relative z-10 w-full">
        <Outlet />
      </div>
    </div>
  );
}

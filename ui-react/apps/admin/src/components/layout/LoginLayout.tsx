import { Outlet } from "react-router-dom";

export default function LoginLayout() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <Outlet />
    </div>
  );
}

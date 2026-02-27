import { lazy, Suspense } from "react";
import { Routes, Route, Navigate } from "react-router-dom";
import {
  ExclamationTriangleIcon,
  UsersIcon,
  Bars3Icon,
} from "@heroicons/react/24/outline";
import Login from "./pages/Login";
import Setup from "./pages/Setup";
import { getConfig } from "./env";
import AppLayout from "./components/layout/AppLayout";
import LoginLayout from "./components/layout/LoginLayout";
import ConnectivityGuard from "./components/common/ConnectivityGuard";
import ProtectedRoute from "./components/common/ProtectedRoute";
import NamespaceGuard from "./components/common/NamespaceGuard";
import SetupGuard from "./components/common/SetupGuard";
import FeatureGate from "./components/common/FeatureGate";

const Dashboard = lazy(() => import("./pages/Dashboard"));
const Devices = lazy(() => import("./pages/devices"));
const Sessions = lazy(() => import("./pages/Sessions"));
const SessionDetails = lazy(() => import("./pages/SessionDetails"));
const NotFound = lazy(() => import("./pages/NotFound"));
const PublicKeys = lazy(() => import("./pages/public-keys"));
const DeviceDetails = lazy(() => import("./pages/DeviceDetails"));
const AddDevice = lazy(() => import("./pages/AddDevice"));
const Team = lazy(() => import("./pages/team"));
const WebEndpoints = lazy(() => import("./pages/WebEndpoints"));
const FirewallRulesPage = lazy(() => import("./pages/firewall-rules"));
const Settings = lazy(() => import("./pages/Settings"));
const BannerEdit = lazy(() => import("./pages/BannerEdit"));
const Profile = lazy(() => import("./pages/Profile"));
const ForgotPassword = lazy(() => import("./pages/ForgotPassword"));
const UpdatePassword = lazy(() => import("./pages/UpdatePassword"));

export default function App() {
  return (
    <Suspense>
      <Routes>
        <Route element={<ConnectivityGuard />}>
          <Route element={<SetupGuard />}>
            <Route element={<LoginLayout />}>
              <Route path="/login" element={<Login />} />
              <Route path="/setup" element={<Setup />} />
              {getConfig().cloud && (
                <>
                  <Route path="/forgot-password" element={<ForgotPassword />} />
                  <Route path="/update-password" element={<UpdatePassword />} />
                </>
              )}
            </Route>
            <Route element={<ProtectedRoute />}>
              <Route element={<NamespaceGuard />}>
                <Route element={<AppLayout />}>
                  <Route
                    path="/"
                    element={<Navigate to="/dashboard" replace />}
                  />
                  <Route path="/dashboard" element={<Dashboard />} />
                  <Route path="/devices" element={<Devices />} />
                  <Route path="/devices/add" element={<AddDevice />} />
                  <Route path="/devices/:uid" element={<DeviceDetails />} />
                  <Route path="/sessions" element={<Sessions />} />
                  <Route path="/sessions/:uid" element={<SessionDetails />} />
                  <Route path="/sshkeys/public-keys" element={<PublicKeys />} />
                  <Route path="/team" element={<Team />} />
                  <Route path="/webendpoints" element={<WebEndpoints />} />
                  <Route path="/settings" element={<Settings />} />
                  <Route path="/settings/banner" element={<BannerEdit />} />
                  <Route path="/profile" element={<Profile />} />
                  <Route
                    path="/firewall/rules"
                    element={
                      <FeatureGate
                        feature="Firewall Rules"
                        description="Control who can access your devices and from where. Define rules based on hostname, username, and source IP to enforce your security policies."
                        highlights={[
                          {
                            icon: (
                              <ExclamationTriangleIcon className="w-5 h-5" />
                            ),
                            title: "Allow & Deny",
                            description:
                              "Create rules to allow or block SSH connections based on your criteria.",
                          },
                          {
                            icon: <UsersIcon className="w-5 h-5" />,
                            title: "User Filtering",
                            description:
                              "Restrict access per username, hostname, or source IP address.",
                          },
                          {
                            icon: <Bars3Icon className="w-5 h-5" />,
                            title: "Priority Order",
                            description:
                              "Organize rules by priority to control evaluation order.",
                          },
                        ]}
                      >
                        <FirewallRulesPage />
                      </FeatureGate>
                    }
                  />
                </Route>
              </Route>
            </Route>
          </Route>
        </Route>
        <Route path="*" element={<NotFound />} />
      </Routes>
    </Suspense>
  );
}

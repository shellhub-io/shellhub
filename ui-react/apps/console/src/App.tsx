import { lazy, Suspense } from "react";
import { Routes, Route, Navigate } from "react-router-dom";
import { getConfig } from "./env";
import Login from "./pages/Login";
import Setup from "./pages/Setup";
import AppLayout from "./components/layout/AppLayout";

const MfaLogin = lazy(() => import("./pages/MfaLogin"));
const MfaRecover = lazy(() => import("./pages/MfaRecover"));
const MfaResetRequest = lazy(() => import("./pages/MfaResetRequest"));
const MfaResetVerify = lazy(() => import("./pages/MfaResetVerify"));
const MfaResetComplete = lazy(() => import("./pages/MfaResetComplete"));
import LoginLayout from "./components/layout/LoginLayout";
import ConnectivityGuard from "./components/common/ConnectivityGuard";
import ProtectedRoute from "./components/common/ProtectedRoute";
import NamespaceGuard from "./components/common/NamespaceGuard";
import SetupGuard from "./components/common/SetupGuard";
import SignUpGuard from "./components/common/SignUpGuard";
import AdminRoute from "./components/common/AdminRoute";
import AdminLayout from "./components/layout/AdminLayout";
import LicenseGuard from "./components/common/LicenseGuard";

const SignUp = lazy(() => import("./pages/SignUp"));
const ConfirmAccount = lazy(() => import("./pages/ConfirmAccount"));
const ValidationAccount = lazy(() => import("./pages/ValidationAccount"));
const Dashboard = lazy(() => import("./pages/Dashboard"));
const Devices = lazy(() => import("./pages/devices"));
const Sessions = lazy(() => import("./pages/sessions"));
const SessionDetails = lazy(() => import("./pages/SessionDetails"));
const NotFound = lazy(() => import("./pages/NotFound"));
const PublicKeys = lazy(() => import("./pages/public-keys"));
const DeviceDetails = lazy(() => import("./pages/DeviceDetails"));
const AddDevice = lazy(() => import("./pages/AddDevice"));
const Team = lazy(() => import("./pages/team"));
const Settings = lazy(() => import("./pages/Settings"));
const BannerEdit = lazy(() => import("./pages/BannerEdit"));
const Profile = lazy(() => import("./pages/Profile"));
const ForgotPassword = lazy(() => import("./pages/ForgotPassword"));
const UpdatePassword = lazy(() => import("./pages/UpdatePassword"));
const SecureVault = lazy(() => import("./pages/secure-vault"));
const AdminDashboard = lazy(() => import("./pages/admin/Dashboard"));
const AdminLicense = lazy(() => import("./pages/admin/License"));
const AdminUnauthorized = lazy(() => import("./pages/admin/Unauthorized"));
const AdminUsers = lazy(() => import("./pages/admin/users"));
const AdminUserDetails = lazy(() => import("./pages/admin/users/UserDetails"));
const AdminNamespaces = lazy(() => import("./pages/admin/namespaces"));
const AdminNamespaceDetails = lazy(
  () => import("./pages/admin/namespaces/NamespaceDetails"),
);

export default function App() {
  return (
    <Suspense>
      <Routes>
        <Route element={<ConnectivityGuard />}>
          <Route element={<SetupGuard />}>
            <Route element={<LoginLayout />}>
              <Route path="/login" element={<Login />} />
              <Route path="/mfa-login" element={<MfaLogin />} />
              <Route path="/mfa-recover" element={<MfaRecover />} />
              <Route path="/mfa-reset-request" element={<MfaResetRequest />} />
              <Route path="/mfa-reset-verify" element={<MfaResetVerify />} />
              <Route path="/reset-mfa" element={<MfaResetComplete />} />
              <Route path="/setup" element={<Setup />} />
              <Route element={<SignUpGuard />}>
                <Route path="/sign-up" element={<SignUp />} />
                <Route path="/confirm-account" element={<ConfirmAccount />} />
                <Route
                  path="/validation-account"
                  element={<ValidationAccount />}
                />
              </Route>
              {getConfig().cloud && (
                <>
                  <Route path="/forgot-password" element={<ForgotPassword />} />
                  <Route path="/update-password" element={<UpdatePassword />} />
                </>
              )}
            </Route>
            <Route element={<ProtectedRoute />}>
              {/* Admin panel — layout wraps all /admin routes including unauthorized */}
              <Route element={<AdminLayout />}>
                <Route
                  path="/admin/unauthorized"
                  element={<AdminUnauthorized />}
                />
                <Route element={<AdminRoute />}>
                  <Route path="/admin/license" element={<AdminLicense />} />
                  <Route element={<LicenseGuard />}>
                    <Route
                      path="/admin"
                      element={<Navigate to="/admin/dashboard" replace />}
                    />
                    <Route
                      path="/admin/dashboard"
                      element={<AdminDashboard />}
                    />
                    <Route path="/admin/users" element={<AdminUsers />} />
                    <Route
                      path="/admin/users/:id"
                      element={<AdminUserDetails />}
                    />
                    <Route
                      path="/admin/namespaces"
                      element={<AdminNamespaces />}
                    />
                    <Route
                      path="/admin/namespaces/:id"
                      element={<AdminNamespaceDetails />}
                    />
                  </Route>
                </Route>
              </Route>

              {/* User console */}
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
                  <Route path="/secure-vault" element={<SecureVault />} />
                  <Route path="/team" element={<Team />} />
                  <Route path="/settings" element={<Settings />} />
                  <Route path="/settings/banner" element={<BannerEdit />} />
                  <Route path="/profile" element={<Profile />} />
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

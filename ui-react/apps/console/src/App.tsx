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

const SignUp = lazy(() => import("./pages/SignUp"));
const ConfirmAccount = lazy(() => import("./pages/ConfirmAccount"));
const ValidationAccount = lazy(() => import("./pages/ValidationAccount"));
const Dashboard = lazy(() => import("./pages/Dashboard"));
const Devices = lazy(() => import("./pages/devices"));
const Sessions = lazy(() => import("./pages/Sessions"));
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

export default function App() {
  return (
    <Suspense>
      <Routes>
        <Route element={<ConnectivityGuard />}>
          <Route element={<SetupGuard />}>
            <Route element={<LoginLayout />}>
              <Route path="/login" element={<Login />} />
              <Route path="/confirm-account" element={<ConfirmAccount />} />
              <Route path="/mfa-login" element={<MfaLogin />} />
              <Route path="/mfa-recover" element={<MfaRecover />} />
              <Route path="/mfa-reset-request" element={<MfaResetRequest />} />
              <Route path="/mfa-reset-verify" element={<MfaResetVerify />} />
              <Route path="/reset-mfa" element={<MfaResetComplete />} />
              <Route path="/setup" element={<Setup />} />
              <Route element={<SignUpGuard />}>
                <Route path="/sign-up" element={<SignUp />} />
                <Route path="/confirm-account" element={<ConfirmAccount />} />
                <Route path="/validation-account" element={<ValidationAccount />} />
              </Route>
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

import { Routes, Route } from "react-router-dom";
import Landing from "./pages/landing";
import GettingStarted from "./pages/getting-started";
import Enterprise from "./pages/enterprise";
import Pricing from "./pages/pricing";
import Features from "./pages/features";
import HowItWorks from "./pages/how-it-works";
import Integrations from "./pages/integrations";
import IotEmbedded from "./pages/use-cases/IotEmbedded";
import EdgeComputing from "./pages/use-cases/EdgeComputing";
import RemoteSupport from "./pages/use-cases/RemoteSupport";
import DevopsCiCd from "./pages/use-cases/DevopsCiCd";
import ContainerManagement from "./pages/use-cases/ContainerManagement";

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<Landing />} />
      <Route path="/getting-started" element={<GettingStarted />} />
      <Route path="/enterprise" element={<Enterprise />} />
      <Route path="/pricing" element={<Pricing />} />
      <Route path="/features" element={<Features />} />
      <Route path="/how-it-works" element={<HowItWorks />} />
      <Route path="/integrations" element={<Integrations />} />
      <Route path="/use-cases/iot-embedded" element={<IotEmbedded />} />
      <Route path="/use-cases/edge-computing" element={<EdgeComputing />} />
      <Route path="/use-cases/remote-support" element={<RemoteSupport />} />
      <Route path="/use-cases/devops-ci-cd" element={<DevopsCiCd />} />
      <Route path="/use-cases/container-management" element={<ContainerManagement />} />
    </Routes>
  );
}

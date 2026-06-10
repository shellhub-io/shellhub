import React from "react";
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

export const routes: { path: string; element: React.ReactElement }[] = [
  { path: "/", element: <Landing /> },
  { path: "/getting-started", element: <GettingStarted /> },
  { path: "/enterprise", element: <Enterprise /> },
  { path: "/pricing", element: <Pricing /> },
  { path: "/features", element: <Features /> },
  { path: "/how-it-works", element: <HowItWorks /> },
  { path: "/integrations", element: <Integrations /> },
  { path: "/use-cases/iot-embedded", element: <IotEmbedded /> },
  { path: "/use-cases/edge-computing", element: <EdgeComputing /> },
  { path: "/use-cases/remote-support", element: <RemoteSupport /> },
  { path: "/use-cases/devops-ci-cd", element: <DevopsCiCd /> },
  { path: "/use-cases/container-management", element: <ContainerManagement /> },
];

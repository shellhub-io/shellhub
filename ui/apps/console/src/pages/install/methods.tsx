import type { JSX } from "react";
import {
  ArchiveBoxIcon,
  CommandLineIcon,
  ComputerDesktopIcon,
  CpuChipIcon,
  CubeIcon,
  ServerStackIcon,
  SparklesIcon,
  WrenchIcon,
} from "@heroicons/react/24/outline";
import { DockerIcon } from "@shellhub/design-system/primitives";

/** An installation method the installer knows how to run, or a platform that needs its own. */
export type Method =
  | "auto"
  | "docker"
  | "podman"
  | "snap"
  | "standalone"
  | "wsl"
  | "yocto"
  | "buildroot"
  | "freebsd";

/** One method as the screens present it: what it is called, what it does, and what it needs. */
export interface MethodInfo {
  id: Method;
  label: string;
  tag?: string;
  description: string;
  icon: JSX.Element;
  /** The installer does not do this one; the platform's own toolchain does. */
  manual?: boolean;
  docsUrl?: string;
}

/**
 * Methods that install without being told a namespace: the agent boots with no credentials,
 * mints a pairing code at runtime and prints an accept URL, so nothing sensitive rides on the
 * command line. Snap is the exception the installer itself enforces (`require_tenant`), and the
 * manual methods never reach the installer at all.
 */
export const PAIRING_METHODS: Method[] = [
  "auto",
  "docker",
  "podman",
  "standalone",
  "wsl",
];

/** Every method, in the order the screens offer them. */
export const METHODS: MethodInfo[] = [
  {
    id: "auto",
    label: "Auto Detect",
    tag: "Recommended",
    description:
      "Automatically detects Docker, Snap, or Standalone and uses the best available method.",
    icon: <SparklesIcon className="w-5 h-5" />,
  },
  {
    id: "docker",
    label: "Docker",
    description:
      "Run the agent as a Docker container. Requires Docker daemon running on the host.",
    icon: <DockerIcon className="w-5 h-5" />,
  },
  {
    id: "standalone",
    label: "Standalone",
    description:
      "Install directly using runc and systemd. No container runtime required.",
    icon: <ServerStackIcon className="w-5 h-5" />,
  },
  {
    id: "podman",
    label: "Podman",
    description:
      "Alternative to Docker with rootless container capabilities. Requires Podman daemon.",
    icon: <CubeIcon className="w-5 h-5" />,
  },
  {
    id: "snap",
    label: "Snap",
    description:
      "Easy installation via Snap store with automatic updates. Requires snapd service.",
    icon: <ArchiveBoxIcon className="w-5 h-5" />,
  },
  {
    id: "wsl",
    label: "WSL",
    description:
      "Optimized for Windows Subsystem for Linux 2 with systemd and mirrored networking.",
    icon: <ComputerDesktopIcon className="w-5 h-5" />,
  },
  {
    id: "yocto",
    label: "Yocto Project",
    tag: "Manual",
    description:
      "For embedded Linux systems built with the Yocto build system.",
    manual: true,
    docsUrl: "https://docs.shellhub.io/overview/supported-platforms/yocto",
    icon: <CpuChipIcon className="w-5 h-5" />,
  },
  {
    id: "buildroot",
    label: "Buildroot",
    tag: "Manual",
    description: "For embedded Linux systems built with Buildroot toolchain.",
    manual: true,
    docsUrl: "https://docs.shellhub.io/overview/supported-platforms/buildroot",
    icon: <WrenchIcon className="w-5 h-5" />,
  },
  {
    id: "freebsd",
    label: "FreeBSD",
    tag: "Manual",
    description:
      "For FreeBSD systems. Requires ports tree and manual compilation.",
    manual: true,
    docsUrl: "https://docs.shellhub.io/overview/supported-platforms/freebsd",
    icon: <CommandLineIcon className="w-5 h-5" />,
  },
];

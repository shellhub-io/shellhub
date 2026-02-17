const distroMap: Record<string, string> = {
  alpine: "fl-alpine",
  arch: "fl-archlinux",
  centos: "fl-centos",
  coreos: "fl-coreos",
  debian: "fl-debian",
  devuan: "fl-devuan",
  elementary: "fl-elementary",
  fedora: "fl-fedora",
  freebsd: "fl-freebsd",
  gentoo: "fl-gentoo",
  linuxmint: "fl-linuxmint",
  mageia: "fl-mageia",
  manjaro: "fl-manjaro",
  mandriva: "fl-mandriva",
  nixos: "fl-nixos",
  opensuse: "fl-opensuse",
  rhel: "fl-redhat",
  sabayon: "fl-sabayon",
  slackware: "fl-slackware",
  ubuntu: "fl-ubuntu",
  "ubuntu-core": "fl-ubuntu",
  ubuntucore: "fl-ubuntu",
  raspbian: "fl-raspberry-pi",
  void: "fl-void",
  docker: "fl-docker",
};

interface DistroIconProps {
  id: string;
  className?: string;
}

export default function DistroIcon({ id, className = "" }: DistroIconProps) {
  const icon = distroMap[id?.toLowerCase()] ?? "fl-tux";
  return <i className={`${icon} ${className}`} aria-hidden="true" />;
}

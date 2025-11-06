<template>
  <i
    :class="iconName"
    style="font-size: 20px"
  />
</template>

<script setup lang="ts">
import { computed } from "vue";

const { icon } = defineProps<{ icon: string }>();

const deviceIconMap: Record<string, string> = {
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
  raspbian: "fl-raspberry-pi",
  "ubuntu-core": "fl-ubuntu",
  void: "fl-void",
};

const iconName = computed(() => deviceIconMap[icon] || "");
</script>

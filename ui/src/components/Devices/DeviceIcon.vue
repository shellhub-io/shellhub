<template>
  <i
    data-test="type-icon"
    :class="iconName + ' mr-1'"
    style="font-size: 20px"
  />
</template>

<script setup lang="ts">
import { computed } from "vue";

const props = defineProps({
  icon: {
    type: String,
    required: true,
  },
});

const deviceIcon = {
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
  ubuntucore: "fl-ubuntu",
  void: "fl-void",
  docker: "fl-docker",
};

const iconName = computed(() => deviceIcon[props.icon] || "fl-tux");
</script>

<template>
  <i v-if="iconName" data-test="type-icon" :class="iconName + ' mr-1'" style="font-size: 18px" />
  <i v-else data-test="type-icon" class="fl-tux mr-1" style="font-size: 18px" />
</template>

<script lang="ts">
import { computed, defineComponent } from "vue";

export default defineComponent({
  props: {
    icon: {
      type: String,
      required: true,
    },
  },
  setup(props) {
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
    };

    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    const iconName = computed(() => deviceIcon[props.icon]);

    return {
      iconName,
    };
  },
});
</script>

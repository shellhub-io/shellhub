<template>
  <v-card class="bg-v-theme-surface mx-auto py-6 border">
    <v-card-title class="text-center d-flex justify-center pa-5">
      <div>
        <v-icon size="x-large" data-test="boxMessage-icon">
          {{ icon() }}
        </v-icon>
      </div>
    </v-card-title>

    <v-card-title class="d-flex justify-center">
      <div class="text-h5" data-test="boxMessage-title">
        Looks like you don't have any {{ title() }}
      </div>
    </v-card-title>

    <div class="d-flex justify-center flex-column">
      <v-list-item
        v-for="(item, i) in text()"
        :key="i"
        class="text-center listText mg-fix"
      >
        <div
          :data-test="i + '-boxMessage-text'"
          v-text="item"
        />
      </v-list-item>
    </div>

    <div class="d-flex justify-center flex-column">
      <!-- eslint-disable vue/no-v-html -->
      <v-list-item
        v-for="(item, index) in textWithLink()"
        :key="index"
        class="text-center listText mg-fix mt-n3"
      >
        <div
          class="justify-center"
          :data-test="index + '-boxMessage-text'"
          v-html="item"
        />
      </v-list-item>
      <!-- eslint-enable vue/no-v-html-->
    </div>

    <v-card-actions class="justify-center pt-8 pb-0">
      <DeviceAdd v-if="typeMessage == 'device'" />

      <span v-if="typeMessage == 'firewall'">
        <FirewallRuleAdd @update="refreshFirewallRule" />
      </span>

      <span v-else-if="typeMessage == 'publicKey'">
        <PublicKeyAdd @update="refreshPublicKey" />
      </span>
    </v-card-actions>
  </v-card>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { useStore } from "../../store";
import DeviceAdd from "../Devices/DeviceAdd.vue";
import FirewallRuleAdd from "../firewall/FirewallRuleAdd.vue";
import PublicKeyAdd from "../PublicKeys/PublicKeyAdd.vue";
import { INotificationsError } from "@/interfaces/INotifications";

const items = {
  device: {
    icon: "mdi-cellphone-link",
    title: "Device",
    text: [
      "In order to register a device on ShellHub, you need to install ShellHub agent onto it.",
    ],
    textWithLink: [
      `<p>The easiest way to install ShellHub agent is with our automatic one-line installation
                script, which works with all Linux distributions that have Docker installed and
                properly set up.
            <a
              target="_blank"
              href="https://docs.shellhub.io/user-guides/devices/adding"
            >See More</a>.</p>`,
    ],
  },
  session: {
    icon: "mdi-history",
    title: "Session",
    text: [
      "An SSH session is created when a connection is made to any registered device.",
    ],
    textWithLink: [
      `<p>Please follow our guide on how to connect to your devices
            <a
              target="_blank"
              href="https://docs.shellhub.io/user-guides/devices/connecting"
            >See More</a>.</p>`,
    ],
  },
  firewall: {
    icon: "mdi-security",
    title: "Firewall Rule",
    text: [
      `ShellHub provides flexible firewall for filtering SSH connections.
              It gives a fine-grained control over which SSH connections reach the devices.`,
      `Using Firewall Rules you can deny or allow SSH connections from specific
              IP address to a specific or a group of devices using a given username.`,
    ],
    textWithLink: [],
  },
  publicKey: {
    icon: "mdi-key",
    title: "Public Keys",
    text: [
      "You can connect to your devices using password-based logins, but we strongly recommend using SSH key pairs instead.",
      "SSH keys are more secure than passwords and can help you log in without having to remember long passwords.",
    ],
    textWithLink: [],
  },
};

export default defineComponent({
  props: {
    typeMessage: {
      type: String,
      default: "firewall",
      validator: (value: string) => ["device", "session", "firewall", "publicKey"].includes(value),
      required: true,
    },
  },
  setup(props) {
    const icon = () => {
      switch (props.typeMessage) {
        case "session":
          return items.session.icon;
        case "firewall":
          return items.firewall.icon;
        case "publicKey":
          return items.publicKey.icon;
        case "device":
          return items.device.icon;
        default:
          return null;
      }
    };
    const title = () => {
      switch (props.typeMessage) {
        case "session":
          return items.session.title;
        case "firewall":
          return items.firewall.title;
        case "publicKey":
          return items.publicKey.title;
        case "device":
          return items.device.title;
        default:
          return null;
      }
    };
    const text = () => {
      switch (props.typeMessage) {
        case "session":
          return items.session.text;
        case "firewall":
          return items.firewall.text;
        case "publicKey":
          return items.publicKey.text;
        case "device":
          return items.device.text;
        default:
          return null;
      }
    };
    const textWithLink = () => {
      switch (props.typeMessage) {
        case "session":
          return items.session.textWithLink;
        case "firewall":
          return items.firewall.textWithLink;
        case "publicKey":
          return items.publicKey.textWithLink;
        case "device":
          return items.device.textWithLink;
        default:
          return null;
      }
    };

    const store = useStore();

    const refreshFirewallRule = async () => {
      try {
        await store.dispatch("firewallRules/refresh");
      } catch (error: any) {
        store.dispatch(
          "snackbar/showSnackbarErrorLoading",
          INotificationsError.firewallRuleList,
        );
        throw new Error(error);
      }
    };

    const refreshPublicKey = async () => {
      try {
        await store.dispatch("publicKeys/refresh");
      } catch (error: any) {
        store.dispatch(
          "snackbar/showSnackbarErrorLoading",
          INotificationsError.publicKeyList,
        );
        throw new Error(error);
      }
    };

    return {
      items,
      icon,
      title,
      text,
      textWithLink,
      refreshFirewallRule,
      refreshPublicKey,
    };
  },
  components: { DeviceAdd, FirewallRuleAdd, PublicKeyAdd },
});
</script>

<style lang="scss">
.listText {
  min-height: 0px !important;
}
.mg-fix {
  margin: 0 auto;
}
</style>

<template>
  <fragment>
    <v-card
      class="mx-auto py-6"
      outlined
    >
      <v-card-title class="justify-center">
        <div>
          <v-icon
            x-large
            data-test="boxMessage-icon"
          >
            {{ icon() }}
          </v-icon>
        </div>
      </v-card-title>

      <v-card-title class="justify-center">
        <div
          class="headline"
          data-test="boxMessage-title"
        >
          Looks like you don't have any {{ title() }}
        </div>
      </v-card-title>

      <v-list-item
        v-for="(item, i) in text()"
        :key="i"
        class="text-center listText"
      >
        <v-list-item-content
          :class="{'justify-center py-0': true,
                   'pt-2': i > 0,
          }"
          :data-test="i+'-boxMessage-text'"
          v-text="item"
        />
      </v-list-item>

      <!-- eslint-disable vue/no-v-html -->
      <v-list-item
        v-for="(item, y) in textWithLink()"
        :key="text().length+y"
        class="text-center listText"
      >
        <v-list-item-content
          class="justify-center py-2"
          :data-test="text().length+y+'-boxMessage-text'"
          v-html="item"
        />
      </v-list-item>
      <!-- eslint-enable vue/no-v-html-->

      <v-card-actions class="justify-center pt-8 pb-0">
        <DeviceAdd v-if="typeMessage == 'device'" />

        <FirewallRuleEdit
          v-if="typeMessage == 'firewall'"
          :create-rule="true"
          @update="refreshFirewallRule"
        />

        <PublicKeyCreate
          v-else-if="typeMessage == 'publicKey'"
          :create-key="true"
          @update="refreshPublicKey"
        />
      </v-card-actions>
    </v-card>
  </fragment>
</template>

<script>

import DeviceAdd from '@/components/device/DeviceAdd';
import FirewallRuleEdit from '@/components/firewall_rule/FirewallRuleFormDialog';
import PublicKeyCreate from '@/components/public_key/KeyFormDialog';

export default {
  name: 'BoxMessageComponent',

  components: {
    DeviceAdd,
    FirewallRuleEdit,
    PublicKeyCreate,
  },

  props: {
    typeMessage: {
      type: String,
      default: 'firewall',
      validator: (value) => ['device', 'session', 'firewall', 'publicKey'].includes(value),
    },
  },

  data() {
    return {
      items:
      {
        device:
        {
          icon: 'devices',
          title: 'Device',
          text: [
            'In order to register a device on ShellHub, you need to install ShellHub agent onto it.',
          ],
          textWithLink: [
            `<p>The easiest way to install ShellHub agent is with our automatic one-line installation
                script, which works with all Linux distributions that have Docker installed and
                properly set up.
            <a
              target="_blank"
              href="https://docs.shellhub.io/getting-started/registering-device/"
            >See More</a>.</p>`,
          ],
        },
        session:
        {
          icon: 'history',
          title: 'Session',
          text: [
            'An SSH session is created when a connection is made to any registered device.',
          ],
          textWithLink: [
            `<p>Please follow our guide on how to connect to your devices
            <a
              target="_blank"
              href="https://docs.shellhub.io/getting-started/connecting-device/"
            >See More</a>.</p>`,
          ],
        },
        firewall:
        {
          icon: 'security',
          title: 'Firewall Rule',
          text: [
            `ShellHub provides flexible firewall for filtering SSH connections.
              It gives a fine-grained control over which SSH connections reach the devices.`,
            `Using Firewall Rules you can deny or allow SSH connections from specific
              IP address to a specific or a group of devices using a given username.`,
          ],
          textWithLink: [],
        },
        publicKey:
        {
          icon: 'vpn_key',
          title: 'Public Keys',
          text: [
            'You can connect to your devices using password-based logins, but we strongly recommend using SSH key pairs instead.',
            'SSH keys are more secure than passwords and can help you log in without having to remember long passwords.',
          ],
          textWithLink: [],
        },
      },
    };
  },

  async created() {
    this.$store.dispatch('boxs/setStatus', true);
  },

  methods: {
    icon() {
      switch (this.typeMessage) {
      case 'session':
        return this.items.session.icon;
      case 'firewall':
        return this.items.firewall.icon;
      case 'publicKey':
        return this.items.publicKey.icon;
      case 'device':
        return this.items.device.icon;
      default:
        return null;
      }
    },

    title() {
      switch (this.typeMessage) {
      case 'session':
        return this.items.session.title;
      case 'firewall':
        return this.items.firewall.title;
      case 'publicKey':
        return this.items.publicKey.title;
      case 'device':
        return this.items.device.title;
      default:
        return null;
      }
    },

    text() {
      switch (this.typeMessage) {
      case 'session':
        return this.items.session.text;
      case 'firewall':
        return this.items.firewall.text;
      case 'publicKey':
        return this.items.publicKey.text;
      case 'device':
        return this.items.device.text;
      default:
        return null;
      }
    },

    textWithLink() {
      switch (this.typeMessage) {
      case 'session':
        return this.items.session.textWithLink;
      case 'firewall':
        return this.items.firewall.textWithLink;
      case 'publicKey':
        return this.items.publicKey.textWithLink;
      case 'device':
        return this.items.device.textWithLink;
      default:
        return null;
      }
    },

    async refreshFirewallRule() {
      try {
        await this.$store.dispatch('firewallrules/refresh');
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.firewallRuleList);
      }
    },

    async refreshPublicKey() {
      try {
        await this.$store.dispatch('publickeys/refresh');
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.publicKeyList);
      }
    },
  },
};

</script>

<style>
.listText {
  min-height: 0px !important;
}
</style>

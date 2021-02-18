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

      <v-card-actions class="justify-center pt-8 pb-0">
        <FirewallRuleFormDialog
          v-if="typeMessage == 'firewall'"
          :create-rule="true"
          @update="refresh"
        />
      </v-card-actions>
    </v-card>
  </fragment>
</template>

<script>

import FirewallRuleFormDialog from '@/components/firewall_rules/FirewallRulesFormDialog';

export default {
  name: 'BoxMessage',

  components: {
    FirewallRuleFormDialog,
  },

  props: {
    typeMessage: {
      type: String,
      default: 'firewall',
      validator: (value) => ['firewall'].includes(value),
    },
  },

  data() {
    return {
      items:
      {
        firewall:
        {
          icon: 'security',
          title: 'Firewall Rule',
          text: [`ShellHub provides flexible firewall for filtering SSH connections.
            It gives a fine-grained control over which SSH connections reach the devices.`,
          `Using Firewall Rules you can deny or allow SSH connections from specific
            IP address to a specific or a group of devices using a given username.`],
        },
      },
    };
  },

  methods: {
    icon() {
      switch (this.typeMessage) {
      case 'firewall':
        return this.items.firewall.icon;
      default:
        return null;
      }
    },

    title() {
      switch (this.typeMessage) {
      case 'firewall':
        return this.items.firewall.title;
      default:
        return null;
      }
    },

    text() {
      switch (this.typeMessage) {
      case 'firewall':
        return this.items.firewall.text;
      default:
        return null;
      }
    },

    moreInformation() {
      switch (this.typeMessage) {
      case 'firewall':
        return this.items.firewall.moreInformation;
      default:
        return null;
      }
    },

    async refresh() {
      try {
        await this.$store.dispatch('firewallrules/refresh');
      } catch (e) {
        this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.firewallRuleList);
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

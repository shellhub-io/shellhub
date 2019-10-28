<template>
<v-card>
    <v-app-bar flat color="transparent">
        <v-toolbar-title>Device Fleet</v-toolbar-title>
    </v-app-bar>
    <v-divider></v-divider>
    <v-card-text class="pa-0">
        <v-data-table :headers="headers" :items="devices" item-key="uid" hide-default-footer>
            <template v-slot:item.online="{ item }">
                <v-icon color="success" v-if="item.online">check_circle</v-icon>
                <v-tooltip bottom v-else>
                    <template #activator="{ on }">
                        <v-icon v-on="on">check_circle</v-icon>
                    </template>
                    <span>last seen {{ item.last_seen | moment("from", "now") }}</span>
                </v-tooltip>
            </template>

            <template v-slot:item.uid="{ item }">
                <v-chip>
                    {{ item.uid }}
                    <v-icon small right @click v-clipboard="item.uid" v-clipboard:success="showCopySnack">mdi-content-copy</v-icon>
                </v-chip>
            </template>

            <template v-slot:item.attributes.pretty_name="{ item }">
                <v-icon left>{{ deviceIcon[item.attributes.id] || 'fl-tux' }}</v-icon>
                {{ item.attributes.pretty_name }}
            </template>

            <template v-slot:item.identity.mac="{ item }">
                <code>{{ item.identity.mac }}</code>
            </template>

            <template v-slot:item.actions="{ item }">
                <TerminalDialog :uid="item.uid" v-if="item.online"></TerminalDialog>

                <v-icon @click="remove(item.uid)">
                    delete
                </v-icon>
            </template>
        </v-data-table>
    </v-card-text>
    <v-snackbar v-model="copySnack" :timeout=3000>Device UID copied to clipboard</v-snackbar>
</v-card>
</template>

<script>
import TerminalDialog from "@/components/TerminalDialog.vue";

export default {
  components: {
    TerminalDialog
  },

  created() {
    this.$store.dispatch("devices/fetch");
  },

  computed: {
    devices() {
      return this.$store.getters["devices/list"];
    }
  },

  methods: {
    copy(device) {
      this.$clipboard(device.uid);
    },

    remove(uid) {
      if (confirm("Are you sure?")) {
        this.$store.dispatch("devices/remove", uid);
      }
    },

    showCopySnack() {
      this.copySnack = true;
    }
  },

  data() {
    return {
      deviceIcon: {
        arch: "fl-archlinux",
        ubuntu: "fl-ubuntu"
      },
      copySnack: false,
      headers: [
        {
          text: "UID",
          value: "uid"
        },
        {
          text: "Operating System",
          value: "attributes.pretty_name"
        },
        {
          text: "MAC",
          value: "identity.mac",
          align: "center"
        },
        {
          text: "Online",
          value: "online",
          align: "center"
        },
        {
          text: "Actions",
          value: "actions",
          align: "center",
          sortable: false
        }
      ]
    };
  }
};
</script>

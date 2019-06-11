<template>
  <div>
    <v-data-table :headers="headers" :items="$store.getters['devices/list']" class="elevation-1">
        <template v-slot:items="props">
            <td class="text-xs-left">{{ props.item.uid }}</td>
            <td>{{ props.item.identity.mac }}</td>
            <td>{{ props.item.last_seen | moment("from", "now") }}</td>
            <td><TerminalDialog :uid="props.item.uid"></TerminalDialog></td>
        </template>
    </v-data-table>
  </div>
</template>

<script>
import TerminalDialog from "@/components/TerminalDialog.vue";

export default {
  name: "DeviceList",

  components: {
    TerminalDialog
  },

  data() {
    return {
      show: false,

      headers: [
        {
          text: "Device UID",
          value: "uid"
        },
        {
          text: "MAC Address",
          value: "identity.mac"
        },
        {
          text: "Last Seen",
          value: "last_seen"
        },
        { text: "Actions", value: "name", sortable: false }
      ]
    };
  },

  created() {
    this.$store.dispatch("devices/fetch");
  }
};
</script>

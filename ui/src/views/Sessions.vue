<template>
  <div>
    <v-data-table :headers="headers" :items="$store.getters['sessions/list']" class="elevation-1">
        <template v-slot:items="props">
            <td class="text-xs-left">{{ props.item.uid }}</td>
            <td>{{ props.item.device }}</td>
            <td>{{ props.item.username }}</td>
            <td>{{ props.item.started_at | moment("ddd D MMM YYYY HH:mm:ss") }}</td>
            <td>{{ props.item.finished_at | moment("ddd D MMM YYYY HH:mm:ss") }}</td>
        </template>
    </v-data-table>
  </div>
</template>

<script>
export default {
  data() {
    return {
      headers: [
        {
          text: "UID",
          value: "uid"
        },
        {
          text: "Device UID",
          value: "device"
        },
        {
          text: "Username",
          value: "username"
        },
        {
          text: "Started At",
          value: "started_at"
        },
        {
          text: "Finished At",
          value: "finished_at"
        }
      ]
    };
  },

  created() {
    this.$store.dispatch("sessions/fetch");
  }
};
</script>

<template>
  <fragment>
    <p class="ml-4 pt-4">
      Dear, {{ $store.getters["auth/currentUser"] }}!
    </p>
    <p class="mt-4 ml-4 mr-4">
      ShellHub is a modern SSH server for remotely accessing Linux devices via CLI or web-based user
      interface, it avoids the inconveniences such as firewall and NAT when getting your IP address,
      automating the access process, and helping you easily, seamlessly and securely connect your
      device to the internet.
    </p>
    <p class="ml-4 pb-4">
      Thank you for choosing the ShellHub.
    </p>
  </fragment>
</template>

<script>

export default {
  name: 'WelcomeFirstScreen',
};

</script>

<style>

</style>
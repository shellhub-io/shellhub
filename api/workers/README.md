# Introduction

TODO

### Connected Devices

The ***connected_devices*** queue manages operations related to devices that are currently online. It includes the following tasks:

- `connected_devices:[increase|decrease]`: This task increases or decreases the count of connected devices with a specified status. The payload must match the format `{tenant}:{status}`. If the device's status is unknown or irrelevant, the UID can be used instead, instructing the task to lookup the database and use the stored status.
